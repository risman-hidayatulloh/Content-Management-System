package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"fullstack-cms/internal/content"
)

func registerContentTypes(r *mux.Router, db *gorm.DB) {
	r.HandleFunc("/content-types", ctCreate(db)).Methods(http.MethodPost)
	r.HandleFunc("/content-types", ctList(db)).Methods(http.MethodGet)
	r.HandleFunc("/content-types/{id}", ctGet(db)).Methods(http.MethodGet)
	r.HandleFunc("/content-types/{id}", ctUpdate(db)).Methods(http.MethodPatch)
	r.HandleFunc("/content-types/{id}", ctDelete(db)).Methods(http.MethodDelete)
}

type fieldReq struct {
	Name       string          `json:"name"`
	Key        string          `json:"key"`
	Type       string          `json:"type"`
	Required   bool            `json:"required"`
	IsUnique   bool            `json:"isUnique"`
	IsRelation bool            `json:"isRelation"`
	RelationTo *string         `json:"relationTo"`
	Config     json.RawMessage `json:"config"`
	Order      int             `json:"order"`
}

func ctCreate(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name   string     `json:"name"`
			Slug   string     `json:"slug"`
			Fields []fieldReq `json:"fields"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", 400)
			return
		}
		if req.Name == "" || req.Slug == "" {
			http.Error(w, "name/slug required", 422)
			return
		}

		ct := content.ContentType{Name: req.Name, Slug: req.Slug}
		for _, f := range req.Fields {
			ct.Fields = append(ct.Fields, content.FieldDef{
				Name: f.Name, Key: f.Key, Type: f.Type, Required: f.Required,
				IsUnique: f.IsUnique, IsRelation: f.IsRelation, RelationTo: f.RelationTo,
				Config: datatypes.JSON(f.Config), Order: f.Order,
			})
		}
		if err := db.Create(&ct).Error; err != nil {
			http.Error(w, "db", 500)
			return
		}
		json.NewEncoder(w).Encode(ct)
	}
}

func ctList(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var items []content.ContentType
		db.Preload("Fields", func(tx *gorm.DB) *gorm.DB { return tx.Order("\"order\" asc") }).Find(&items)
		json.NewEncoder(w).Encode(items)
	}
}

func ctGet(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		var ct content.ContentType
		if err := db.Preload("Fields", func(tx *gorm.DB) *gorm.DB { return tx.Order("\"order\" asc") }).First(&ct, "id = ?", id).Error; err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(ct)
	}
}

func ctUpdate(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		var req struct {
			Name   *string     `json:"name"`
			Slug   *string     `json:"slug"`
			Fields *[]fieldReq `json:"fields"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", 400)
			return
		}
		var ct content.ContentType
		if err := db.Preload("Fields").First(&ct, "id = ?", id).Error; err != nil {
			http.Error(w, "not found", 404)
			return
		}
		if req.Name != nil {
			ct.Name = *req.Name
		}
		if req.Slug != nil {
			ct.Slug = *req.Slug
		}
		if req.Fields != nil {
			// replace all fields (simple approach for take-home)
			if err := db.Where("content_type_id = ?", ct.ID).Delete(&content.FieldDef{}).Error; err != nil {
				http.Error(w, "db", 500)
				return
			}
			fields := make([]content.FieldDef, 0, len(*req.Fields))
			for _, f := range *req.Fields {
				fields = append(fields, content.FieldDef{
					ContentTypeID: ct.ID, Name: f.Name, Key: f.Key, Type: f.Type,
					Required: f.Required, IsUnique: f.IsUnique, IsRelation: f.IsRelation,
					RelationTo: f.RelationTo, Config: datatypes.JSON(f.Config), Order: f.Order,
				})
			}
			ct.Fields = fields
		}
		if err := db.Save(&ct).Error; err != nil {
			http.Error(w, "db", 500)
			return
		}
		json.NewEncoder(w).Encode(ct)
	}
}

func ctDelete(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		// soft delete entries by GORM automatically when deleting content type? We explicit deny if entries exist.
		var n int64
		db.Model(&content.Entry{}).Where("content_type_id = ?", id).Count(&n)
		if n > 0 {
			http.Error(w, "cannot delete: entries exist", 409)
			return
		}
		if err := db.Delete(&content.ContentType{}, "id = ?", id).Error; err != nil {
			http.Error(w, "db", 500)
			return
		}
		w.WriteHeader(204)
	}
}
