// Package pricing estimates a deliberately limited steady-state subtotal using
// user-supplied offline rates. It never exports catalogue contents.
package pricing

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const maxBytes = 4 << 20
const maxSafe = int64(9007199254740991)

var invalid = errors.New("invalid pricing catalogue")
var regionPattern = regexp.MustCompile(`^[a-z]{2}(-[a-z]+)+-[0-9]+$`)
var decimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]{1,6})?$`)
var unitPricePattern = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]{1,12})?$`)
var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
var fields = map[string][]string{
	"aws_instance":    {"instance_type", "ami", "tenancy"},
	"aws_db_instance": {"instance_class", "engine", "multi_az", "license_model"},
}

type rate struct {
	ResourceType string            `json:"resource_type"`
	Region       string            `json:"region"`
	Selectors    map[string]string `json:"selectors"`
	HourlyPrice  string            `json:"hourly_price,omitempty"`
	Component    string            `json:"component,omitempty"`
	Unit         string            `json:"unit,omitempty"`
	UnitPrice    string            `json:"unit_price,omitempty"`
}

type Catalogue struct {
	currency, asOf string
	rates          map[string]int64
	version        string
	usage          map[string]map[string]map[string]int64
}

// Load accepts only bounded regular local files. Errors never include paths.
func Load(path string) (*Catalogue, error) {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBytes {
		return nil, errors.New("cannot read pricing catalogue: expected a bounded local regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("cannot read pricing catalogue")
	}
	defer f.Close()
	return Read(f)
}

func Read(r io.Reader) (*Catalogue, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil || len(data) > maxBytes {
		return nil, invalid
	}
	// Disallow duplicate JSON keys as well as duplicate rate identities.
	scan := json.NewDecoder(bytes.NewReader(data))
	if !uniqueJSON(scan, 0) {
		return nil, invalid
	}
	if _, err := scan.Token(); err != io.EOF {
		return nil, invalid
	}
	var doc struct {
		Version  string `json:"version"`
		Currency string `json:"currency"`
		AsOf     string `json:"as_of"`
		Rates    []rate `json:"rates"`
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if dec.Decode(&doc) != nil || !catalogueShape(data, doc.Version) || (doc.Version != "1" && doc.Version != "2") || !currencyPattern.MatchString(doc.Currency) || len(doc.Rates) > 10000 || doc.Rates == nil {
		return nil, invalid
	}
	date, err := time.Parse("2006-01-02", doc.AsOf)
	if err != nil || date.Format("2006-01-02") != doc.AsOf {
		return nil, invalid
	}
	c := &Catalogue{version: doc.Version, currency: doc.Currency, asOf: doc.AsOf, rates: map[string]int64{}}
	for _, r := range doc.Rates {
		keys, ok := fields[r.ResourceType]
		if doc.Version == "2" {
			_, ok = services[r.ResourceType]
			keys = selectorKeys(r.ResourceType)
		}
		if !ok || !regionPattern.MatchString(r.Region) || len(r.Region) > 64 || r.Selectors == nil || len(r.Selectors) != len(keys) || !validSelectors(r.ResourceType, r.Selectors) {
			return nil, invalid
		}
		text := r.HourlyPrice
		if doc.Version == "1" {
			if r.Component != "" || r.Unit != "" || r.UnitPrice != "" {
				return nil, invalid
			}
		} else {
			comp, ok := lookupComponent(r.ResourceType, r.Component)
			if !ok || r.Unit != comp.unit || r.HourlyPrice != "" {
				return nil, invalid
			}
			text = r.UnitPrice
		}
		micro, ok := price(text)
		if doc.Version == "2" {
			micro, ok = unitPrice(text)
		}
		if !ok {
			return nil, invalid
		}
		key := identity(r.ResourceType, r.Region, r.Selectors)
		if doc.Version == "2" {
			key += "|" + r.Component
		}
		if _, exists := c.rates[key]; exists {
			return nil, invalid
		}
		c.rates[key] = micro
	}
	return c, nil
}

func identity(typ, region string, selectors map[string]string) string {
	// JSON length-prefixes strings unambiguously, without exposing them.
	parts := []string{typ, region}
	for _, key := range selectorKeys(typ) {
		parts = append(parts, selectors[key])
	}
	b, _ := json.Marshal(parts)
	return string(b)
}

func price(s string) (int64, bool) {
	if len(s) > 20 || !decimalPattern.MatchString(s) {
		return 0, false
	}
	whole, fraction, _ := strings.Cut(s, ".")
	n, err := strconv.ParseInt(whole+fraction+strings.Repeat("0", 6-len(fraction)), 10, 64)
	return n, err == nil && n <= 1000000000000
}

// unitPrice retains sub-micro-unit request rates in integer pico-units.
func unitPrice(s string) (int64, bool) {
	if len(s) > 24 || !unitPricePattern.MatchString(s) {
		return 0, false
	}
	whole, fraction, _ := strings.Cut(s, ".")
	n, err := strconv.ParseInt(whole+fraction+strings.Repeat("0", 12-len(fraction)), 10, 64)
	return n, err == nil && n <= 1000000000000000000
}

func monthly(micro int64) (int64, bool) {
	// One currency unit = 1e6 micro-units = 100 cents; half-up once per resource.
	if micro < 0 || micro > 1000000000000 {
		return 0, false
	}
	n := (micro*730 + 5000) / 10000
	return n, n <= maxSafe
}

func catalogueShape(data []byte, version string) bool {
	doc, ok := object(data)
	if !ok || !exactKeys(doc, []string{"version", "currency", "as_of", "rates"}) {
		return false
	}
	var rates []json.RawMessage
	if json.Unmarshal(doc["rates"], &rates) != nil {
		return false
	}
	keys := []string{"resource_type", "region", "selectors", "hourly_price"}
	if version == "2" {
		keys = []string{"resource_type", "region", "selectors", "component", "unit", "unit_price"}
	}
	for _, raw := range rates {
		r, ok := object(raw)
		if !ok || !exactKeys(r, keys) {
			return false
		}
	}
	return true
}
func exactKeys(m map[string]json.RawMessage, keys []string) bool {
	if len(m) != len(keys) {
		return false
	}
	for _, key := range keys {
		raw, ok := m[key]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return false
		}
	}
	return true
}

func uniqueJSON(dec *json.Decoder, depth int) bool {
	if depth > 8 {
		return false
	}
	token, err := dec.Token()
	if err != nil {
		return false
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return true
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			key, err := dec.Token()
			s, ok := key.(string)

			if err != nil || !ok || seen[s] {
				return false
			}
			seen[s] = true
			if !uniqueJSON(dec, depth+1) {
				return false
			}
		}
	case '[':
		for dec.More() {
			if !uniqueJSON(dec, depth+1) {
				return false
			}
		}
	default:
		return false
	}
	end, err := dec.Token()
	return err == nil && ((delim == '{' && end == json.Delim('}')) || (delim == '[' && end == json.Delim(']')))
}
