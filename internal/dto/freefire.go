package dto

// FreeFire Response DTO
type FreeFireResponse struct {
	Success  bool              `json:"success"`
	Nickname string            `json:"nickname"`
	PlayerID string            `json:"playerId"`
	UID      string            `json:"uid"`
	Level    int               `json:"level,omitempty"`
	Region   string            `json:"region,omitempty"`
	AvatarID int               `json:"avatarId,omitempty"`
	AvatarURL string           `json:"avatarUrl,omitempty"`
	Source   string            `json:"source"` // qual API foi usada
}

// Garena Fast API Response
type GarenaFastAPIResponse struct {
	Success bool                  `json:"success"`
	Source  string                `json:"source"`
	Data    GarenaFastAPIData     `json:"data"`
}

type GarenaFastAPIData struct {
	Nickname string `json:"nickname"`
	PlayerID string `json:"playerId"`
	Level    int    `json:"level"`
	AvatarID int    `json:"avatarId"`
	AvatarURL string `json:"avatarUrl"`
}

// Region API Response (XZA e Get-Region)
type RegionAPIResponse struct {
	Nickname string `json:"nickname"`
	Region   string `json:"region"`
	UID      string `json:"uid"`
}
