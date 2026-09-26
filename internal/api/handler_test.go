package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Varun0856/goqueue/internal/job"
	"github.com/Varun0856/goqueue/internal/store"
	"github.com/Varun0856/goqueue/internal/worker"
)

func TestCreateJob(t *testing.T) {
	s := store.NewStore()
	pool := worker.NewPool(2, 10, s)
	pool.Start()
	defer pool.Stop()

	handler := NewHandler(pool, s)

	req := httptest.NewRequest(
		http.MethodPost,
		"/jobs",
		strings.NewReader(`{"type":"uppercase", "payload":"hello"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.CreateJob(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, rec.Code)
	}

	var response struct {
		ID     string        `json:"id"`
		Status job.JobStatus `json:"status"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response.ID == "" {
		t.Fatal("expected a job ID")
	}

	if response.Status != job.Queued {
		t.Errorf("expected queued, got %q", response.Status)
	}
}

func TestGetJobNotFound(t *testing.T) {
	s := store.NewStore()
	pool := worker.NewPool(1, 10, s)
	pool.Start()
	defer pool.Stop()

	handler := NewHandler(pool, s)

	req := httptest.NewRequest(http.MethodGet, "/jobs/unknown-id", nil)
	req.SetPathValue("id", "unknown-id")

	rec := httptest.NewRecorder()
	handler.GetJob(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetJob(t *testing.T) {
	s := store.NewStore()

	pool := worker.NewPool(4, 100, s)
	pool.Start()
	defer pool.Stop()

	j, err := job.New("uppercase", "hello")
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}
	s.Create(*j)

	handler := NewHandler(pool, s)

	req := httptest.NewRequest(http.MethodGet, "/jobs/"+j.ID, nil)
	req.SetPathValue("id", j.ID)

	rec := httptest.NewRecorder()
	handler.GetJob(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got job.Job

	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.ID != j.ID {
		t.Errorf("expected job ID %q, got %q", j.ID, got.ID)
	}

	if got.Status != job.Queued {
		t.Errorf("expected status %q, got %q", job.Queued, got.Status)
	}
}
