-- +goose Up
ALTER TABLE stages ALTER COLUMN project_id DROP NOT NULL;

-- +goose Down
ALTER TABLE stages ALTER COLUMN project_id SET NOT NULL;
