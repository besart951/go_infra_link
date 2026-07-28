package fielddevice

import (
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestNormalizeMultiCreateResultAddsStablePerItemFailureDetails(t *testing.T) {
	requestedID := uuid.MustParse("00000000-0000-0000-0000-000000000901")
	result := &domainFacility.FieldDeviceMultiCreateResult{
		Results: []domainFacility.FieldDeviceCreateResult{{
			Index:      0,
			Error:      "ApparatNr is already assigned",
			ErrorField: "fielddevice.apparat_nr",
		}},
		TotalRequests: 1,
		FailureCount:  1,
	}

	normalizeMultiCreateResult(result, []domainFacility.FieldDeviceCreateItem{{
		FieldDevice: &domainFacility.FieldDevice{Base: domain.Base{ID: requestedID}},
	}})

	item := result.Results[0]
	if item.ID != requestedID ||
		item.ErrorCode != itemErrorCodeConflict ||
		item.ErrorField != "fielddevice.apparat_nr" ||
		item.Reason != "ApparatNr is already assigned" {
		t.Fatalf("normalized create result: %+v", item)
	}
}

func TestNormalizeBulkResultUsesDeterministicPrimaryField(t *testing.T) {
	requestedID := uuid.MustParse("00000000-0000-0000-0000-000000000902")
	result := &domainFacility.BulkOperationResult{
		Results: []domainFacility.BulkOperationResultItem{{
			Fields: map[string]string{
				"specification.type":  "invalid type",
				"bacnet_objects.0.ga": "invalid GA",
			},
		}},
		TotalCount:   1,
		FailureCount: 1,
	}

	normalizeBulkResult(result, []uuid.UUID{requestedID})

	item := result.Results[0]
	if item.ID != requestedID ||
		item.ErrorCode != itemErrorCodeValidation ||
		item.ErrorField != "bacnet_objects.0.ga" ||
		item.Error != "invalid GA" ||
		item.Reason != "invalid GA" {
		t.Fatalf("normalized bulk result: %+v", item)
	}
}

func TestFailedBulkResultFromErrorPreservesCommitConstraintField(t *testing.T) {
	requestedID := uuid.MustParse("00000000-0000-0000-0000-000000000903")
	apparatNr := 2
	result := failedBulkResultFromError(
		[]domainFacility.BulkFieldDeviceUpdate{{
			ID:        requestedID,
			ApparatNr: &apparatNr,
		}},
		domain.NewValidationError().
			Add("fielddevice.apparat_nr", "apparatnummer ist bereits vergeben"),
	)

	item := result.Results[0]
	if item.ID != requestedID ||
		item.ErrorCode != itemErrorCodeConflict ||
		item.ErrorField != "fielddevice.apparat_nr" ||
		item.Reason != "apparatnummer ist bereits vergeben" ||
		item.Fields["fielddevice.apparat_nr"] != item.Reason {
		t.Fatalf("commit constraint result: %+v", item)
	}
}
