package image_tools

import (
	"MiniBot/utils/cache"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetImageFormat(t *testing.T) {
	data, _ := cache.GetAvatar(741433361)
	format, _ := GetImageFormat(data)
	assert.Equal(t, "jpeg", format)
}
