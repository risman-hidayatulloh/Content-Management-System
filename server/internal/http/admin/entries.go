package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"fullstack-cms/internal/content"
)

func registerEntries(r *mux.Router, db *gorm.DB) {
	e := r.PathPrefix("/content/{type}").Subrouter()
	e.HandleFunc("/entries", entryList(db)).Methods(http.MethodGet)
	e.HandleFunc("/entries", entryCreate(db)).Methods(http.MethodPost)
	e.HandleFunc("/entries/{id}", entryGet(db)).Methods(http.MethodGet)
	e.HandleFunc("/entries/{id}", entryUpdate(db)).Methods(http.MethodPatch)
	e.HandleFunc("/entries/{id}", entryDelete(db)).Methods(http.MethodDelete)
	e.HandleFunc("/entries/{id}/publish", entryPublish(db)).Methods(http.MethodPost)
	e.HandleFunc("/entries/{id}/rollback", entryRollback(db)).Methods(http.MethodPost)
}

func findCTBySlug(db *gorm.DB, slug string) (content.ContentType, error) {
	var ct content.ContentType
	err := db.Preload("Fields", func(tx *gorm.DB) *gorm.DB {
		return tx.Order(clause.OrderByColumn{
			Column: clause.Column{Table: "field_defs", Name: "order"},
			Desc:   false,
		})
	}).First(&ct, "slug = ?", slug).Error
	return ct, err
}

func entryList(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctSlug := mux.Vars(r)["type"]
		fmt.Println("ctSlug:", ctSlug)
		ct, err := findCTBySlug(db, ctSlug)
		if err != nil {
			http.Error(w, "content type not found", 404)
			return
		}

		page, size := 1, 10
		if v := r.URL.Query().Get("page"); v != "" {
			fmt.Sscanf(v, "%d", &page)
		}
		if v := r.URL.Query().Get("pageSize"); v != "" {
			fmt.Sscanf(v, "%d", &size)
		}
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 10
		}

		var total int64
		db.Model(&content.Entry{}).Where("content_type_id = ?", ct.ID).Count(&total)

		var items []content.Entry
		db.Where("content_type_id = ?", ct.ID).
			Order(clause.OrderByColumn{Column: clause.Column{Name: "updated_at"}, Desc: true}).
			Offset((page - 1) * size).Limit(size).Find(&items)

		json.NewEncoder(w).Encode(map[string]any{
			"data": items, "meta": map[string]any{"page": page, "pageSize": size, "total": total},
		})
	}
}

func entryCreate(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctSlug := mux.Vars(r)["type"]
		ct, err := findCTBySlug(db, ctSlug)
		if err != nil {
			http.Error(w, "content type not found", 404)
			return
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad json", 400)
			return
		}

		normalized, err := content.ValidateAndNormalize(db, ct, payload)
		if err != nil {
			http.Error(w, err.Error(), 422)
			return
		}

		// slug mgmt: jika ada "title" → generate; bisa override jika ada "slug" di payload
		var slugStr *string
		if v, ok := normalized["slug"]; ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				ss := content.Slugify(s)
				slugStr = &ss
			}
		} else if title, ok := normalized["title"].(string); ok && title != "" {
			ss := content.Slugify(title)
			slugStr = &ss
		}

		b, _ := json.Marshal(normalized)
		entry := content.Entry{
			ContentTypeID: ct.ID,
			Slug:          slugStr,
			Status:        "draft",
			Data:          datatypes.JSON(b),
			Version:       1,
			CreatedBy:     currentUserID(r),
			UpdatedBy:     currentUserID(r),
		}
		if err := db.Create(&entry).Error; err != nil {
			http.Error(w, "db", 500)
			return
		}

		createVersion(db, entry.ID, 1, normalized, "create", currentUserID(r))
		logAudit(db, "entry.create", "entry:"+entry.ID, normalized, currentUserID(r))

		json.NewEncoder(w).Encode(entry)
	}
}

func entryGet(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		var e content.Entry
		if err := db.First(&e, "id = ?", id).Error; err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(e)
	}
}

func entryUpdate(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctSlug := mux.Vars(r)["type"]
		ct, err := findCTBySlug(db, ctSlug)
		if err != nil {
			http.Error(w, "content type not found", 404)
			return
		}

		id := mux.Vars(r)["id"]
		var e content.Entry
		if err := db.First(&e, "id = ?", id).Error; err != nil {
			http.Error(w, "not found", 404)
			return
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad json", 400)
			return
		}

		normalized, err := content.ValidateAndNormalize(db, ct, payload)
		if err != nil {
			http.Error(w, err.Error(), 422)
			return
		}

		// optional update slug
		if s, ok := normalized["slug"].(string); ok && s != "" {
			ss := content.Slugify(s)
			e.Slug = &ss
		}

		b, _ := json.Marshal(normalized)
		e.Data = datatypes.JSON(b)
		e.UpdatedBy = currentUserID(r)
		e.Version += 1

		if err := db.Save(&e).Error; err != nil {
			http.Error(w, "db", 500)
			return
		}
		createVersion(db, e.ID, e.Version, normalized, "update", currentUserID(r))
		logAudit(db, "entry.update", "entry:"+e.ID, normalized, currentUserID(r))

		json.NewEncoder(w).Encode(e)
	}
}

func entryDelete(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if err := db.Delete(&content.Entry{}, "id = ?", id).Error; err != nil {
			http.Error(w, "db", 500)
			return
		}
		logAudit(db, "entry.delete", "entry:"+id, nil, currentUserID(r))
		w.WriteHeader(204)
	}
}

func entryPublish(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		atStr := r.URL.Query().Get("at")

		// schedule sederhana: kalau at di masa depan → set ScheduledAt + goroutine
		if atStr != "" {
			when, err := time.Parse(time.RFC3339, atStr)
			if err != nil {
				http.Error(w, "bad time", 400)
				return
			}
			var e content.Entry
			if err := db.First(&e, "id = ?", id).Error; err != nil {
				http.Error(w, "not found", 404)
				return
			}
			e.ScheduledAt = &when
			if err := db.Save(&e).Error; err != nil {
				http.Error(w, "db", 500)
				return
			}

			go func(entryID string, t time.Time) {
				sleep := time.Until(t)
				if sleep > 0 {
					time.Sleep(sleep)
				}
				_ = publishNow(db, entryID, currentUserID(r)) // naive, diganti Asynq nanti
			}(e.ID, when)

			logAudit(db, "entry.schedule", "entry:"+e.ID, map[string]any{"at": when}, currentUserID(r))
			json.NewEncoder(w).Encode(map[string]any{"scheduled": when.UTC().Format(time.RFC3339)})
			return
		}

		// publish sekarang
		if err := publishNow(db, id, currentUserID(r)); err != nil {
			http.Error(w, "db", 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}
}

func publishNow(db *gorm.DB, id string, actor string) error {
	now := time.Now().UTC()
	return db.Model(&content.Entry{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": "published", "published_at": now, "updated_by": actor}).Error
}

func entryRollback(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		versionStr := r.URL.Query().Get("version")
		if versionStr == "" {
			http.Error(w, "version required", 400)
			return
		}
		ver, _ := strconv.Atoi(versionStr)

		var v content.EntryVersion
		if err := db.First(&v, "entry_id = ? AND version = ?", id, ver).Error; err != nil {
			http.Error(w, "version not found", 404)
			return
		}
		if err := db.Model(&content.Entry{}).
			Where("id = ?", id).
			Updates(map[string]any{"data": v.Data, "updated_by": currentUserID(r)}).Error; err != nil {
			http.Error(w, "db", 500)
			return
		}
		logAudit(db, "entry.rollback", "entry:"+id, map[string]any{"toVersion": ver}, currentUserID(r))
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}
}

// helpers versioning & audit
func createVersion(db *gorm.DB, entryID string, ver int, data map[string]any, summary string, actor string) {
	b, _ := json.Marshal(data)
	ev := content.EntryVersion{EntryID: entryID, Version: ver, Data: datatypes.JSON(b), ActorID: actor}
	if summary != "" {
		ev.Summary = &summary
	}
	_ = db.Create(&ev).Error
}
func logAudit(db *gorm.DB, action, target string, meta map[string]any, actor string) {
	var m datatypes.JSON
	if meta != nil {
		b, _ := json.Marshal(meta)
		m = datatypes.JSON(b)
	}
	_ = db.Create(&content.AuditLog{ActorID: &actor, Action: action, Target: target, Meta: m}).Error
}
