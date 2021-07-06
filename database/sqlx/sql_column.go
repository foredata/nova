package sqlx

import (
	"fmt"
	"reflect"

	"github.com/foredata/nova/pkg/reflects"
)

type sqlColumn struct {
	Key   string
	Value interface{}
}

func toColumns(v interface{}) ([]*sqlColumn, error) {
	switch x := v.(type) {
	case M:
		return toMapColumns(x)
	case map[string]interface{}:
		return toMapColumns(x)
	case D:
		var columns []*sqlColumn
		for _, k := range x {
			columns = append(columns, &sqlColumn{
				Key:   k.Key,
				Value: k.Value,
			})
		}
	}

	rv := reflect.ValueOf(v)
	rt := reflects.Deref(rv.Type())
	switch rt.Kind() {
	case reflect.Map:
		if rt.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("sql: invalid data")
		}

		var columns []*sqlColumn
		for _, k := range rv.MapKeys() {
			v := rv.MapIndex(k)
			columns = append(columns, &sqlColumn{
				Key:   k.String(),
				Value: v.Interface(),
			})
		}

		return columns, nil
	case reflect.Struct:
		m, err := getModel(v)
		if err != nil {
			return nil, err
		}

		res := make([]*sqlColumn, 0, len(m.Fields))
		for _, f := range m.Fields {
			c := &sqlColumn{
				Key:   f.Name,
				Value: rv.Field(f.Index).Interface(),
			}
			res = append(res, c)
		}

		return res, err
	default:
		return nil, fmt.Errorf("sql: not support type")
	}
}

func toMapColumns(m map[string]interface{}) ([]*sqlColumn, error) {
	var columns []*sqlColumn
	for k, v := range m {
		columns = append(columns, &sqlColumn{
			Key:   k,
			Value: v,
		})
	}

	// order ?
	// sort.Slice(columns, func(i, j int) bool {
	// 	return columns[i].Name < columns[j].Name
	// })

	return columns, nil
}

func toMap(v interface{}) (M, error) {
	switch x := v.(type) {
	case M:
		return x, nil
	case map[string]interface{}:
		return x, nil
	case D:
		res := make(M, len(x))
		for _, e := range x {
			res[e.Key] = e.Value
		}
		return res, nil
	}

	rv := reflect.ValueOf(v)
	rt := reflects.Deref(rv.Type())
	switch rt.Kind() {
	case reflect.Map:
		if rt.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("sql: invalid data")
		}

		res := make(M, rv.Len())
		for _, k := range rv.MapKeys() {
			v := rv.MapIndex(k)
			res[k.String()] = v.Interface()
		}

		return res, nil
	case reflect.Struct:
		m, err := getModel(v)
		if err != nil {
			return nil, err
		}

		res := make(M, len(m.Fields))
		for _, f := range m.Fields {
			res[f.Name] = rv.Field(f.Index).Interface()
		}
		return res, nil
	default:
		return nil, errInvalidData
	}
}
