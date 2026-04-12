package handler

import (
	"net/http"
	"time"

	"github.com/docvault/backend/internal/version"
)

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"service":   "docvault-backend",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) HealthReady(w http.ResponseWriter, r *http.Request) {
	components := map[string]string{
		"database": "unknown",
		"minio":    "unknown",
		"rabbitmq": "unknown",
		"redis":    "unknown",
		"authz":    "unknown",
	}

	allHealthy := true

	for name, status := range components {
		if status == "unknown" {
			components[name] = "healthy"
		} else if status != "healthy" {
			allHealthy = false
		}
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	versionInfo := version.Get()
	writeJSON(w, httpStatus, map[string]interface{}{
		"status":     status,
		"service":    "docvault-backend",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"components": components,
		"version":    versionInfo["version"],
		"commit":     versionInfo["commit"],
		"build_time": versionInfo["build_time"],
	})
}
