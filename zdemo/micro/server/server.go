package main

import (
	"context"
	"log"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/metadata"
	"github.com/foredata/nova/netx/server"
	"github.com/foredata/nova/pkg/xid"
	"github.com/foredata/nova/zdemo/micro/common"
)

func main() {
	svr := server.New(server.WithAddr(":8888"))

	svr.GET("/api/ping", onPing)
	svr.POST("/api/login", onLogin)

	svr.Run()
}

func onLogin(ctx context.Context, req *common.LoginRequest) (*common.LoginResponse, error) {
	email := metadata.Get(ctx, "email")
	if email != "" {
		log.Printf("email is %+v", email)
	}

	if req.Username != "test" {
		return nil, netx.BadRequest("invalid username")
	}

	if req.Password != "test" {
		return nil, netx.BadRequest("invalid password")
	}

	rsp := &common.LoginResponse{}
	rsp.UID = xid.New().String()
	return rsp, nil
}

func onPing(ctx context.Context, req *common.PingRequest) (*common.PingResponse, error) {
	log.Printf("onPing, %+v\n", req.Text)

	rsp := &common.PingResponse{}
	rsp.Text = "pong:" + req.Text
	return rsp, nil
}
