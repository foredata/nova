package httpc

import (
	"testing"
)

func TestReplacePath(t *testing.T) {
	res := replacePath("im/v1/chats/:chat_id/:not_found/:Upper", map[string]string{"chat_id": "aaa", "Upper": "upper"})
	t.Log(res)
}
