package go2rtc

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Modular per-stream management via go2rtc HTTP API.
//
//   - PUT /api/streams?name={id}&src={rtspURL} creates/replaces ONE stream
//     in the running config AND persists it to go2rtc.yaml (no restart needed).
//   - DELETE /api/streams?src={id} removes ONE stream the same way.
//
// This keeps go2rtc in sync on every camera create/update/delete, while
// SyncFromDB remains as startup reconciliation (full rewrite).

func streamsClient() *http.Client { return &http.Client{Timeout: 5 * time.Second} }

// UpsertStream creates or replaces a single go2rtc stream.
func UpsertStream(go2rtcURL, name, streamURL string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(streamURL) == "" {
		return fmt.Errorf("go2rtc upsert: empty name or url")
	}
	// DELETE first (best-effort) so a re-PUT cleanly replaces an existing stream.
	_ = DeleteStream(go2rtcURL, name)
	q := url.Values{}
	q.Set("name", name)
	q.Add("src", streamURL)
	req, _ := http.NewRequest("PUT", strings.TrimRight(go2rtcURL, "/")+"/api/streams?"+q.Encode(), nil)
	resp, err := streamsClient().Do(req)
	if err != nil {
		return fmt.Errorf("go2rtc upsert %s: %w", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("go2rtc upsert %s: %s %s", name, resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

// DeleteStream removes a single go2rtc stream. Missing streams are not an error.
func DeleteStream(go2rtcURL, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("go2rtc delete: empty name")
	}
	q := url.Values{}
	q.Set("src", name)
	req, _ := http.NewRequest("DELETE", strings.TrimRight(go2rtcURL, "/")+"/api/streams?"+q.Encode(), nil)
	resp, err := streamsClient().Do(req)
	if err != nil {
		return fmt.Errorf("go2rtc delete %s: %w", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("go2rtc delete %s: %s %s", name, resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}
