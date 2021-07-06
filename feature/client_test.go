package feature_test

import (
	"context"
	"testing"

	"github.com/foredata/nova/feature"
)

func TestToggled(t *testing.T) {
	demoOn, _ := feature.IsToggled(context.Background(), "demo", nil, false)
	if demoOn {
		t.Logf("demo on")
	} else {
		t.Logf("demo off")
	}

	variant, _ := feature.Evaluate(context.Background(), "variant", nil, "v1")
	switch variant {
	case "v1":
		t.Log("hit v1")
	case "v2":
		t.Log("hit v2")
	case "v3":
		t.Log("hit v3")
	}
}
