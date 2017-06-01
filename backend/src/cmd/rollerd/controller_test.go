package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRequestIP(t *testing.T) {
	testCases := []struct {
		remoteAddr     string
		xForwardedFor  string
		expectedOutput string
	}{
		{"", "", ""},
		{"1.1.1.1:12345", "", "1.1.1.1"},
		{"1.1.1.1:12345", "2.2.2.2.2", "1.1.1.1"},
		{"1.1.1.1:12345", "2.2.2.2", "2.2.2.2"},
		{"1.1.1.1:12345", "3.3.3.3, 4.4.4.4", "3.3.3.3"},
	}

	for _, tc := range testCases {
		r, _ := http.NewRequest("POST", "/v1/update", nil)
		r.RemoteAddr = tc.remoteAddr
		r.Header.Set("X-Forwarded-For", tc.xForwardedFor)
		assert.Equal(t, tc.expectedOutput, getRequestIP(r))
	}
}
