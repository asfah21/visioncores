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
		log.Printf("onvif %s: body not XML ct=%s len=%d body=%.400s", op, resp.Header.Get("Content-Type"), len(body), body)
		return err
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
