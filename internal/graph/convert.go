package graph

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	dbsqlc "github.com/donamo/todo-backend/internal/db"
	"github.com/donamo/todo-backend/internal/graph/model"
)

func toEpic(epic dbsqlc.Epic) *model.Epic {
	return &model.Epic{
		ID:          epic.ID.String(),
		Name:        epic.Name,
		Description: nullStringPtr(epic.Description),
		Color:       nullStringPtr(epic.Color),
		Position:    int(epic.Position),
		CreatedAt:   epic.CreatedAt,
		UpdatedAt:   epic.UpdatedAt,
	}
}

func toProject(project dbsqlc.Project) *model.Project {
	return &model.Project{
		ID:          project.ID.String(),
		EpicID:      nullUUIDPtr(project.EpicID),
		Name:        project.Name,
		Description: nullStringPtr(project.Description),
		Status:      model.ProjectStatus(project.Status),
		StartDate:   nullTimePtr(project.StartDate),
		TargetDate:  nullTimePtr(project.TargetDate),
		Position:    int(project.Position),
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func toStage(stage dbsqlc.Stage) *model.Stage {
	return &model.Stage{
		ID:          stage.ID.String(),
		ProjectID:   stage.ProjectID.String(),
		Name:        stage.Name,
		Description: nullStringPtr(stage.Description),
		Status:      model.StageStatus(stage.Status),
		StartDate:   nullTimePtr(stage.StartDate),
		TargetDate:  nullTimePtr(stage.TargetDate),
		Position:    int(stage.Position),
		CreatedAt:   stage.CreatedAt,
		UpdatedAt:   stage.UpdatedAt,
	}
}

func toTodo(todo dbsqlc.Todo, labels []*model.Label) *model.Todo {
	return &model.Todo{
		ID:              todo.ID.String(),
		ProjectID:       nullUUIDPtr(todo.ProjectID),
		StageID:         nullUUIDPtr(todo.StageID),
		Title:           todo.Title,
		Description:     nullStringPtr(todo.Description),
		Priority:        model.TodoPriority(todo.Priority),
		Status:          model.TodoStatus(todo.Status),
		Labels:          labels,
		StartDate:       nullTimePtr(todo.StartDate),
		DueDate:         nullTimePtr(todo.DueDate),
		EstimatedEffort: nullStringPtr(todo.EstimatedEffort),
		Position:        int(todo.Position),
		NextAction:      todo.NextAction,
		Milestone:       todo.Milestone,
		Recurrence:      model.Recurrence(todo.Recurrence),
		CompletedAt:     nullTimePtr(todo.CompletedAt),
		CreatedAt:       todo.CreatedAt,
		UpdatedAt:       todo.UpdatedAt,
	}
}

func toLabel(label dbsqlc.Label) *model.Label {
	return &model.Label{
		ID:        label.ID.String(),
		Name:      label.Name,
		Color:     nullStringPtr(label.Color),
		CreatedAt: label.CreatedAt,
		UpdatedAt: label.UpdatedAt,
	}
}

func toProjectNote(note dbsqlc.ProjectNote) *model.ProjectNote {
	return &model.ProjectNote{
		ID:        note.ID.String(),
		ProjectID: note.ProjectID.String(),
		Body:      note.Body,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}
}

func toTodoDependency(dep dbsqlc.TodoDependency) *model.TodoDependency {
	return &model.TodoDependency{
		TodoID:          dep.TodoID.String(),
		DependsOnTodoID: dep.DependsOnTodoID.String(),
		CreatedAt:       dep.CreatedAt,
	}
}

func toAIProposal(proposal dbsqlc.AiProposal) *model.AIProposal {
	return &model.AIProposal{
		ID:           proposal.ID.String(),
		ParentType:   model.AIProposalParentType(proposal.ParentType),
		ParentID:     proposal.ParentID.String(),
		MagicText:    proposal.MagicText,
		Summary:      proposal.Summary,
		ProposalJSON: string(proposal.ProposalJson),
		Status:       model.AIProposalStatus(proposal.Status),
		CreatedAt:    proposal.CreatedAt,
		AppliedAt:    nullTimePtr(proposal.AppliedAt),
	}
}

func nullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	year, month, day := value.Date()
	return sql.NullTime{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC), Valid: true}
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func nullInt32(value *int) sql.NullInt32 {
	if value == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*value), Valid: true}
}

func nullBool(value *bool) sql.NullBool {
	if value == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *value, Valid: true}
}

func nullUUID(value *string) (uuid.NullUUID, error) {
	if value == nil {
		return uuid.NullUUID{}, nil
	}
	id, err := uuid.Parse(*value)
	if err != nil {
		return uuid.NullUUID{}, err
	}
	return uuid.NullUUID{UUID: id, Valid: true}, nil
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}

func nullUUIDPtr(value uuid.NullUUID) *string {
	if !value.Valid {
		return nil
	}
	str := value.UUID.String()
	return &str
}

func enumString[T interface{ String() string }](value *T) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: (*value).String(), Valid: true}
}
