package dto

import (
	"time"

	"github.com/google/uuid"
)

// Facility DTOs - StateText

type CreateStateTextRequest struct {
	RefNumber   int     `json:"ref_number" binding:"required"`
	StateText1  *string `json:"state_text1"`
	StateText2  *string `json:"state_text2"`
	StateText3  *string `json:"state_text3"`
	StateText4  *string `json:"state_text4"`
	StateText5  *string `json:"state_text5"`
	StateText6  *string `json:"state_text6"`
	StateText7  *string `json:"state_text7"`
	StateText8  *string `json:"state_text8"`
	StateText9  *string `json:"state_text9"`
	StateText10 *string `json:"state_text10"`
	StateText11 *string `json:"state_text11"`
	StateText12 *string `json:"state_text12"`
	StateText13 *string `json:"state_text13"`
	StateText14 *string `json:"state_text14"`
	StateText15 *string `json:"state_text15"`
	StateText16 *string `json:"state_text16"`
}

type UpdateStateTextRequest struct {
	RefNumber   *int    `json:"ref_number"`
	StateText1  *string `json:"state_text1"`
	StateText2  *string `json:"state_text2"`
	StateText3  *string `json:"state_text3"`
	StateText4  *string `json:"state_text4"`
	StateText5  *string `json:"state_text5"`
	StateText6  *string `json:"state_text6"`
	StateText7  *string `json:"state_text7"`
	StateText8  *string `json:"state_text8"`
	StateText9  *string `json:"state_text9"`
	StateText10 *string `json:"state_text10"`
	StateText11 *string `json:"state_text11"`
	StateText12 *string `json:"state_text12"`
	StateText13 *string `json:"state_text13"`
	StateText14 *string `json:"state_text14"`
	StateText15 *string `json:"state_text15"`
	StateText16 *string `json:"state_text16"`
}

type StateTextResponse struct {
	ID         uuid.UUID `json:"id"`
	RefNumber  int       `json:"ref_number"`
	StateText1 *string   `json:"state_text1"`
	// Include only first few for lightness or all? User just wants search.
	// But detailed response usually contains all.
	StateText2  *string   `json:"state_text2"`
	StateText3  *string   `json:"state_text3"`
	StateText4  *string   `json:"state_text4"`
	StateText5  *string   `json:"state_text5"`
	StateText6  *string   `json:"state_text6"`
	StateText7  *string   `json:"state_text7"`
	StateText8  *string   `json:"state_text8"`
	StateText9  *string   `json:"state_text9"`
	StateText10 *string   `json:"state_text10"`
	StateText11 *string   `json:"state_text11"`
	StateText12 *string   `json:"state_text12"`
	StateText13 *string   `json:"state_text13"`
	StateText14 *string   `json:"state_text14"`
	StateText15 *string   `json:"state_text15"`
	StateText16 *string   `json:"state_text16"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StateTextListResponse struct {
	Items      []StateTextResponse `json:"items"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	TotalPages int                 `json:"total_pages"`
}

// Facility DTOs - NotificationClass

type CreateNotificationClassRequest struct {
	EventCategory        string `json:"event_category" binding:"required"`
	Nc                   int    `json:"nc" binding:"required"`
	ObjectDescription    string `json:"object_description" binding:"required"`
	InternalDescription  string `json:"internal_description" binding:"required"`
	Meaning              string `json:"meaning" binding:"required"`
	AckRequiredNotNormal bool   `json:"ack_required_not_normal"`
	AckRequiredError     bool   `json:"ack_required_error"`
	AckRequiredNormal    bool   `json:"ack_required_normal"`
	NormNotNormal        int    `json:"norm_not_normal"`
	NormError            int    `json:"norm_error"`
	NormNormal           int    `json:"norm_normal"`
}

type UpdateNotificationClassRequest struct {
	EventCategory        *string `json:"event_category"`
	Nc                   *int    `json:"nc"`
	ObjectDescription    *string `json:"object_description"`
	InternalDescription  *string `json:"internal_description"`
	Meaning              *string `json:"meaning"`
	AckRequiredNotNormal *bool   `json:"ack_required_not_normal"`
	AckRequiredError     *bool   `json:"ack_required_error"`
	AckRequiredNormal    *bool   `json:"ack_required_normal"`
	NormNotNormal        *int    `json:"norm_not_normal"`
	NormError            *int    `json:"norm_error"`
	NormNormal           *int    `json:"norm_normal"`
}

type NotificationClassResponse struct {
	ID                   uuid.UUID `json:"id"`
	EventCategory        string    `json:"event_category"`
	Nc                   int       `json:"nc"`
	ObjectDescription    string    `json:"object_description"`
	InternalDescription  string    `json:"internal_description"`
	Meaning              string    `json:"meaning"`
	AckRequiredNotNormal bool      `json:"ack_required_not_normal"`
	AckRequiredError     bool      `json:"ack_required_error"`
	AckRequiredNormal    bool      `json:"ack_required_normal"`
	NormNotNormal        int       `json:"norm_not_normal"`
	NormError            int       `json:"norm_error"`
	NormNormal           int       `json:"norm_normal"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type NotificationClassListResponse struct {
	Items      []NotificationClassResponse `json:"items"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	TotalPages int                         `json:"total_pages"`
}

// Facility DTOs - AlarmDefinition

type CreateAlarmDefinitionRequest struct {
	Name        string     `json:"name" binding:"required"`
	AlarmNote   *string    `json:"alarm_note"`
	AlarmTypeID *uuid.UUID `json:"alarm_type_id"`
}

type UpdateAlarmDefinitionRequest struct {
	Name        *string    `json:"name"`
	AlarmNote   *string    `json:"alarm_note"`
	AlarmTypeID *uuid.UUID `json:"alarm_type_id"`
}

type AlarmDefinitionResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	AlarmNote   *string    `json:"alarm_note"`
	AlarmTypeID *uuid.UUID `json:"alarm_type_id,omitempty"`
	IsActive    bool       `json:"is_active"`
	Scope       string     `json:"scope"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type AlarmDefinitionListResponse struct {
	Items      []AlarmDefinitionResponse `json:"items"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	TotalPages int                       `json:"total_pages"`
}

// Facility DTOs - SPSControllerSystemType

// This represents the join of SPS Controller and System Type
type SPSControllerSystemTypeResponse struct {
	ID              uuid.UUID `json:"id"`
	SPSControllerID uuid.UUID `json:"sps_controller_id"`
	SystemTypeID    uuid.UUID `json:"system_type_id"`

	// Pre-filled names for display in combobox
	SPSControllerName string `json:"sps_controller_name"`
	SystemTypeName    string `json:"system_type_name"`

	Number       *int      `json:"number"`
	DocumentName *string   `json:"document_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SPSControllerSystemTypeListResponse struct {
	Items      []SPSControllerSystemTypeResponse `json:"items"`
	Total      int64                             `json:"total"`
	Page       int                               `json:"page"`
	TotalPages int                               `json:"total_pages"`
}
