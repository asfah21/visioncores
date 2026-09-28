package recordings

import (
	"database/sql"
	"strconv"
	"time"
)

type Recording struct {
	ID          int        `json:"id"`
	CameraID    string     `json:"camera_id"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	DurationSec int        `json:"duration"`
	FilePath    string     `json:"file_path"`
	FileSize    int64      `json:"file_size"`
	Format      string     `json:"format"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Store struct{ DB *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{DB: db} }

func (s *Store) Create(cameraID string, start time.Time, filePath string) (int, error) {
	var id int
	err := s.DB.QueryRow(`INSERT INTO recordings(camera_id,start_time,file_path,format,status)
		VALUES($1,$2,$3,'mp4','recording') RETURNING id`, cameraID, start, filePath).Scan(&id)
	return id, err
}

func (s *Store) Finish(id int, end time.Time, size int64) {
	dur := int(end.Sub(s.StartOf(id)).Seconds())
	if dur < 0 {
		dur = 0
	}
	_, _ = s.DB.Exec(`UPDATE recordings SET end_time=$2,duration_sec=$3,file_size=$4,status='completed' WHERE id=$1`, id, end, dur, size)
}

func (s *Store) StartOf(id int) time.Time {
	var t time.Time
	_ = s.DB.QueryRow(`SELECT start_time FROM recordings WHERE id=$1`, id).Scan(&t)
	return t
}

func (s *Store) Fail(id int) {
	_, _ = s.DB.Exec(`UPDATE recordings SET status='failed', end_time=now() WHERE id=$1`, id)
}

func (s *Store) Get(id int) (Recording, error) {
	var r Recording
	err := s.DB.QueryRow(`SELECT id,camera_id,start_time,end_time,duration_sec,file_path,file_size,format,status,created_at FROM recordings WHERE id=$1`, id).
		Scan(&r.ID, &r.CameraID, &r.StartTime, &r.EndTime, &r.DurationSec, &r.FilePath, &r.FileSize, &r.Format, &r.Status, &r.CreatedAt)
	return r, err
}

func (s *Store) List(cameraID string, from, to *time.Time, limit int) ([]Recording, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	q := `SELECT id,camera_id,start_time,end_time,duration_sec,file_path,file_size,format,status,created_at FROM recordings WHERE camera_id=$1`
	args := []interface{}{cameraID}
	idx := 2
	if from != nil {
		q += ` AND start_time >= $` + itoa(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		q += ` AND start_time < $` + itoa(idx)
		args = append(args, *to)
		idx++
	}
	q += ` ORDER BY start_time DESC LIMIT $` + itoa(idx)
	args = append(args, limit)
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Recording
	for rows.Next() {
		var r Recording
		if err := rows.Scan(&r.ID, &r.CameraID, &r.StartTime, &r.EndTime, &r.DurationSec, &r.FilePath, &r.FileSize, &r.Format, &r.Status, &r.CreatedAt); err != nil {
			continue
		}
		// hide absolute path from API; frontend uses /stream endpoint
		r.FilePath = ""
		out = append(out, r)
	}
	return out, nil
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
