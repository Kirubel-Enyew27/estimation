package handler

import (
	"estimation/service"
	"estimation/store"
)

// Handler wires HTTP endpoints to the service layer. Keep this as a thin
// adapter; route registration and actual HTTP logic will be implemented
// in later phases.
type Handler struct {
    Service service.EstimationService
    Store   store.MaterialStore
}

// New creates a new Handler with required dependencies.
func New(svc service.EstimationService, st store.MaterialStore) *Handler {
    return &Handler{Service: svc, Store: st}
}
