package server

import "testing"

func TestFuncName(t *testing.T) {
	t.Log(funcName(toMethodPath))
}
