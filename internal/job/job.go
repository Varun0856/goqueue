package job

import (
	"fmt"
	"github.com/google/uuid"
)

type JobStatus string

const (
	Queued    JobStatus = "queued"
	Running   JobStatus = "running"
	Completed JobStatus = "completed"
	Failed    JobStatus = "failed"
)

type Job struct {
	ID      string
	Type    string
	Payload string
	Status  JobStatus
	Result  string
	Error   string
}

func (j *Job) Transition(next JobStatus) error {
	switch j.Status {
	case Queued:
		if next != Running {
			return fmt.Errorf("status 'queued' can only transition to 'running' (attempted: '%s')", next)
		}

	case Running:
		if next != Completed && next != Failed {
			return fmt.Errorf("status 'running' can only transition to 'completed' or 'failed' (attempted: '%s')", next)
		}

	case Completed, Failed:
		return fmt.Errorf("cannot transition from final status '%s' to '%s'", j.Status, next)

	default:
		return fmt.Errorf("unknown current status '%s'", j.Status)
	}

	j.Status = next
	return nil
}

func New(jobType string, payload string) (*Job, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	if jobType == "" || payload == "" {
		return nil, fmt.Errorf("job type and payload cannot be empty")
	}

	return &Job{
		ID:      id.String(),
		Type:    jobType,
		Payload: payload,
		Status:  Queued,
	}, nil
}
