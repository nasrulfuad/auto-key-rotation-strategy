package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOptionsDefaultsToAFiveSecondHTTPClient(t *testing.T) {
	assert.Equal(t, 5*time.Second, newOptions(nil).httpClient.Timeout)
}

func TestNewOptionsAppliesHTTPClientAndIgnoresNil(t *testing.T) {
	client := &http.Client{Timeout: time.Minute}

	assert.Same(t, client, newOptions([]Option{WithHTTPClient(client)}).httpClient)
	assert.Equal(t, 5*time.Second, newOptions([]Option{WithHTTPClient(nil)}).httpClient.Timeout)
}
