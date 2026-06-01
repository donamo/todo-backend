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
		snapshot.Parent = toEpic(epic)
		snapshot.Projects = projects
		stages, todos, err := childrenForProjects(ctx, q, userID, projects)
		if err != nil {
			return snapshot, err
		}
		snapshot.Stages = stages
		snapshot.Todos = todos
	case model.AIProposalParentTypeProject.String():
		project, err := q.GetProject(ctx, dbsqlc.GetProjectParams{ID: parentID, UserID: userID})
		if err != nil {
			return snapshot, err
		}
		stages, err := q.ListStages(ctx, dbsqlc.ListStagesParams{ProjectID: uuid.NullUUID{UUID: parentID, Valid: true}, UserID: userID})
		if err != nil {
			return snapshot, err
		}
		todos, err := q.ListTodos(ctx, dbsqlc.ListTodosParams{
			UserID:    userID,
			ProjectID: uuid.NullUUID{UUID: parentID, Valid: true},
		})
		if err != nil {
			return snapshot, err
		}
		snapshot.Parent = toProject(project)
		snapshot.Projects = []dbsqlc.Project{project}
		snapshot.Stages = stages
		snapshot.Todos = todos
	default:
		return snapshot, fmt.Errorf("unsupported ai proposal parent type %q", parentType)
	}
	return snapshot, nil
}

func childrenForProjects(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, projects []dbsqlc.Project) ([]dbsqlc.Stage, []dbsqlc.Todo, error) {
	var stages []dbsqlc.Stage
	var todos []dbsqlc.Todo
	for _, project := range projects {
		projectStages, err := q.ListStages(ctx, dbsqlc.ListStagesParams{ProjectID: uuid.NullUUID{UUID: project.ID, Valid: true}, UserID: userID})
		if err != nil {
			return nil, nil, err
		}
		stages = append(stages, projectStages...)
		projectTodos, err := q.ListTodos(ctx, dbsqlc.ListTodosParams{
			UserID:    userID,
			ProjectID: uuid.NullUUID{UUID: project.ID, Valid: true},
		})
		if err != nil {
			return nil, nil, err
		}
		todos = append(todos, projectTodos...)
	}
	return stages, todos, nil
}

func (r *Resolver) applyAIPlan(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, proposal dbsqlc.AiProposal, plan ai.Plan) error {
	projectIDs := map[string]uuid.UUID{}
	stageIDs := map[string]uuid.UUID{}
	slog.Debug("ai proposal apply plan",
		"proposalID", proposal.ID,
		"parentType", proposal.ParentType,
		"parentID", proposal.ParentID,
		"plan", jsonForGraphLog(plan),
	)

	for index, change := range plan.Projects {
		if change.Action == "UPDATE" {
			id, err := parseUUID(change.ID)
			if err != nil {
				slog.Error("ai proposal project update id parse failed", "proposalID", proposal.ID, "index", index, "change", jsonForGraphLog(change), "err", err)
				return err
			}
			_, err = q.UpdateProject(ctx, dbsqlc.UpdateProjectParams{
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
				slog.Error("ai proposal project update failed", "proposalID", proposal.ID, "index", index, "projectID", id, "change", jsonForGraphLog(change), "err", err)
				return err
			}
			continue
		}

		epicID := uuid.NullUUID{}
		if proposal.ParentType == model.AIProposalParentTypeEpic.String() {
			epicID = uuid.NullUUID{UUID: proposal.ParentID, Valid: true}
		}
		project, err := q.CreateProject(ctx, dbsqlc.CreateProjectParams{
			UserID:      userID,
			EpicID:      epicID,
			Name:        requiredName(change.Name, "AI project"),
			Description: nullString(change.Description),
			Status:      optionalStringPtr(change.Status),
			StartDate:   dateString(change.StartDate),
			TargetDate:  dateString(change.TargetDate),
			Position:    optionalInt(change.Position),
		})
		if err != nil {
			slog.Error("ai proposal project create failed", "proposalID", proposal.ID, "index", index, "change", jsonForGraphLog(change), "err", err)
			return err
		}
		if tempID := meaningfulAIRef(change.TempID); tempID != "" {
			projectIDs[tempID] = project.ID
		}
		slog.Debug("ai proposal project created", "proposalID", proposal.ID, "index", index, "projectID", project.ID, "tempID", change.TempID)
	}

	for index, change := range plan.Stages {
		if change.Action == "UPDATE" {
			id, err := parseUUID(change.ID)
			if err != nil {
				slog.Error("ai proposal stage update id parse failed", "proposalID", proposal.ID, "index", index, "change", jsonForGraphLog(change), "err", err)
				return err
			}
			_, err = q.UpdateStage(ctx, dbsqlc.UpdateStageParams{
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
				slog.Error("ai proposal stage update failed", "proposalID", proposal.ID, "index", index, "stageID", id, "change", jsonForGraphLog(change), "err", err)
				return err
			}
			continue
		}

		projectID, err := stageProjectID(change, proposal, projectIDs)
		if err != nil {
			slog.Error("ai proposal stage project resolve failed", "proposalID", proposal.ID, "index", index, "change", jsonForGraphLog(change), "projectTempIDs", len(projectIDs), "err", err)
			return err
		}
		if err := validateAIProject(ctx, q, userID, projectID); err != nil {
			slog.Error("ai proposal stage project validation failed", "proposalID", proposal.ID, "index", index, "projectID", projectID.UUID, "projectValid", projectID.Valid, "change", jsonForGraphLog(change), "err", err)
			return err
		}
		stage, err := q.CreateStage(ctx, dbsqlc.CreateStageParams{
			UserID:      userID,
			ProjectID:   projectID,
			Name:        requiredName(change.Name, "AI stage"),
			Description: nullString(change.Description),
			Status:      optionalStringPtr(change.Status),
			StartDate:   dateString(change.StartDate),
			TargetDate:  dateString(change.TargetDate),
			Position:    optionalInt(change.Position),
		})
		if err != nil {
			slog.Error("ai proposal stage create failed", "proposalID", proposal.ID, "index", index, "projectID", projectID.UUID, "projectValid", projectID.Valid, "change", jsonForGraphLog(change), "err", err)
			return err
		}
		if tempID := meaningfulAIRef(change.TempID); tempID != "" {
			stageIDs[tempID] = stage.ID
		}
		slog.Debug("ai proposal stage created", "proposalID", proposal.ID, "index", index, "stageID", stage.ID, "projectID", projectID.UUID, "projectValid", projectID.Valid, "tempID", change.TempID)
	}

	for index, change := range plan.Todos {
		if change.Action == "UPDATE" {
			id, err := parseUUID(change.ID)
			if err != nil {
				slog.Error("ai proposal todo update id parse failed", "proposalID", proposal.ID, "index", index, "change", jsonForGraphLog(change), "err", err)
				return err
			}
			if boolPtrTrue(change.NextAction) {
				todo, err := q.GetTodo(ctx, dbsqlc.GetTodoParams{ID: id, UserID: userID})
				if err != nil {
					slog.Error("ai proposal todo lookup failed", "proposalID", proposal.ID, "index", index, "todoID", id, "change", jsonForGraphLog(change), "err", err)
					return err
				}
				if todo.ProjectID.Valid {
					if err := q.ClearProjectNextActions(ctx, dbsqlc.ClearProjectNextActionsParams{UserID: userID, ProjectID: todo.ProjectID, ID: id}); err != nil {
						slog.Error("ai proposal todo next action clear failed", "proposalID", proposal.ID, "index", index, "todoID", id, "projectID", todo.ProjectID.UUID, "err", err)
						return err
					}
				}
			}
			_, err = q.UpdateTodo(ctx, dbsqlc.UpdateTodoParams{
				ID:              id,
				UserID:          userID,
				Title:           nullableString(change.Title),
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
				slog.Error("ai proposal todo update failed", "proposalID", proposal.ID, "index", index, "todoID", id, "change", jsonForGraphLog(change), "err", err)
				return err
			}
			continue
		}

		projectID, err := todoProjectID(change, proposal, projectIDs)
		if err != nil {
			slog.Error("ai proposal todo project resolve failed", "proposalID", proposal.ID, "index", index, "change", jsonForGraphLog(change), "projectTempIDs", len(projectIDs), "err", err)
			return err
		}
		stageID, err := todoStageID(change, stageIDs)
		if err != nil {
			slog.Error("ai proposal todo stage resolve failed", "proposalID", proposal.ID, "index", index, "change", jsonForGraphLog(change), "stageTempIDs", len(stageIDs), "err", err)
			return err
		}
		if err := validateAIProject(ctx, q, userID, projectID); err != nil {
			slog.Error("ai proposal todo project validation failed", "proposalID", proposal.ID, "index", index, "projectID", projectID.UUID, "projectValid", projectID.Valid, "change", jsonForGraphLog(change), "err", err)
			return err
		}
		if err := validateAIStage(ctx, q, userID, stageID, projectID); err != nil {
			slog.Error("ai proposal todo stage validation failed", "proposalID", proposal.ID, "index", index, "projectID", projectID.UUID, "projectValid", projectID.Valid, "stageID", stageID.UUID, "stageValid", stageID.Valid, "change", jsonForGraphLog(change), "err", err)
			return err
		}
		if boolPtrTrue(change.NextAction) && projectID.Valid {
			if err := q.ClearProjectNextActions(ctx, dbsqlc.ClearProjectNextActionsParams{UserID: userID, ProjectID: projectID, ID: uuid.Nil}); err != nil {
				slog.Error("ai proposal todo next action clear failed", "proposalID", proposal.ID, "index", index, "projectID", projectID.UUID, "change", jsonForGraphLog(change), "err", err)
				return err
			}
		}
		todo, err := q.CreateTodo(ctx, dbsqlc.CreateTodoParams{
			UserID:          userID,
			ProjectID:       projectID,
			StageID:         stageID,
			Title:           requiredName(change.Title, "AI todo"),
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
			slog.Error("ai proposal todo create failed", "proposalID", proposal.ID, "index", index, "projectID", projectID.UUID, "projectValid", projectID.Valid, "stageID", stageID.UUID, "stageValid", stageID.Valid, "change", jsonForGraphLog(change), "err", err)
			return err
		}
		slog.Debug("ai proposal todo created", "proposalID", proposal.ID, "index", index, "todoID", todo.ID, "projectID", projectID.UUID, "projectValid", projectID.Valid, "stageID", stageID.UUID, "stageValid", stageID.Valid)
	}

	return nil
}

func validateAIProject(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, projectID uuid.NullUUID) error {
	if !projectID.Valid {
		return nil
	}
	if _, err := q.GetProject(ctx, dbsqlc.GetProjectParams{ID: projectID.UUID, UserID: userID}); err != nil {
		return fmt.Errorf("ai proposal references unavailable project %s: %w", projectID.UUID, err)
	}
	return nil
}

func validateAIStage(ctx context.Context, q *dbsqlc.Queries, userID uuid.UUID, stageID uuid.NullUUID, projectID uuid.NullUUID) error {
	if !stageID.Valid {
		return nil
	}
	stage, err := q.GetStage(ctx, dbsqlc.GetStageParams{ID: stageID.UUID, UserID: userID})
	if err != nil {
		return fmt.Errorf("ai proposal references unavailable stage %s: %w", stageID.UUID, err)
	}
	if !projectID.Valid {
		return nil
	}
	if !stage.ProjectID.Valid {
		return fmt.Errorf("ai proposal references stage %s under project %s, but stage has no project", stageID.UUID, projectID.UUID)
	}
	if stage.ProjectID.UUID != projectID.UUID {
		return fmt.Errorf("ai proposal references stage %s under project %s, but stage belongs to project %s", stageID.UUID, projectID.UUID, stage.ProjectID.UUID)
	}
	return nil
}

func stageProjectID(change ai.StageChange, proposal dbsqlc.AiProposal, projectIDs map[string]uuid.UUID) (uuid.NullUUID, error) {
	if proposal.ParentType == model.AIProposalParentTypeProject.String() {
		return uuid.NullUUID{UUID: proposal.ParentID, Valid: true}, nil
	}
	if projectTempID := meaningfulAIRef(change.ProjectTempID); projectTempID != "" {
		if id, ok := projectIDs[projectTempID]; ok {
			return uuid.NullUUID{UUID: id, Valid: true}, nil
		}
		return uuid.NullUUID{}, fmt.Errorf("unknown project temp id %q", change.ProjectTempID)
	}
	if projectID := meaningfulAIRef(change.ProjectID); projectID != "" {
		id, err := parseUUID(projectID)
		return uuid.NullUUID{UUID: id, Valid: err == nil}, err
	}
	if proposal.ParentType == model.AIProposalParentTypeEpic.String() && len(projectIDs) == 1 {
		for _, id := range projectIDs {
			return uuid.NullUUID{UUID: id, Valid: true}, nil
		}
	}
	return uuid.NullUUID{}, nil
}

func todoProjectID(change ai.TodoChange, proposal dbsqlc.AiProposal, projectIDs map[string]uuid.UUID) (uuid.NullUUID, error) {
	if proposal.ParentType == model.AIProposalParentTypeProject.String() {
		return uuid.NullUUID{UUID: proposal.ParentID, Valid: true}, nil
	}
	if projectTempID := meaningfulAIRef(change.ProjectTempID); projectTempID != "" {
		if id, ok := projectIDs[projectTempID]; ok {
			return uuid.NullUUID{UUID: id, Valid: true}, nil
		}
		return uuid.NullUUID{}, fmt.Errorf("unknown project temp id %q", change.ProjectTempID)
	}
	if projectID := meaningfulAIRef(change.ProjectID); projectID != "" {
		id, err := parseUUID(projectID)
		return uuid.NullUUID{UUID: id, Valid: err == nil}, err
	}
	if proposal.ParentType == model.AIProposalParentTypeEpic.String() && len(projectIDs) == 1 {
		for _, id := range projectIDs {
			return uuid.NullUUID{UUID: id, Valid: true}, nil
		}
	}
	return uuid.NullUUID{}, nil
}

func todoStageID(change ai.TodoChange, stageIDs map[string]uuid.UUID) (uuid.NullUUID, error) {
	if stageTempID := meaningfulAIRef(change.StageTempID); stageTempID != "" {
		if id, ok := stageIDs[stageTempID]; ok {
			return uuid.NullUUID{UUID: id, Valid: true}, nil
		}
		return uuid.NullUUID{}, fmt.Errorf("unknown stage temp id %q", change.StageTempID)
	}
	if stageID := meaningfulAIRef(change.StageID); stageID != "" {
		id, err := parseUUID(stageID)
		return uuid.NullUUID{UUID: id, Valid: err == nil}, err
	}
	return uuid.NullUUID{}, nil
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

func requiredName(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func boolPtrTrue(value *bool) bool {
	return value != nil && *value
}
