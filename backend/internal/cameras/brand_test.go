package cameras

import "testing"

func TestBrandPTZProfile(t *testing.T) {
	if got := BrandPTZProfile("bardi"); got != "IPCProfilesToken0" {
		t.Fatalf("bardi = %q", got)
	}
	if got := BrandPTZProfile(" Tuya "); got != "IPCProfilesToken0" {
		t.Fatalf("tuya = %q", got)
	}
	for _, b := range []string{"", "generic", "hikvision", "dahua", "unknown-brand"} {
		if got := BrandPTZProfile(b); got != "" {
			t.Fatalf("%q should use discovery, got %q", b, got)
		}
	}
}

func TestBrandLabel(t *testing.T) {
	if BrandLabel("bardi") != "Bardi / Tuya" {
		t.Fatal("bardi label")
	}
	if BrandLabel("") != "Generic ONVIF" {
		t.Fatal("default label")
	}
}
