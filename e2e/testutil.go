package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/donamo/todo-backend/db/migrations"
	"github.com/donamo/todo-backend/internal/app"
	"github.com/donamo/todo-backend/internal/config"
)

type Suite struct {
	DB     *sql.DB
	Server *httptest.Server
	Cookie *http.Cookie
}

const testSessionSecret = "e2e-session-secret"

func NewSuite(t *testing.T) *Suite {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL is required for DB-backed e2e tests")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		if !isMissingDatabase(err) {
			t.Fatalf("connect test database: %v", err)
		}
		if err := createTestDatabase(ctx, dbURL); err != nil {
			t.Fatalf("create test database: %v", err)
		}
		if err := db.PingContext(ctx); err != nil {
			t.Fatalf("connect created test database: %v", err)
		}
	}

	if _, err := db.ExecContext(ctx, "DROP SCHEMA public CASCADE"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA public"); err != nil {
		t.Fatal(err)
	}

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "."); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(app.New(config.Config{
		FrontendURL:           "http://localhost:5173",
		SessionSecret:         testSessionSecret,
		HTTPShutdownTimeout:   10 * time.Second,
		HTTPReadHeaderTimeout: 5 * time.Second,
	}, db))
	t.Cleanup(server.Close)

	userID := createTestUser(t, db)
	return &Suite{DB: db, Server: server, Cookie: loginCookie(t, userID)}
}

func isMissingDatabase(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 3D000")
}

func createTestDatabase(ctx context.Context, dbURL string) error {
	u, err := url.Parse(dbURL)
	if err != nil {
		return err
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return fmt.Errorf("database name is empty")
	}
	if dbName == "postgres" || dbName == "template0" || dbName == "template1" {
		return fmt.Errorf("refusing to create reserved database %q", dbName)
	}

	u.Path = "/postgres"
	adminDB, err := sql.Open("pgx", u.String())
	if err != nil {
		return err
	}
	defer adminDB.Close()

	if err := adminDB.PingContext(ctx); err != nil {
		return err
	}
	_, err = adminDB.ExecContext(ctx, "CREATE DATABASE "+quoteIdent(dbName))
	if err != nil && !strings.Contains(err.Error(), "SQLSTATE 42P04") {
		return err
	}
	return nil
}

func quoteIdent(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

type gqlResponse[T any] struct {
	Data   T `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (s *Suite) GQL(t *testing.T, query string, variables map[string]any, out any) {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, s.Server.URL+"/graphql", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.Cookie != nil {
		req.AddCookie(s.Cookie)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("graphql status = %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}

func createTestUser(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := db.QueryRow(`
		INSERT INTO users (google_subject, email, name, approved)
		VALUES ('e2e-user', 'e2e@example.com', 'E2E User', true)
		RETURNING id
	`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func createPendingUser(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := db.QueryRow(`
		INSERT INTO users (google_subject, email, name, approved)
		VALUES ('pending-user', 'pending@example.com', 'Pending User', false)
		RETURNING id
	`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func loginCookie(t *testing.T, userID uuid.UUID) *http.Cookie {
	t.Helper()
	store := sessions.NewCookieStore([]byte(testSessionSecret))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	sess, err := store.Get(req, "todo_session")
	if err != nil {
		t.Fatal(err)
	}
	sess.Values["user_id"] = userID.String()
	if err := sess.Save(req, rec); err != nil {
		t.Fatal(err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login cookie was not created")
	}
	return cookies[0]
}
