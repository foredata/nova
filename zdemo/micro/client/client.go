package main

import (
	"context"
	"fmt"
	"log"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/client"
	"github.com/foredata/nova/netx/discovery/static"
	"github.com/foredata/nova/netx/protocol/rpc"
	"github.com/foredata/nova/zdemo/micro/common"
)

func main() {
	cli := client.New(
		client.WithProtocol(rpc.New()),
		client.WithResolver(static.New()),
	)
	pingReq := &common.PingRequest{}
	pingReq.Text = "asdfsafasdfsf"

	req := netx.NewRequest()
	req.SetService("127.0.0.1:8888")
	req.SetURI("onPing")
	req.Encode(netx.CodecTypeJson, pingReq)
	rsp, err := cli.Call(context.Background(), req)
	if err != nil {
		log.Print(err)
		return
	}

	pingRsp := &common.PingResponse{}
	if err := rsp.Decode(pingRsp); err != nil {
		log.Print(err)
		return
	}
	fmt.Println(pingRsp.Text)
}
