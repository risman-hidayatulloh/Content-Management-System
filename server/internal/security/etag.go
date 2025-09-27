package security

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

func WriteJSONWithETag(w http.ResponseWriter, r *http.Request, status int, payload any, etagSrc string) {
	if match := r.Header.Get("If-None-Match"); match != "" {
		// weak compare is fine (we compute fresh anyway)
	}
	body, _ := json.Marshal(payload)
	sum := sha1.Sum([]byte(etagSrc))
	etag := `W/"` + hex.EncodeToString(sum[:]) + `"`

	if inm := r.Header.Get("If-None-Match"); inm == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=60, s-maxage=600, stale-while-revalidate=120")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(body)
}
