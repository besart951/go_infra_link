package facility

import (
	"context"
	"errors"
	"testing"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type deleteImpactRepositoryFake struct {
	impacts []domainFacility.DeleteImpact
	err     error
	ids     []uuid.UUID
}

func (f *deleteImpactRepositoryFake) DeleteImpacts(_ context.Context, _ domainFacility.DeleteImpactResource, ids []uuid.UUID) ([]domainFacility.DeleteImpact, error) {
	f.ids = append([]uuid.UUID(nil), ids...)
	return f.impacts, f.err
}

func TestDeleteImpactServiceEnsureDeleteAllowed(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name    string
		impacts []domainFacility.DeleteImpact
		err     error
		wantErr error
	}{
		{name: "allows unreferenced item", impacts: []domainFacility.DeleteImpact{{ID: id}}},
		{name: "blocks referenced item", impacts: []domainFacility.DeleteImpact{{ID: id, Blockers: []domainFacility.DeleteImpactBlocker{{Resource: "field_devices", Count: 1}}}}, wantErr: domainFacility.ErrReferenceInUse},
		{name: "propagates repository error", err: errors.New("database unavailable")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &deleteImpactRepositoryFake{impacts: tt.impacts, err: tt.err}
			err := NewDeleteImpactService(repo).EnsureDeleteAllowed(context.Background(), domainFacility.DeleteImpactResourceApparat, id)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && tt.err == nil && err != nil {
				t.Fatalf("error = %v, want nil", err)
			}
			if tt.err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("error = %v, want %v", err, tt.err)
			}
		})
	}
}

func TestDeleteImpactServiceListDeduplicatesIDs(t *testing.T) {
	id := uuid.New()
	repo := &deleteImpactRepositoryFake{}
	_, err := NewDeleteImpactService(repo).List(context.Background(), domainFacility.DeleteImpactResourceSystemPart, []uuid.UUID{id, id, uuid.Nil})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(repo.ids) != 1 || repo.ids[0] != id {
		t.Fatalf("repository IDs = %v, want [%s]", repo.ids, id)
	}
}
