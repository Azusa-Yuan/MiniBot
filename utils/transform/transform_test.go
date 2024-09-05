package transform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBidAndId(t *testing.T) {
	combine := BidWithidInt64Custom(123456, 789413212)
	bid, id := GetBidAndId(combine)
	assert.Equal(t, int64(123456), bid)
	assert.Equal(t, int64(789413212), id)
}
