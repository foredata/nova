package httpc

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
)

func newError(err error) *Response {
	return &Response{err: err}
}

func newResponse(raw *http.Response, reqConentType string, err error) *Response {
	return &Response{Response: raw, reqConentType: reqConentType, err: err}
}

type Response struct {
	*http.Response
	err           error
	reqConentType string
}

func (rsp *Response) Error() error {
	return rsp.err
}

// Decode 解析消息
func (rsp *Response) Decode(out interface{}) error {
	if rsp.err != nil {
		return rsp.err
	}
	contentType := rsp.Header.Get("Content-Type")
	if contentType != "" {
		idx := strings.LastIndexByte(contentType, ';')
		if idx != -1 {
			contentType = contentType[0:idx]
		}
	}
	if contentType == "" {
		contentType = rsp.reqConentType
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	rsp.Body.Close()
	rsp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return decode(contentType, body, out)
}
