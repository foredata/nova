package sqlx

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/foredata/nova/pkg/reflects"
	"github.com/foredata/nova/pkg/strx"
)

var (
	errInvalidModelType = errors.New("sql: invalid model type,must struct")
)

var (
	models    = make(map[string]*sqlModel)
	modelsMux sync.RWMutex
)

type sqlModel struct {
	Table  string      // 表名
	Fields []*sqlField // 列信息
}

type sqlField struct {
	Name       string
	Type       int
	Index      int
	PrimaryKey bool
	Unique     bool
	Value      interface{} // 仅解析map时使用
}

type sqlTableName interface {
	TableName() string
}

// RegisterModel 注册model
func RegisterModel(m interface{}) error {
	table := getTableName(m)
	if table == "" {
		return fmt.Errorf("sql: cannot parse table name, %+v", m)
	}

	_, err := registerModel(table, m)
	return err
}

func registerModel(table string, m interface{}) (*sqlModel, error) {
	res := &sqlModel{}
	rt := reflects.Deref(reflect.TypeOf(m))
	if rt.Kind() != reflect.Struct {
		return nil, errInvalidModelType
	}
	for i := 0; i < rt.NumField(); i++ {
		rf := rt.Field(i)
		tags := rf.Tag.Get("db")
		if len(tags) == 0 {
			continue
		}

		f := &sqlField{Index: i}

		tokens := strings.Split(tags, ";")
		for j, t := range tokens {
			t = strings.TrimSpace(t)
			if t == "" {
				if j == 0 {
					f.Name = strx.ToSnake(rf.Name)
				}
				continue
			}
			tt := strings.SplitN(t, ":", 2)
			t0 := tt[0]
			t1 := ""
			if len(tt) > 1 {
				t1 = tt[1]
			}
			switch t0 {
			case "column":
				f.Name = t1
			case "primaryKey":
				f.PrimaryKey = true
			case "unique":
				f.Unique = true
			}

			if f.Name == "" {
				return nil, fmt.Errorf("sql: invalid field name, name=%s, tags=%s", rf.Name, tags)
			}
		}

		res.Fields = append(res.Fields, f)
	}
	modelsMux.Lock()
	models[table] = res
	modelsMux.Unlock()
	return res, nil
}

func getTableName(m interface{}) string {
	if tn, ok := m.(sqlTableName); ok {
		return tn.TableName()
	} else {
		t := reflects.Deref(reflect.TypeOf(m))
		return strx.ToSnake(t.Name())
	}
}

func getModel(m interface{}) (*sqlModel, error) {
	table := getTableName(m)
	if table == "" {
		return nil, nil
	}

	modelsMux.RLock()
	model := models[table]
	modelsMux.RUnlock()
	if model != nil {
		return model, nil
	}

	return registerModel(table, m)
}
