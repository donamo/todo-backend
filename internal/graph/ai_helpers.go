package graph

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/donamo/todo-backend/internal/ai"
	dbsqlc "github.com/donamo/todo-backend/internal/db"
	"github.com/donamo/todo-backend/internal/graph/model"
)

func (r *Resolver) aiContext(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, parentType string, parentID uuid.UUID) (ai.Context, error) {
	snapshot := ai.Context{
		ParentType: parentType,
		ParentID:   parentID.String(),
	}
	switch parentType {
	case model.AIProposalParentTypeEpic.String():
		epic, err := q.GetEpic(ctx, dbsqlc.GetEpicParams{ID: parentID, UserID: userID})
		if err != nil {
			return snapshot, err
		}
		projects, err := q.ListProjects(ctx, dbsqlc.ListProjectsParams{
			UserID: userID,
			EpicID: uuid.NullUUID{UUID: parentID, Valid: true},
		})
		if err != nil {
			return snapshot, err
		}
		snapshot.ID = epic.ID.String()
		snapshot.Name = epic.Name
		snapshot.Description = stringFromNull(epic.Description)
		snapshot.Color = stringFromNull(epic.Color)
		snapshot.Projects, err = projectContexts(ctx, q, userID, projects)
		if err != nil {
			return snapshot, err
		}
	case model.AIProposalParentTypeProject.String():
		project, err := q.GetProject(ctx, dbsqlc.GetProjectParams{ID: parentID, UserID: userID})
		if err != nil {
			return snapshot, err
		}
		snapshot.ID = project.ID.String()
		snapshot.Name = project.Name
		snapshot.Description = stringFromNull(project.Description)
		if project.EpicID.Valid {
			epic, err := q.GetEpic(ctx, dbsqlc.GetEpicParams{ID: project.EpicID.UUID, UserID: userID})
			if err != nil {
				return snapshot, err
			}
			snapshot.ID = epic.ID.String()
			snapshot.Name = epic.Name
			snapshot.Description = stringFromNull(epic.Description)
			snapshot.Color = stringFromNull(epic.Color)
		}
		snapshot.Projects, err = projectContexts(ctx, q, userID, []dbsqlc.Project{project})
		if err != nil {
			return snapshot, err
		}
	default:
		return snapshot, fmt.Errorf("unsupported ai proposal parent type %q", parentType)
	}
	return snapshot, nil
}

func projectContexts(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, projects []dbsqlc.Project) ([]ai.ProjectContext, error) {
	result := make([]ai.ProjectContext, 0, len(projects))
	for _, project := range projects {
		projectStages, err := q.ListStages(ctx, dbsqlc.ListStagesParams{ProjectID: uuid.NullUUID{UUID: project.ID, Valid: true}, UserID: userID})
		if err != nil {
			return nil, err
		}
		projectTodos, err := q.ListTodos(ctx, dbsqlc.ListTodosParams{
			UserID:    userID,
			ProjectID: uuid.NullUUID{UUID: project.ID, Valid: true},
		})
		if err != nil {
			return nil, err
		}
		result = append(result, toProjectContext(project, projectStages, projectTodos))
	}
	return result, nil
}

func toProjectContext(project dbsqlc.Project, stages []dbsqlc.Stage, todos []dbsqlc.Todo) ai.ProjectContext {
	todosByStage := map[uuid.UUID][]ai.TodoContext{}
	for _, todo := range todos {
		if todo.StageID.Valid {
			todosByStage[todo.StageID.UUID] = append(todosByStage[todo.StageID.UUID], toTodoContext(todo))
		}
	}
	stageContexts := make([]ai.StageContext, 0, len(stages))
	for _, stage := range stages {
		stageContexts = append(stageContexts, toStageContext(stage, todosByStage[stage.ID]))
	}
	return ai.ProjectContext{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: stringFromNull(project.Description),
		Status:      project.Status,
		StartDate:   timeFromNull(project.StartDate),
		TargetDate:  timeFromNull(project.TargetDate),
		Position:    project.Position,
		Stages:      stageContexts,
	}
}

func toStageContext(stage dbsqlc.Stage, todos []ai.TodoContext) ai.StageContext {
	return ai.StageContext{
		ID:          stage.ID.String(),
		Name:        stage.Name,
		Description: stringFromNull(stage.Description),
		Status:      stage.Status,
		StartDate:   timeFromNull(stage.StartDate),
		TargetDate:  timeFromNull(stage.TargetDate),
		Position:    stage.Position,
		Todos:       todos,
	}
}

func toTodoContext(todo dbsqlc.Todo) ai.TodoContext {
	return ai.TodoContext{
		ID:              todo.ID.String(),
		Name:            todo.Title,
		Description:     stringFromNull(todo.Description),
		Priority:        todo.Priority,
		Status:          todo.Status,
		StartDate:       timeFromNull(todo.StartDate),
		DueDate:         timeFromNull(todo.DueDate),
		EstimatedEffort: stringFromNull(todo.EstimatedEffort),
		Position:        todo.Position,
		NextAction:      todo.NextAction,
		Milestone:       todo.Milestone,
		Recurrence:      todo.Recurrence,
	}
}

func stringFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func timeFromNull(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.Format(time.RFC3339)
	return &formatted
}

func (r *Resolver) applyAIPlan(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, proposal dbsqlc.AiProposal, plan ai.Plan) error {
	slog.Debug("ai proposal apply plan",
		"proposalID", proposal.ID,
		"parentType", proposal.ParentType,
		"parentID", proposal.ParentID,
		"plan", jsonForGraphLog(plan),
	)

	if proposal.ParentType == model.AIProposalParentTypeProject.String() {
		projectID := proposal.ParentID
		for stageIndex, stageChange := range plan.Stages {
			stageID, err := applyAIStage(ctx, q, userID, proposal.ID, projectID, stageIndex, stageChange)
			if err != nil {
				return err
			}
			for todoIndex, todoChange := range stageChange.Todos {
				if err := applyAITodo(ctx, q, userID, proposal.ID, projectID, stageID, stageIndex, todoIndex, todoChange); err != nil {
					return err
				}
			}
		}
		return nil
	}

	for projectIndex, projectChange := range plan.Projects {
		var projectID uuid.UUID
		if projectRef := meaningfulAIRef(projectChange.ID); projectRef != "" {
			id, err := parseUUID(projectRef)
			if err != nil {
				slog.Error("ai proposal project update id parse failed", "proposalID", proposal.ID, "index", projectIndex, "change", jsonForGraphLog(projectChange), "err", err)
				return err
			}
			project, err := q.UpdateProject(ctx, dbsqlc.UpdateProjectParams{
				ID:          id,
				UserID:      userID,
				Name:        nullableString(projectChange.Name),
				Description: nullString(projectChange.Description),
				Status:      nullableStringPtr(projectChange.Status),
				StartDate:   dateString(projectChange.StartDate),
				TargetDate:  dateString(projectChange.TargetDate),
				Position:    nullInt32(projectChange.Position),
			})
			if err != nil {
				slog.Error("ai proposal project update failed", "proposalID", proposal.ID, "index", projectIndex, "projectID", id, "change", jsonForGraphLog(projectChange), "err", err)
				return err
			}
			projectID = project.ID
			slog.Debug("ai proposal project updated", "proposalID", proposal.ID, "index", projectIndex, "projectID", projectID)
		} else {
			project, err := q.CreateProject(ctx, dbsqlc.CreateProjectParams{
				UserID:      userID,
				EpicID:      uuid.NullUUID{UUID: proposal.ParentID, Valid: true},
				Name:        projectChange.Name,
				Description: nullString(projectChange.Description),
				Status:      optionalStringPtr(projectChange.Status),
				StartDate:   dateString(projectChange.StartDate),
				TargetDate:  dateString(projectChange.TargetDate),
				Position:    optionalInt(projectChange.Position),
			})
			if err != nil {
				slog.Error("ai proposal project create failed", "proposalID", proposal.ID, "index", projectIndex, "change", jsonForGraphLog(projectChange), "err", err)
				return err
			}
			projectID = project.ID
			slog.Debug("ai proposal project created", "proposalID", proposal.ID, "index", projectIndex, "projectID", projectID)
		}

		for stageIndex, stageChange := range projectChange.Stages {
			stageID, err := applyAIStage(ctx, q, userID, proposal.ID, projectID, stageIndex, stageChange)
			if err != nil {
				return err
			}
			for todoIndex, todoChange := range stageChange.Todos {
				if err := applyAITodo(ctx, q, userID, proposal.ID, projectID, stageID, stageIndex, todoIndex, todoChange); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func applyAIStage(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, proposalID uuid.UUID, projectID uuid.UUID, index int, change ai.StageChange) (uuid.UUID, error) {
	if stageRef := meaningfulAIRef(change.ID); stageRef != "" {
		id, err := parseUUID(stageRef)
		if err != nil {
			slog.Error("ai proposal stage update id parse failed", "proposalID", proposalID, "index", index, "change", jsonForGraphLog(change), "err", err)
			return uuid.Nil, err
		}
		stage, err := q.UpdateStage(ctx, dbsqlc.UpdateStageParams{
			ID:          id,
			UserID:      userID,
			Name:        nullableString(change.Name),
			Description: nullString(change.Description),
			Status:      nullableStringPtr(change.Status),
			StartDate:   dateString(change.StartDate),
			TargetDate:  dateString(change.TargetDate),
			Position:    nullInt32(change.Position),
		})
		if err != nil {
			slog.Error("ai proposal stage update failed", "proposalID", proposalID, "index", index, "stageID", id, "change", jsonForGraphLog(change), "err", err)
			return uuid.Nil, err
		}
		slog.Debug("ai proposal stage updated", "proposalID", proposalID, "index", index, "stageID", stage.ID, "projectID", projectID)
		return stage.ID, nil
	}

	stage, err := q.CreateStage(ctx, dbsqlc.CreateStageParams{
		UserID:      userID,
		ProjectID:   uuid.NullUUID{UUID: projectID, Valid: true},
		Name:        change.Name,
		Description: nullString(change.Description),
		Status:      optionalStringPtr(change.Status),
		StartDate:   dateString(change.StartDate),
		TargetDate:  dateString(change.TargetDate),
		Position:    optionalInt(change.Position),
	})
	if err != nil {
		slog.Error("ai proposal stage create failed", "proposalID", proposalID, "index", index, "projectID", projectID, "change", jsonForGraphLog(change), "err", err)
		return uuid.Nil, err
	}
	slog.Debug("ai proposal stage created", "proposalID", proposalID, "index", index, "stageID", stage.ID, "projectID", projectID)
	return stage.ID, nil
}

func applyAITodo(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, proposalID uuid.UUID, projectID uuid.UUID, stageID uuid.UUID, stageIndex int, todoIndex int, change ai.TodoChange) error {
	if todoRef := meaningfulAIRef(change.ID); todoRef != "" {
		id, err := parseUUID(todoRef)
		if err != nil {
			slog.Error("ai proposal todo update id parse failed", "proposalID", proposalID, "stageIndex", stageIndex, "todoIndex", todoIndex, "change", jsonForGraphLog(change), "err", err)
			return err
		}
		if boolPtrTrue(change.NextAction) {
			if err := q.ClearProjectNextActions(ctx, dbsqlc.ClearProjectNextActionsParams{UserID: userID, ProjectID: uuid.NullUUID{UUID: projectID, Valid: true}, ID: id}); err != nil {
				slog.Error("ai proposal todo next action clear failed", "proposalID", proposalID, "stageIndex", stageIndex, "todoIndex", todoIndex, "todoID", id, "projectID", projectID, "err", err)
				return err
			}
		}
		_, err = q.UpdateTodo(ctx, dbsqlc.UpdateTodoParams{
			ID:              id,
			UserID:          userID,
			Title:           nullableString(change.Name),
			Description:     nullString(change.Description),
			Priority:        nullableStringPtr(change.Priority),
			Status:          nullableStringPtr(change.Status),
			StartDate:       dateString(change.StartDate),
			DueDate:         dateString(change.DueDate),
			EstimatedEffort: nullString(change.EstimatedEffort),
			Position:        nullInt32(change.Position),
			NextAction:      nullBool(change.NextAction),
			Milestone:       nullBool(change.Milestone),
			Recurrence:      nullableStringPtr(change.Recurrence),
		})
		if err != nil {
			slog.Error("ai proposal todo update failed", "proposalID", proposalID, "stageIndex", stageIndex, "todoIndex", todoIndex, "todoID", id, "change", jsonForGraphLog(change), "err", err)
			return err
		}
		slog.Debug("ai proposal todo updated", "proposalID", proposalID, "stageIndex", stageIndex, "todoIndex", todoIndex, "todoID", id, "projectID", projectID, "stageID", stageID)
		return nil
	}

	if boolPtrTrue(change.NextAction) {
		if err := q.ClearProjectNextActions(ctx, dbsqlc.ClearProjectNextActionsParams{UserID: userID, ProjectID: uuid.NullUUID{UUID: projectID, Valid: true}, ID: uuid.Nil}); err != nil {
			slog.Error("ai proposal todo next action clear failed", "proposalID", proposalID, "stageIndex", stageIndex, "todoIndex", todoIndex, "projectID", projectID, "change", jsonForGraphLog(change), "err", err)
			return err
		}
	}
	todo, err := q.CreateTodo(ctx, dbsqlc.CreateTodoParams{
		UserID:          userID,
		ProjectID:       uuid.NullUUID{UUID: projectID, Valid: true},
		StageID:         uuid.NullUUID{UUID: stageID, Valid: true},
		Title:           change.Name,
		Description:     nullString(change.Description),
		Priority:        optionalStringPtr(change.Priority),
		Status:          optionalStringPtr(change.Status),
		StartDate:       dateString(change.StartDate),
		DueDate:         dateString(change.DueDate),
		EstimatedEffort: nullString(change.EstimatedEffort),
		Position:        optionalInt(change.Position),
		NextAction:      optionalBool(change.NextAction),
		Milestone:       optionalBool(change.Milestone),
		Recurrence:      optionalStringPtr(change.Recurrence),
	})
	if err != nil {
		slog.Error("ai proposal todo create failed", "proposalID", proposalID, "stageIndex", stageIndex, "todoIndex", todoIndex, "projectID", projectID, "stageID", stageID, "change", jsonForGraphLog(change), "err", err)
		return err
	}
	slog.Debug("ai proposal todo created", "proposalID", proposalID, "stageIndex", stageIndex, "todoIndex", todoIndex, "todoID", todo.ID, "projectID", projectID, "stageID", stageID)
	return nil
}

func meaningfulAIRef(value string) string {
	normalized := strings.TrimSpace(value)
	switch strings.ToLower(normalized) {
	case "", "none", "null", "nil", "n/a", "na", "undefined", "-":
		return ""
	default:
		return normalized
	}
}

func nullableString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullableStringPtr(value *string) sql.NullString {
	if value == nil || *value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func optionalStringPtr(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}

func dateString(value *string) sql.NullTime {
	if value == nil || *value == "" {
		return sql.NullTime{}
	}
	parsed, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: parsed, Valid: true}
}

func boolPtrTrue(value *bool) bool {
	return value != nil && *value
}
