package middleware

// It is inspired by the rate limiter implementation in the Upspin project.
// See: github.com/upspin/upspin
import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
)

const maxEntries = 100000

type rateLimiter struct {
	backoff     time.Duration
	mu          sync.Mutex
	m           map[string]*request
	max         time.Duration
	first, last *request
}

type request struct {
	ipAdd      string
	seen       time.Time
	backoff    time.Duration
	prev, next *request
}

func NewRateLimiter() *rateLimiter {
	return &rateLimiter{
		backoff: 1 * time.Second,
		max:     24 * time.Hour,
	}

}

func (r *rateLimiter) pass(now time.Time, key string) (bool, time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.m == nil {
		r.m = make(map[string]*request)
	}

	req, ok := r.m[key]
	if !ok {
		req = &request{
			ipAdd:   key,
			seen:    now,
			backoff: r.backoff,
		}
		r.m[key] = req

		if r.last != nil {
			r.last.next = req
			req.prev = r.last
		}

		r.last = req
		// if no elements in the list add the first element
		if r.first == nil {
			r.first = req
		}

		// if there is an element just push them to the end
	} else {
		reset := req.seen.Add(r.max)

		if now.After(reset) {
			req.backoff = r.backoff
		} else {
			passTime := req.seen.Add(r.backoff)
			if now.After(passTime) {
				req.backoff *= 2
				if req.backoff > r.max {
					req.backoff = r.backoff
				}
			} else {
				return false, passTime.Sub(now)
			}

		}
		req.seen = now

		if r.last != req {
			if req.prev != nil {
				req.prev.next = req.next
			} else {
				r.first = req.next
				req.next.prev = nil
			}
			if req.next != nil {
				req.next.prev = req.prev
			}
			r.last.next = req
			req.prev = r.last
			req.next = nil
			r.last = req
		}
	}

	r.rateCleaner(now)
	return true, 0
}

func (r *rateLimiter) rateCleaner(now time.Time) {
	cut := 0
	if len(r.m) > maxEntries {
		cut = len(r.m) - maxEntries
	}

	for req, i := r.first, 0; req != nil; req, i = req.next, i+1 {
		if !now.After(req.seen.Add(r.max)) && i >= cut {
			break
		}
		delete(r.m, req.ipAdd)
		// jump to the next element in the list
		r.first = req.next
		if req.next != nil {
			req.next.prev = nil
		}
	}
}

func RateLimiter(next http.Handler, rl *rateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, retryAfter := rl.pass(time.Now(), getKey(r))
		if !ok {
			w.Header().Set("Retry-After", retryAfter.String())
			apiErr := apierror.Build(apierror.RateLimitCode,
				"too many requests",
				apierror.WithTarget("ip"),
				apierror.WithInnerError("RateLimitExceeded"),
			)
			apiresponse.Send(w, http.StatusTooManyRequests, apiErr)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getKey(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	var ip string
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		ip = strings.TrimSpace(ips[0])
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip + r.URL.Path
}
