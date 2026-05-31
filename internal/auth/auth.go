package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/donamo/todo-backend/internal/config"
	dbsqlc "github.com/donamo/todo-backend/internal/db"
)

type Handler struct {
	store       *sessions.CookieStore
	queries     *dbsqlc.Queries
	oauth2      *oauth2.Config
	frontendURL string
}

func NewHandler(db *sql.DB, store *sessions.CookieStore) (*Handler, error) {
	var queries *dbsqlc.Queries
	if db != nil {
		queries = dbsqlc.New(db)
	}

	cfg := &oauth2.Config{
		ClientID:     config.String("GOOGLE_CLIENT_ID", ""),
		ClientSecret: config.String("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  config.String("GOOGLE_REDIRECT_URL", ""),
		Endpoint:     google.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	return &Handler{
		store:       store,
		queries:     queries,
		oauth2:      cfg,
		frontendURL: config.String("FRONTEND_URL", "http://localhost:5173"),
	}, nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database is not configured", http.StatusServiceUnavailable)
		return
	}
	if h.oauth2.ClientID == "" || h.oauth2.ClientSecret == "" || h.oauth2.RedirectURL == "" {
		http.Error(w, "google auth is not configured", http.StatusServiceUnavailable)
		return
	}
	state, err := randomState()
	if err != nil {
		slog.Error("oauth state generation failed", "err", err)
		http.Error(w, "state generation failed", http.StatusInternalServerError)
		return
	}
	sess, err := h.store.Get(r, sessionName)
	if err != nil {
		slog.Error("session get failed", "err", err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	sess.Values["oauth_state"] = state
	if err := sess.Save(r, w); err != nil {
		slog.Error("session save failed", "err", err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, h.oauth2.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database is not configured", http.StatusServiceUnavailable)
		return
	}
	if h.oauth2.ClientID == "" || h.oauth2.ClientSecret == "" || h.oauth2.RedirectURL == "" {
		http.Error(w, "google auth is not configured", http.StatusServiceUnavailable)
		return
	}
	provider, err := oidc.NewProvider(r.Context(), "https://accounts.google.com")
	if err != nil {
		slog.Error("oidc provider init failed", "err", err)
		http.Error(w, "google auth is not available", http.StatusServiceUnavailable)
		return
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: h.oauth2.ClientID})
	sess, err := h.store.Get(r, sessionName)
	if err != nil {
		slog.Error("session get failed", "err", err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	savedState, _ := sess.Values["oauth_state"].(string)
	if savedState == "" || savedState != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	delete(sess.Values, "oauth_state")

	token, err := h.oauth2.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		slog.Error("oauth token exchange failed", "err", err)
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "missing id_token", http.StatusInternalServerError)
		return
	}

	idToken, err := verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		slog.Error("id token verification failed", "err", err)
		http.Error(w, "invalid id_token", http.StatusInternalServerError)
		return
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		slog.Error("id token claims decode failed", "err", err)
		http.Error(w, "claims error", http.StatusInternalServerError)
		return
	}

	user, err := h.queries.UpsertUser(r.Context(), dbsqlc.UpsertUserParams{
		GoogleSubject: claims.Sub,
		Email:         claims.Email,
		Name:          claims.Name,
		Approved:      IsAdmin(claims.Email),
	})
	if err != nil {
		slog.Error("user upsert failed", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	sess.Values[sessionUserKey] = user.ID.String()
	if err := sess.Save(r, w); err != nil {
		slog.Error("session save failed", "err", err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, h.frontendURL, http.StatusTemporaryRedirect)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, err := h.store.Get(r, sessionName)
	if err != nil {
		slog.Error("session get failed", "err", err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		slog.Error("session save failed", "err", err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"id":       user.ID,
		"email":    user.Email,
		"name":     user.Name,
		"approved": user.Approved || IsAdmin(user.Email),
		"isAdmin":  IsAdmin(user.Email),
	}); err != nil {
		slog.Error("auth me response encode failed", "err", err)
	}
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
