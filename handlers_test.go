package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetHost(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want string
	}{
		{
			name: "host in header",
			req: &http.Request{
				Host: "example.com",
				Header: http.Header{
					"Host": []string{"example.com"},
				},
			},
			want: "example.com",
		},
		{
			name: "host in request",
			req: &http.Request{
				Host: "example.com",
			},
			want: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getHost(tt.req)
			if got != tt.want {
				t.Errorf("getHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostDispatchingHandler(t *testing.T) {
	handler := NewHostDispatchingHandler()
	
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	handler.HandleHost("example.com", testHandler)
	
	tests := []struct {
		name       string
		host       string
		wantStatus int
	}{
		{
			name:       "known host",
			host:       "example.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "unknown host",
			host:       "unknown.com",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Host = tt.host
			rr := httptest.NewRecorder()
			
			handler.ServeHTTP(rr, req)
			
			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned status %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}

