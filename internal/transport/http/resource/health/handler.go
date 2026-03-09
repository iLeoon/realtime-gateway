package health

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/jackc/pgx/v5/pgxpool"
)

type componentStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type healthResponse struct {
	Status     string                     `json:"status"`
	Components map[string]componentStatus `json:"components"`
}

type Handler struct {
	db   *pgxpool.Pool
	conf *config.Config
}

func NewHandler(db *pgxpool.Pool, conf *config.Config) *Handler {
	return &Handler{db: db, conf: conf}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Check)
	return mux
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	components := make(map[string]componentStatus)
	healthy := true

	// Check database
	dbCtx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := h.db.Ping(dbCtx); err != nil {
		components["database"] = componentStatus{Status: "unhealthy", Message: err.Error()}
		healthy = false
	} else {
		stat := h.db.Stat()
		components["database"] = componentStatus{
			Status:  "healthy",
			Message: formatPoolStat(stat),
		}
	}

	// Check TCP server
	conn, err := net.DialTimeout("tcp", h.conf.TcpPort, 2*time.Second)
	if err != nil {
		components["tcp_server"] = componentStatus{Status: "unhealthy", Message: err.Error()}
		healthy = false
	} else {
		conn.Close()
		components["tcp_server"] = componentStatus{Status: "healthy"}
	}

	overall := "healthy"
	statusCode := http.StatusOK
	if !healthy {
		overall = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	apiresponse.Send(w, statusCode, healthResponse{
		Status:     overall,
		Components: components,
	})
}

func formatPoolStat(s *pgxpool.Stat) string {
	return "total=" + itoa(int(s.TotalConns())) +
		" idle=" + itoa(int(s.IdleConns())) +
		" acquired=" + itoa(int(s.AcquiredConns()))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
