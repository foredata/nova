package sqlx

import (
	"reflect"
	"regexp"
	"strings"
	"sync"
)

const (
	dbTagKey = "db"
)

var (
	dbMap = make(map[reflect.Type]*dbStruct)
	dbMux sync.RWMutex
)

type dbField struct {
	Indexes []int // 普通字段只会有1个值,nested struct field会有多个值
}

// dbStruct 反射解析struct中含有db tag的字段
type dbStruct struct {
	Fields map[string]*dbField
}

func getStruct(t reflect.Type) *dbStruct {
	if t.Kind() != reflect.Struct {
		return nil
	}

	dbMux.RLock()
	res := dbMap[t]
	dbMux.RUnlock()

	if res != nil {
		return res
	}

	dbMux.Lock()
	res = dbMap[t]
	if res == nil {
		res = parseStruct(t)
		dbMap[t] = res
	}
	dbMux.Unlock()

	return res
}

func parseStruct(rt reflect.Type) *dbStruct {
	result := &dbStruct{
		Fields: make(map[string]*dbField, rt.NumField()),
	}

	type traverse struct {
		Type         reflect.Type
		IndexPrefix  []int
		ColumnPrefix string
	}

	var queue = []*traverse{
		{Type: rt, IndexPrefix: nil, ColumnPrefix: ""},
	}

	for len(queue) > 0 {
		traversal := queue[0]
		queue = queue[1:]
		structType := traversal.Type
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				// Field is unexported, skip it.
				continue
			}

			tag, tagPresent := field.Tag.Lookup(dbTagKey)
			if tagPresent {
				tag = strings.Split(tag, ",")[0]
			}
			if tag == "-" {
				continue
			}

			index := make([]int, 0, len(traversal.IndexPrefix)+len(field.Index))
			index = append(index, traversal.IndexPrefix...)
			index = append(index, field.Index...)

			columnPart := tag
			if !tagPresent {
				columnPart = toSnakeCase(field.Name)
			}

			if !field.Anonymous {
				column := buildColumn(traversal.ColumnPrefix, columnPart)
				if _, exists := result.Fields[column]; !exists {
					result.Fields[column] = &dbField{Indexes: index}
				}
			}

			childType := field.Type
			if field.Type.Kind() == reflect.Ptr {
				childType = field.Type.Elem()
			}

			if childType.Kind() == reflect.Struct {
				if field.Anonymous {
					// If "db" tag is present for embedded struct
					// use it with "." to prefix all column from the embedded struct.
					// the default behavior is to propagate columns as is.
					columnPart = tag
				}
				columnPrefix := buildColumn(traversal.ColumnPrefix, columnPart)
				queue = append(queue, &traverse{
					Type:         childType,
					IndexPrefix:  index,
					ColumnPrefix: columnPrefix,
				})
			}
		}
	}

	return result
}

func buildColumn(prefix string, part string) string {
	if prefix != "" {
		return prefix + "." + part
	}

	return part
}

var (
	matchFirstCapRe = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCapRe   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func toSnakeCase(str string) string {
	snake := matchFirstCapRe.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCapRe.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
