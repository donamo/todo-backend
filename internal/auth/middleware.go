package auth

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"

	dbsqlc "github.com/donamo/todo-backend/internal/db"
)

const (
	sessionName    = "todo_session"
	sessionUserKey = "user_id"
)

type contextKey string

const userContextKey contextKey = "user"

func IsAdmin(email string) bool {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	return adminEmail != "" && email == adminEmail
}

func SessionMiddleware(db *sql.DB, store *sessions.CookieStore) func(http.Handler) http.Handler {
	var queries *dbsqlc.Queries
	if db != nil {
		queries = dbsqlc.New(db)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if queries == nil {
				next.ServeHTTP(w, r)
				return
			}

			sess, err := store.Get(r, sessionName)
			if err != nil {
				slog.Error("session get failed", "err", err)
				next.ServeHTTP(w, r)
				return
			}

			userIDStr, ok := sess.Values[sessionUserKey].(string)
			if !ok || userIDStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				slog.Error("session user id parse failed", "err", err)
				next.ServeHTTP(w, r)
				return
			}

			user, err := queries.GetUserByID(r.Context(), userID)
			if err != nil {
				slog.Error("session user lookup failed", "err", err, "userID", userID)
				next.ServeHTTP(w, r)
				return
			}

			if !user.Approved && !IsAdmin(user.Email) {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserFromContext(ctx context.Context) *dbsqlc.User {
	u, _ := ctx.Value(userContextKey).(*dbsqlc.User)
	return u
}
