package pricing

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/danushkastanley/tfviz/internal/input"
)

var invalidUsage = errors.New("invalid usage data")

// LoadUsage reads only a bounded local regular file; errors contain no user input.
func (c *Catalogue) LoadUsage(path string, resources []input.Resource) error {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBytes {
		return errors.New("cannot read usage data: expected a bounded local regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return errors.New("cannot read usage data")
	}
	defer f.Close()
	return c.ReadUsage(f, resources)
}

func object(raw json.RawMessage) (map[string]json.RawMessage, bool) {
	var m map[string]json.RawMessage
	err := json.Unmarshal(raw, &m)
	return m, err == nil && m != nil
}

// ReadUsage validates every entry before installing any usage. Quantities are
// monthly micro-units, not copied between before and after.
func (c *Catalogue) ReadUsage(r io.Reader, resources []input.Resource) error {
	if c.version != "2" {
		return errors.New("usage data requires a version 2 pricing catalogue")
	}
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil || len(data) > maxBytes {
		return invalidUsage
	}
	scan := json.NewDecoder(bytes.NewReader(data))
	if !uniqueJSON(scan, 0) {
		return invalidUsage
	}
	if _, err := scan.Token(); err != io.EOF {
		return invalidUsage
	}
	doc, ok := object(data)
	if !ok || len(doc) != 2 || string(doc["version"]) != `"1"` {
		return invalidUsage
	}
	entries, ok := object(doc["resources"])
	if !ok || len(entries) > 10000 {
		return invalidUsage
	}
	types := map[string]string{}
	for _, r := range resources {
		if _, ok := services[r.Type]; ok && r.Mode == "managed" && r.Provider == "hashicorp/aws" && !r.Importing {
			types[r.Address] = r.Type
		}
	}
	usage := map[string]map[string]map[string]int64{}
	for address, raw := range entries {
		typ, ok := types[address]
		if !ok {
			return invalidUsage
		}
		sides, ok := object(raw)
		if !ok || len(sides) == 0 || len(sides) > 2 {
			return invalidUsage
		}
		usage[address] = map[string]map[string]int64{}
		for side, raw := range sides {
			if side != "before" && side != "after" {
				return invalidUsage
			}
			metrics, ok := object(raw)
			if !ok {
				return invalidUsage
			}
			usage[address][side] = map[string]int64{}
			for name, raw := range metrics {
				comp, ok := lookupComponent(typ, name)
				if !ok || comp.name == services[typ].base.name {
					return invalidUsage
				}
				var text string
				if json.Unmarshal(raw, &text) != nil || len(raw) == 0 || raw[0] != '"' {
					return invalidUsage
				}
				n, ok := quantity(text)
				if !ok {
					return invalidUsage
				}
				usage[address][side][name] = n
			}
		}
	}
	c.usage = usage
	return nil
}
func quantity(s string) (int64, bool) {
	if len(s) > 24 || !decimalPattern.MatchString(s) {
		return 0, false
	}
	// Same six-place fixed point as prices, but up to one billion monthly units.
	whole, fraction, _ := strings.Cut(s, ".")
	n, err := strconv.ParseInt(whole+fraction+strings.Repeat("0", 6-len(fraction)), 10, 64)
	return n, err == nil && n <= 1000000000000000
}
func fmtInteger(n int64) string { return strconv.FormatInt(n, 10) }
func validMemory(s string) bool {
	n, e := strconv.ParseInt(s, 10, 64)
	return e == nil && n >= 128 && n <= 10240 && fmtInteger(n) == s
}

// Products can exceed int64 even when final cents are safe. Stdlib big.Int
// keeps exact fixed-point arithmetic and rounds only once per resource.
func addProduct(total *big.Int, price, quantity int64) {
	total.Add(total, new(big.Int).Mul(big.NewInt(price), big.NewInt(quantity)))
}
func cents(total *big.Int) (*int64, bool) {
	// Pico-unit prices × micro-unit quantities become cents at a 1e16 divisor.
	n := new(big.Int).Add(total, big.NewInt(5000000000000000))
	n.Quo(n, big.NewInt(10000000000000000))
	if !n.IsInt64() || n.Sign() < 0 || n.Int64() > maxSafe {
		return nil, false
	}
	return ptr(n.Int64()), true
}
