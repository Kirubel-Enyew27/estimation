package handler

import (
	"encoding/json"
	"errors"
	"estimation/domain"
	"estimation/service"
	"fmt"
	"io"
	"net/http"
)

type Handler struct {
	estimator service.EstimationService
	materials service.MaterialService
}

func New(estimator service.EstimationService, materials service.MaterialService) *Handler {
	return &Handler{
		estimator: estimator,
		materials: materials,
	}
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
	if h.estimator == nil {
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

	result, err := h.estimator.Estimate(r.Context(), req.CalculationRequest)
	if err != nil {
		if errors.Is(err, service.ErrMaterialNotFound) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to calculate project estimate")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusBadRequest, "GET method required")
		return
	}
	if h.materials == nil {
		writeError(w, http.StatusInternalServerError, "material service is not configured")
		return
	}

	materials, err := h.materials.ListMaterials(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list materials")
		return
	}
	writeJSON(w, http.StatusOK, materials)
}

type calculateProjectRequest struct {
	domain.CalculationRequest
}

func (r calculateProjectRequest) Validate() error {
	return r.CalculationRequest.Validate()
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
		return errors.New("invalid JSON request body: multiple JSON values are not allowed")
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
