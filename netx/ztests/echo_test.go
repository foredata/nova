package ztests

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/transport/gpc"
	"github.com/foredata/nova/netx/transport/nio"
	"github.com/foredata/nova/pkg/bytex"
)

func TestEcho(t *testing.T) {
	runEcho(t, gpc.New)
}

func TestNio(t *testing.T) {
	runEcho(t, nio.New)
}

func runEcho(t *testing.T, fact netx.Factory) {
	tran := fact()
	tran.AddFilters(
		&echoFilter{},
	)
	// start server
	t.Logf("start server")
	l, err := tran.Listen(":6789")
	if err != nil {
		t.Fatalf("listen fail, %+v", err)
	}

	defer l.Close()
	conn, err := tran.Dial(
		"localhost:6789",
		netx.WithDialCallback(func(conn netx.Conn, err error) {
			if err == nil {
				_ = conn.Send(&echoMsg{Text: "ping"})
			}
			fmt.Printf("dial finish, %+v, %+v\n", conn.ID(), err)
		}),
	)
	if err != nil {
		t.Fatalf("dial fail, %+v", err)
	}
	time.Sleep(time.Second)
	conn.Close()
	l.Close()
	t.Logf("echo finish")
}

type echoMsg struct {
	Text string
}

// filter
type echoFilter struct {
	netx.BaseFilter
}

func (f *echoFilter) Name() string {
	return "echo"
}

func (f *echoFilter) HandleRead(ctx netx.FilterCtx) error {
	buff, ok := ctx.Data().(bytex.Buffer)
	if !ok {
		ctx.Abort()
		return nil
	}
	// 粘包处理
	_, _ = buff.Seek(0, io.SeekStart)
	var len uint16

	if err := bytex.ReadUint16LE(buff, &len); err != nil {
		return err
	}
	fmt.Printf("read len, %+v\n", len)
	msgdata := buff.ReadN(int(len))
	if msgdata == nil {
		fmt.Printf("no enough data, %+v", buff.Len())
		return nil
	}
	buff.Discard()
	fmt.Printf("read msg:%+v\n", msgdata.String())
	req := &echoMsg{}
	dec := json.NewDecoder(msgdata)
	if err := dec.Decode(req); err != nil {
		fmt.Printf("decode fail, %+v\n", err)
		return nil
	}
	rsp := &echoMsg{}
	if strings.HasPrefix(req.Text, "ping") {
		rsp.Text = "pong"
	} else {
		rsp.Text = "ping"
	}
	log.Printf("recv: %s", req.Text)
	log.Printf("send: %s", rsp.Text)
	return ctx.Conn().Send(rsp)
}

func (f *echoFilter) HandleWrite(ctx netx.FilterCtx) error {
	buff := bytex.NewBuffer()
	enc := json.NewEncoder(buff)
	if err := enc.Encode(ctx.Data()); err != nil {
		ctx.Abort()
		return nil
	}
	fmt.Printf("format msg, %+v\n", buff.String())
	//
	size := buff.Len()
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, uint16(size))
	_ = buff.Prepend(data)
	_, _ = buff.Seek(0, io.SeekStart)
	fmt.Printf("write len=%+v\n", size)
	ctx.SetData(buff)
	return nil
}
