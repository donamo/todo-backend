package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/sessions"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/donamo/todo-backend/internal/auth"
	"github.com/donamo/todo-backend/internal/config"
	"github.com/donamo/todo-backend/internal/graph"
)

func New(cfg config.Config, db *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}))

	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
		Path:     "/",
	}
	authHandler, err := auth.NewHandler(db, store)
	if err != nil {
		slog.Error("auth handler init failed", "err", err)
		authHandler = nil
	}
	r.Use(auth.SessionMiddleware(db, store))

	r.Get("/health", health)
	if authHandler != nil {
		r.Get("/auth/google/login", authHandler.Login)
		r.Get("/auth/callback/google", authHandler.Callback)
		r.Post("/auth/logout", authHandler.Logout)
		r.With(auth.RequireAuth).Get("/auth/me", authHandler.Me)
	} else {
		r.Get("/auth/google/login", notImplemented("auth handler is not available"))
		r.Get("/auth/callback/google", notImplemented("auth handler is not available"))
		r.Post("/auth/logout", notImplemented("auth handler is not available"))
		r.Get("/auth/me", unauthenticated)
	}

	gqlSrv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: graph.NewResolver(db),
	}))
	gqlSrv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		slog.Error("graphql error", "err", err)
		return graphql.DefaultErrorPresenter(ctx, err)
	})
	gqlSrv.SetRecoverFunc(func(ctx context.Context, err any) error {
		slog.Error("graphql panic recovered", "err", err)
		return fmt.Errorf("internal server error")
	})
	r.With(auth.RequireAuth).Handle("/graphql", gqlSrv)
	r.Handle("/playground", playground.Handler("GraphQL", "/graphql"))

	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func unauthenticated(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
}

func notImplemented(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": message})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("json response encode failed", "err", err)
	}
}
