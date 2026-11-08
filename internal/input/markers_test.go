package input

import "testing"

func TestSensitivityWinsOverUnknownAndOmittedContainers(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		raw, sensitive, unknown any
	}{
		{"scalar with object sensitivity", "known", map[string]any{"nested": true}, nil},
		{"list with object sensitivity", []any{}, map[string]any{"nested": true}, nil},
		{"unknown sensitive leaf", nil, true, true},
		{"unknown sensitive subtree", nil, map[string]any{"arn": true}, true},
		{"omitted sensitive field", map[string]any{}, map[string]any{"arn": true}, map[string]any{"arn": true}},
		{"empty sensitive list", []any{}, []any{map[string]any{"market_type": true}}, nil},
		{"partly omitted sensitive list", []any{"known"}, []any{false, true}, []any{false, true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			value := buildValue(tt.raw, tt.sensitive, tt.unknown, nil, nil)
			if !value.Contains(KindSensitive) {
				t.Fatal("sensitivity marker lost")
			}
			if value.Comparison() != ComparisonUnavailable {
				t.Fatal("unknown comparison claimed")
			}
		})
	}
}

func TestOmittedUnknownListItemsRemainUnknown(t *testing.T) {
	for _, raw := range []any{[]any{}, "known"} {
		value := buildValue(raw, nil, []any{map[string]any{"market_type": true}}, nil, nil)
		if !value.Contains(KindUnknown) {
			t.Fatal("unknown marker lost")
		}
	}
}
