package mw

import (
	"net/http"

	"github.com/digitalcircle-com-br/gw/lib/base"
)

func XFrame(xframe string, next http.HandlerFunc) http.HandlerFunc {

	if xframe == "" || xframe == "*" {
		xframe = "SAMEORIGIN"
	}
	base.Log("Setting XFrame: %s", xframe)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		(w).Header().Set("X-Frame-Options:", xframe)

		next.ServeHTTP(w, r)
	})
}
