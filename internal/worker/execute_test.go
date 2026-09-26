package worker

import (
	"testing"
)

func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		jobType string
		payload string
		result  string
		wantErr bool
	}{
		{"uppercase", "uppercase", "hello", "HELLO", false},
		{"reverse", "reverse", "hello", "olleh", false},
		{"invalid job type", "lowercase", "hello", "", true},
		{"empty inputs", "", "", "", true},
		{"empty jobtype", "", "hello", "", true},
		{"empty payload", "uppercase", "", "", false},
		{"test unicode reverse", "reverse", "héllo", "olléh", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Execute(tt.jobType, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected err, but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.result {
				t.Errorf("expected %q, but got %q", tt.result, result)
			}
		})
	}
}
