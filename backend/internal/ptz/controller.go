package ptz

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"backend/internal/cameras"
	"backend/internal/onvif"
)

// Controller is GSIVision's PTZ abstraction. No ONVIF library types leak here.
type Controller interface {
	Move(ctx context.Context, cameraID string, pan, tilt, zoom float64, durationMs int) error
	Stop(ctx context.Context, cameraID string) error
	Info(ctx context.Context, cameraID string) (onvif.DeviceInfo, []onvif.Profile, onvif.Capabilities, error)
	Status(ctx context.Context, cameraID string) (onvif.PTZStatus, error)
	ListPresets(ctx context.Context, cameraID string) ([]onvif.Preset, error)
	GotoPreset(ctx context.Context, cameraID, presetToken string) error
	SetPreset(ctx context.Context, cameraID, name string) (string, error)
	GoHome(ctx context.Context, cameraID string) error
}

type controller struct {
	cams   *cameras.Store
	client onvif.OnvifClient

	mu       sync.Mutex
	profiles map[string]cachedProfile
	timers   map[string]*time.Timer
}

type cachedProfile struct {
	token string
	exp   time.Time
}

func NewController(cams *cameras.Store, client onvif.OnvifClient) Controller {
	return &controller{cams: cams, client: client, profiles: map[string]cachedProfile{}, timers: map[string]*time.Timer{}}
}

func (c *controller) device(cameraID string) (cameras.Camera, string, onvif.Device, error) {
	cam, pw, err := c.cams.Get(cameraID)
	if err != nil {
		return cameras.Camera{}, "", onvif.Device{}, fmt.Errorf("camera not found")
	}
	return cam, pw, onvif.Device{Host: cam.Host, OnvifPort: cam.OnvifPort, Username: cam.Username, Password: pw}, nil
}

func (c *controller) profileToken(ctx context.Context, cameraID string, cam cameras.Camera, pw string, dev onvif.Device) (string, error) {
	c.mu.Lock()
	if p, ok := c.profiles[cameraID]; ok && time.Now().Before(p.exp) && p.token != "" {
		tok := p.token
		c.mu.Unlock()
		return tok, nil
	}
	c.mu.Unlock()
	// Manual override: skip GetProfiles discovery entirely. Some firmwares
	// return malformed XAddrs or truncated discovery responses while PTZ
	// itself works fine with a known token (e.g. IPCProfilesToken0).
	if tok := strings.TrimSpace(cam.PTZProfile); tok != "" {
		return tok, nil
	}
	ctx2, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	profs, err := c.client.Profiles(ctx2, dev)
	if err != nil {
		// surface the real cause (host only, never credentials).
		// Mention credentials only when the failure actually looks like auth.
		return "", fmt.Errorf("onvif profiles for %s: %w %s", dev.Host, err, profileHint(err))
	}
	if len(profs) == 0 {
		return "", fmt.Errorf("no media profiles for %s (device returned an empty profile list; set a manual PTZ profile for this camera)", dev.Host)
	}
	tok := profs[0].Token
	c.mu.Lock()
	c.profiles[cameraID] = cachedProfile{token: tok, exp: time.Now().Add(5 * time.Minute)}
	c.mu.Unlock()
	return tok, nil
}

func validSpeed(v float64) bool { return v >= -1 && v <= 1 }

// profileHint tailors the guidance: credentials are mentioned only when the
// failure actually looks like auth; otherwise point at discovery/XAddr issues.
func profileHint(err error) string {
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "401") || strings.Contains(s, "unauthor") || strings.Contains(s, "forbidden") || strings.Contains(s, "auth") {
		return "(check ONVIF credentials)"
	}
	return "(profile discovery failed: malformed XAddr or broken GetProfiles response; set a manual PTZ profile for this camera)"
}

// Move sends ContinuousMove then guarantees Stop after durationMs.
// durationMs is clamped to 100..5000 so PTZ can never run unbounded.
func (c *controller) Move(ctx context.Context, cameraID string, pan, tilt, zoom float64, durationMs int) error {
	if !validSpeed(pan) || !validSpeed(tilt) || !validSpeed(zoom) {
		return fmt.Errorf("pan/tilt/zoom must be in [-1,1]")
	}
	if durationMs <= 0 {
		durationMs = 800
	}
	if durationMs < 100 {
		durationMs = 100
	}
	if durationMs > 5000 {
		durationMs = 5000
	}
	cam, _, dev, err := c.device(cameraID)
	if err != nil {
		return err
	}
	_ = cam
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	token, err := c.profileToken(ctx2, cameraID, cam, "", dev)
	if err != nil {
		return err
	}
	if err := c.client.Move(ctx2, dev, token, pan, tilt, zoom); err != nil {
		return err
	}
	// schedule guaranteed stop (serialised per camera)
	c.mu.Lock()
	if t, ok := c.timers[cameraID]; ok && t != nil {
		t.Stop()
	}
	camID := cameraID
	c.timers[cameraID] = time.AfterFunc(time.Duration(durationMs)*time.Millisecond, func() {
		sctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		_ = c.client.Stop(sctx, dev, token)
		c.mu.Lock()
		delete(c.timers, camID)
		c.mu.Unlock()
	})
	c.mu.Unlock()
	return nil
}

func (c *controller) Stop(ctx context.Context, cameraID string) error {
	cam, _, dev, err := c.device(cameraID)
	if err != nil {
		return err
	}
	_ = cam
	c.mu.Lock()
	if t, ok := c.timers[cameraID]; ok && t != nil {
		t.Stop()
		delete(c.timers, cameraID)
	}
	c.mu.Unlock()
	ctx2, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	token, err := c.profileToken(ctx2, cameraID, cam, "", dev)
	if err != nil {
		return err
	}
	return c.client.Stop(ctx2, dev, token)
}

func (c *controller) Info(ctx context.Context, cameraID string) (onvif.DeviceInfo, []onvif.Profile, onvif.Capabilities, error) {
	cam, _, dev, err := c.device(cameraID)
	if err != nil {
		return onvif.DeviceInfo{}, nil, onvif.Capabilities{}, err
	}
	_ = cam
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	info, err := c.client.DeviceInfo(ctx2, dev)
	if err != nil {
		return onvif.DeviceInfo{}, nil, onvif.Capabilities{}, err
	}
	profs, _ := c.client.Profiles(ctx2, dev)
	caps, _ := c.client.Capabilities(ctx2, dev)
	return info, profs, caps, nil
}

func (c *controller) Status(ctx context.Context, cameraID string) (onvif.PTZStatus, error) {
	cam, _, dev, err := c.device(cameraID)
	if err != nil {
		return onvif.PTZStatus{}, err
	}
	_ = cam
	ctx2, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	token, err := c.profileToken(ctx2, cameraID, cam, "", dev)
	if err != nil {
		return onvif.PTZStatus{}, err
	}
	return c.client.Status(ctx2, dev, token)
}

func (c *controller) ListPresets(ctx context.Context, cameraID string) ([]onvif.Preset, error) {
	cam, _, dev, err := c.device(cameraID)
	if err != nil {
		return nil, err
	}
	_ = cam
	ctx2, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	token, err := c.profileToken(ctx2, cameraID, cam, "", dev)
	if err != nil {
		return nil, err
	}
	return c.client.Presets(ctx2, dev, token)
}

func (c *controller) GotoPreset(ctx context.Context, cameraID, presetToken string) error {
	if presetToken == "" {
		return fmt.Errorf("preset token is required")
	}
	cam, _, dev, err := c.device(cameraID)
	if err != nil {
		return err
	}
	_ = cam
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	token, err := c.profileToken(ctx2, cameraID, cam, "", dev)
	if err != nil {
		return err
	}
	return c.client.GotoPreset(ctx2, dev, token, presetToken)
}

func (c *controller) SetPreset(ctx context.Context, cameraID, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("preset name is required")
	}
	cam, _, dev, err := c.device(cameraID)
	if err != nil {
		return "", err
	}
	_ = cam
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	token, err := c.profileToken(ctx2, cameraID, cam, "", dev)
	if err != nil {
		return "", err
	}
	return c.client.SetPreset(ctx2, dev, token, name)
}

func (c *controller) GoHome(ctx context.Context, cameraID string) error {
	cam, _, dev, err := c.device(cameraID)
	if err != nil {
		return err
	}
	_ = cam
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	token, err := c.profileToken(ctx2, cameraID, cam, "", dev)
	if err != nil {
		return err
	}
	return c.client.GotoHome(ctx2, dev, token)
}
