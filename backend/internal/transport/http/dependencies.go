package handler

import (
	"github.com/casbin/casbin/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (h *Handler) SetDB(db *pgxpool.Pool) {
	h.db = db
}

func (h *Handler) SetAuthorizationEnforcer(enforcer *casbin.Enforcer) {
	h.authzEnforcer = enforcer
}
