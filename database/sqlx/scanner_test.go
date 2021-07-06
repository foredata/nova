package sqlx

import (
	"context"
	"testing"
)

func TestScan(t *testing.T) {
	type demoItem struct {
		Name int
	}
	d, err := Open("mysql", "xx")
	if err != nil {
		t.Error(err)
	}
	ctx := context.Background()
	db, err := d.Database(ctx, "demo")
	if err != nil {
		t.Error(err)
	}
	var out []*demoItem
	s, err := NewScanner(ctx, &out, db, "demo", M{"name": "aaa"}, 100, nil)
	if err != nil {
		t.Error(err)
	}
	for s.Next() {
		v := s.Value()
		t.Log(v)
	}
}
