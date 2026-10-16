package api

import (
	"encoding/json"
	"net/http"

	"github.com/Varun0856/goqueue/internal/job"
	"github.com/Varun0856/goqueue/internal/store"
	"github.com/Varun0856/goqueue/internal/worker"
)

type Handler struct {
	pool  *worker.Pool
	store *store.Store
}

func NewHandler(pool *worker.Pool, s *store.Store) *Handler {
	return &Handler{
		pool:  pool,
		store: s,
	}
}

type createJobRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type createJobResponse struct {
	ID     string        `json:"ID"`
	Status job.JobStatus `json:"status"`
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	j, err := job.New(req.Type, req.Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.pool.Enqueue(j); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(createJobResponse{
		ID:     j.ID,
		Status: j.Status,
	})
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	j, err := h.store.Get(id)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(j)
}

func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs := h.store.List()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		return
	}
}
