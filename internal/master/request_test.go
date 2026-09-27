package master

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequireRequestIDBounds(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		status int
		ok     bool
	}{
		{name: "missing", status: 400},
		{name: "whitespace", value: " \t", status: 400},
		{name: "too long", value: strings.Repeat("x", 129), status: 400},
		{name: "maximum", value: strings.Repeat("x", 128), ok: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", nil)
			if test.value != "" {
				r.Header.Set("Idempotency-Key", test.value)
			}
			w := httptest.NewRecorder()
			if got := requireRequestID(w, r); got != test.ok {
				t.Fatalf("requireRequestID()=%v, want %v", got, test.ok)
			}
			if test.ok {
				if w.Code != 200 {
					t.Fatalf("valid key status=%d, want untouched recorder", w.Code)
				}
				return
			}
			if w.Code != test.status {
				t.Fatalf("status=%d, want %d", w.Code, test.status)
			}
		})
	}
}
