package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"gorm.io/gorm"
)

var (
	ErrValidation = errors.New("validation error")
)

func ValidateAndNormalize(db *gorm.DB, ct ContentType, payload map[string]any) (map[string]any, error) {
	out := map[string]any{}
	errs := map[string]string{}
	policy := bluemonday.UGCPolicy()

	for _, f := range ct.Fields {
		val, ok := payload[f.Key]
		if f.Required && (!ok || isEmpty(val)) {
			errs[f.Key] = "required"
			continue
		}
		if !ok {
			continue // field opsional tak ada → skip
		}

		switch f.Type {
		case "string":
			s, ok := toString(val)
			if !ok {
				errs[f.Key] = "must be string"
				continue
			}
			out[f.Key] = s

		case "markdown", "richtext":
			s, ok := toString(val)
			if !ok {
				errs[f.Key] = "must be string"
				continue
			}
			// sanitasi XSS untuk HTML (markdown nantinya dirender → tetap aman)
			out[f.Key] = policy.Sanitize(s)

		case "number":
			n, ok := toNumber(val)
			if !ok {
				errs[f.Key] = "must be number"
				continue
			}
			out[f.Key] = n

		case "boolean":
			b, ok := toBool(val)
			if !ok {
				errs[f.Key] = "must be boolean"
				continue
			}
			out[f.Key] = b

		case "datetime":
			s, ok := toString(val)
			if !ok {
				errs[f.Key] = "must be RFC3339 string"
				continue
			}
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				errs[f.Key] = "invalid RFC3339"
				continue
			}
			out[f.Key] = t.UTC().Format(time.RFC3339)

		case "relation":
			// expect string id
			id, ok := toString(val)
			if !ok {
				errs[f.Key] = "must be id string"
				continue
			}
			if f.RelationTo == nil || *f.RelationTo == "" {
				errs[f.Key] = "relationTo missing"
				continue
			}
			switch *f.RelationTo {
			case "media":
				var m Media
				if err := db.First(&m, "id = ?", id).Error; err != nil {
					errs[f.Key] = "related media not found"
					continue
				}
			default:
				// asumsikan slug ContentType lain
				var ctid string
				if err := db.Model(&ContentType{}).Where("slug = ?", *f.RelationTo).Pluck("id", &ctid).Error; err != nil || ctid == "" {
					errs[f.Key] = "related content type not found"
					continue
				}
				var rel Entry
				if err := db.First(&rel, "id = ? AND content_type_id = ?", id, ctid).Error; err != nil {
					errs[f.Key] = "related entry not found"
					continue
				}
			}
			out[f.Key] = id

		default:
			errs[f.Key] = "unsupported type"
		}

		// cek unique
		if f.IsUnique {
			var count int64
			// JSONB query: (data->>'key') = value AND content_type_id = ?
			db.Model(&Entry{}).
				Where("content_type_id = ? AND (data->>?) = ?", ct.ID, f.Key, fmt.Sprintf("%v", out[f.Key])).
				Count(&count)
			if count > 0 {
				errs[f.Key] = "must be unique"
			}
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrValidation, errs)
	}
	return out, nil
}

func isEmpty(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return t == ""
	default:
		return false
	}
}

func toString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case fmt.Stringer:
		return x.String(), true
	default:
		return "", false
	}
}
func toNumber(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
func toBool(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	default:
		return false, false
	}
}
