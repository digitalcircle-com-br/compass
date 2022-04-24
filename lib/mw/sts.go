package mw

import (
	"net/http"

	"github.com/digitalcircle-com-br/service"
)

func STS(sts string, next http.HandlerFunc) http.HandlerFunc {

	if sts == "" || sts == "*" {
		sts = "max-age=31536000; includeSubDomains"
	}
	service.Log("Setting STS: %s", sts)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		(w).Header().Set("Strict-Transport-Security", sts)

		next.ServeHTTP(w, r)
	})
}
