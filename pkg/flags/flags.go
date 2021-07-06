package flags

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/foredata/nova/pkg/strx"
)

// Bind 解析并绑定参数,有flag标记的tag使用可选参数,否则安顺序绑定参数
func Bind(out interface{}, args []string) error {
	params, options, err := Parse(args)
	if err != nil {
		return err
	}

	// TODO: cache reflect struct?
	index := 0
	rv := reflect.ValueOf(out)
	rt := rv.Type()
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("non-struct type")
	}

	for i := 0; i < rt.NumField(); i++ {
		ftyp := rt.Field(i)
		fval := rv.Field(i)
		tags := ftyp.Tag.Get("flag")
		if len(tags) == 0 {
			// bind params
			if index >= len(params) {
				return fmt.Errorf("bind flag fail, params overflow, %+v, %+v", ftyp.Name, args)
			}

			if err := strx.Bindv(params[index], fval); err != nil {
				return err
			}
			continue
		}

		// bind options
		tokens := strings.Split(tags, "|")
		for _, v := range tokens {
			v = strings.TrimSpace(v)
			vals, ok := options[v]
			if !ok {
				continue
			}

			if err := strx.BindSlicev(vals, fval); err != nil {
				return err
			}
		}
	}

	return nil
}

// Parse 类似官方flag,解析CommandLine生成必须参数和可选参数
// 1:-表示shortcut,可以多个合并,例如,-h 表示help, -czvx，表示-c -z -v -x,如果带参数,则只会设置给最后一个
// 2:--表示全称，例如--help
// 3:后边可以紧跟一个参数,可以使用=连载一起写,也可以空格分隔
// 4:可以重复,相同的则合并成1个处理
//
// https://github.com/simonleung8/flags
// https://github.com/jessevdk/go-flags
func Parse(args []string) ([]string, map[string][]string, error) {
	var params []string
	options := make(map[string][]string)
	for idx := 0; idx < len(args); idx++ {
		token := args[idx]
		if len(token) == 0 {
			continue
		}

		if token[0] != '-' {
			params = append(params, token)
			continue
		}
		// the last maybe -
		if len(token) == 1 {
			if idx < len(args)-1 {
				return nil, nil, errors.New("invalid -, not the last")
			}
			options["-"] = nil
			return params, options, nil
		}

		var short bool
		var key string
		var val string
		if token[1] == '-' {
			short = false
			key = token[2:]
		} else {
			short = true
			key = token[1:]
		}

		if strings.ContainsRune(key, '=') {
			values := strings.SplitN(key, "=", 2)
			key = values[0]
			val = values[1]
		} else if idx+1 < len(args) && args[idx+1][0] != '-' {
			idx++
			val = args[idx]
		}

		if short && len(key) > 1 {
			// multiple short key, like -aux
			for _, ch := range key {
				k := string(ch)
				if _, ok := options[k]; ok {
					return nil, nil, fmt.Errorf("duplicate shot key, %+v", k)
				}

				options[k] = nil
			}
		} else if val == "" {
			if s, ok := options[key]; ok {
				return nil, nil, fmt.Errorf("duplicate key, %+v, %+v", key, s)
			}
			options[key] = nil
		} else {
			options[key] = append(options[key], val)
		}
	}

	return params, options, nil
}
