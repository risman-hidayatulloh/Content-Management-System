package admin

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"fullstack-cms/config"
	"fullstack-cms/internal/content"
)

func uploadMedia(db *gorm.DB, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.App.MaxUploadBytes)
		file, hdr, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "no file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		id := uuid.NewString()
		ext := filepath.Ext(hdr.Filename)
		base := id + ext
		dst := filepath.Join(cfg.App.MediaDir, base)

		out, err := createFile(dst)
		if err != nil {
			http.Error(w, "save error", 500)
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, "write error", 500)
			return
		}

		variants := map[string]string{}
		if img, err := imaging.Open(dst); err == nil {
			thumb := imaging.Thumbnail(img, 320, 320, imaging.Lanczos)
			thumbName := id + "_thumb" + ext
			thumbPath := filepath.Join(cfg.App.MediaDir, thumbName)
			_ = imaging.Save(thumb, thumbPath)
			variants["thumb"] = path.Join("/media", thumbName)
		}

		b, _ := json.Marshal(variants)
		m := content.Media{
			Filename:   hdr.Filename,
			Mime:       hdr.Header.Get("Content-Type"),
			Size:       hdr.Size,
			Variants:   datatypes.JSON(b),
			URL:        path.Join("/media", base),
			StorageKey: base,
			CreatedBy:  "system",
		}
		if err := db.Create(&m).Error; err != nil {
			http.Error(w, "db", 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

// small helper (could add mkdir -p)
func createFile(path string) (*os.File, error) {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	return os.Create(path)
}
