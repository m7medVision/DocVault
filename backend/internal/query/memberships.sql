-- name: CreateMembership :exec
INSERT INTO memberships (id, user_id, org_id, role, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: ListMembershipsByTenant :many
SELECT m.id AS membership_id, m.user_id, m.org_id, u.email, u.display_name, m.role, m.created_at
FROM memberships m
JOIN users u ON u.id = m.user_id
JOIN organizations o ON o.id = m.org_id
WHERE o.tenant_id = $1
ORDER BY m.created_at ASC;

-- name: ListMembershipsByOrg :many
SELECT m.id AS membership_id, m.user_id, m.org_id, u.email, u.display_name, m.role, m.created_at
FROM memberships m
JOIN users u ON u.id = m.user_id
JOIN organizations o ON o.id = m.org_id
WHERE o.tenant_id = $1 AND m.org_id = $2
ORDER BY m.created_at ASC;

-- name: GetMembershipByID :one
SELECT m.id AS membership_id, m.user_id, m.org_id, u.email, u.display_name, m.role, m.created_at
FROM memberships m
JOIN users u ON u.id = m.user_id
JOIN organizations o ON o.id = m.org_id
WHERE o.tenant_id = $1 AND m.id = $2;

-- name: UpdateMembershipRole :execrows
UPDATE memberships
SET role = $1
WHERE id = $2;

-- name: GetPrimaryMembershipByUserID :one
SELECT org_id, role
FROM memberships
WHERE user_id = $1
ORDER BY created_at ASC
LIMIT 1;
