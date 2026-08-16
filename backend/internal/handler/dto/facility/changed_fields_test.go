package facility

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestUpdateRequestChangedFieldsMatchTheHTTPContract(t *testing.T) {
	value := "value"
	fieldDeviceID := uuid.New()
	visible := true
	tests := []struct {
		name string
		got  []string
		want []string
	}{
		{
			name: "field device",
			got: UpdateFieldDeviceRequest{
				BMK:                       &value,
				TextIndividuell:           &value,
				ApparatID:                 uuid.New(),
				SPSControllerSystemTypeID: uuid.New(),
			}.ChangedFields(),
			want: []string{"bmk", "text_fix", "sps_controller_system_type_id", "apparat_id"},
		},
		{
			name: "controller",
			got: UpdateSPSControllerRequest{
				DeviceName:     "SPS-1",
				DeviceLocation: &value,
				SystemTypes:    &[]SPSControllerSystemTypeInput{},
			}.ChangedFields(),
			want: []string{"device_name", "device_location", "system_types"},
		},
		{
			name: "bacnet object",
			got: UpdateBacnetObjectRequest{
				FieldDeviceID: &fieldDeviceID,
				BacnetObjectPatchInput: BacnetObjectPatchInput{
					Description: &value,
					GMSVisible:  &visible,
				},
			}.ChangedFields(),
			want: []string{"description", "gms_visible", "field_device_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("ChangedFields() = %#v, want %#v", tt.got, tt.want)
			}
		})
	}
}

func TestBulkFieldDeviceChangedFieldsAreDeduplicatedInRequestOrder(t *testing.T) {
	req := BulkUpdateFieldDeviceRequest{Updates: []BulkUpdateFieldDeviceItem{
		{BMK: OptionalString{Set: true}, Specification: &SpecificationInput{}},
		{BMK: OptionalString{Set: true}, Description: OptionalString{Set: true}, BacnetObjects: &[]BacnetObjectBulkPatchInput{}},
	}}

	want := []string{"bmk", "specification", "description", "bacnet_objects"}
	if got := req.ChangedFields(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ChangedFields() = %#v, want %#v", got, want)
	}
}
