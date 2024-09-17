package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.yarn.social/types"
)

func TestNoCustomTime(t *testing.T) {
	assert.NotEqual(t, "just now", NoCustomTime(time.Now()))
}

func TestCustomTime(t *testing.T) {
	assert.NotEmpty(t, CustomTime(time.Now()))
}

func TestCustomRelTime(t *testing.T) {
	assert.NotEmpty(t, CustomRelTime(time.Now(), time.Now(), "ago", "from now"))
}

func TestFormatTwt(t *testing.T) {
	twter := types.NilTwt.Twter()
	assert.NotEmpty(t, FormatTwt(types.MakeTwt(twter, time.Now(), "test")))
}

func TestFormatTwtEmpty(t *testing.T) {
	assert.EqualValues(t, "<p></p>", FormatTwt(types.NilTwt))
}

func TestFormatTwtLink(t *testing.T) {
	assert.EqualValues(t, "<p><p><a href=\"https://example.com\" rel=\"nofollow\">https://example.com</a></p>\n</p>", FormatTwt(types.MakeTwt(types.NilTwt.Twter(), time.Now(), "https://example.com")))
}

func TestFormatTwtLinkEmpty(t *testing.T) {
	assert.EqualValues(t, "<p></p>", FormatTwt(types.MakeTwt(types.NilTwt.Twter(), time.Now(), "")))
}

func TestFormatTwtLinkInvalid(t *testing.T) {
	assert.EqualValues(t, "<p><p>invalid</p>\n</p>", FormatTwt(types.MakeTwt(types.NilTwt.Twter(), time.Now(), "invalid")))
}
