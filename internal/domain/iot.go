package domain

import "time"

type IoTDevice struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	CampusID        string                 `json:"campus_id"`
	DeviceType      string                 `json:"device_type"`
	Location        map[string]interface{} `json:"location"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	Status          string                 `json:"status"`
	SecurityState   string                 `json:"security_state"`
	FirmwareVersion string                 `json:"firmware_version"`
	LastSeenAt      time.Time              `json:"last_seen_at"`
}

type IoTCommand struct {
	ID              string                 `json:"id"`
	DeviceID        string                 `json:"device_id"`
	RequestedBy     string                 `json:"requested_by"`
	ApprovedBy      string                 `json:"approved_by"`
	CommandPayload  map[string]interface{} `json:"command_payload"`
	SafetyClearance bool                   `json:"safety_clearance"`
	Status          string                 `json:"status"`
}

type ResearchPublication struct {
	ID            string        `json:"id"`
	TenantID      string        `json:"tenant_id"`
	Title         string        `json:"title"`
	Abstract      string        `json:"abstract"`
	Authors       []interface{} `json:"authors"`
	DOI           string        `json:"doi"`
	PublishedDate time.Time     `json:"published_date"`
}
