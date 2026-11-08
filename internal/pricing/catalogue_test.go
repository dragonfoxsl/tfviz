package pricing

import (
	"strings"
	"testing"
)

func TestRejectMalformedCatalogue(t *testing.T) {
	for name, text := range map[string]string{
		"trailing":            synthetic + ` {}`,
		"unknown":             strings.Replace(synthetic, `"version":"1"`, `"version":"1","extra":true`, 1),
		"rate unknown":        strings.Replace(synthetic, `"hourly_price":`, `"extra":true,"hourly_price":`, 1),
		"duplicate key":       strings.Replace(synthetic, `"version":"1"`, `"version":"1","version":"1"`, 1),
		"duplicate selector":  strings.Replace(synthetic, `"tenancy":"default"`, `"tenancy":"default","tenancy":"default"`, 1),
		"version":             strings.Replace(synthetic, `"version":"1"`, `"version":"2"`, 1),
		"currency":            strings.Replace(synthetic, `"USD"`, `"usd"`, 1),
		"date":                strings.Replace(synthetic, `"2026-01-01"`, `"2026-02-30"`, 1),
		"date width":          strings.Replace(synthetic, `"2026-01-01"`, `"2026-1-1"`, 1),
		"no rates":            strings.Replace(synthetic, `"rates":[`, `"rates":null,"ignored":[`, 1),
		"missing selector":    strings.Replace(synthetic, `"tenancy":"default"`, `"other":"default"`, 1),
		"extra selector":      strings.Replace(synthetic, `"tenancy":"default"`, `"tenancy":"default","os":"linux"`, 1),
		"unsupported tenancy": strings.Replace(synthetic, `"default"`, `"host"`, 1),
		"unsupported type":    strings.Replace(synthetic, `"aws_instance"`, `"aws_s3_bucket"`, 1),
		"bad region":          strings.Replace(synthetic, `"us-east-1"`, `"canary-bad-region"`, 1),
		"oversize":            strings.Repeat(" ", maxBytes+1),
		"depth":               strings.Repeat("[", 10) + strings.Repeat("]", 10),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Read(strings.NewReader(text)); err == nil || err.Error() != "invalid pricing catalogue" {
				t.Fatalf("error = %v", err)
			}
		})
	}
	for _, text := range []string{"-1", "NaN", "Infinity", "1e-3", "+1", "01", "1.", ".1", "0.1234567", "1000000.000001", "9999999999999999999999"} {
		if _, err := Read(strings.NewReader(strings.Replace(synthetic, "0.0005", text, 1))); err == nil {
			t.Errorf("accepted %q", text)
		}
	}
	repeated := strings.Replace(synthetic, `}]}`, `},`+strings.Split(synthetic, `"rates":[`)[1], 1)
	if _, err := Read(strings.NewReader(repeated)); err == nil {
		t.Fatal("duplicate rate")
	}
	for _, tc := range []struct {
		price string
		cents int64
	}{
		{"0", 0}, {"0.000001", 0}, {"0.000499", 36}, {"0.000500", 37}, {"1.234567", 90123}, {"1000000", 73000000000},
	} {
		micro, ok := price(tc.price)
		if !ok {
			t.Fatal(tc.price)
		}
		got, ok := monthly(micro)
		if !ok || got != tc.cents {
			t.Errorf("%s = %d, want %d", tc.price, got, tc.cents)
		}
	}
	if _, ok := monthly(-1); ok {
		t.Fatal("negative")
	}
	if _, ok := monthly(1000000000001); ok {
		t.Fatal("overflow")
	}
}
