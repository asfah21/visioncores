package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	JWTSecret          string
	InternalAPIToken   string
	RecordingPath      string
	RetentionDays      int
	SegmentSeconds     int
	Go2rtcURL          string
	Go2rtcConfigPath   string
	CredentialKey      string
	FFmpegBin          string
	CameraSeedPath     string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// Load reads all backend configuration from environment with safe defaults.
// RECORDING_PATH defaults to ./recordings for local dev, /recordings in docker.
func Load() Config {
	return Config{
		DBHost:         getenv("DB_HOST", "localhost"),
		DBPort:         getenv("DB_PORT", "5433"),
		DBUser:         getenv("DB_USER", "vision"),
		DBPass:         getenv("DB_PASS", "vision123"),
		DBName:         getenv("DB_NAME", "visiondb"),
		JWTSecret:      getenv("JWT_SECRET", "visioncore-super-secret-key-change-in-production"),
		InternalAPIToken: getenv("INTERNAL_API_TOKEN", ""),
		RecordingPath:  getenv("RECORDING_PATH", "./recordings"),
		RetentionDays:  getint("RECORDING_RETENTION_DAYS", 7),
		SegmentSeconds: getint("RECORDING_SEGMENT_SECONDS", 300),
		Go2rtcURL:      getenv("GO2RTC_URL", "http://media:1984"),
		Go2rtcConfigPath: getenv("GO2RTC_CONFIG_PATH", "../media/go2rtc.yaml"),
		CredentialKey:  getenv("CREDENTIAL_KEY", "visioncore-credential-key-change-me-32b"),
		FFmpegBin:      getenv("FFMPEG_BIN", "ffmpeg"),
		CameraSeedPath: getenv("CAMERA_SEED_PATH", "../config/cameras.yaml"),
	}
}
