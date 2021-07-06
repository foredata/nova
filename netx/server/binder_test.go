package server

import (
	"net/url"
	"testing"
)

func TestBind(t *testing.T) {
	var req struct {
		Username string `query:"username"`
		Password string `query:"password"`
	}
	rawurl := "/api/test_query?username=test&password=pwd"
	u, err := url.Parse(rawurl)
	if err != nil {
		t.Fatal(err)
	}
	if err := bindStruct(&req, u.Query(), "query"); err != nil {
		t.Fatal(err)
	}

	t.Log(req)
}
