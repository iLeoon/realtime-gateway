package apiresponse

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

// Send writes a JSON response with safety buffering.
// If encoding fails, it automatically logs the error and sends a 500.
func Send(w http.ResponseWriter, status int, data interface{}) {
	const path errors.PathName = "resource/apiresponse/apiresponse"
	const op errors.Op = "apiresponse.Send"
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		log.Error.Printf("%s: %s: failed to encode response: %v", path, op, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		apiErr := apierror.Build(apierror.InternalServerErrorCode,
			"unexpected error occured while trying to process response",
			apierror.WithTarget("response"),
			apierror.WithInnerError("FailedToEncodeRespBody"),
		)
		_ = json.NewEncoder(w).Encode(apiErr)
		return
	}
	// Create a write timeout.
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
		log.Error.Println("failed to set a write timeout")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			log.Error.Printf("%s: %s: network write timeout User's connection was too slow: %v", path, op, err)
			return
		}

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Error.Printf("%s: %s: unexpected timeout error: %v", path, op, err)
			return
		}
		log.Error.Printf("%s: %s: unexpected error occured when sending the resposne message: %v", path, op, err)
		return
	}

	_ = rc.SetWriteDeadline(time.Time{})
}
