package pricing

import (
	"strings"
	"testing"
)

const synthetic = `{"version":"1","currency":"USD","as_of":"2026-01-01","rates":[{"resource_type":"aws_instance","region":"us-east-1","selectors":{"instance_type":"t3.micro","ami":"ami-synthetic","tenancy":"default"},"hourly_price":"0.0005"}]}`

func TestCatalogueAndRounding(t *testing.T) {
	c, err := Read(strings.NewReader(synthetic))
	if err != nil {
		t.Fatal(err)
	}
	if c.currency != "USD" {
		t.Fatal("currency")
	}
	if got, ok := monthly(500); !ok || got != 37 {
		t.Fatalf("round half up = %d, %v", got, ok)
	}
}
