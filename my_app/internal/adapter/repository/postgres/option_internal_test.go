package postgres

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/store"
)

func TestNewOptionsDefaultsToQueriesOnTheGivenDatabase(t *testing.T) {
	db := &sql.DB{}

	assert.Equal(t, store.New(db), newOptions(db, nil).querier)
}

func TestNewOptionsAppliesQuerierAndIgnoresNil(t *testing.T) {
	db := &sql.DB{}
	querier := store.New(&sql.Tx{})

	assert.Same(t, querier, newOptions(db, []Option{WithQuerier(querier)}).querier)
	assert.Equal(t, store.New(db), newOptions(db, []Option{WithQuerier(nil)}).querier)
}
