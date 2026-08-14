package domain

// DomainKey identifies a domain for command routing.
type DomainKey string

const (
	DomainControlCabinet DomainKey = "control_cabinet"
	DomainSPSController  DomainKey = "sps_controller"
	DomainFieldDevice    DomainKey = "field_device"
	DomainBACnetObject   DomainKey = "bacnet_object"
)
