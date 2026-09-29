package cameras

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	DB            *sql.DB
	CredentialKey string
}

func NewStore(db *sql.DB, credKey string) *Store { return &Store{DB: db, CredentialKey: credKey} }

func (s *Store) List(onlyEnabled bool) ([]Camera, error) {
	q := `SELECT id,name,description,host,rtsp_port,rtsp_url,username,password_enc,onvif_port,enabled,auto_record,created_at,updated_at FROM cameras ORDER BY id`
	if onlyEnabled {
		q = `SELECT id,name,description,host,rtsp_port,rtsp_url,username,password_enc,onvif_port,enabled,auto_record,created_at,updated_at FROM cameras WHERE enabled=true ORDER BY id`
	}
	rows, err := s.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// init non-nil agar JSON selalu [] bukan null
	out := []Camera{}
	for rows.Next() {
		var c Camera
		var enc string
		var tmpl string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Host, &c.RTSPPort, &tmpl, &c.Username, &enc, &c.OnvifPort, &c.Enabled, &c.AutoRecord, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		if pw, err := Decrypt(s.CredentialKey, enc); err == nil && pw != "" {
			c.RTSPURL = InjectPassword(tmplWithCreds(tmpl, c.Username), c.Username, pw)
		} else {
			c.RTSPURL = tmpl
		}
		out = append(out, c)
	}
	return out, nil
}

func tmplWithCreds(tmpl, username string) string {
	// stored template has no userinfo; keep as-is, InjectPassword adds it.
	return tmpl
}

func (s *Store) Get(id string) (Camera, string, error) {
	var c Camera
	var enc, tmpl string
	err := s.DB.QueryRow(`SELECT id,name,description,host,rtsp_port,rtsp_url,username,password_enc,onvif_port,enabled,auto_record,created_at,updated_at FROM cameras WHERE id=$1`, id).
		Scan(&c.ID, &c.Name, &c.Description, &c.Host, &c.RTSPPort, &tmpl, &c.Username, &enc, &c.OnvifPort, &c.Enabled, &c.AutoRecord, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, "", err
	}
	pw, _ := Decrypt(s.CredentialKey, enc)
	if pw != "" {
		c.RTSPURL = InjectPassword(tmpl, c.Username, pw)
	} else {
		c.RTSPURL = tmpl
	}
	return c, pw, nil
}

type Validated struct {
	Input       UpsertInput
	Host        string
	RTSPPort    int
	TemplateURL string // stored without password
	EncPassword string
}

func (s *Store) Validate(in UpsertInput) (Validated, error) {
	var v Validated
	in.ID = strings.TrimSpace(in.ID)
	if !ValidID(in.ID) {
		return v, fmt.Errorf("invalid id: use letters, numbers, _ or -")
	}
	if strings.TrimSpace(in.Name) == "" {
		return v, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(in.RTSPURL) == "" {
		return v, fmt.Errorf("rtsp_url is required")
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(in.RTSPURL)), "rtsp://") {
		return v, fmt.Errorf("rtsp_url must start with rtsp://")
	}
	host, port := ParseHostPort(in.RTSPURL, 8554)
	if in.Host != "" {
		host = strings.TrimSpace(in.Host)
	}
	rtspPort := port
	if in.RTSPPort > 0 {
		rtspPort = in.RTSPPort
	}
	onvifPort := in.OnvifPort
	if onvifPort == 0 {
		onvifPort = 8000
	}
	enc := ""
	if in.Password != "" {
		e, err := Encrypt(s.CredentialKey, in.Password)
		if err != nil {
			return v, err
		}
		enc = e
	}
	v = Validated{Input: in, Host: host, RTSPPort: rtspPort, TemplateURL: StripPassword(strings.TrimSpace(in.RTSPURL)), EncPassword: enc}
	// overlay ports
	v.Input.OnvifPort = onvifPort
	return v, nil
}

func (s *Store) Create(in UpsertInput) (Camera, error) {
	// auto-generate id (cam1, cam2, ...) when the client omits it
	if strings.TrimSpace(in.ID) == "" {
		id, err := s.nextCameraID()
		if err != nil {
			return Camera{}, err
		}
		in.ID = id
	}
	v, err := s.Validate(in)
	if err != nil {
		return Camera{}, err
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	autoRec := false
	if in.AutoRecord != nil {
		autoRec = *in.AutoRecord
	}
	name := strings.TrimSpace(in.Name)
	username := strings.TrimSpace(in.Username)
	_, err = s.DB.Exec(`INSERT INTO cameras(id,name,description,host,rtsp_port,rtsp_url,username,password_enc,onvif_port,enabled,auto_record,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now())`,
		in.ID, name, strings.TrimSpace(in.Description), v.Host, v.RTSPPort, v.TemplateURL, username, v.EncPassword, v.Input.OnvifPort, enabled, autoRec)
	if err != nil {
		return Camera{}, err
	}
	c, _, _ := s.Get(in.ID)
	return c, nil
}

func (s *Store) Update(id string, in UpsertInput) (Camera, error) {
	cur, curPw, err := s.Get(id)
	if err != nil {
		return Camera{}, err
	}
	in.ID = id
	if strings.TrimSpace(in.Name) == "" {
		in.Name = cur.Name
	}
	if strings.TrimSpace(in.RTSPURL) == "" {
		// keep existing template; reconstruct display url from cur for validation
		in.RTSPURL = cur.RTSPURL
	}
	if strings.TrimSpace(in.Host) == "" {
		in.Host = cur.Host
	}
	if in.RTSPPort == 0 {
		in.RTSPPort = cur.RTSPPort
	}
	if in.OnvifPort == 0 {
		in.OnvifPort = cur.OnvifPort
	}
	if strings.TrimSpace(in.Username) == "" {
		in.Username = cur.Username
	}
	pw := in.Password
	enc := ""
	if pw == "" {
		// keep old encrypted password
		var old string
		_ = s.DB.QueryRow(`SELECT password_enc FROM cameras WHERE id=$1`, id).Scan(&old)
		enc = old
		pw = curPw
		_ = pw
		v, err := s.Validate(in)
		if err != nil {
			return Camera{}, err
		}
		enabled := cur.Enabled
		if in.Enabled != nil {
			enabled = *in.Enabled
		}
		autoRec := cur.AutoRecord
		if in.AutoRecord != nil {
			autoRec = *in.AutoRecord
		}
		_, err = s.DB.Exec(`UPDATE cameras SET name=$2,description=$3,host=$4,rtsp_port=$5,rtsp_url=$6,username=$7,password_enc=$8,onvif_port=$9,enabled=$10,auto_record=$11,updated_at=now() WHERE id=$1`,
			id, strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), v.Host, v.RTSPPort, v.TemplateURL, strings.TrimSpace(in.Username), enc, v.Input.OnvifPort, enabled, autoRec)
		if err != nil {
			return Camera{}, err
		}
		c, _, _ := s.Get(id)
		return c, nil
	}
	v, err := s.Validate(in)
	if err != nil {
		return Camera{}, err
	}
	enabled := cur.Enabled
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	autoRec := cur.AutoRecord
	if in.AutoRecord != nil {
		autoRec = *in.AutoRecord
	}
	_, err = s.DB.Exec(`UPDATE cameras SET name=$2,description=$3,host=$4,rtsp_port=$5,rtsp_url=$6,username=$7,password_enc=$8,onvif_port=$9,enabled=$10,auto_record=$11,updated_at=now() WHERE id=$1`,
		id, strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), v.Host, v.RTSPPort, v.TemplateURL, strings.TrimSpace(in.Username), v.EncPassword, v.Input.OnvifPort, enabled, autoRec)
	if err != nil {
		return Camera{}, err
	}
	c, _, _ := s.Get(id)
	return c, nil
}

func (s *Store) Delete(id string) error {
	_, err := s.DB.Exec(`DELETE FROM cameras WHERE id=$1`, id)
	return err
}

// nextCameraID returns the first free camN id (cam1, cam2, ...).
// The PRIMARY KEY constraint is the final arbiter against races;
// callers surface the conflict error so the client can retry.
func (s *Store) nextCameraID() (string, error) {
	rows, err := s.DB.Query(`SELECT id FROM cameras`)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	taken := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			taken[id] = true
		}
	}
	for n := 1; n < 10000; n++ {
		id := fmt.Sprintf("cam%d", n)
		if !taken[id] {
			return id, nil
		}
	}
	return "", fmt.Errorf("no free camera id")
}

func (s *Store) Touch(id string) {
	_, _ = s.DB.Exec(`UPDATE cameras SET updated_at=now() WHERE id=$1`, id)
}

var _ = time.Now
