package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/foredata/nova/netx/registry"
)

const (
	prefixSvc = "svc_"
)

type entry struct {
	Expired time.Time         // 过期时间
	Service *registry.Service // 服务内容
}

func (e *entry) Encode() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Expired.Format(time.RFC3339))
	buf.WriteByte('\t')

	data, err := json.Marshal(e.Service)
	if err != nil {
		return nil, fmt.Errorf("encode service fail, %+v", err)
	}
	buf.Write(data)
	return buf.Bytes(), nil
}

func (e *entry) Decode(data []byte) error {
	text := string(data)
	index := strings.IndexByte(text, '\t')
	if index == -1 {
		return fmt.Errorf("invalid sep")
	}
	t, err := time.Parse(time.RFC3339, text[:index])
	if err != nil {
		return fmt.Errorf("parse time fail, %+v", err)
	}
	e.Expired = t

	if err := json.Unmarshal([]byte(text[index+1:]), e.Service); err != nil {
		return fmt.Errorf("parse service fail, %+v", err)
	}

	return nil
}

func fromFile(filename string) (*entry, []byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	e := &entry{}
	if err := e.Decode(data); err != nil {
		return nil, nil, err
	}

	return e, data, nil
}

func toFilename(rootDir string, noteId string) string {
	name := prefixSvc + noteId
	return path.Join(rootDir, name)
}

func toNodeId(filename string) string {
	return strings.TrimPrefix(filename, prefixSvc)
}

func isServiceFile(name string) bool {
	return strings.HasPrefix(name, prefixSvc)
}
