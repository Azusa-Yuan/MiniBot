package filter

import (
	"encoding/json"
	"testing"

	zero "ZeroBot"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestFilter(t *testing.T) {
	e := map[string]interface{}{
		"post_type": "notice",
		"user_id":   "notice",
	}
	b, _ := json.Marshal(e)
	rawEvent := gjson.ParseBytes(b)
	event := &zero.Event{
		RawEvent: rawEvent,
	}
	result := Filter(
		func(ctx *zero.Ctx) gjson.Result {
			return ctx.Event.RawEvent
		},
		Field("post_type").Any(
			Equal("notice"),
			Not(
				In("message"),
			),
		),
		Field("user_id").All(
			NotEqual("abs"),
		),
	)(&zero.Ctx{Event: event})
	assert.Equal(t, true, result)
}
