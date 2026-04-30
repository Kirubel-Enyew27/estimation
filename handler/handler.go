package handler

import (
	"estimation/service"
	"estimation/store"
)

type Handler struct {
	Service service.EstimationService
	Store store.MaterialStore
}

func New(svc service.EstimationService, st store.MaterialStore) *Handler {
	return &Handler{Service: svc, Store: st}
}
