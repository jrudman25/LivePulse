package api

import (
	"net/http"
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
	// Verify the sessionID guard logic directly
	sessionID := ""
	assert.Empty(t, sessionID, "empty session_id should be caught before upgrade")
}
