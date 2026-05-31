-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    google_subject TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT '',
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE epics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    color TEXT,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE epics ADD CONSTRAINT epics_id_user_id_unique UNIQUE (id, user_id);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    epic_id UUID REFERENCES epics(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    start_date DATE,
    target_date DATE,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE projects ADD CONSTRAINT projects_id_user_id_unique UNIQUE (id, user_id);
ALTER TABLE projects ADD CONSTRAINT projects_epic_user_fk FOREIGN KEY (epic_id, user_id) REFERENCES epics(id, user_id);

CREATE TABLE stages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'PLANNED',
    start_date DATE,
    target_date DATE,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE stages ADD CONSTRAINT stages_id_user_id_unique UNIQUE (id, user_id);
ALTER TABLE stages ADD CONSTRAINT stages_project_user_fk FOREIGN KEY (project_id, user_id) REFERENCES projects(id, user_id) ON DELETE CASCADE;

CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    stage_id UUID REFERENCES stages(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT,
    priority TEXT NOT NULL DEFAULT 'NORMAL',
    status TEXT NOT NULL DEFAULT 'OPEN',
    start_date DATE,
    due_date DATE,
    estimated_effort TEXT,
    position INTEGER NOT NULL DEFAULT 0,
    next_action BOOLEAN NOT NULL DEFAULT FALSE,
    milestone BOOLEAN NOT NULL DEFAULT FALSE,
    recurrence TEXT NOT NULL DEFAULT 'NONE',
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE todos ADD CONSTRAINT todos_id_user_id_unique UNIQUE (id, user_id);
ALTER TABLE todos ADD CONSTRAINT todos_project_user_fk FOREIGN KEY (project_id, user_id) REFERENCES projects(id, user_id) ON DELETE CASCADE;
ALTER TABLE todos ADD CONSTRAINT todos_stage_user_fk FOREIGN KEY (stage_id, user_id) REFERENCES stages(id, user_id);

CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    color TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE labels ADD CONSTRAINT labels_id_user_id_unique UNIQUE (id, user_id);

CREATE TABLE todo_labels (
    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (todo_id, label_id)
);

CREATE TABLE project_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE project_notes ADD CONSTRAINT project_notes_project_user_fk FOREIGN KEY (project_id, user_id) REFERENCES projects(id, user_id) ON DELETE CASCADE;

CREATE TABLE todo_dependencies (
    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    depends_on_todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (todo_id, depends_on_todo_id),
    CHECK (todo_id <> depends_on_todo_id)
);

CREATE UNIQUE INDEX labels_user_name_idx ON labels(user_id, name);

CREATE UNIQUE INDEX todos_one_next_action_per_project_idx
    ON todos(user_id, project_id)
    WHERE next_action = TRUE AND project_id IS NOT NULL;

CREATE INDEX epics_user_id_idx ON epics(user_id);
CREATE INDEX projects_user_id_idx ON projects(user_id);
CREATE INDEX projects_epic_id_idx ON projects(epic_id);
CREATE INDEX stages_user_id_idx ON stages(user_id);
CREATE INDEX stages_project_id_idx ON stages(project_id);
CREATE INDEX todos_user_id_idx ON todos(user_id);
CREATE INDEX todos_project_id_idx ON todos(project_id);
CREATE INDEX todos_stage_id_idx ON todos(stage_id);
CREATE INDEX todos_status_idx ON todos(status);
CREATE INDEX todos_due_date_idx ON todos(due_date);
CREATE INDEX todo_labels_label_id_idx ON todo_labels(label_id);
CREATE INDEX project_notes_user_id_idx ON project_notes(user_id);
CREATE INDEX project_notes_project_id_idx ON project_notes(project_id);
CREATE INDEX todo_dependencies_depends_on_idx ON todo_dependencies(depends_on_todo_id);

-- +goose Down
DROP TABLE todo_dependencies;
DROP TABLE project_notes;
DROP TABLE todo_labels;
DROP TABLE labels;
DROP TABLE todos;
DROP TABLE stages;
DROP TABLE projects;
DROP TABLE epics;
DROP TABLE users;
