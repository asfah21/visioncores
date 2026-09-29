// Package onvif exposes GSIVision's own ONVIF abstraction.
// Only goclient.go is allowed to import github.com/use-go/onvif.
// Every other package must depend on the interfaces below.
package onvif

import (
	"context"
	"time"
)

type Device struct {
	Host      string
	OnvifPort int
	Username  string
	Password  string
}

func (d Device) XAddr() string {
	if d.OnvifPort == 0 {
		d.OnvifPort = 8000
	}
	return joinHostPort(d.Host, d.OnvifPort)
}

type DeviceInfo struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Firmware     string `json:"firmware"`
	Serial       string `json:"serial,omitempty"`
	HardwareID   string `json:"hardware_id,omitempty"`
}

type Profile struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type PTZStatus struct {
	Pan  float64 `json:"pan"`
	Tilt float64 `json:"tilt"`
	Zoom float64 `json:"zoom"`
}

type Capabilities struct {
	PTZ bool `json:"ptz"`
}

// OnvifClient is the single seam for all ONVIF traffic.
type Preset struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type OnvifClient interface {
	DeviceInfo(ctx context.Context, dev Device) (DeviceInfo, error)
	Profiles(ctx context.Context, dev Device) ([]Profile, error)
	Capabilities(ctx context.Context, dev Device) (Capabilities, error)
	Status(ctx context.Context, dev Device, profileToken string) (PTZStatus, error)
	Move(ctx context.Context, dev Device, profileToken string, pan, tilt, zoom float64) error
	Stop(ctx context.Context, dev Device, profileToken string) error
	Presets(ctx context.Context, dev Device, profileToken string) ([]Preset, error)
	GotoPreset(ctx context.Context, dev Device, profileToken, presetToken string) error
	SetPreset(ctx context.Context, dev Device, profileToken, name string) (string, error)
	GotoHome(ctx context.Context, dev Device, profileToken string) error
}

func Timeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}
