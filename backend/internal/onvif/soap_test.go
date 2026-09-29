package onvif

import (
	"strings"
	"testing"

	"github.com/beevik/etree"
)

func mustParse(t *testing.T, body string) {
	t.Helper()
	doc := etree.NewDocument()
	if err := doc.ReadFromString(body); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if doc.Root() == nil {
		t.Fatal("no root element")
	}
}

func TestSanitizeXMLPassthrough(t *testing.T) {
	good := `<?xml version="1.0" encoding="UTF-8"?><SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope"><SOAP-ENV:Body><x>a&amp;b</x></SOAP-ENV:Body></SOAP-ENV:Envelope>`
	if got := string(sanitizeXML([]byte(good))); got != good {
		t.Fatal("well-formed document must pass through byte-identical")
	}
	mustParse(t, good)
}

func TestSanitizeXMLControlChars(t *testing.T) {
	// \x01 and \x0b are illegal in XML 1.0; strict parsers reject the doc.
	bad := "<?xml version=\"1.0\"?><root><name>Cam\x01 \x0b 1</name></root>"
	doc := etree.NewDocument()
	if err := doc.ReadFromString(bad); err == nil {
		t.Skip("parser accepts control chars; sanitize path not exercised")
	}
	clean := sanitizeXML([]byte(bad))
	if strings.ContainsAny(string(clean), "\x01\x0b") {
		t.Fatal("illegal chars not removed")
	}
	mustParse(t, string(clean))
}

func TestSanitizeXMLBadUTF8(t *testing.T) {
	bad := append([]byte(`<?xml version="1.0"?><root><name>Cam `), 0xff, 0xfe)
	bad = append(bad, []byte(` 1</name></root>`)...)
	doc := etree.NewDocument()
	if err := doc.ReadFromString(string(bad)); err == nil {
		t.Skip("parser accepts bad utf-8; sanitize path not exercised")
	}
	mustParse(t, string(sanitizeXML(bad)))
}
