package go2rtc

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/internal/cameras"
	"backend/internal/config"
)

// SyncFromDB regenerates go2rtc.yaml streams from DB (full RTSP) and best-effort reloads media.
// When DB holds no usable streams it leaves the existing file untouched so a
// working go2rtc.yaml (e.g. ${CAMx_RTSP_URL} expanded by the media container)
// is never wiped by an empty seed.
func SyncFromDB(cfg config.Config, list []cameras.Camera) error {
	count := 0
	for _, c := range list {
		if c.Enabled && c.RTSPURL != "" {
			count++
		}
	}
	if count == 0 {
		log.Println("go2rtc sync: no enabled cameras with rtsp_url in DB, keeping existing go2rtc.yaml")
		return nil
	}
	var b strings.Builder
	b.WriteString("# GENERATED FILE - DO NOT EDIT. Source: PostgreSQL cameras (synced by backend)\n")
	b.WriteString("streams:\n")
	for _, c := range list {
		if !c.Enabled || c.RTSPURL == "" {
			continue
		}
		// quote to keep YAML valid
		b.WriteString(fmt.Sprintf("  %s: \"%s\"\n", c.ID, c.RTSPURL))
	}
	b.WriteString("\napi:\n  listen: \":1984\"\n\nrtsp:\n  listen: \":8554\"\n  protocols:\n    - tcp  # Force TCP\n")
	if cfg.Go2rtcConfigPath != "" {
		_ = os.WriteFile(cfg.Go2rtcConfigPath, []byte(b.String()), 0644)
	}
	// best-effort reload; ignore errors (media may restart on its own)
	client := &http.Client{Timeout: 3 * time.Second}
	req, _ := http.NewRequest("POST", strings.TrimRight(cfg.Go2rtcURL, "/")+"/api/reload", nil)
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
	}
	return nil
}

// StreamURLs builds frontend-facing URLs from GO2RTC_URL (public) for a camera id.
func StreamURLs(cfg config.Config, cameraID string) map[string]string {
	base := strings.TrimRight(cfg.Go2rtcURL, "/")
	if base == "" {
		base = "http://localhost:1984"
	}
	return map[string]string{
		"webrtc": base + "/stream.html?src=" + cameraID,
		"mse":    base + "/api/stream.mse?src=" + cameraID,
		"hls":    base + "/api/stream.m3u8?src=" + cameraID,
		"rtsp":   "rtsp://" + cameraID,
	}
}
