package recorder

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"backend/internal/cameras"
	"backend/internal/config"
	"backend/internal/recordings"
)

type running struct {
	cmd   *exec.Cmd
	recID int
	dir   string
	start time.Time
}

// Manager orchestrates one ffmpeg process per camera writing MP4 segments.
type Manager struct {
	cfg  config.Config
	cams *cameras.Store
	recs *recordings.Store

	mu sync.Mutex
	m  map[string]*running
}

func NewManager(cfg config.Config, cams *cameras.Store, recs *recordings.Store) *Manager {
	return &Manager{cfg: cfg, cams: cams, recs: recs, m: map[string]*running{}}
}

func (m *Manager) IsRunning(cameraID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.m[cameraID]
	return ok
}

// safeDir builds RECORDING_PATH/camera-id/YYYY/MM/DD/HH with traversal guard.
func (m *Manager) safeDir(cameraID string, t time.Time) (string, error) {
	if !cameras.ValidID(cameraID) {
		return "", fmt.Errorf("invalid camera id")
	}
	base, err := filepath.Abs(m.cfg.RecordingPath)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, cameraID, fmt.Sprintf("%04d", t.Year()), fmt.Sprintf("%02d", int(t.Month())), fmt.Sprintf("%02d", t.Day()), fmt.Sprintf("%02d", t.Hour()))
	clean := filepath.Clean(dir)
	rel, err := filepath.Rel(base, clean)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path traversal blocked")
	}
	if err := os.MkdirAll(clean, 0755); err != nil {
		return "", err
	}
	return clean, nil
}

func (m *Manager) Start(cameraID string) error {
	m.mu.Lock()
	if _, ok := m.m[cameraID]; ok {
		m.mu.Unlock()
		return fmt.Errorf("already recording")
	}
	m.mu.Unlock()

	cam, _, err := m.cams.Get(cameraID)
	if err != nil {
		return fmt.Errorf("camera not found")
	}
	if cam.RTSPURL == "" {
		return fmt.Errorf("camera has no rtsp_url")
	}
	now := time.Now()
	dir, err := m.safeDir(cameraID, now)
	if err != nil {
		return err
	}
	outPattern := filepath.Join(dir, "%M-%S.mp4")
	// placeholder row: file_path stores directory + session marker; segments append per file.
	// We store one row per session dir for simplicity; actual segments listed from FS on playback.
	sessionPath := filepath.Join(dir, fmt.Sprintf("session-%s.txt", now.Format("150405")))
	recID, err := m.recs.Create(cameraID, now, sessionPath)
	if err != nil {
		return err
	}
	seg := m.cfg.SegmentSeconds
	if seg <= 0 {
		seg = 300
	}
	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-i", cam.RTSPURL,
		"-c:v", "copy", "-an",
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%d", seg),
		"-segment_format", "mp4",
		"-strftime", "1",
		outPattern,
	}
	cmd := exec.Command(m.cfg.FFmpegBin, args...)
	cmd.Stdout = nil
	// NOTE: stderr is intentionally discarded to avoid logging RTSP url/credentials.
	if err := cmd.Start(); err != nil {
		m.recs.Fail(recID)
		return fmt.Errorf("ffmpeg start: %w", err)
	}
	m.mu.Lock()
	m.m[cameraID] = &running{cmd: cmd, recID: recID, dir: dir, start: now}
	m.mu.Unlock()
	log.Println("recording started:", cameraID)
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		r := m.m[cameraID]
		delete(m.m, cameraID)
		m.mu.Unlock()
		if r == nil {
			return
		}
		if err != nil {
			log.Println("recording ffmpeg exit:", cameraID, err)
			m.recs.Fail(r.recID)
			return
		}
		var total int64
		entries, _ := os.ReadDir(r.dir)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".mp4") {
				if fi, err := e.Info(); err == nil {
					total += fi.Size()
				}
			}
		}
		m.recs.Finish(r.recID, time.Now(), total)
	}()
	return nil
}

func (m *Manager) Stop(cameraID string) error {
	m.mu.Lock()
	r, ok := m.m[cameraID]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("not recording")
	}
	if r.cmd.Process != nil {
		_ = r.cmd.Process.Kill()
	}
	return nil
}

// AutoStart resumes auto_record cameras on boot.
func (m *Manager) AutoStart() {
	list, err := m.cams.List(true)
	if err != nil {
		return
	}
	for _, c := range list {
		if c.AutoRecord {
			if err := m.Start(c.ID); err != nil {
				log.Println("autostart", c.ID, err)
			}
		}
	}
}
