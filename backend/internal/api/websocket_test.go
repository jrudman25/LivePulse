package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckOrigin_AllowsLocalhost(t *testing.T) {
	r := &http.Request{Header: http.Header{"Origin": []string{"http://localhost:3000"}}}
	assert.True(t, upgrader.CheckOrigin(r), "localhost:3000 should be allowed")
}

func TestCheckOrigin_AllowsProductionDomain(t *testing.T) {
	r := &http.Request{Header: http.Header{"Origin": []string{"https://livepulse-hq.vercel.app"}}}
	assert.True(t, upgrader.CheckOrigin(r), "production Vercel domain should be allowed")
}

func TestCheckOrigin_AllowsEmptyOrigin(t *testing.T) {
	r := &http.Request{Header: http.Header{}}
	assert.True(t, upgrader.CheckOrigin(r), "empty origin (e.g. non-browser client) should be allowed")
}

func TestCheckOrigin_RejectsMaliciousDomain(t *testing.T) {
	cases := []string{
		"https://evil.com",
		"http://localhost:8080",
		"https://livepulse-hq.vercel.app.evil.com",
		"http://localhost:3001",
	}
	for _, origin := range cases {
		r := &http.Request{Header: http.Header{"Origin": []string{origin}}}
		assert.False(t, upgrader.CheckOrigin(r), "origin %q should be rejected", origin)
	}
}

func TestHandleWebSocket_RejectsMissingSessionID(t *testing.T) {
	server := &Server{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	server.HandleWebSocket(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleWebSocket_RejectsInvalidSessionID(t *testing.T) {
	server := &Server{}
	for _, id := range []string{"", "bad id!", strings.Repeat("a", 200), "../etc"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ws?session_id="+url.QueryEscape(id), nil)
		server.HandleWebSocket(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code, "session_id %q should be rejected", id)
	}
}

func TestIsValidSessionID(t *testing.T) {
	for _, id := range []string{"Z7r9jZ1Adb8f1", "550e8400-e29b-41d4-a716-446655440000", "wsbench-local", "room_42"} {
		assert.True(t, isValidSessionID(id), "session_id %q should be valid", id)
	}
}
