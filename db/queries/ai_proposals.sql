-- name: CreateAIProposal :one
INSERT INTO ai_proposals (user_id, parent_type, parent_id, magic_text, summary, proposal_json)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, parent_type, parent_id, magic_text, summary, proposal_json, status, created_at, applied_at;

-- name: GetAIProposal :one
SELECT id, user_id, parent_type, parent_id, magic_text, summary, proposal_json, status, created_at, applied_at
FROM ai_proposals
WHERE id = $1 AND user_id = $2;

-- name: ListAIProposals :many
SELECT id, user_id, parent_type, parent_id, magic_text, summary, proposal_json, status, created_at, applied_at
FROM ai_proposals
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('parent_type')::text IS NULL OR parent_type = sqlc.narg('parent_type')::text)
  AND (sqlc.narg('parent_id')::uuid IS NULL OR parent_id = sqlc.narg('parent_id')::uuid)
ORDER BY created_at DESC;

-- name: MarkAIProposalApplied :one
UPDATE ai_proposals
SET status = 'APPLIED',
    applied_at = now()
WHERE id = $1 AND user_id = $2 AND status = 'DRAFT'
RETURNING id, user_id, parent_type, parent_id, magic_text, summary, proposal_json, status, created_at, applied_at;
