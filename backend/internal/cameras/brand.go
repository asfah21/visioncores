package cameras

import "strings"

// Brand catalog for PTZ behavior. Selecting a brand auto-configures the
// matching ONVIF profile token so users never type tokens manually.
// Brands without a known token use live discovery (empty = auto).
//
//	bardi/tuya -> IPCProfilesToken0 (discovery responses truncated on this firmware)
var brandProfileTokens = map[string]string{
	"bardi": "IPCProfilesToken0",
	"tuya":  "IPCProfilesToken0",
}

// BrandPTZProfile returns the known ONVIF profile token for a brand,
// or "" when the brand uses live discovery.
func BrandPTZProfile(brand string) string {
	return brandProfileTokens[strings.ToLower(strings.TrimSpace(brand))]
}

// KnownBrands lists selectable brands: generic first, then presets.
func KnownBrands() []string {
	return []string{"", "bardi"}
}

// BrandLabel is the display name for a brand value.
func BrandLabel(brand string) string {
	switch strings.ToLower(strings.TrimSpace(brand)) {
	case "bardi", "tuya":
		return "Bardi / Tuya"
	default:
		return "Generic ONVIF"
	}
}
