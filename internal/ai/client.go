package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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
	ParentType string `json:"parentType"`
	ParentID   string `json:"parentId"`
	Parent     any    `json:"parent"`
	Projects   any    `json:"projects"`
	Stages     any    `json:"stages"`
	Todos      any    `json:"todos"`
}

type Plan struct {
	Summary  string          `json:"summary"`
	Projects []ProjectChange `json:"projects"`
	Stages   []StageChange   `json:"stages"`
	Todos    []TodoChange    `json:"todos"`
}

type ProjectChange struct {
	Action      string  `json:"action"`
	ID          string  `json:"id"`
	TempID      string  `json:"tempId"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	StartDate   *string `json:"startDate"`
	TargetDate  *string `json:"targetDate"`
	Position    *int    `json:"position"`
}

type StageChange struct {
	Action        string  `json:"action"`
	ID            string  `json:"id"`
	TempID        string  `json:"tempId"`
	ProjectID     string  `json:"projectId"`
	ProjectTempID string  `json:"projectTempId"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	Status        *string `json:"status"`
	StartDate     *string `json:"startDate"`
	TargetDate    *string `json:"targetDate"`
	Position      *int    `json:"position"`
}

type TodoChange struct {
	Action          string  `json:"action"`
	ID              string  `json:"id"`
	ProjectID       string  `json:"projectId"`
	ProjectTempID   string  `json:"projectTempId"`
	StageID         string  `json:"stageId"`
	StageTempID     string  `json:"stageTempId"`
	Title           string  `json:"title"`
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
		cfg.Timeout = 30 * time.Second
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

	input, err := json.Marshal(map[string]any{
		"magicText": magicText,
		"context":   appContext,
	})
	if err != nil {
		return Plan{}, err
	}

	body, err := json.Marshal(map[string]any{
		"model": c.model,
		"input": []map[string]any{
			{
				"role":    "system",
				"content": "You generate safe project planning proposals as JSON. Only propose changes under the provided parent. Prefer creating new items unless the user clearly asks to modify existing items. Use action CREATE or UPDATE. For new items use stable tempId values. Do not delete anything.",
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
				"strict":      false,
				"schema":      proposalSchema(),
			},
		},
	})
	if err != nil {
		return Plan{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return Plan{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if c.projectID != "" {
		req.Header.Set("OpenAI-Project", c.projectID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return Plan{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Plan{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Plan{}, fmt.Errorf("openai response status %d: %s", resp.StatusCode, string(respBody))
	}

	text, err := responseText(respBody)
	if err != nil {
		return Plan{}, err
	}
	var plan Plan
	if err := json.Unmarshal([]byte(text), &plan); err != nil {
		return Plan{}, fmt.Errorf("decode ai proposal: %w", err)
	}
	if plan.Summary == "" {
		return Plan{}, fmt.Errorf("ai proposal summary is empty")
	}
	return plan, nil
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

func proposalSchema() map[string]any {
	changeProps := map[string]any{
		"action":          map[string]any{"type": "string", "enum": []string{"CREATE", "UPDATE"}},
		"id":              map[string]any{"type": "string"},
		"tempId":          map[string]any{"type": "string"},
		"name":            map[string]any{"type": "string"},
		"title":           map[string]any{"type": "string"},
		"description":     map[string]any{"type": "string"},
		"status":          map[string]any{"type": "string"},
		"priority":        map[string]any{"type": "string"},
		"recurrence":      map[string]any{"type": "string"},
		"startDate":       map[string]any{"type": "string"},
		"targetDate":      map[string]any{"type": "string"},
		"dueDate":         map[string]any{"type": "string"},
		"estimatedEffort": map[string]any{"type": "string"},
		"position":        map[string]any{"type": "integer"},
		"projectId":       map[string]any{"type": "string"},
		"projectTempId":   map[string]any{"type": "string"},
		"stageId":         map[string]any{"type": "string"},
		"stageTempId":     map[string]any{"type": "string"},
		"nextAction":      map[string]any{"type": "boolean"},
		"milestone":       map[string]any{"type": "boolean"},
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary":  map[string]any{"type": "string"},
			"projects": map[string]any{"type": "array", "items": map[string]any{"type": "object", "properties": changeProps}},
			"stages":   map[string]any{"type": "array", "items": map[string]any{"type": "object", "properties": changeProps}},
			"todos":    map[string]any{"type": "array", "items": map[string]any{"type": "object", "properties": changeProps}},
		},
		"required": []string{"summary", "projects", "stages", "todos"},
	}
}
