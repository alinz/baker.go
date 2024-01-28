package rule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alinz/baker.go/pkg/log"
	"github.com/alinz/baker.go/rule/internal/rate"
)

type WindowDuration struct {
	time.Duration
}

// MarshalJSON implements the json.Marshaler interface for WindowDuration.
func (d WindowDuration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for WindowDuration.
func (d *WindowDuration) UnmarshalJSON(data []byte) error {
	if len(data) < 2 {
		d.Duration = 0
		return nil
	}

	duration, err := time.ParseDuration(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}

	d.Duration = duration
	return nil
}

type RateLimiter struct {
	RequestLimit   int            `json:"request_limit"`
	WindowDuration WindowDuration `json:"window_duration"`
	middle         func(next http.Handler) http.Handler
}

var _ Middleware = (*RateLimiter)(nil)

func (r *RateLimiter) IsCachable() bool {
	return true
}

func (r *RateLimiter) UpdateMiddelware(newImpl Middleware) Middleware {
	newR, ok := newImpl.(*RateLimiter)
	if !ok {
		log.Error().
			Str("type", "RateLimiter").
			Msg("failed to update middleware")
		return r
	}

	if r != nil {
		if r.RequestLimit != newR.RequestLimit ||
			r.WindowDuration != newR.WindowDuration {
			log.Warn().
				Str("type", "RateLimiter").
				Int("request_limit", r.RequestLimit).
				Dur("window_duration", r.WindowDuration.Duration).
				Msg("middleware is already initialized with different values")
		}
		return r
	}

	if r.RequestLimit == newR.RequestLimit &&
		r.WindowDuration == newR.WindowDuration &&
		r.middle != nil {
		return r
	}

	r.RequestLimit = newR.RequestLimit
	r.WindowDuration = newR.WindowDuration

	r.middle = rate.LimitByIP(r.RequestLimit, r.WindowDuration.Duration)

	return r
}

func (r *RateLimiter) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func NewRateLimiter(requestLimit int, windowDuration time.Duration) struct {
	Type string `json:"type"`
	Args any    `json:"args"`
} {
	return struct {
		Type string `json:"type"`
		Args any    `json:"args"`
	}{
		Type: "RateLimiter",
		Args: RateLimiter{
			RequestLimit: requestLimit,
			WindowDuration: WindowDuration{
				Duration: windowDuration,
			},
		},
	}
}

func RegisterRateLimiter() RegisterFunc {
	return func(m map[string]BuilderFunc) error {
		m["RateLimiter"] = func(raw json.RawMessage) (Middleware, error) {
			rateLimiter := &RateLimiter{}
			err := json.Unmarshal(raw, rateLimiter)
			if err != nil {
				return nil, err
			}
			return rateLimiter, nil
		}

		return nil
	}
}
