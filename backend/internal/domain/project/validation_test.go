package project

import (
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

func TestProjectValidate(t *testing.T) {
	tests := []struct {
		name       string
		project    Project
		wantFields map[string]string
	}{
		{
			name: "valid",
			project: Project{
				Name:      "New plant",
				Status:    StatusPlanned,
				PhaseID:   uuid.New(),
				CreatorID: uuid.New(),
			},
		},
		{
			name:    "reports every invalid input",
			project: Project{Status: ProjectStatus("unknown")},
			wantFields: map[string]string{
				"project.name":       "required",
				"project.status":     "oneof",
				"project.phase_id":   "required",
				"project.creator_id": "required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if len(tt.wantFields) == 0 {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}

			validationErr, ok := domain.AsValidationError(err)
			if !ok {
				t.Fatalf("Validate() error = %T, want *domain.ValidationError", err)
			}
			for path, code := range tt.wantFields {
				if got := validationErr.Codes[path]; got != code {
					t.Errorf("code for %q = %q, want %q", path, got, code)
				}
			}
		})
	}
}
