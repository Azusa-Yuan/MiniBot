package pcrjjc3

import (
	"MiniBot/utils/tests"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserQuery(t *testing.T) {
	_, err := userQuery("2748597743")
	assert.NoError(t, err)
}

func TestCalKnightRank(t *testing.T) {
	rank := calKnightRank(2748597743)
	assert.Equal(t, 201, rank)
}

func TestDelAttention(t *testing.T) {
	mockClient := tests.CreatMockClient()

	msg := "删除关注 1"
	mockClient.Send(msg)
	msg = mockClient.Get()
	ok := strings.Contains(msg, "您没有绑定任何游戏账号")
	assert.Equal(t, true, ok)

	mockClient.SetUid(1043728417)
	msg = "删除关注 2"
	mockClient.Send(msg)
	msg = mockClient.Get()
	ok = strings.Contains(msg, "请输入正确的关注序号")
	assert.Equal(t, true, ok)

	mockClient.SetUid(1043728417)
	msg = "删除关注 1"
	mockClient.Send(msg)
	msg = mockClient.Get()
	ok = strings.Contains(msg, "删除成功")
	assert.Equal(t, true, ok)
}
