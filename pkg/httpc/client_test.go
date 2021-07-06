package httpc_test

import (
	"context"
	"testing"

	"github.com/foredata/nova/pkg/httpc"
)

func TestClient(t *testing.T) {
	var text string
	err := httpc.Get(context.Background(), "http://www.baidu.com").Decode(&text)
	if err != nil {
		t.Error(err)
	}
	t.Log(text)
}
