package cameras

import (
	"strings"
	"testing"
)

func TestValidID(t *testing.T) {
	if !ValidID("cam1") || !ValidID("cam-1_2") {
		t.Fatal("expected valid ids")
	}
	if ValidID("cam 1") || ValidID("cam/1") || ValidID("") {
		t.Fatal("expected invalid ids rejected")
	}
}

func TestSanitizeNeverLeaksPassword(t *testing.T) {
	raw := "rtsp://admin:s3cret@10.0.0.5:8554/Streaming/Channels/101"
	s := SanitizeRTSPForLog(raw)
	if strings.Contains(s, "s3cret") {
		t.Fatal("password leaked in log string:", s)
	}
	red := RedactedHost(raw)
	if strings.Contains(red, "admin") || strings.Contains(red, "s3cret") {
		t.Fatal("redacted host still has userinfo:", red)
	}
}

func TestEncryptRoundtrip(t *testing.T) {
	enc, err := Encrypt("test-key-12345678901234567890", "p@ss:w0rd")
	if err != nil || enc == "" {
		t.Fatal("encrypt failed", err)
	}
	if strings.Contains(enc, "p@ss") {
		t.Fatal("ciphertext leaks plaintext")
	}
	dec, err := Decrypt("test-key-12345678901234567890", enc)
	if err != nil || dec != "p@ss:w0rd" {
		t.Fatal("decrypt roundtrip failed", dec, err)
	}
}

func TestStripAndInject(t *testing.T) {
	raw := "rtsp://admin:secret@host:8554/Streaming/Channels/101"
	tmpl := StripPassword(raw)
	if strings.Contains(tmpl, "secret") {
		t.Fatal("template still contains password")
	}
	full := InjectPassword(tmpl, "admin", "secret")
	if !strings.Contains(full, "host") || !strings.Contains(full, "Streaming") {
		t.Fatal("inject broke url:", full)
	}
}
