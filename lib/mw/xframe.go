package mw

import (
	"net/http"

	"github.com/digitalcircle-com-br/service"
)

func XFrame(xframe string, next http.HandlerFunc) http.HandlerFunc {

	if xframe == "" || xframe == "*" {
		xframe = "SAMEORIGIN"
	}
	service.Log("Setting XFrame: %s", xframe)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		(w).Header().Set("X-Frame-Options:", xframe)

		next.ServeHTTP(w, r)
	})
}
