package flags

import "testing"

func TestParse(t *testing.T) {
	args := []string{"gosu", "--watch", "foo@release", "some quoted string 'inside'"}
	params, flags, err := Parse(args)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(params, len(params))
		t.Log(flags)
	}
}

func TestBind(t *testing.T) {
	type Request struct {
		Text  string
		Watch string `flag:"watch"`
	}

	args := []string{"gosu", "--watch", "foo@release", "some quoted string 'inside'"}
	var req Request
	if err := Bind(&req, args); err != nil {
		t.Error(err)
	} else {
		t.Log(req)
	}
}
