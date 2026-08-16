package facility

// DetailRelationItem is an intentionally label-first relationship. The ID is
// retained solely for navigation and API requests; clients must never render it
// as a fallback label.
type DetailRelationItem struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Subtitle string `json:"subtitle,omitempty"`
}

// DetailRelation is a paginated, permission-filtered part of a facility
// hierarchy detail. Relationships without read permission are omitted.
type DetailRelation struct {
	Key        string               `json:"key"`
	Label      string               `json:"label"`
	Resource   string               `json:"resource"`
	Count      int64                `json:"count"`
	Items      []DetailRelationItem `json:"items"`
	Page       int                  `json:"page"`
	TotalPages int                  `json:"total_pages"`
}

// DetailCapabilities describes local actions after all applicable permissions
// have been evaluated. Project details additionally require the matching
// project capability and global permission before CanUpdate becomes true.
type DetailCapabilities struct {
	CanUpdate bool `json:"can_update"`
}

type BuildingDetailResponse struct {
	Building     BuildingResponse   `json:"building"`
	Relations    []DetailRelation   `json:"relations"`
	Capabilities DetailCapabilities `json:"capabilities"`
}

type ControlCabinetDetailResponse struct {
	ControlCabinet ControlCabinetResponse `json:"control_cabinet"`
	Relations      []DetailRelation       `json:"relations"`
	Capabilities   DetailCapabilities     `json:"capabilities"`
}

type SPSControllerDetailResponse struct {
	SPSController SPSControllerResponse `json:"sps_controller"`
	Relations     []DetailRelation      `json:"relations"`
	Capabilities  DetailCapabilities    `json:"capabilities"`
}

type SPSControllerSystemTypeDetailResponse struct {
	SPSControllerSystemType SPSControllerSystemTypeResponse `json:"sps_controller_system_type"`
	Relations               []DetailRelation                `json:"relations"`
	Capabilities            DetailCapabilities              `json:"capabilities"`
}

type FieldDeviceDetailResponse struct {
	FieldDevice  FieldDeviceResponse `json:"field_device"`
	Relations    []DetailRelation    `json:"relations"`
	Capabilities DetailCapabilities  `json:"capabilities"`
}
