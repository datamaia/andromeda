package config

import "strconv"

// flatten converts a nested map (as parsed from TOML) into a flat map keyed by dotted paths.
// Nested tables become dotted keys (a.b.c); arrays are kept as whole values under their key.
// A value that is already flat (a non-map) is copied through.
func flatten(nested map[string]any) map[string]any {
	out := map[string]any{}
	flattenInto("", nested, out)
	return out
}

func flattenInto(prefix string, v map[string]any, out map[string]any) {
	for k, val := range v {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch child := val.(type) {
		case map[string]any:
			flattenInto(key, child, out)
		default:
			out[key] = val
		}
	}
}

// unflatten is the inverse of flatten: it rebuilds a nested map from dotted keys. Used when
// re-emitting configuration (e.g. writing a resolved profile). Numeric-looking segments are
// kept as map keys (configuration tables, not arrays).
func unflatten(flat map[string]any) map[string]any {
	root := map[string]any{}
	for k, v := range flat {
		segs := splitDotted(k)
		cur := root
		for i, s := range segs {
			if i == len(segs)-1 {
				cur[s] = v
				continue
			}
			next, ok := cur[s].(map[string]any)
			if !ok {
				next = map[string]any{}
				cur[s] = next
			}
			cur = next
		}
	}
	return root
}

func splitDotted(k string) []string {
	var out []string
	start := 0
	for i := 0; i < len(k); i++ {
		if k[i] == '.' {
			out = append(out, k[start:i])
			start = i + 1
		}
	}
	out = append(out, k[start:])
	return out
}

// asString renders a scalar config value as a string (helper for typed accessors added later).
func asString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case bool:
		return strconv.FormatBool(t), true
	case int64:
		return strconv.FormatInt(t, 10), true
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64), true
	default:
		return "", false
	}
}
