package public

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.uber.org/zap"

	"fullstack-cms/config"
	"fullstack-cms/internal/content"
	"fullstack-cms/internal/security"
)

func Register(r *mux.Router, db *gorm.DB, _ *zap.Logger, _ config.Config) {
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/{type}", list(db)).Methods(http.MethodGet)
	api.HandleFunc("/{type}/{slug}", detail(db)).Methods(http.MethodGet)
}

func list(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := mux.Vars(r)["type"]
		page, size := getPage(r)
		sortField, desc := getSort(r, "published_at", true)

		q := db.Model(&content.Entry{}).
			Where("status = 'published' AND content_type_id = (SELECT id FROM content_types WHERE slug = ?)", ct)

		if val := r.URL.Query().Get("filter[title][contains]"); val != "" {
			q = q.Where("(data->>'title') ILIKE ?", "%"+val+"%")
		}

		var total int64
		q.Count(&total)

		var items []content.Entry
		q.Order(clause.OrderByColumn{Column: clause.Column{Name: sortField}, Desc: desc}).
			Offset((page - 1) * size).Limit(size).Find(&items)

		type out struct {
			ID   string                 `json:"id"`
			Slug *string                `json:"slug"`
			Data map[string]interface{} `json:"data"`
		}
		resp := make([]out, 0, len(items))
		for _, e := range items {
			var m map[string]interface{}
			_ = json.Unmarshal(e.Data, &m)
			resp = append(resp, out{ID: e.ID, Slug: e.Slug, Data: m})
		}

		etagSrc := fmt.Sprintf("%d:%s", total, maxUpdatedAt(items).Format(time.RFC3339Nano))
		security.WriteJSONWithETag(w, r, 200, map[string]any{
			"data": resp, "meta": map[string]any{"page": page, "pageSize": size, "total": total},
		}, etagSrc)
	}
}

func detail(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := mux.Vars(r)["type"]
		slug := mux.Vars(r)["slug"]

		var e content.Entry
		if err := db.Where("status='published' AND slug = ? AND content_type_id = (SELECT id FROM content_types WHERE slug = ?)", slug, ct).First(&e).Error; err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var m map[string]interface{}
		_ = json.Unmarshal(e.Data, &m)
		security.WriteJSONWithETag(w, r, 200, map[string]any{"id": e.ID, "slug": e.Slug, "data": m}, e.UpdatedAt.Format(time.RFC3339Nano))
	}
}

func getPage(r *http.Request) (int, int) {
	p := 1
	s := 10
	if v := r.URL.Query().Get("page"); v != "" {
		fmt.Sscanf(v, "%d", &p)
	}
	if v := r.URL.Query().Get("pageSize"); v != "" {
		fmt.Sscanf(v, "%d", &s)
	}
	if p < 1 {
		p = 1
	}
	if s < 1 || s > 100 {
		s = 10
	}
	return p, s
}
func getSort(r *http.Request, def string, defDesc bool) (string, bool) {
	sf := def
	desc := defDesc
	if v := r.URL.Query().Get("sort"); v != "" {
		var dir string
		fmt.Sscanf(v, "%[^:]:%s", &sf, &dir)
		desc = (dir == "desc")
	}
	return sf, desc
}
func maxUpdatedAt(entries []content.Entry) time.Time {
	var max time.Time
	for _, e := range entries {
		if e.UpdatedAt.After(max) {
			max = e.UpdatedAt
		}
	}
	return max
}
