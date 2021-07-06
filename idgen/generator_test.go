package idgen_test

import (
	"testing"

	"github.com/foredata/nova/idgen"
)

func TestID(t *testing.T) {
	gen := idgen.NewGenerator(1, idgen.WithSecond())
	id, _ := gen.Next()
	t.Log(id.Format())
}
