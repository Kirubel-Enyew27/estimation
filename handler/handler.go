package handler

import (
	"encoding/json"
	"errors"
	"estimation/domain"
	"estimation/service"
	"estimation/store"
	"fmt"
	"io"
	"net/http"
)

type Handler struct {
	Service service.EstimationService
	Store   store.MaterialStore
}

func New(svc service.EstimationService, st store.MaterialStore) *Handler {
	return &Handler{Service: svc, Store: st}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /calculate/project", h.CalculateProject)
	mux.HandleFunc("GET /materials", h.ListMaterials)
	return mux
}

func (h *Handler) CalculateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusBadRequest, "POST method required")
		return
	}
	if h.Service == nil {
		writeError(w, http.StatusInternalServerError, "estimation service is not configured")
		return
	}

	var req calculateProjectRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.Service.Estimate(r.Context(), req.CalcualtionRequest)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to calculate projet estimate")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusBadRequest, "GET method required")
		return
	}
    if h.Store == nil {
		writeError(w, http.StatusInternalServerError, "material store is not configured")
		return
	}

	materials, err := h.Store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list materials")
		return
	}
	writeJSON(w, http.StatusOK, materials)
}

type calculateProjectRequest struct {
	domain.CalcualtionRequest
}

func (r calculateProjectRequest) Validate() error {
	return r.CalcualtionRequest.Validate()
}

type errorResponse struct {
	Error string `json:"error"`
}

func readJSON(r *http.Request, destination any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body is required")
		}
		return fmt.Errorf("invalid JSON request body: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("invalid JSON request body: multiple json values are not allowed")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}