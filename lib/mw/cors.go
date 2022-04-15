package mw

import (
	"net/http"
	"strings"

	"github.com/digitalcircle-com-br/service"
)

func init() {
	service.Log("Initiating CORS")
	corsmode := "*"
	if corsmode == "*" {
		corsmode = "SAME"
	}
}

var corsmode string

func CORS(next http.HandlerFunc) http.HandlerFunc {
	corsmode = "SAME"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if next == nil {
			service.Log("Empty handler at %s", r.URL.Path)
			return
		}
		if corsmode == "SAME" {
			orig := r.Header.Get("Origin")
			if orig == "" {
				orig = strings.Split(r.Host, ":")[0]
			}
			service.Log("Using Origin (SAME): %s", orig)
			(w).Header().Set("Access-Control-Allow-Origin", orig)
		} else {
			service.Log("Using Origin (FIXED): %s", corsmode)
			(w).Header().Set("Access-Control-Allow-Origin", corsmode)
		}

		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Last-Modified, Expires, Accept, Cache-Control, Content-Type, Content-Language,Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Pragma")
		(w).Header().Set("Access-Control-Allow-Credentials", "true")
		if (*r).Method == http.MethodOptions {
			service.Log("Sending pre-flight cors: %+v", w.Header())
			return
		}

		next.ServeHTTP(w, r)
	})
}
