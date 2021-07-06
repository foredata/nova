package strx

import (
	"testing"
)

func TestBind(t *testing.T) {
	a := "123"
	var b *int
	if err := Bind(a, &b); err != nil {
		t.Error(err)
	} else {
		t.Logf("%+v\n", *b)
	}
}

func TestBindSlice(t *testing.T) {
	a := []string{"123", "456"}
	var b []int
	if err := BindSlice(a, &b); err != nil {
		t.Error(err)
	} else {
		t.Log(b)
	}
}
