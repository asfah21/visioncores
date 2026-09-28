package cameras

import (
	"log"
	"os"
	"regexp"
	"strings"
)

// SeedFromYAML imports config/cameras.yaml once when the cameras table is empty.
// Expected minimal format (same as repo):
//
//	cameras:
//	  - id: cam1
//	    label: Cam 1
//	    description: Ruang Tengah
//	    rtsp_url: "${CAM1_RTSP_URL}" or literal rtsp://...
//
// ${VAR} is expanded from environment so secrets never live in git.
func (s *Store) SeedFromYAML(path string) {
	var n int
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM cameras`).Scan(&n)
	if n > 0 {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Println("camera seed: no seed file:", path)
		return
	}
	text := os.ExpandEnv(string(raw))
	blocks := regexp.MustCompile(`(?m)^\s*-\s*id:\s*([^\r\n#]+)`).FindAllStringSubmatchIndex(text, -1)
	if len(blocks) == 0 {
		return
	}
	for i, b := range blocks {
		id := strings.TrimSpace(text[b[2]:b[3]])
		end := len(text)
		if i+1 < len(blocks) {
			end = blocks[i+1][0]
		}
		body := text[b[1]:end]
		label := pick(body, "label")
		desc := pick(body, "description")
		rtsp := pick(body, "rtsp_url")
		rtsp = strings.Trim(rtsp, `"' `)
		if id == "" || rtsp == "" || strings.Contains(rtsp, "${") {
			log.Println("camera seed: skip", id, "(missing rtsp_url env?)")
			continue
		}
		if !ValidID(id) {
			continue
		}
		host, port := ParseHostPort(rtsp, 8554)
		tmpl := StripPassword(rtsp)
		// try to split userinfo for username storage (password stays in template? no -> encrypt)
		username, password := splitUser(rtsp)
		enc := ""
		if password != "" {
			enc, _ = Encrypt(s.CredentialKey, password)
		}
		if label == "" {
			label = id
		}
		_, err := s.DB.Exec(`INSERT INTO cameras(id,name,description,host,rtsp_port,rtsp_url,username,password_enc,onvif_port,enabled,auto_record)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,8000,true,false) ON CONFLICT (id) DO NOTHING`,
			id, label, desc, host, port, tmpl, username, enc)
		if err != nil {
			log.Println("camera seed insert:", err)
			continue
		}
		log.Println("camera seed: imported", id)
	}
}

func pick(body, key string) string {
	re := regexp.MustCompile(`(?m)^\s*` + key + `:\s*(?:"([^"]*)"|'([^']*)'|(.+?))\s*$`)
	m := re.FindStringSubmatch(body)
	if m == nil {
		return ""
	}
	for _, g := range m[1:] {
		if strings.TrimSpace(g) != "" {
			return strings.TrimSpace(g)
		}
	}
	return ""
}

func splitUser(raw string) (string, string) {
	// rtsp://user:pass@host/...
	re := regexp.MustCompile(`^rtsp://([^:/@]+)(?::([^@]*))?@`)
	m := re.FindStringSubmatch(raw)
	if m == nil {
		return "", ""
	}
	return m[1], m[2]
}
