package cli

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/foredata/nova/pkg/strx"
)

// toAction 通过反射解析Action,支持的函数签名为:
// func(ctx Context, xxx *Cmd) (interface{}, error)
func toAction(fn interface{}) (action Action, flags []*Flag) {
	if fn == nil {
		return nil, nil
	}

	if act, ok := fn.(Action); ok {
		return act, nil
	}

	rv := reflect.ValueOf(fn)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		panic(fmt.Errorf("cli: invalid action, must func"))
	}

	if rt.NumOut() < 1 || rt.NumOut() > 2 {
		panic(fmt.Errorf("cli: invalid cmd out arg num, must 1 or 2"))
	}

	if !isError(rt.Out(rt.NumOut() - 1)) {
		panic(fmt.Errorf("cli: the last out arg not error"))
	}

	switch rt.NumIn() {
	case 1:
		if !isStructPtr(rt.In(0)) {
			panic(fmt.Errorf("cli: invalid cmd in arg, must be struct ptr"))
		}
	case 2:
		if isCliContext(rt.In(0)) {
			panic(fmt.Errorf("cli: invalid cmd in arg, param0 must be Context"))
		}

		if !isStructPtr(rt.In(1)) {
			panic(fmt.Errorf("cli: invalid cmd in arg, must be struct ptr"))
		}
	}

	desc := parseCmdTag(rt.In(rt.NumIn() - 1).Elem())
	flags = toFlags(desc.Fields)
	//
	switch rt.NumIn() {
	case 1:
		action = func(ctx Context) (interface{}, error) {
			cmd := reflect.New(rt.In(0).Elem())

			if err := bindCmd(ctx, cmd.Elem(), desc); err != nil {
				return nil, err
			}

			in := []reflect.Value{cmd}
			out := rv.Call(in)
			return toActionResult(out)
		}
	case 2:
		action = func(ctx Context) (interface{}, error) {
			cmd := reflect.New(rt.In(1)).Elem()

			if err := bindCmd(ctx, cmd.Elem(), desc); err != nil {
				return nil, err
			}

			in := []reflect.Value{reflect.ValueOf(ctx), cmd}
			out := rv.Call(in)
			return toActionResult(out)
		}
	}

	return
}

type field struct {
	FieldName  string
	FieldIndex int
	ParamIndex int
	ParamSlice bool // 是否是数组,只有最后一个可以是数组
	Flag       *Flag
}

type description struct {
	Fields []*field
}

func parseCmdTag(rt reflect.Type) *description {
	desc := &description{}

	paramIndex := 0

	for i, numFields := 0, rt.NumField(); i < numFields; i++ {
		f := rt.Field(i)
		tags := f.Tag.Get("flag")
		if len(tags) == 0 {
			ff := &field{FieldName: f.Name, FieldIndex: i, ParamIndex: paramIndex, ParamSlice: f.Type.Kind() == reflect.Slice}
			desc.Fields = append(desc.Fields, ff)
			paramIndex++
			continue
		}
		tokens := split(tags)
		if len(tokens) == 0 {
			panic(fmt.Errorf("cli: invalid tag"))
		}

		flag := &Flag{}
		desc.Fields = append(desc.Fields, &field{FieldName: f.Name, FieldIndex: i, Flag: flag})

		// parse name
		t0 := strings.TrimSpace(tokens[0])
		if t0 == "-" {
			flag.Name = strx.ToSnake(f.Name)
		} else {
			names := strings.Split(t0, "|")
			flag.Name = names[0]
			flag.Aliases = names[1:]
		}

		for j := 1; j < len(tokens); j++ {
			key, val := parseKV(strings.TrimSpace(tokens[j]))
			switch key {
			case "required":
				flag.Required = toBool(val)
			case "hidden":
				flag.Hidden = toBool(val)
			case "default":
				flag.Default = val
			case "usage":
				flag.Usage = val
			}
		}

	}

	if err := isValidParams(desc.Fields); err != nil {
		panic(err)
	}

	return desc
}

func bindCmd(ctx Context, out reflect.Value, desc *description) error {
	for _, f := range desc.Fields {
		if f.Flag != nil {
			values := ctx.Flag(f.Flag.Name)
			if err := strx.BindSlicev(values, out.Field(f.FieldIndex)); err != nil {
				return err
			}
		} else if f.ParamIndex < ctx.NArg() {
			if f.ParamSlice {
				values := ctx.Tail(f.ParamIndex)
				if err := strx.BindSlicev(values, out.Field(f.FieldIndex)); err != nil {
					return err
				}
			} else {
				value := ctx.Arg(f.ParamIndex)
				if err := strx.Bindv(value, out.Field(f.FieldIndex)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func split(str string) []string {
	quoted := false
	return strings.FieldsFunc(str, func(r rune) bool {
		if r == '\'' {
			quoted = !quoted
		}

		return !quoted && r == ','
	})
}

func parseKV(str string) (string, string) {
	idx := strings.IndexByte(str, '=')
	if idx == -1 {
		return str, ""
	}

	key := strings.TrimSpace(str[:idx])
	val := strings.TrimSpace(str[idx+1:])
	val = strings.Trim(val, "'")

	return key, val
}

func toBool(str string) bool {
	if str == "" {
		return true
	}
	res, _ := strconv.ParseBool(str)
	return res
}

func toActionResult(out []reflect.Value) (interface{}, error) {
	last := out[len(out)-1]
	if !last.IsNil() {
		return nil, last.Interface().(error)
	}

	if len(out) == 1 || out[0].IsNil() {
		return nil, nil
	}

	return out[0].Interface(), nil
}

// toFlags 提取类型为Flag的字段
func toFlags(flags []*field) []*Flag {
	if len(flags) == 0 {
		return nil
	}

	res := make([]*Flag, 0, len(flags))
	for _, f := range flags {
		if f.Flag != nil {
			res = append(res, f.Flag)
		}
	}

	return res
}

// isValidParams 校验params是否合法,只有最后一个Params可以是数组
func isValidParams(fields []*field) error {
	isSlice := false
	for _, f := range fields {
		if f.Flag != nil {
			continue
		}
		if f.ParamSlice {
			if isSlice {
				return fmt.Errorf("cli: invalid param, only the last field can be slice")
			}
			isSlice = true
		}
	}

	return nil
}

var errType = reflect.TypeOf((*error)(nil)).Elem()
var cliCtxType = reflect.TypeOf((*Context)(nil)).Elem()

func isError(t reflect.Type) bool {
	return t.Implements(errType)
}

func isCliContext(t reflect.Type) bool {
	return t.Implements(cliCtxType)
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}
