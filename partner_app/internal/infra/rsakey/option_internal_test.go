package rsakey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGeneratorDefaultsToMinimumBits(t *testing.T) {
	generator, err := NewGenerator()

	require.NoError(t, err)
	assert.Equal(t, MinimumBits, generator.bits)
}

func TestNewGeneratorAppliesBitsAndRejectsWeakKeys(t *testing.T) {
	generator, err := NewGenerator(WithBits(3072))
	require.NoError(t, err)
	assert.Equal(t, 3072, generator.bits)

	_, err = NewGenerator(WithBits(1024))
	assert.Error(t, err)
}
