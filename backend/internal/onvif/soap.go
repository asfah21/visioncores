package onvif

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/beevik/etree"
)

// readSOAP unwraps a SOAP envelope and decodes the first Body child into out.
// op labels the operation for diagnostics. Failures are logged server-side
// with a body snippet (responses never carry passwords) so wire-level
// incompatibilities with picky devices can be diagnosed from backend logs.
func readSOAP(resp *http.Response, out interface{}, op string) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("onvif %s: http %d ct=%s body=%.300s", op, resp.StatusCode, resp.Header.Get("Content-Type"), body)
		return fmt.Errorf("onvif http %d: %s", resp.StatusCode, truncate(string(body), 300))
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromString(string(body)); err != nil {
		log.Printf("onvif %s: body not XML ct=%s len=%d body=%q", op, resp.Header.Get("Content-Type"), len(body), body)
		// Some devices emit bytes illegal in XML 1.0 (bad UTF-8, control
		// chars in strings). Strict parsers (Go, etree-validate) reject the
		// whole document while lenient ones (Python/zeep) accept it.
		// Retry once with a sanitized copy.
		if clean := sanitizeXML(body); len(clean) > 0 {
			if err2 := doc.ReadFromString(string(clean)); err2 == nil {
				log.Printf("onvif %s: recovered by sanitizing %d->%d bytes", op, len(body), len(clean))
				body = clean
			} else {
				return err
			}
		} else {
			return err
		}
	}
	root := doc.Root()
	if root == nil {
		return fmt.Errorf("empty soap response")
	}
	// find Body element regardless of namespace prefix
	var bodyEl *etree.Element
	for _, ch := range root.ChildElements() {
		if strings.HasSuffix(ch.Tag, "Body") {
			bodyEl = ch
			break
		}
	}
	if bodyEl == nil {
		return fmt.Errorf("soap Body not found")
	}
	children := bodyEl.ChildElements()
	if len(children) == 0 {
		return fmt.Errorf("soap Body empty")
	}
	sub := etree.NewDocument()
	sub.SetRoot(children[0].Copy())
	payload, err := sub.WriteToString()
	if err != nil {
		return err
	}
	// strip outer response wrapper: decode inner first response struct
	// use-go types are shaped as <GetXResponse>...</GetXResponse> directly.
	if err := xml.Unmarshal([]byte(payload), out); err != nil {
		log.Printf("onvif %s: soap decode failed ct=%s body=%.400s", op, resp.Header.Get("Content-Type"), body)
		return fmt.Errorf("soap decode: %w", err)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// sanitizeXML drops bytes illegal in XML 1.0 so a document with device-side
// garbage still parses. Well-formed documents pass through byte-identical.
func sanitizeXML(b []byte) []byte {
	s := strings.ToValidUTF8(string(b), "")
	return []byte(strings.Map(func(r rune) rune {
		switch {
		case r == 0x9 || r == 0xA || r == 0xD:
			return r
		case r < 0x20 || (r >= 0x7F && r <= 0x84) || (r >= 0x86 && r <= 0x9F):
			return -1
		default:
			return r
		}
	}, s))
}
