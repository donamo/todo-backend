package graph

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/donamo/todo-backend/internal/auth"
	dbsqlc "github.com/donamo/todo-backend/internal/db"
	"github.com/donamo/todo-backend/internal/graph/model"
)

func (r *Resolver) requireAdminQueries(ctx context.Context) (*dbsqlc.Queries, error) {
	q, err := r.requireQueries()
	if err != nil {
		return nil, err
	}
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}
	if !auth.IsAdmin(user.Email) {
		return nil, errors.New("forbidden")
	}
	return q, nil
}

func (r *Resolver) todoList(ctx context.Context, load func(*dbsqlc.Queries, uuid.UUID) ([]dbsqlc.Todo, error)) ([]*model.Todo, error) {
	q, userID, err := r.requireUserQueries(ctx)
	if err != nil {
		return nil, err
	}
	todos, err := load(q, userID)
	if err != nil {
		return nil, err
	}
	return r.todosWithLabels(ctx, q, userID, todos)
}

func (r *Resolver) getTodoModel(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, todoID uuid.UUID) (*model.Todo, error) {
	todo, err := q.GetTodo(ctx, dbsqlc.GetTodoParams{ID: todoID, UserID: userID})
	if err != nil {
		return nil, err
	}
	return r.todoWithLabels(ctx, q, userID, todo)
}

func (r *Resolver) todoWithLabels(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, todo dbsqlc.Todo) (*model.Todo, error) {
	labels, err := q.ListTodoLabels(ctx, dbsqlc.ListTodoLabelsParams{TodoID: todo.ID, UserID: userID})
	if err != nil {
		return nil, err
	}
	result := make([]*model.Label, 0, len(labels))
	for _, label := range labels {
		result = append(result, toLabel(label))
	}
	return toTodo(todo, result), nil
}

func (r *Resolver) todosWithLabels(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, todos []dbsqlc.Todo) ([]*model.Todo, error) {
	result := make([]*model.Todo, 0, len(todos))
	for _, todo := range todos {
		item, err := r.todoWithLabels(ctx, q, userID, todo)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *Resolver) setTodoNextAction(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, todoID uuid.UUID, nextAction bool) (*model.Todo, error) {
	todo, err := q.GetTodo(ctx, dbsqlc.GetTodoParams{ID: todoID, UserID: userID})
	if err != nil {
		return nil, err
	}
	if nextAction && todo.ProjectID.Valid {
		if err := q.ClearProjectNextActions(ctx, dbsqlc.ClearProjectNextActionsParams{UserID: userID, ProjectID: todo.ProjectID, ID: todoID}); err != nil {
			return nil, err
		}
	}
	todo, err = q.UpdateTodo(ctx, dbsqlc.UpdateTodoParams{
		ID:         todoID,
		UserID:     userID,
		NextAction: sql.NullBool{Bool: nextAction, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return r.todoWithLabels(ctx, q, userID, todo)
}

func progressModel(total int32, done int32) *model.Progress {
	percent := 0
	if total > 0 {
		percent = int(done * 100 / total)
	}
	return &model.Progress{
		Total:   int(total),
		Done:    int(done),
		Percent: percent,
	}
}

func optionalInt(value *int) any {
	if value == nil {
		return nil
	}
	return int32(*value)
}

func optionalBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func optionalProjectStatus(value *model.ProjectStatus) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func nullProjectStatus(value *model.ProjectStatus) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
}

func optionalStageStatus(value *model.StageStatus) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func nullStageStatus(value *model.StageStatus) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
}

func optionalTodoPriority(value *model.TodoPriority) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func nullTodoPriority(value *model.TodoPriority) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
}

func optionalTodoStatus(value *model.TodoStatus) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func nullTodoStatus(value *model.TodoStatus) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
}

func optionalRecurrence(value *model.Recurrence) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func nullRecurrence(value *model.Recurrence) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
}
