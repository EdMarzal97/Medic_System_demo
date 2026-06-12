package http

import (
	"encoding/json"
	"net/http"
	"vitract-assessment/internal/patient"
	"vitract-assessment/internal/sample"
)

type Handler struct {
	patientRepo *patient.Repository
	sampleRepo  *sample.Repository
	jobQueue    chan sample.Job
}

func NewHandler(pRepo *patient.Repository, sRepo *sample.Repository, queue chan sample.Job) *Handler {
	return &Handler{
		patientRepo: pRepo,
		sampleRepo:  sRepo,
		jobQueue:    queue,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /patients", h.handleListPatients)
	mux.HandleFunc("GET /samples", h.handleListSamples)
	mux.HandleFunc("GET /samples/{id}/result", h.handleGetSampleResult)
	mux.HandleFunc("POST /samples/ingest-mock", h.handleMockIngest)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	h.respondWithJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

func (h *Handler) handleListPatients(w http.ResponseWriter, r *http.Request) {
	patients, err := h.patientRepo.GetAll(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondWithJSON(w, http.StatusOK, patients)
}

func (h *Handler) handleListSamples(w http.ResponseWriter, r *http.Request) {
	samples, err := h.sampleRepo.GetAllSummaries(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondWithJSON(w, http.StatusOK, samples)
}

func (h *Handler) handleGetSampleResult(w http.ResponseWriter, r *http.Request) {
	sampleID := r.PathValue("id")
	if sampleID == "" {
		h.respondWithError(w, http.StatusBadRequest, "missing sample id")
		return
	}

	result, err := h.sampleRepo.GetFullResult(r.Context(), sampleID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "sample result not found or not processed yet")
		return
	}
	h.respondWithJSON(w, http.StatusOK, result)
}

func (h *Handler) handleMockIngest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SampleID string `json:"sample_id"`
		FilePath string `json:"file_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.SampleID == "" || body.FilePath == "" {
		h.respondWithError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	h.jobQueue <- sample.Job{
		SampleID: body.SampleID,
		FilePath: body.FilePath,
	}

	h.respondWithJSON(w, http.StatusAccepted, map[string]string{
		"message":   "sample ingestion job accepted and queued asynchronously",
		"sample_id": body.SampleID,
		"status":    "pending",
	})
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) respondWithError(w http.ResponseWriter, status int, message string) {
	h.respondWithJSON(w, status, map[string]string{"error": message})
}