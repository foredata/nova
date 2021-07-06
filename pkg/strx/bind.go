package strx

import (
	"fmt"
	"reflect"
	"strconv"
)

// Bindv 将字符串绑定到reflect.Value上
func Bindv(str string, val reflect.Value) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("%w", recover())
		}
	}()

	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	switch typ.Kind() {
	case reflect.String:
		val.SetString(str)
	case reflect.Bool:
		v, err := strconv.ParseBool(toZero(str))
		if err != nil {
			return err
		}
		val.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(toZero(str), 10, 64)
		if err != nil {
			return err
		}
		val.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(toZero(str), 10, 64)
		if err != nil {
			return err
		}
		val.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(toZero(str), 64)
		if err != nil {
			return err
		}
		val.SetFloat(v)
	default:
		return fmt.Errorf("bind string fail, not support type, %+v", typ.Kind())
	}

	return nil
}

// BindSlicev 绑定字符串至reflect.Value上
func BindSlicev(values []string, val reflect.Value) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("%w", recover())
		}
	}()

	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() == reflect.Slice {
		numElems := len(values)
		slice := reflect.MakeSlice(typ, numElems, numElems)
		for j := 0; j < numElems; j++ {
			if err := Bindv(values[j], slice.Index(j)); err != nil {
				return err
			}
		}

		val.Set(slice)
		return nil
	}

	if len(values) == 0 {
		return nil
	}

	return Bindv(values[0], val)
}

func toZero(value string) string {
	if value == "" {
		return "0"
	}

	return value
}

// Bind 将字符串绑定到对应类型
func Bind(str string, out interface{}) error {
	switch x := out.(type) {
	case *string:
		*x = str
	case **string:
		*x = &str
	case *bool:
		if v, err := strconv.ParseBool(str); err == nil {
			*x = v
		} else {
			return err
		}
	case **bool:
		if v, err := strconv.ParseBool(str); err == nil {
			*x = &v
		} else {
			return err
		}
	case *int:
		if v, err := strconv.Atoi(str); err == nil {
			*x = v
		} else {
			return err
		}
	case **int:
		if v, err := strconv.Atoi(str); err == nil {
			*x = &v
		} else {
			return err
		}
	case *int8:
		if v, err := strconv.Atoi(str); err == nil {
			*x = int8(v)
		} else {
			return err
		}
	case **int8:
		if v, err := strconv.Atoi(str); err == nil {
			vv := int8(v)
			*x = &vv
		} else {
			return err
		}
	case *int16:
		if v, err := strconv.Atoi(str); err == nil {
			*x = int16(v)
		} else {
			return err
		}
	case **int16:
		if v, err := strconv.Atoi(str); err == nil {
			vv := int16(v)
			*x = &vv
		} else {
			return err
		}
	case *int32:
		if v, err := strconv.Atoi(str); err == nil {
			*x = int32(v)
		} else {
			return err
		}
	case **int32:
		if v, err := strconv.Atoi(str); err == nil {
			vv := int32(v)
			*x = &vv
		} else {
			return err
		}
	case *int64:
		if v, err := strconv.ParseInt(str, 10, 64); err == nil {
			*x = v
		} else {
			return err
		}
	case **int64:
		if v, err := strconv.ParseInt(str, 10, 64); err == nil {
			*x = &v
		} else {
			return err
		}
	case *uint:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			*x = uint(v)
		} else {
			return err
		}
	case **uint:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			vv := uint(v)
			*x = &vv
		} else {
			return err
		}
	case *uint8:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			*x = uint8(v)
		} else {
			return err
		}
	case **uint8:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			vv := uint8(v)
			*x = &vv
		} else {
			return err
		}
	case *uint16:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			*x = uint16(v)
		} else {
			return err
		}
	case **uint16:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			vv := uint16(v)
			*x = &vv
		} else {
			return err
		}
	case *uint32:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			*x = uint32(v)
		} else {
			return err
		}
	case **uint32:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			vv := uint32(v)
			*x = &vv
		} else {
			return err
		}
	case *uint64:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			*x = v
		} else {
			return err
		}
	case **uint64:
		if v, err := strconv.ParseUint(str, 10, 64); err == nil {
			*x = &v
		} else {
			return err
		}
	case *float32:
		if v, err := strconv.ParseFloat(str, 32); err == nil {
			*x = float32(v)
		} else {
			return err
		}
	case **float32:
		if v, err := strconv.ParseFloat(str, 32); err == nil {
			vv := float32(v)
			*x = &vv
		} else {
			return err
		}
	case *float64:
		if v, err := strconv.ParseFloat(str, 64); err == nil {
			*x = v
		} else {
			return err
		}
	case **float64:
		if v, err := strconv.ParseFloat(str, 64); err == nil {
			*x = &v
		} else {
			return err
		}
	case reflect.Value:
		return Bindv(str, x)
	case *reflect.Value:
		return Bindv(str, *x)
	default:
		return fmt.Errorf("not support")
	}

	return nil
}

// BindSlice 绑定数组
func BindSlice(str []string, out interface{}) error {
	switch x := out.(type) {
	case *[]string:
		*x = str
	case **[]string:
		*x = &str
	case *[]bool:
		res, err := toBoolSlice(str)
		*x = res
		return err
	case **[]bool:
		res, err := toBoolSlice(str)
		*x = &res
		return err
	case *[]int:
		res, err := toIntSlice(str)
		*x = res
		return err
	case **[]int:
		res, err := toIntSlice(str)
		*x = &res
		return err
	case *[]int8:
		res, err := toInt8Slice(str)
		*x = res
		return err
	case **[]int8:
		res, err := toInt8Slice(str)
		*x = &res
		return err
	case *[]int16:
		res, err := toInt16Slice(str)
		*x = res
		return err
	case **[]int16:
		res, err := toInt16Slice(str)
		*x = &res
		return err
	case *[]int32:
		res, err := toInt32Slice(str)
		*x = res
		return err
	case **[]int32:
		res, err := toInt32Slice(str)
		*x = &res
		return err
	case *[]int64:
		res, err := toInt64Slice(str)
		*x = res
		return err
	case **[]int64:
		res, err := toInt64Slice(str)
		*x = &res
		return err
	case *[]uint:
		res, err := toUintSlice(str)
		*x = res
		return err
	case **[]uint:
		res, err := toUintSlice(str)
		*x = &res
		return err
	case *[]uint8:
		res, err := toUint8Slice(str)
		*x = res
		return err
	case **[]uint8:
		res, err := toUint8Slice(str)
		*x = &res
		return err
	case *[]uint16:
		res, err := toUint16Slice(str)
		*x = res
		return err
	case **[]uint16:
		res, err := toUint16Slice(str)
		*x = &res
		return err
	case *[]uint32:
		res, err := toUint32Slice(str)
		*x = res
		return err
	case **[]uint32:
		res, err := toUint32Slice(str)
		*x = &res
		return err
	case *[]uint64:
		res, err := toUint64Slice(str)
		*x = res
		return err
	case **[]uint64:
		res, err := toUint64Slice(str)
		*x = &res
		return err
	case *[]float32:
		res, err := toFloat32Slice(str)
		*x = res
		return err
	case **[]float32:
		res, err := toFloat32Slice(str)
		*x = &res
		return err
	case *[]float64:
		res, err := toFloat64Slice(str)
		*x = res
		return err
	case **[]float64:
		res, err := toFloat64Slice(str)
		*x = &res
		return err
	default:
		t := reflect.TypeOf(out)
		if t.Kind() != reflect.Slice {
			v := ""
			if len(str) > 0 {
				v = str[0]
			}
			return Bind(v, out)
		}

		return fmt.Errorf("not support, %+v", t.Kind())
	}
	return nil
}

func toBoolSlice(str []string) ([]bool, error) {
	res := make([]bool, 0, len(str))
	for _, s := range str {
		if v, err := strconv.ParseBool(s); err == nil {
			res = append(res, v)
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toIntSlice(str []string) ([]int, error) {
	res := make([]int, 0, len(str))
	for _, s := range str {
		if v, err := strconv.Atoi(s); err == nil {
			res = append(res, v)
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toInt8Slice(str []string) ([]int8, error) {
	res := make([]int8, 0, len(str))
	for _, s := range str {
		if v, err := strconv.Atoi(s); err == nil {
			res = append(res, int8(v))
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toInt16Slice(str []string) ([]int16, error) {
	res := make([]int16, 0, len(str))
	for _, s := range str {
		if v, err := strconv.Atoi(s); err == nil {
			res = append(res, int16(v))
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toInt32Slice(str []string) ([]int32, error) {
	res := make([]int32, 0, len(str))
	for _, s := range str {
		if v, err := strconv.Atoi(s); err == nil {
			res = append(res, int32(v))
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toInt64Slice(str []string) ([]int64, error) {
	res := make([]int64, 0, len(str))
	for _, s := range str {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			res = append(res, int64(v))
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toUintSlice(str []string) ([]uint, error) {
	res := make([]uint, 0, len(str))
	for _, s := range str {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			res = append(res, uint(v))
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toUint8Slice(str []string) ([]uint8, error) {
	res := make([]uint8, 0, len(str))
	for _, s := range str {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			res = append(res, uint8(v))
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toUint16Slice(str []string) ([]uint16, error) {
	res := make([]uint16, 0, len(str))
	for _, s := range str {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			res = append(res, uint16(v))
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toUint32Slice(str []string) ([]uint32, error) {
	res := make([]uint32, 0, len(str))
	for _, s := range str {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			res = append(res, uint32(v))
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toUint64Slice(str []string) ([]uint64, error) {
	res := make([]uint64, 0, len(str))
	for _, s := range str {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			res = append(res, v)
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toFloat32Slice(str []string) ([]float32, error) {
	res := make([]float32, 0, len(str))
	for _, s := range str {
		if v, err := strconv.ParseFloat(s, 32); err == nil {
			res = append(res, float32(v))
		} else {
			return nil, err
		}
	}

	return res, nil
}

func toFloat64Slice(str []string) ([]float64, error) {
	res := make([]float64, 0, len(str))
	for _, s := range str {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			res = append(res, v)
		} else {
			return nil, err
		}
	}

	return res, nil
}
