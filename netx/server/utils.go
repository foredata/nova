package server

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/foredata/nova/netx"
)

// funcName reflect func name
func funcName(h interface{}) string {
	v := reflect.ValueOf(h)
	if v.Type().Kind() == reflect.Func {
		name := runtime.FuncForPC(v.Pointer()).Name()
		index := strings.LastIndexByte(name, '.')
		if index != -1 {
			return name[index+1:]
		}
		return name
	}

	return v.Type().String()
}

func toMethodPath(method netx.Method, path string) string {
	return fmt.Sprintf("[%s]%s", method.String(), path)
}
