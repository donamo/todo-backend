package graph

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/donamo/todo-backend/internal/ai"
	"github.com/donamo/todo-backend/internal/auth"
	"github.com/donamo/todo-backend/internal/config"
	dbsqlc "github.com/donamo/todo-backend/internal/db"
)

type Resolver struct {
	db      *sql.DB
	queries *dbsqlc.Queries
	ai      *ai.Client
}

func NewResolver(db *sql.DB) *Resolver {
	r := &Resolver{db: db}
	if db != nil {
		r.queries = dbsqlc.New(db)
	}
	timeout, err := config.Duration("OPENAI_TIMEOUT", 90*time.Second)
	if err != nil {
		slog.Error("invalid duration env, using default", "name", "OPENAI_TIMEOUT", "default", "90s", "err", err)
		timeout = 90 * time.Second
	}
	r.ai = ai.NewClient(ai.Config{
		APIKey:    config.String("OPENAI_API_KEY", ""),
		ProjectID: config.String("OPENAI_PROJECT_ID", ""),
		BaseURL:   config.String("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		Model:     config.String("OPENAI_MODEL", "gpt-5.4-mini"),
		Timeout:   timeout,
	})
	return r
}

func (r *Resolver) requireQueries() (*dbsqlc.Queries, error) {
	if r.queries == nil {
		return nil, fmt.Errorf("database is not configured")
	}
	return r.queries, nil
}

func (r *Resolver) requireUserQueries(ctx context.Context) (*dbsqlc.Queries, uuid.UUID, error) {
	q, err := r.requireQueries()
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, uuid.UUID{}, fmt.Errorf("unauthorized")
	}
	return q, user.ID, nil
}
