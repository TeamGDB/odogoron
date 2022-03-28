package odogoron

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewApp(t *testing.T) {
	app := New()
	assert.False(t, app.isRunning)
}
