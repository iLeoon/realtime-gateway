package apiresponse

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

// Send writes a JSON response with safety buffering.
// If encoding fails, it automatically logs the error and sends a 500.
func Send(w http.ResponseWriter, status int, data interface{}) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		logger.Error("Failed to encode response", "Error", err)
		apiErr := apierror.Build(apierror.InternalServerErrorCode,
			"unexpected error occured while trying to process response",
			apierror.WithTarget("response"),
			apierror.WithInnerError("FailedToEncodeRespBody"),
		)

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(apiErr)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		logger.Error("Faild to write the the response", "Error", err)
	}
}
