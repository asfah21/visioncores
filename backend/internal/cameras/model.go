package cameras

import "time"

// Camera is the DB source of truth. rtsp_url in DB always holds the FULL url
// (with credentials). API responses never include it — see ToPublic.
type Camera struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Host        string    `json:"host"`
	RTSPPort    int       `json:"rtsp_port"`
	RTSPURL     string    `json:"rtsp_url,omitempty"`
	Username    string    `json:"username"`
	OnvifPort   int       `json:"onvif_port"`
	Enabled     bool      `json:"enabled"`
	AutoRecord  bool      `json:"auto_record"`
	// PTZProfile overrides ONVIF profile discovery (GetProfiles) when set.
	// For devices whose discovery is broken but whose PTZ works with a known token.
	PTZProfile  string    `json:"ptz_profile,omitempty"`
	// Brand selects a known PTZ preset (e.g. bardi -> IPCProfilesToken0).
	// Empty = Generic ONVIF (live discovery).
	Brand       string    `json:"brand,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PublicCamera is safe to send to the frontend (no secrets, no full RTSP).
type PublicCamera struct {
	Camera
	RTSPURL string `json:"rtsp_url,omitempty"`
}

func (c Camera) ToPublic() PublicCamera {
	p := PublicCamera{Camera: c}
	p.RTSPURL = ""
	cam := p.Camera
	cam.RTSPURL = ""
	p.Camera = cam
	return p
}

type UpsertInput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Host        string `json:"host"`
	RTSPPort    int    `json:"rtsp_port"`
	RTSPURL     string `json:"rtsp_url"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	OnvifPort   int    `json:"onvif_port"`
	Enabled     *bool  `json:"enabled"`
	AutoRecord  *bool  `json:"auto_record"`
	PTZProfile  string `json:"ptz_profile"`
	Brand       string `json:"brand"`
}
