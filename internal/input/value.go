package input

import (
	"encoding/json"
	"sort"
	"strconv"
)

// Kind classifies a Value.
type Kind uint8

const (
	KindNull Kind = iota
	KindString
	KindNumber
	KindBool
	KindList
	KindMap
	// KindUnknown is a value known only after apply.
	KindUnknown
	// KindSensitive is a value the producer marked sensitive. It carries no
	// payload: the reader discards the raw value when it builds the tree.
	KindSensitive
)

// Comparison is the reader's verdict on whether a sensitive value changed.
// It is established from the raw values before they are discarded.
type Comparison uint8

const (
	ComparisonUnavailable Comparison = iota
	ComparisonUnchanged
	ComparisonChanged
)

// Value is a producer attribute value with sensitivity and unknown markers
// applied. Downstream code can only read known, non-sensitive payloads.
type Value struct {
	kind       Kind
	text       string // string or number text
	boolean    bool
	items      []Value
	fields     map[string]Value
	comparison Comparison
}

// Absent is the zero Value: no value was recorded.
var Absent = Value{kind: KindNull}

func (v Value) Kind() Kind { return v.kind }

// Comparison reports whether a sensitive value changed. It is meaningful
// only for KindSensitive.
func (v Value) Comparison() Comparison { return v.comparison }

// Field returns the attribute with the given name, or Absent. Fields of a
// sensitive or unknown object inherit its kind, so a lookup can never
// expose what a marker hides.
func (v Value) Field(name string) Value {
	switch v.kind {
	case KindSensitive, KindUnknown:
		return v
	case KindMap:
		if f, ok := v.fields[name]; ok {
			return f
		}
	}
	return Absent
}

// Path follows nested fields, treating a single-element block list as its
// element (Terraform encodes nested blocks as lists).
func (v Value) Path(names ...string) Value {
	cur := v
	for _, name := range names {
		if cur.kind == KindList && len(cur.items) == 1 {
			cur = cur.items[0]
		}
		cur = cur.Field(name)
	}
	return cur
}

// Items returns list elements; sensitive or unknown lists yield nothing.
func (v Value) Items() []Value {
	if v.kind != KindList {
		return nil
	}
	return v.items
}

// Keys returns the sorted field names of a known map or object.
func (v Value) Keys() []string {
	if v.kind != KindMap {
		return nil
	}
	keys := make([]string, 0, len(v.fields))
	for k := range v.fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// String returns a known string payload.
func (v Value) String() (string, bool) {
	if v.kind != KindString {
		return "", false
	}
	return v.text, true
}

// Number returns a known number payload.
func (v Value) Number() (float64, bool) {
	if v.kind != KindNumber {
		return 0, false
	}
	n, err := strconv.ParseFloat(v.text, 64)
	return n, err == nil
}

// Bool returns a known boolean payload.
func (v Value) Bool() (bool, bool) {
	if v.kind != KindBool {
		return false, false
	}
	return v.boolean, true
}

// Strings returns the string elements of a known list, skipping elements
// that are unknown, sensitive or not strings. The second result is false if
// any element was skipped.
func (v Value) Strings() ([]string, bool) {
	if v.kind != KindList {
		return nil, false
	}
	out := make([]string, 0, len(v.items))
	complete := true
	for _, item := range v.items {
		if s, ok := item.String(); ok {
			out = append(out, s)
		} else {
			complete = false
		}
	}
	return out, complete
}

// Contains reports whether any part of the value has the given kind.
func (v Value) Contains(kind Kind) bool {
	if v.kind == kind {
		return true
	}
	for _, item := range v.items {
		if item.Contains(kind) {
			return true
		}
	}
	for _, field := range v.fields {
		if field.Contains(kind) {
			return true
		}
	}
	return false
}

// buildValue converts decoded JSON (decoded with UseNumber) into a Value,
// replacing every subtree marked by sensitive or unknown. When other is
// supplied, sensitive leaves record whether raw differs from other.
func buildValue(raw, sensitive, unknown any, other *any, otherUnknown any) Value {
	if marked(sensitive) || ((marked(unknown) || markerShapeMismatch(raw, sensitive) || markerShapeMismatch(raw, unknown)) && containsMark(sensitive)) {
		comparison := ComparisonUnavailable
		if !containsMark(unknown) && !markerShapeMismatch(raw, sensitive) {
			comparison = compare(raw, other, otherUnknown)
		}
		return Value{kind: KindSensitive, comparison: comparison}
	}
	if marked(unknown) || (markerShapeMismatch(raw, unknown) && containsMark(unknown)) {
		return Value{kind: KindUnknown}
	}
	switch r := raw.(type) {
	case nil:
		// Omitted sensitive containers still carry evidence, never payloads.
		if containsMark(sensitive) {
			return Value{kind: KindSensitive}
		}
		// A structure omitted from the value but partly marked unknown is
		// rebuilt from its markers rather than reported as absent.
		switch marks := unknown.(type) {
		case map[string]any:
			if containsMark(marks) {
				return buildValue(map[string]any{}, sensitive, marks, other, otherUnknown)
			}
		case []any:
			if containsMark(marks) {
				return buildValue(make([]any, len(marks)), sensitive, marks, other, otherUnknown)
			}
		}
		return Absent
	case string:
		return Value{kind: KindString, text: r}
	case json.Number:
		return Value{kind: KindNumber, text: r.String()}
	case bool:
		return Value{kind: KindBool, boolean: r}
	case []any:
		// Producers can omit list elements that still carry sensitivity/unknown
		// markers. Preserve that evidence rather than claiming an empty list.
		sensitiveItems, _ := sensitive.([]any)
		unknownItems, _ := unknown.([]any)
		items := make([]Value, max(len(r), len(sensitiveItems), len(unknownItems)))
		for i := range items {
			var rawItem any
			if i < len(r) {
				rawItem = r[i]
			}
			items[i] = buildValue(rawItem, index(sensitive, i), index(unknown, i), indexOther(other, i), index(otherUnknown, i))
		}
		return Value{kind: KindList, items: items}
	case map[string]any:
		fields := make(map[string]Value, len(r))
		for k, item := range r {
			fields[k] = buildValue(item, field(sensitive, k), field(unknown, k), fieldOther(other, k), field(otherUnknown, k))
		}
		// Preserve sensitivity evidence even when a marked attribute is omitted.
		if marks, ok := sensitive.(map[string]any); ok {
			for k, mark := range marks {
				if _, present := r[k]; !present && containsMark(mark) {
					fields[k] = buildValue(nil, mark, field(unknown, k), fieldOther(other, k), field(otherUnknown, k))
				}
			}
		}
		// Producers omit unknown attributes from the value itself and list
		// them only in the unknown markers; they must read as unknown, not absent.
		if marks, ok := unknown.(map[string]any); ok {
			for k, mark := range marks {
				if _, present := r[k]; !present && containsMark(mark) {
					fields[k] = buildValue(nil, field(sensitive, k), mark, fieldOther(other, k), field(otherUnknown, k))
				}
			}
		}
		return Value{kind: KindMap, fields: fields}
	default:
		return Absent
	}
}

// compare establishes a sensitive value's change only when both sides are
// recorded and the other side is known.
func compare(raw any, other *any, otherUnknown any) Comparison {
	if other == nil || marked(otherUnknown) || containsMark(otherUnknown) {
		return ComparisonUnavailable
	}
	if equalJSON(raw, *other) {
		return ComparisonUnchanged
	}
	return ComparisonChanged
}

// Mismatched producer marker shapes fail closed instead of losing evidence.
func markerShapeMismatch(raw, marker any) bool {
	if raw == nil {
		return false // omitted values are rebuilt from their markers
	}
	switch marker.(type) {
	case map[string]any:
		_, ok := raw.(map[string]any)
		return !ok
	case []any:
		_, ok := raw.([]any)
		return !ok
	}
	return false
}

func marked(m any) bool {
	b, ok := m.(bool)
	return ok && b
}

func containsMark(m any) bool {
	switch t := m.(type) {
	case bool:
		return t
	case []any:
		for _, v := range t {
			if containsMark(v) {
				return true
			}
		}
	case map[string]any:
		for _, v := range t {
			if containsMark(v) {
				return true
			}
		}
	}
	return false
}

func index(m any, i int) any {
	if l, ok := m.([]any); ok {
		if i < len(l) {
			return l[i]
		}
		return nil
	}
	return m // a whole-list marker applies to every element
}

func field(m any, k string) any {
	if o, ok := m.(map[string]any); ok {
		return o[k]
	}
	return m
}

func indexOther(other *any, i int) *any {
	if other == nil {
		return nil
	}
	if l, ok := (*other).([]any); ok && i < len(l) {
		return &l[i]
	}
	return nil
}

func fieldOther(other *any, k string) *any {
	if other == nil {
		return nil
	}
	if o, ok := (*other).(map[string]any); ok {
		if v, present := o[k]; present {
			return &v
		}
	}
	return nil
}
