package http

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging"
)

func TestNewOptionsDefaultsToNopLoggerAndNoDocs(t *testing.T) {
	o := newOptions(nil)

	assert.Equal(t, logging.NopLogger{}, o.logger)
	assert.Nil(t, o.docs)
}

func TestNewRouterLeavesDocsRoutesOutUnlessWithDocsIsGiven(t *testing.T) {
	docs := NewDocsController([]byte("{}"))

	assert.Nil(t, NewRouter(nil, nil, nil, nil, nil, Audiences{}).docs)
	assert.Same(t, docs, NewRouter(nil, nil, nil, nil, nil, Audiences{}, WithDocs(docs)).docs)
	assert.Nil(t, NewRouter(nil, nil, nil, nil, nil, Audiences{}, WithDocs(nil)).docs)
}
