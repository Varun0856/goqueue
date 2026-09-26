package job

import "testing"

func TestTransition(t *testing.T) {
	tests := []struct {
		name    string
		current JobStatus
		next    JobStatus
		wantErr bool
	}{
		{"queued to running", Queued, Running, false},
		{"queued to completed", Queued, Completed, true},
		{"queued to failed", Queued, Failed, true},
		{"running to completed", Running, Completed, false},
		{"running to failed", Running, Failed, false},
		{"running to queued", Running, Queued, true},
		{"completed to failed", Completed, Failed, true},
		{"completed to queued", Completed, Queued, true},
		{"completed to running", Completed, Running, true},
		{"failed to completed", Failed, Completed, true},
		{"failed to queued", Failed, Queued, true},
		{"failed to running", Failed, Running, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{
				Status: tt.current,
			}
			err := job.Transition(tt.next)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			wantStatus := tt.next
			if tt.wantErr {
				wantStatus = tt.current
			}
			if job.Status != wantStatus {
				t.Errorf("unexpected status: got %v, want %v", job.Status, wantStatus)
			}
		})
	}
}

func TestNewJobCreation(t *testing.T) {
	tests := []struct {
		name    string
		jobType string
		payload string
		wantErr bool
	}{
		{"valid uppercase job", "uppercase", "hello", false},
		{"valid reverse job", "reverse", "golang", false},
		{"empty type", "", "hello", true},
		{"empty payload", "uppercase", "", true},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job, err := New(tt.jobType, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if job == nil {
				t.Fatalf("expected a job, got nil")
			}

			if job.Status != Queued {
				t.Errorf("expected job status to be 'queued' but got %q", job.Status)
			}

			if job.Type != tt.jobType {
				t.Errorf("expected type %q, got %q", tt.jobType, job.Type)
			}

			if job.Payload != tt.payload {
				t.Errorf("expected payload %q, got %q", tt.payload, job.Payload)
			}

			if job.ID == "" {
				t.Errorf("expected a non empty job id")
			}

			if job.Result != "" {
				t.Errorf("expected job result to be nil at creation but got %q", job.Result)
			}

			if job.Error != "" {
				t.Errorf("expected job error to be empty but got %q", job.Error)
			}

		})
	}
}

func TestNewJobUniqueID(t *testing.T) {
	job1, err1 := New("uppercase", "hello")
	job2, err2 := New("reverse", "golang")

	if err1 != nil {
		t.Fatalf("unexpected error: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}

	if job1.ID == "" || job2.ID == "" {
		t.Fatalf("job1 or job2 id is empty")
	}

	if job1.ID == job2.ID {
		t.Fatalf("expected unique IDs, but got identical IDs")
	}
}
