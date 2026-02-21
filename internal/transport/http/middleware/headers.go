package middleware

import (
	"net/http"
	"strings"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

func ValidateHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptHeader := r.Header.Get("Accept")
		if acceptHeader != "" {
			lowerCase := strings.ToLower(acceptHeader)

			valid := strings.Contains(lowerCase, "application/json") ||
				strings.Contains(lowerCase, "*/*") ||
				strings.Contains(lowerCase, "text/javascript")

			if !valid {
				log.Error.Println("Request contains invalid accept header format", "Header", acceptHeader)
				apiErr := apierror.Build(apierror.StatusNotAcceptedCode, "Using invalid accept header format")
				apiresponse.Send(w, http.StatusNotAcceptable, apiErr)
				return
			}

		}
		next.ServeHTTP(w, r)
	})
}
