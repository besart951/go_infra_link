package facility

import (
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
)

func TestAggregateValidationReportsStableFieldCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want map[string]string
	}{
		{
			name: "control cabinet",
			err:  (ControlCabinet{}).Validate(),
			want: map[string]string{
				"controlcabinet.building_id":        "required",
				"controlcabinet.control_cabinet_nr": "required",
			},
		},
		{
			name: "SPS controller",
			err:  (SPSController{}).Validate(),
			want: map[string]string{
				"spscontroller.control_cabinet_id": "required",
				"spscontroller.device_name":        "required",
				"spscontroller.ga_device":          "required",
			},
		},
		{
			name: "SPS controller system type",
			err:  (SPSControllerSystemType{}).Validate("spscontroller.system_types[0]"),
			want: map[string]string{
				"spscontroller.system_types[0].sps_controller_id": "required",
				"spscontroller.system_types[0].system_type_id":    "required",
			},
		},
		{
			name: "field device",
			err:  (FieldDevice{}).Validate(),
			want: map[string]string{
				"fielddevice.sps_controller_system_type_id": "required",
				"fielddevice.system_part_id":                "required",
				"fielddevice.apparat_id":                    "required",
				"fielddevice.apparat_nr":                    "range",
			},
		},
		{
			name: "BACnet object",
			err: (BacnetObject{
				SoftwareType: BacnetSoftwareType("invalid"),
				HardwareType: BacnetHardwareType("invalid"),
			}).Validate("bacnet_object"),
			want: map[string]string{
				"bacnet_object.text_fix":      "required",
				"bacnet_object.software_type": "oneof",
				"bacnet_object.hardware_type": "oneof",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErr, ok := domain.AsValidationError(tt.err)
			if !ok {
				t.Fatalf("error = %T, want *domain.ValidationError", tt.err)
			}
			for path, code := range tt.want {
				if got := validationErr.Codes[path]; got != code {
					t.Errorf("code for %q = %q, want %q", path, got, code)
				}
			}
		})
	}
}

func TestBacnetTypesValidateKnownValues(t *testing.T) {
	if !BacnetSoftwareTypeAI.Valid() {
		t.Fatal("AI software type should be valid")
	}
	if !BacnetHardwareTypeEMPTY.Valid() {
		t.Fatal("empty hardware type should be valid")
	}
}
