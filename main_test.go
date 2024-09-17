package main

import (
	"testing"
	"time"

	"git.mills.io/yarnsocial/yarn"
	"github.com/stretchr/testify/assert"
)

func TestFullVersion(t *testing.T) {
	assert.NotEmpty(t, yarn.FullVersion())
}

func TestNoCustomTime(t *testing.T) {
	assert.NotEqual(t, "just now", NoCustomTime(time.Now()))
}

func TestCustomTime(t *testing.T) {
	assert.NotEmpty(t, CustomTime(time.Now()))
}
