package tracing_test

import (
	"io/ioutil"
	"testing"

	"github.com/foredata/nova/debug/tracing"
)

func TestTracer(t *testing.T) {
	span := tracing.StartSpan("get.data")
	defer span.Finish()

	child := tracing.StartSpan("read.file", tracing.WithParent(span))
	child.SetTag(tracing.ResourceName, "test.json")

	_, err := ioutil.ReadFile("./test.json")
	child.Finish()
	if err != nil {
		t.Fatal(err)
	}
}
