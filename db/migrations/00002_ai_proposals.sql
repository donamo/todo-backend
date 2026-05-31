-- +goose Up
CREATE TABLE IF NOT EXISTS ai_proposals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_type TEXT NOT NULL,
    parent_id UUID NOT NULL,
    magic_text TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    proposal_json JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'DRAFT',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    applied_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS ai_proposals_user_id_idx ON ai_proposals(user_id);
CREATE INDEX IF NOT EXISTS ai_proposals_parent_idx ON ai_proposals(user_id, parent_type, parent_id);

-- +goose Down
DROP TABLE ai_proposals;
