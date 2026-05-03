package facility

import (
	"strings"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

func generatedSPSControllerDeviceName(
	controlCabinet *domainFacility.ControlCabinet,
	building *domainFacility.Building,
	gaDevice *string,
) (string, bool) {
	if gaDevice == nil {
		return "", false
	}

	ga := strings.ToUpper(strings.TrimSpace(*gaDevice))
	if ga == "" {
		return "", false
	}

	iwsCode := ""
	if building != nil {
		iwsCode = strings.TrimSpace(building.IWSCode)
	}

	cabinetNr := ""
	if controlCabinet != nil && controlCabinet.ControlCabinetNr != nil {
		cabinetNr = strings.TrimSpace(*controlCabinet.ControlCabinetNr)
	}

	if iwsCode == "" || cabinetNr == "" {
		return ga, true
	}

	return strings.ToUpper(iwsCode + "_" + cabinetNr + "_" + ga), true
}
