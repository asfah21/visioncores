package onvif

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/media"
	"github.com/use-go/onvif/ptz"
	"github.com/use-go/onvif/xsd"
	xonvif "github.com/use-go/onvif/xsd/onvif"
)

// GoClient is the ONLY type allowed to touch use-go/onvif.
// All SOAP/XML specifics stay inside this file.
type GoClient struct {
	HTTP *http.Client
}

func NewGoClient() *GoClient {
	return &GoClient{HTTP: &http.Client{Timeout: 10 * time.Second}}
}

func (g *GoClient) open(dev Device) (*onvif.Device, error) {
	params := onvif.DeviceParams{
		Xaddr:      dev.XAddr(),
		Username:   dev.Username,
		Password:   dev.Password,
		HttpClient: g.HTTP,
	}
	d, err := onvif.NewDevice(params)
	if err != nil {
		return nil, fmt.Errorf("onvif connect %s: %w", dev.Host, err)
	}
	if debugONVIF() {
		log.Printf("onvif endpoints for %s: %v", dev.Host, d.GetServices())
	}
	return d, nil
}

var (
	debugOnce sync.Once
	debugVal  bool
)

func debugONVIF() bool {
	debugOnce.Do(func() { debugVal = os.Getenv("ONVIF_DEBUG") != "" })
	return debugVal
}

func (g *GoClient) DeviceInfo(ctx context.Context, dev Device) (DeviceInfo, error) {
	d, err := g.open(dev)
	if err != nil {
		return DeviceInfo{}, err
	}
	_ = ctx
	resp, err := d.CallMethod(device.GetDeviceInformation{})
	if err != nil {
		return DeviceInfo{}, err
	}
	defer resp.Body.Close()
	var out device.GetDeviceInformationResponse
	if err := readSOAP(resp, &out, "DeviceInfo"); err != nil {
		return DeviceInfo{}, err
	}
	return DeviceInfo{
		Manufacturer: out.Manufacturer,
		Model:        out.Model,
		Firmware:     out.FirmwareVersion,
		Serial:       out.SerialNumber,
		HardwareID:   out.HardwareId,
	}, nil
}

func (g *GoClient) Profiles(ctx context.Context, dev Device) ([]Profile, error) {
	d, err := g.open(dev)
	if err != nil {
		return nil, err
	}
	_ = ctx
	resp, err := d.CallMethod(media.GetProfiles{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out media.GetProfilesResponse
	if err := readSOAP(resp, &out, "GetProfiles"); err != nil {
		return nil, err
	}
	profs := make([]Profile, 0, len(out.Profiles))
	for _, p := range out.Profiles {
		profs = append(profs, Profile{Token: string(p.Token), Name: string(p.Name)})
	}
	return profs, nil
}

func (g *GoClient) Capabilities(ctx context.Context, dev Device) (Capabilities, error) {
	d, err := g.open(dev)
	if err != nil {
		return Capabilities{}, err
	}
	_ = ctx
	resp, err := d.CallMethod(device.GetCapabilities{Category: "PTZ"})
	if err != nil {
		return Capabilities{}, err
	}
	defer resp.Body.Close()
	caps := Capabilities{}
	for k := range d.GetServices() {
		if k == "PTZ" || k == "ptz" {
			caps.PTZ = true
		}
	}
	if !caps.PTZ {
		if profs, err := g.Profiles(ctx, dev); err == nil && len(profs) > 0 {
			caps.PTZ = true
		}
	}
	return caps, nil
}

func (g *GoClient) Status(ctx context.Context, dev Device, profileToken string) (PTZStatus, error) {
	d, err := g.open(dev)
	if err != nil {
		return PTZStatus{}, err
	}
	_ = ctx
	resp, err := d.CallMethod(ptz.GetStatus{ProfileToken: xonvif.ReferenceToken(profileToken)})
	if err != nil {
		return PTZStatus{}, err
	}
	defer resp.Body.Close()
	var out ptz.GetStatusResponse
	if err := readSOAP(resp, &out, "GetStatus"); err != nil {
		return PTZStatus{}, err
	}
	return PTZStatus{
		Pan:  out.PTZStatus.Position.PanTilt.X,
		Tilt: out.PTZStatus.Position.PanTilt.Y,
		Zoom: out.PTZStatus.Position.Zoom.X,
	}, nil
}

func (g *GoClient) Move(ctx context.Context, dev Device, profileToken string, pan, tilt, zoom float64) error {
	d, err := g.open(dev)
	if err != nil {
		return err
	}
	_ = ctx
	req := ptz.ContinuousMove{
		ProfileToken: xonvif.ReferenceToken(profileToken),
		Velocity: xonvif.PTZSpeed{
			PanTilt: xonvif.Vector2D{X: pan, Y: tilt},
			Zoom:    xonvif.Vector1D{X: zoom},
		},
	}
	resp, err := d.CallMethod(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ptz move http %d", resp.StatusCode)
	}
	return nil
}

func (g *GoClient) Stop(ctx context.Context, dev Device, profileToken string) error {
	d, err := g.open(dev)
	if err != nil {
		return err
	}
	_ = ctx
	resp, err := d.CallMethod(ptz.Stop{
		ProfileToken: xonvif.ReferenceToken(profileToken),
		PanTilt:      true,
		Zoom:         true,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// presetsResponse decodes ALL <Preset> elements. (The lib's own
// GetPresetsResponse holds a single Preset and would drop all but the last.)
type presetsResponse struct {
	Presets []presetEntry `xml:"Preset"`
}

type presetEntry struct {
	Token string `xml:"token,attr"`
	Name  string `xml:"Name"`
}

func (g *GoClient) Presets(ctx context.Context, dev Device, profileToken string) ([]Preset, error) {
	d, err := g.open(dev)
	if err != nil {
		return nil, err
	}
	_ = ctx
	resp, err := d.CallMethod(ptz.GetPresets{ProfileToken: xonvif.ReferenceToken(profileToken)})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out presetsResponse
	if err := readSOAP(resp, &out, "GetPresets"); err != nil {
		return nil, err
	}
	presets := make([]Preset, 0, len(out.Presets))
	for _, p := range out.Presets {
		presets = append(presets, Preset{Token: p.Token, Name: p.Name})
	}
	return presets, nil
}

func (g *GoClient) GotoPreset(ctx context.Context, dev Device, profileToken, presetToken string) error {
	d, err := g.open(dev)
	if err != nil {
		return err
	}
	_ = ctx
	resp, err := d.CallMethod(ptz.GotoPreset{
		ProfileToken: xonvif.ReferenceToken(profileToken),
		PresetToken:  xonvif.ReferenceToken(presetToken),
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ptz goto preset http %d", resp.StatusCode)
	}
	return nil
}

func (g *GoClient) SetPreset(ctx context.Context, dev Device, profileToken, name string) (string, error) {
	d, err := g.open(dev)
	if err != nil {
		return "", err
	}
	_ = ctx
	resp, err := d.CallMethod(ptz.SetPreset{
		ProfileToken: xonvif.ReferenceToken(profileToken),
		PresetName:   xsd.String(name),
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out ptz.SetPresetResponse
	if err := readSOAP(resp, &out, "SetPreset"); err != nil {
		return "", err
	}
	return string(out.PresetToken), nil
}

func (g *GoClient) GotoHome(ctx context.Context, dev Device, profileToken string) error {
	d, err := g.open(dev)
	if err != nil {
		return err
	}
	_ = ctx
	resp, err := d.CallMethod(ptz.GotoHomePosition{
		ProfileToken: xonvif.ReferenceToken(profileToken),
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ptz go home http %d", resp.StatusCode)
	}
	return nil
}
