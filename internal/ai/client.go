package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const logPayloadLimit = 4000

type Client struct {
	apiKey    string
	projectID string
	baseURL   string
	model     string
	http      *http.Client
}

type Config struct {
	APIKey    string
	ProjectID string
	BaseURL   string
	Model     string
	Timeout   time.Duration
}

type Context struct {
	ParentType  string           `json:"parentType"`
	ParentID    string           `json:"parentId"`
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	Color       *string          `json:"color"`
	Projects    []ProjectContext `json:"projects"`
}

type ProjectContext struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	Status      string         `json:"status"`
	StartDate   *string        `json:"startDate"`
	TargetDate  *string        `json:"targetDate"`
	Position    int32          `json:"position"`
	Stages      []StageContext `json:"stages"`
}

type StageContext struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	Status      string        `json:"status"`
	StartDate   *string       `json:"startDate"`
	TargetDate  *string       `json:"targetDate"`
	Position    int32         `json:"position"`
	Todos       []TodoContext `json:"todos"`
}

type TodoContext struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	Priority        string  `json:"priority"`
	Status          string  `json:"status"`
	StartDate       *string `json:"startDate"`
	DueDate         *string `json:"dueDate"`
	EstimatedEffort *string `json:"estimatedEffort"`
	Position        int32   `json:"position"`
	NextAction      bool    `json:"nextAction"`
	Milestone       bool    `json:"milestone"`
	Recurrence      string  `json:"recurrence"`
}

type Plan struct {
	Summary  string          `json:"summary"`
	Projects []ProjectChange `json:"projects,omitempty"`
	Stages   []StageChange   `json:"stages,omitempty"`
}

func (p Plan) ProjectCount() int {
	return len(p.Projects)
}

func (p Plan) StageCount() int {
	count := len(p.Stages)
	for _, project := range p.Projects {
		count += len(project.Stages)
	}
	return count
}

func (p Plan) TodoCount() int {
	count := 0
	for _, stage := range p.Stages {
		count += len(stage.Todos)
	}
	for _, project := range p.Projects {
		for _, stage := range project.Stages {
			count += len(stage.Todos)
		}
	}
	return count
}

type ProjectChange struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	Status      *string       `json:"status"`
	StartDate   *string       `json:"startDate"`
	TargetDate  *string       `json:"targetDate"`
	Position    *int          `json:"position"`
	Stages      []StageChange `json:"stages"`
}

type StageChange struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description *string      `json:"description"`
	Status      *string      `json:"status"`
	StartDate   *string      `json:"startDate"`
	TargetDate  *string      `json:"targetDate"`
	Position    *int         `json:"position"`
	Todos       []TodoChange `json:"todos"`
}

type TodoChange struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	Priority        *string `json:"priority"`
	Status          *string `json:"status"`
	StartDate       *string `json:"startDate"`
	DueDate         *string `json:"dueDate"`
	EstimatedEffort *string `json:"estimatedEffort"`
	Position        *int    `json:"position"`
	NextAction      *bool   `json:"nextAction"`
	Milestone       *bool   `json:"milestone"`
	Recurrence      *string `json:"recurrence"`
}

func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-5.4-mini"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 90 * time.Second
	}
	return &Client{
		apiKey:    cfg.APIKey,
		projectID: cfg.ProjectID,
		baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
		model:     cfg.Model,
		http:      &http.Client{Timeout: cfg.Timeout},
	}
}

func (c *Client) GeneratePlan(ctx context.Context, magicText string, appContext Context) (Plan, error) {
	if c.apiKey == "" {
		return Plan{}, fmt.Errorf("OPENAI_API_KEY is required")
	}
	started := time.Now()
	slog.Info("openai proposal request started",
		"model", c.model,
		"baseURL", c.baseURL,
		"parentType", appContext.ParentType,
		"parentID", appContext.ParentID,
		"projectCount", collectionLen(appContext.Projects),
		"stageCount", countContextStages(appContext),
		"todoCount", countContextTodos(appContext),
		"magicTextChars", len([]rune(magicText)),
	)
	slog.Debug("openai proposal request content",
		"magicText", truncateLogValue(magicText),
		"context", jsonForLog(appContext),
	)

	input, err := json.Marshal(map[string]any{
		"magicText": magicText,
		"context":   appContext,
	})
	if err != nil {
		slog.Error("openai proposal input encode failed", "err", err)
		return Plan{}, err
	}

	body, err := json.Marshal(map[string]any{
		"model": c.model,
		"input": []map[string]any{
			{
				"role":    "system",
				"content": "You generate safe project planning proposals as JSON. Only propose changes under the provided parent tree. Use the same nested hierarchy as the input context. If an object has an id, it means updating that existing object; only use ids that are present in the input context under the selected parent. If an object has no id or id is null, it means creating a new object. Do not use temp ids. Do not use action fields. Do not delete anything. Use name for projects, stages, and todos. When parentType is EPIC, the response root must contain projects, and every stage must be nested inside its project, every todo inside its stage. When parentType is PROJECT, the response root must contain stages only, and every todo must be nested inside its stage. Do not send extra fields. Use null for optional values that do not apply.",
			},
			{
				"role":    "user",
				"content": string(input),
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":        "json_schema",
				"name":        "todo_ai_proposal",
				"description": "Project, stage, and todo changes proposed for user review.",
				"strict":      true,
				"schema":      proposalSchema(appContext.ParentType),
			},
		},
	})
	if err != nil {
		slog.Error("openai proposal request encode failed", "err", err)
		return Plan{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		slog.Error("openai proposal request create failed", "err", err)
		return Plan{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if c.projectID != "" {
		req.Header.Set("OpenAI-Project", c.projectID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		slog.Error("openai proposal request failed", "model", c.model, "duration", time.Since(started).String(), "err", err)
		return Plan{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("openai proposal response read failed", "status", resp.StatusCode, "duration", time.Since(started).String(), "err", err)
		return Plan{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyText := string(respBody)
		slog.Error("openai proposal response status failed",
			"status", resp.StatusCode,
			"duration", time.Since(started).String(),
			"body", truncateLogValue(bodyText),
		)
		return Plan{}, fmt.Errorf("openai response status %d: %s", resp.StatusCode, truncateLogValue(bodyText))
	}

	text, err := responseText(respBody)
	if err != nil {
		slog.Error("openai proposal response parse failed",
			"status", resp.StatusCode,
			"duration", time.Since(started).String(),
			"body", truncateLogValue(string(respBody)),
			"err", err,
		)
		return Plan{}, err
	}
	slog.Debug("openai proposal response content", "response", truncateLogValue(text))

	var plan Plan
	if err := json.Unmarshal([]byte(text), &plan); err != nil {
		slog.Error("openai proposal json decode failed", "duration", time.Since(started).String(), "err", err)
		return Plan{}, fmt.Errorf("decode ai proposal: %w", err)
	}
	if plan.Summary == "" {
		slog.Error("openai proposal validation failed", "duration", time.Since(started).String(), "err", "empty summary")
		return Plan{}, fmt.Errorf("ai proposal summary is empty")
	}
	if err := ValidatePlan(plan, appContext); err != nil {
		slog.Error("openai proposal validation failed", "duration", time.Since(started).String(), "err", err, "plan", jsonForLog(plan))
		return Plan{}, err
	}
	slog.Info("openai proposal request completed",
		"model", c.model,
		"duration", time.Since(started).String(),
		"projectChanges", plan.ProjectCount(),
		"stageChanges", plan.StageCount(),
		"todoChanges", plan.TodoCount(),
		"summary", truncateLogValue(plan.Summary),
	)
	return plan, nil
}

func collectionLen(value any) int {
	if value == nil {
		return 0
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return 0
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return v.Len()
	default:
		return 0
	}
}

func countContextStages(appContext Context) int {
	count := 0
	for _, project := range appContext.Projects {
		count += len(project.Stages)
	}
	return count
}

func countContextTodos(appContext Context) int {
	count := 0
	for _, project := range appContext.Projects {
		for _, stage := range project.Stages {
			count += len(stage.Todos)
		}
	}
	return count
}

func jsonForLog(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("<json encode failed: %v>", err)
	}
	return truncateLogValue(string(body))
}

func truncateLogValue(value string) string {
	runes := []rune(value)
	if len(runes) <= logPayloadLimit {
		return value
	}
	return string(runes[:logPayloadLimit]) + "...(truncated)"
}

func responseText(body []byte) (string, error) {
	var parsed struct {
		Output []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	for _, output := range parsed.Output {
		for _, content := range output.Content {
			if content.Text != "" {
				return content.Text, nil
			}
		}
	}
	return "", fmt.Errorf("openai response did not contain output text")
}

func ValidatePlan(plan Plan, appContext Context) error {
	if strings.TrimSpace(plan.Summary) == "" {
		return fmt.Errorf("ai proposal summary is empty")
	}

	switch appContext.ParentType {
	case "EPIC":
		if len(plan.Projects) == 0 {
			return fmt.Errorf("ai proposal must contain projects under EPIC parent")
		}
		if len(plan.Stages) != 0 {
			return fmt.Errorf("ai proposal must not contain root stages under EPIC parent")
		}
		contextProjects := map[string]ProjectContext{}
		for _, project := range appContext.Projects {
			contextProjects[project.ID] = project
		}
		for projectIndex, project := range plan.Projects {
			contextProject, existingProject := contextProjects[meaningfulPlanRef(project.ID)]
			if err := validateProjectNode(project, existingProject, projectIndex); err != nil {
				return err
			}
			if err := validateStageNodes(project.Stages, contextProject.Stages, fmt.Sprintf("project %d", projectIndex)); err != nil {
				return err
			}
		}
	case "PROJECT":
		if len(plan.Projects) != 0 {
			return fmt.Errorf("ai proposal must not contain root projects under PROJECT parent")
		}
		if len(appContext.Projects) != 1 {
			return fmt.Errorf("ai project context must contain exactly one selected project")
		}
		if len(plan.Stages) == 0 {
			return fmt.Errorf("ai proposal must contain stages under PROJECT parent")
		}
		if err := validateStageNodes(plan.Stages, appContext.Projects[0].Stages, "project parent"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported ai proposal parent type %q", appContext.ParentType)
	}
	return nil
}

func validateProjectNode(project ProjectChange, exists bool, index int) error {
	id := meaningfulPlanRef(project.ID)
	if id != "" && !exists {
		return fmt.Errorf("project %d: id must reference an existing project from context", index)
	}
	if strings.TrimSpace(project.Name) == "" {
		return fmt.Errorf("project %d: name is required", index)
	}
	return nil
}

func validateStageNodes(stages []StageChange, contextStages []StageContext, parentPath string) error {
	contextByID := map[string]StageContext{}
	for _, stage := range contextStages {
		contextByID[stage.ID] = stage
	}
	for stageIndex, stage := range stages {
		id := meaningfulPlanRef(stage.ID)
		contextStage, exists := contextByID[id]
		if id != "" && !exists {
			return fmt.Errorf("%s stage %d: id must reference an existing stage from context", parentPath, stageIndex)
		}
		if strings.TrimSpace(stage.Name) == "" {
			return fmt.Errorf("%s stage %d: name is required", parentPath, stageIndex)
		}
		if err := validateTodoNodes(stage.Todos, contextStage.Todos, fmt.Sprintf("%s stage %d", parentPath, stageIndex)); err != nil {
			return err
		}
	}
	return nil
}

func validateTodoNodes(todos []TodoChange, contextTodos []TodoContext, parentPath string) error {
	contextByID := map[string]struct{}{}
	for _, todo := range contextTodos {
		contextByID[todo.ID] = struct{}{}
	}
	for todoIndex, todo := range todos {
		id := meaningfulPlanRef(todo.ID)
		if id != "" {
			if _, exists := contextByID[id]; !exists {
				return fmt.Errorf("%s todo %d: id must reference an existing todo from context", parentPath, todoIndex)
			}
		}
		if strings.TrimSpace(todo.Name) == "" {
			return fmt.Errorf("%s todo %d: name is required", parentPath, todoIndex)
		}
	}
	return nil
}

func meaningfulPlanRef(value string) string {
	trimmed := strings.TrimSpace(value)
	switch strings.ToLower(trimmed) {
	case "", "none", "null", "nil", "n/a", "na", "undefined", "-":
		return ""
	default:
		return trimmed
	}
}

func proposalSchema(parentType string) map[string]any {
	if parentType == "PROJECT" {
		return strictObjectSchema(map[string]any{
			"summary": map[string]any{"type": "string"},
			"stages":  map[string]any{"type": "array", "items": stageChangeSchema()},
		})
	}
	return strictObjectSchema(map[string]any{
		"summary":  map[string]any{"type": "string"},
		"projects": map[string]any{"type": "array", "items": projectChangeSchema()},
	})
}

func projectChangeSchema() map[string]any {
	props := map[string]any{
		"id":          nullableStringSchema("Existing project id from context for update. Empty or null for create."),
		"name":        nullableStringSchema("Project name. Required by backend validation."),
		"description": nullableStringSchema("Optional project description."),
		"status":      nullableEnumSchema([]string{"ACTIVE", "PAUSED", "DONE", "ARCHIVED"}),
		"startDate":   nullableStringSchema("Optional ISO-8601 start date."),
		"targetDate":  nullableStringSchema("Optional ISO-8601 target date."),
		"position":    nullableIntegerSchema(),
		"stages":      map[string]any{"type": "array", "items": stageChangeSchema()},
	}
	return strictObjectSchema(props)
}

func stageChangeSchema() map[string]any {
	props := map[string]any{
		"id":          nullableStringSchema("Existing stage id from context for update. Empty or null for create."),
		"name":        nullableStringSchema("Stage name. Required by backend validation."),
		"description": nullableStringSchema("Optional stage description."),
		"status":      nullableEnumSchema([]string{"PLANNED", "IN_PROGRESS", "DONE"}),
		"startDate":   nullableStringSchema("Optional ISO-8601 start date."),
		"targetDate":  nullableStringSchema("Optional ISO-8601 target date."),
		"position":    nullableIntegerSchema(),
		"todos":       map[string]any{"type": "array", "items": todoChangeSchema()},
	}
	return strictObjectSchema(props)
}

func todoChangeSchema() map[string]any {
	props := map[string]any{
		"id":              nullableStringSchema("Existing todo id from context for update. Empty or null for create."),
		"name":            nullableStringSchema("Todo name. Required by backend validation."),
		"description":     nullableStringSchema("Optional todo description."),
		"priority":        nullableEnumSchema([]string{"LOW", "NORMAL", "HIGH", "CRITICAL"}),
		"status":          nullableEnumSchema([]string{"OPEN", "IN_PROGRESS", "DONE", "BLOCKED"}),
		"startDate":       nullableStringSchema("Optional ISO-8601 start date."),
		"dueDate":         nullableStringSchema("Optional ISO-8601 due date."),
		"estimatedEffort": nullableStringSchema("Optional effort estimate."),
		"position":        nullableIntegerSchema(),
		"nextAction":      nullableBooleanSchema(),
		"milestone":       nullableBooleanSchema(),
		"recurrence":      nullableEnumSchema([]string{"NONE", "DAILY", "WEEKLY", "MONTHLY", "YEARLY"}),
	}
	return strictObjectSchema(props)
}

func strictObjectSchema(properties map[string]any) map[string]any {
	required := make([]string, 0, len(properties))
	for name := range properties {
		required = append(required, name)
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
		"required":             required,
	}
}

func nullableStringSchema(description string) map[string]any {
	return map[string]any{"type": []string{"string", "null"}, "description": description}
}

func nullableIntegerSchema() map[string]any {
	return map[string]any{"type": []string{"integer", "null"}}
}

func nullableBooleanSchema() map[string]any {
	return map[string]any{"type": []string{"boolean", "null"}}
}

func nullableEnumSchema(values []string) map[string]any {
	enumValues := make([]any, 0, len(values)+1)
	for _, value := range values {
		enumValues = append(enumValues, value)
	}
	enumValues = append(enumValues, nil)
	return map[string]any{"type": []string{"string", "null"}, "enum": enumValues}
}
