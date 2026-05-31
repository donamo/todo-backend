package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGraphQLRequiresAuth(t *testing.T) {
	s := NewSuite(t)

	resp, err := http.Post(s.Server.URL+"/graphql", "application/json", bytes.NewBufferString(`{"query":"{ dashboard { openTodos } }"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestAIProposalWorkflowWithMockOpenAI(t *testing.T) {
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("unexpected openai path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-openai-key" {
			t.Fatalf("unexpected authorization header: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("OpenAI-Project") != "test-project-id" {
			t.Fatalf("unexpected project header: %q", r.Header.Get("OpenAI-Project"))
		}
		w.Header().Set("Content-Type", "application/json")
		plan := map[string]any{
			"summary":  "Stage és todo javaslat a magic text alapján.",
			"projects": []any{},
			"stages": []any{
				map[string]any{
					"action": "CREATE",
					"tempId": "stage-1",
					"name":   "Tervezés",
				},
			},
			"todos": []any{
				map[string]any{
					"action":      "CREATE",
					"stageTempId": "stage-1",
					"title":       "Specifikáció pontosítása",
					"priority":    "HIGH",
					"nextAction":  true,
				},
			},
		}
		text, err := json.Marshal(plan)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.NewEncoder(w).Encode(map[string]any{
			"output": []any{
				map[string]any{
					"content": []any{
						map[string]any{
							"type": "output_text",
							"text": string(text),
						},
					},
				},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer mockOpenAI.Close()

	t.Setenv("OPENAI_API_KEY", "test-openai-key")
	t.Setenv("OPENAI_PROJECT_ID", "test-project-id")
	t.Setenv("OPENAI_BASE_URL", mockOpenAI.URL)
	t.Setenv("OPENAI_MODEL", "test-model")

	s := NewSuite(t)
	projectID := createProjectForAITest(t, s)

	var existingNextAction gqlResponse[struct {
		CreateTodo struct {
			ID         string `json:"id"`
			NextAction bool   `json:"nextAction"`
		} `json:"createTodo"`
	}]
	s.GQL(t, `
		mutation CreateExistingNextAction($input: CreateTodoInput!) {
			createTodo(input: $input) { id nextAction }
		}
	`, map[string]any{
		"input": map[string]any{
			"projectId":  projectID,
			"title":      "Meglévő következő lépés",
			"nextAction": true,
		},
	}, &existingNextAction)
	if len(existingNextAction.Errors) > 0 {
		t.Fatalf("create existing next action errors: %+v", existingNextAction.Errors)
	}
	if !existingNextAction.Data.CreateTodo.NextAction {
		t.Fatal("existing todo should start as next action")
	}

	var generate gqlResponse[struct {
		GenerateAIProposal struct {
			ID           string `json:"id"`
			Status       string `json:"status"`
			Summary      string `json:"summary"`
			ProposalJSON string `json:"proposalJson"`
		} `json:"generateAIProposal"`
	}]
	s.GQL(t, `
		mutation Generate($input: GenerateAIProposalInput!) {
			generateAIProposal(input: $input) {
				id
				status
				summary
				proposalJson
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"parentType": "PROJECT",
			"parentId":   projectID,
			"magicText":  "Készíts tervezési szakaszt és első feladatot",
		},
	}, &generate)
	if len(generate.Errors) > 0 {
		t.Fatalf("generateAIProposal errors: %+v", generate.Errors)
	}
	if generate.Data.GenerateAIProposal.Status != "DRAFT" {
		t.Fatalf("proposal status = %q", generate.Data.GenerateAIProposal.Status)
	}
	if generate.Data.GenerateAIProposal.ProposalJSON == "" {
		t.Fatal("proposalJson is empty")
	}

	var accept gqlResponse[struct {
		AcceptAIProposal struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"acceptAIProposal"`
	}]
	s.GQL(t, `
		mutation Accept($id: ID!) {
			acceptAIProposal(id: $id) { id status }
		}
	`, map[string]any{"id": generate.Data.GenerateAIProposal.ID}, &accept)
	if len(accept.Errors) > 0 {
		t.Fatalf("acceptAIProposal errors: %+v", accept.Errors)
	}
	if accept.Data.AcceptAIProposal.Status != "APPLIED" {
		t.Fatalf("accepted status = %q", accept.Data.AcceptAIProposal.Status)
	}

	var query gqlResponse[struct {
		Stages []struct {
			Name string `json:"name"`
		} `json:"stages"`
		Todos []struct {
			Title      string `json:"title"`
			Priority   string `json:"priority"`
			NextAction bool   `json:"nextAction"`
		} `json:"todos"`
	}]
	s.GQL(t, `
		query Check($projectId: ID!) {
			stages(projectId: $projectId) { name }
			todos(filter: { projectId: $projectId }) { title priority nextAction }
		}
	`, map[string]any{"projectId": projectID}, &query)
	if len(query.Errors) > 0 {
		t.Fatalf("query errors: %+v", query.Errors)
	}
	if len(query.Data.Stages) != 1 || query.Data.Stages[0].Name != "Tervezés" {
		t.Fatalf("stages = %+v", query.Data.Stages)
	}
	if len(query.Data.Todos) != 2 {
		t.Fatalf("todos = %+v", query.Data.Todos)
	}
	nextActions := 0
	foundGenerated := false
	for _, todo := range query.Data.Todos {
		if todo.NextAction {
			nextActions++
		}
		if todo.Title == "Specifikáció pontosítása" && todo.Priority == "HIGH" && todo.NextAction {
			foundGenerated = true
		}
	}
	if nextActions != 1 || !foundGenerated {
		t.Fatalf("todos = %+v", query.Data.Todos)
	}
}

func TestGraphQLTodoWorkflow(t *testing.T) {
	s := NewSuite(t)

	var createEpic gqlResponse[struct {
		CreateEpic struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"createEpic"`
	}]
	s.GQL(t, `
		mutation CreateEpic($input: CreateEpicInput!) {
			createEpic(input: $input) { id name }
		}
	`, map[string]any{
		"input": map[string]any{
			"name":  "Fejlesztés",
			"color": "#2563eb",
		},
	}, &createEpic)
	if len(createEpic.Errors) > 0 {
		t.Fatalf("createEpic errors: %+v", createEpic.Errors)
	}
	epicID := createEpic.Data.CreateEpic.ID

	var createProject gqlResponse[struct {
		CreateProject struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"createProject"`
	}]
	s.GQL(t, `
		mutation CreateProject($input: CreateProjectInput!) {
			createProject(input: $input) { id name status }
		}
	`, map[string]any{
		"input": map[string]any{
			"epicId":     epicID,
			"name":       "Todo backend",
			"startDate":  "2026-06-10T00:00:00Z",
			"targetDate": "2026-06-15T00:00:00Z",
		},
	}, &createProject)
	if len(createProject.Errors) > 0 {
		t.Fatalf("createProject errors: %+v", createProject.Errors)
	}
	projectID := createProject.Data.CreateProject.ID
	if createProject.Data.CreateProject.Status != "ACTIVE" {
		t.Fatalf("project status = %q", createProject.Data.CreateProject.Status)
	}

	var createStage gqlResponse[struct {
		CreateStage struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"createStage"`
	}]
	s.GQL(t, `
		mutation CreateStage($input: CreateStageInput!) {
			createStage(input: $input) { id status }
		}
	`, map[string]any{
		"input": map[string]any{
			"projectId": projectID,
			"name":      "Fejlesztés",
		},
	}, &createStage)
	if len(createStage.Errors) > 0 {
		t.Fatalf("createStage errors: %+v", createStage.Errors)
	}
	stageID := createStage.Data.CreateStage.ID

	var createLabel gqlResponse[struct {
		CreateLabel struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"createLabel"`
	}]
	s.GQL(t, `
		mutation CreateLabel($input: CreateLabelInput!) {
			createLabel(input: $input) { id name }
		}
	`, map[string]any{
		"input": map[string]any{
			"name":  "backend",
			"color": "#16a34a",
		},
	}, &createLabel)
	if len(createLabel.Errors) > 0 {
		t.Fatalf("createLabel errors: %+v", createLabel.Errors)
	}
	labelID := createLabel.Data.CreateLabel.ID

	var createTodo gqlResponse[struct {
		CreateTodo struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			Priority   string `json:"priority"`
			Status     string `json:"status"`
			NextAction bool   `json:"nextAction"`
		} `json:"createTodo"`
	}]
	s.GQL(t, `
		mutation CreateTodo($input: CreateTodoInput!) {
			createTodo(input: $input) { id title priority status nextAction }
		}
	`, map[string]any{
		"input": map[string]any{
			"projectId":  projectID,
			"stageId":    stageID,
			"title":      "GraphQL resolverek bekötése",
			"priority":   "HIGH",
			"dueDate":    "2026-06-15T00:00:00Z",
			"nextAction": true,
			"milestone":  true,
		},
	}, &createTodo)
	if len(createTodo.Errors) > 0 {
		t.Fatalf("createTodo errors: %+v", createTodo.Errors)
	}
	todoID := createTodo.Data.CreateTodo.ID
	if !createTodo.Data.CreateTodo.NextAction {
		t.Fatal("todo should be next action")
	}

	var addLabel gqlResponse[struct {
		AddTodoLabel struct {
			ID     string `json:"id"`
			Labels []struct {
				Name string `json:"name"`
			} `json:"labels"`
		} `json:"addTodoLabel"`
	}]
	s.GQL(t, `
		mutation AddTodoLabel($todoId: ID!, $labelId: ID!) {
			addTodoLabel(todoId: $todoId, labelId: $labelId) {
				id
				labels { name }
			}
		}
	`, map[string]any{
		"todoId":  todoID,
		"labelId": labelID,
	}, &addLabel)
	if len(addLabel.Errors) > 0 {
		t.Fatalf("addTodoLabel errors: %+v", addLabel.Errors)
	}
	if len(addLabel.Data.AddTodoLabel.Labels) != 1 || addLabel.Data.AddTodoLabel.Labels[0].Name != "backend" {
		t.Fatalf("labels = %+v", addLabel.Data.AddTodoLabel.Labels)
	}

	var markDone gqlResponse[struct {
		MarkTodoDone struct {
			ID          string  `json:"id"`
			Status      string  `json:"status"`
			CompletedAt *string `json:"completedAt"`
		} `json:"markTodoDone"`
	}]
	s.GQL(t, `
		mutation MarkDone($id: ID!) {
			markTodoDone(id: $id) { id status completedAt }
		}
	`, map[string]any{"id": todoID}, &markDone)
	if len(markDone.Errors) > 0 {
		t.Fatalf("markTodoDone errors: %+v", markDone.Errors)
	}
	if markDone.Data.MarkTodoDone.Status != "DONE" || markDone.Data.MarkTodoDone.CompletedAt == nil {
		t.Fatalf("done todo = %+v", markDone.Data.MarkTodoDone)
	}

	var query gqlResponse[struct {
		ProjectProgress struct {
			Total   int `json:"total"`
			Done    int `json:"done"`
			Percent int `json:"percent"`
		} `json:"projectProgress"`
		Dashboard struct {
			NextActions int `json:"nextActions"`
		} `json:"dashboard"`
		DoneTodos []struct {
			ID string `json:"id"`
		} `json:"doneTodos"`
	}]
	s.GQL(t, `
		query Check($projectId: ID!) {
			projectProgress(projectId: $projectId) { total done percent }
			dashboard { nextActions }
			doneTodos { id }
		}
	`, map[string]any{"projectId": projectID}, &query)
	if len(query.Errors) > 0 {
		t.Fatalf("query errors: %+v", query.Errors)
	}
	if query.Data.ProjectProgress.Total != 1 || query.Data.ProjectProgress.Done != 1 || query.Data.ProjectProgress.Percent != 100 {
		t.Fatalf("project progress = %+v", query.Data.ProjectProgress)
	}
	if query.Data.Dashboard.NextActions != 1 {
		t.Fatalf("dashboard nextActions = %d", query.Data.Dashboard.NextActions)
	}
	if len(query.Data.DoneTodos) != 1 || query.Data.DoneTodos[0].ID != todoID {
		t.Fatalf("doneTodos = %+v", query.Data.DoneTodos)
	}
}

func createProjectForAITest(t *testing.T, s *Suite) string {
	t.Helper()
	var createEpic gqlResponse[struct {
		CreateEpic struct {
			ID string `json:"id"`
		} `json:"createEpic"`
	}]
	s.GQL(t, `
		mutation CreateEpic($input: CreateEpicInput!) {
			createEpic(input: $input) { id }
		}
	`, map[string]any{"input": map[string]any{"name": "AI Epic"}}, &createEpic)
	if len(createEpic.Errors) > 0 {
		t.Fatalf("createEpic errors: %+v", createEpic.Errors)
	}

	var createProject gqlResponse[struct {
		CreateProject struct {
			ID string `json:"id"`
		} `json:"createProject"`
	}]
	s.GQL(t, `
		mutation CreateProject($input: CreateProjectInput!) {
			createProject(input: $input) { id }
		}
	`, map[string]any{
		"input": map[string]any{
			"epicId": createEpic.Data.CreateEpic.ID,
			"name":   "AI Project",
		},
	}, &createProject)
	if len(createProject.Errors) > 0 {
		t.Fatalf("createProject errors: %+v", createProject.Errors)
	}
	return createProject.Data.CreateProject.ID
}
