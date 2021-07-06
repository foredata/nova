package job

import (
	"context"
	"fmt"
	"testing"
)

func TestJob(t *testing.T) {
	Register("echo", onEcho)

	jobId, err := Run(context.Background(), "echo", []string{"hello"})
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("jobId:%+v\n", jobId)
	}
}

// 绑定流程:如果有flag标记则根据flag名绑定,否则根据字段顺序绑定参数且为必须参数
type echoReq struct {
	Text string
	Opt  string `flag:"o|opt"`
}

func onEcho(ctx context.Context, req *echoReq) error {
	fmt.Printf("echo job:%+v\n", req.Text)
	return nil
}
