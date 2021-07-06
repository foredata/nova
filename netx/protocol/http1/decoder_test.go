package http1

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

func newBuffer(data string) bytex.Buffer {
	data = strings.TrimPrefix(data, "\n")
	buf := bytex.NewBuffer()
	_ = buf.Append(data)
	_, _ = buf.Seek(0, io.SeekStart)
	return buf
}

func TestDecodeGet(t *testing.T) {
	req := `GET /sugrec?prod=pc_his&from=pc_web&json=1&sid=33985_31254_33848_33758_33607_26350&hisdata=%5B%7B%22time%22%3A1594129926%2C%22kw%22%3A%22%E6%9F%94%E6%80%A7%E7%94%B5%E8%B7%AF%E6%9D%BF%20%E9%BE%99%E5%A4%B4%22%7D%2C%7B%22time%22%3A1594130474%2C%22kw%22%3A%22%E6%9F%94%E6%80%A7%E7%94%B5%E8%B7%AF%E6%9D%BF%22%2C%22fq%22%3A2%7D%2C%7B%22time%22%3A1594208014%2C%22kw%22%3A%22redis%20hget%22%7D%2C%7B%22time%22%3A1594212699%2C%22kw%22%3A%22%E9%93%BE%E5%AE%B6%22%7D%2C%7B%22time%22%3A1594259000%2C%22kw%22%3A%22%E8%93%9D%E8%8B%B1%E8%A3%85%E5%A4%87%20%E5%8D%93%E8%83%9C%E5%BE%AE%22%7D%2C%7B%22time%22%3A1594259552%2C%22kw%22%3A%22%E8%93%9D%E8%8B%B1%E8%A3%85%E5%A4%87%20%E5%8D%93%E8%83%9C%E5%BE%AE%20%E6%96%AF%E8%BE%BE%E5%8D%8A%E5%AF%BC%22%7D%2C%7B%22time%22%3A1594262518%2C%22kw%22%3A%22%E7%A7%91%E7%91%9E%E6%8A%80%E6%9C%AF%22%7D%2C%7B%22time%22%3A1594262736%2C%22kw%22%3A%22%E5%88%86%E6%97%B6%20%E5%9D%87%E7%BA%BF%22%7D%2C%7B%22time%22%3A1594274450%2C%22kw%22%3A%22rocketmq%20%E9%A1%BA%E5%BA%8F%E6%B6%88%E8%B4%B9%20%E6%AD%BB%E4%BF%A1%E9%98%9F%E5%88%97%22%7D%2C%7B%22time%22%3A1607579577%2C%22kw%22%3A%22kotlin%20%E6%B3%9B%E5%9E%8B%22%7D%5D&_t=1620385213065&req=2&csor=0 HTTP/1.1
Host: www.baidu.com
Connection: keep-alive
Pragma: no-cache
Cache-Control: no-cache
sec-ch-ua: " Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"
Accept: application/json, text/javascript, */*; q=0.01
sec-ch-ua-mobile: ?0
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36
Sec-Fetch-Site: same-origin
Sec-Fetch-Mode: cors
Sec-Fetch-Dest: empty
Referer: https://www.baidu.com/
Accept-Encoding: gzip, deflate, br
Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
Cookie: BIDUPSID=43A407FED4DEFC3DC8EBADF59847476C; PSTM=1584362439; BD_UPN=123253; BDUSS=mtVYn4tWFhsc3I3SVczMk9peGM4TGxVMklmZDMxejF6a21HRFJlc3djVlVoTmRmRVFBQUFBJCQAAAAAAAAAAAEAAABjKgNJAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFT3r19U969fc; BDUSS_BFESS=mtVYn4tWFhsc3I3SVczMk9peGM4TGxVMklmZDMxejF6a21HRFJlc3djVlVoTmRmRVFBQUFBJCQAAAAAAAAAAAEAAABjKgNJAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFT3r19U969fc; BAIDUID=69A6ABF9FDB1A5C8FE0BC3839CA4779F:FG=1; MCITY=-%3A; ispeed_lsm=2; __yjs_duid=1_c23be3bf312f69ade5c14e48f7c64d331619611721600; BAIDUID_BFESS=69A6ABF9FDB1A5C8FE0BC3839CA4779F:FG=1; Hm_lvt_aec699bb6442ba076c8981c6dc490771=1619316622,1619401184,1619508729,1619866504; BD_HOME=1; H_PS_PSSID=33985_31254_33848_33758_33607_26350; BDRCVFR[feWj1Vr5u3D]=I67x6TjHwwYf0; delPer=0; BD_CK_SAM=1; PSINO=2; BDORZ=B490B5EBF6F3CD402E515D22BCDA1598; rsv_jmp_slow=1620286467475; shifen[251732087453_47469]=1620355164; BCLID=11108962122801013026; BDSFRCVID=nMIOJeC62iveCYrerWQ8bCcWlm8614TTH6f3t1EO2xyM1F6faJ5uEG0PsU8g0Kub67DLogKK0mOTHv-F_2uxOjjg8UtVJeC6EG0Ptf8g0f5; H_BDCLCKID_SF=tR4toDPaJCL3H48k-4QEbbQH-UnLqMjlLgOZ04n-ah05bKow-PoM24IO-4b2L-7LW23BbnOm3UTdfh76Wh35K5tTQP6rLf6N3Hb4KKJxbpb_hR5l0KcFXT8nhUJiB5OLBan7_qvIXKohJh7FM4tW3J0ZyxomtfQxtNRJ0DnjtpChbCLljjKaDToM5pJfetQe2CvXsJO8fMJEsl7_bf--D6c0XfvIt4QJ3jvi3qOJQb4bjl7D2xnxy5K_hUbIQULD-JCq_J5DLxO2856HQT3m345bbN3i-CrrMGQlWb3cWKJV8UbS5CcPBTD02-nBat-OQ6npaJ5nJq5nhMJmb67JDMr0eGKtq6-qtR4jVbDMt6rajtOd5tTD-tRH-UnLq-Qp22OZ0l8KtJR1sx8mhPoky5F8-4b2L-7LWbIHhxomWIQHDnC65b6jQ6ODMxvmh6v45ar4KKJxahCWeIJo5t5ObxtkhUJiB5OLBan7Lj6IXKohJh7FM4tW3J0ZyxomtfQxtNRJ0DnjtnLhbCDr-R-_-4_tbh_X5-RLfKOyop7F54nKDp0Re-7_M4LLM45a5fJ2QKtJ_xTwMn7xsMTs5MnbWh8yKabr0MTrQeQ-5KQN3KJmfM865-RsBIukyhOb2-biWbRL2MbdJqvP_IoG2Mn8M4bb3qOpBtQmJeTxoUJ25DnJhhCGe6-KD6o-ja8eqbTtKC_XBnrEat3SK4bvK5R_XfFgyxomtjj0bDTmh4oH-f3GbboJ-pJShjvBDHOnLUkq5J7Jol5vttnB8DjDj4jHy6L0QttjQnJPfIkja-5tJD3U8b7TyU42hf47yhDL0q4Hb6b9BJcjfU5MSlcNLTjpQT8r5MDOK5OhJRQ2QJ8BtCDhhC5P; BCLID_BFESS=11108962122801013026; BDSFRCVID_BFESS=nMIOJeC62iveCYrerWQ8bCcWlm8614TTH6f3t1EO2xyM1F6faJ5uEG0PsU8g0Kub67DLogKK0mOTHv-F_2uxOjjg8UtVJeC6EG0Ptf8g0f5; H_BDCLCKID_SF_BFESS=tR4toDPaJCL3H48k-4QEbbQH-UnLqMjlLgOZ04n-ah05bKow-PoM24IO-4b2L-7LW23BbnOm3UTdfh76Wh35K5tTQP6rLf6N3Hb4KKJxbpb_hR5l0KcFXT8nhUJiB5OLBan7_qvIXKohJh7FM4tW3J0ZyxomtfQxtNRJ0DnjtpChbCLljjKaDToM5pJfetQe2CvXsJO8fMJEsl7_bf--D6c0XfvIt4QJ3jvi3qOJQb4bjl7D2xnxy5K_hUbIQULD-JCq_J5DLxO2856HQT3m345bbN3i-CrrMGQlWb3cWKJV8UbS5CcPBTD02-nBat-OQ6npaJ5nJq5nhMJmb67JDMr0eGKtq6-qtR4jVbDMt6rajtOd5tTD-tRH-UnLq-Qp22OZ0l8KtJR1sx8mhPoky5F8-4b2L-7LWbIHhxomWIQHDnC65b6jQ6ODMxvmh6v45ar4KKJxahCWeIJo5t5ObxtkhUJiB5OLBan7Lj6IXKohJh7FM4tW3J0ZyxomtfQxtNRJ0DnjtnLhbCDr-R-_-4_tbh_X5-RLfKOyop7F54nKDp0Re-7_M4LLM45a5fJ2QKtJ_xTwMn7xsMTs5MnbWh8yKabr0MTrQeQ-5KQN3KJmfM865-RsBIukyhOb2-biWbRL2MbdJqvP_IoG2Mn8M4bb3qOpBtQmJeTxoUJ25DnJhhCGe6-KD6o-ja8eqbTtKC_XBnrEat3SK4bvK5R_XfFgyxomtjj0bDTmh4oH-f3GbboJ-pJShjvBDHOnLUkq5J7Jol5vttnB8DjDj4jHy6L0QttjQnJPfIkja-5tJD3U8b7TyU42hf47yhDL0q4Hb6b9BJcjfU5MSlcNLTjpQT8r5MDOK5OhJRQ2QJ8BtCDhhC5P; COOKIE_SESSION=1558_0_9_2_35_51_0_2_9_5_0_2_490342_0_220_0_1620356851_0_1620356631%7C9%232062190_306_1620355073%7C9; BA_HECTOR=ahah21ag0g242l24d41g9a7d70q; sugstore=1

`
	buf := newBuffer(req)
	dec := newDecoder()
	frame, err := dec.Decode(buf)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(frame)
	}
}

func TestDecodePost(t *testing.T) {
	req := `
POST /v1/list HTTP/1.1
Host: mcs.snssdk.com
Connection: keep-alive
Content-Length: 197
Pragma: no-cache
Cache-Control: no-cache
sec-ch-ua: " Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"
X-MCS-AppKey: 566f58151b0ed37e
sec-ch-ua-mobile: ?0
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36
Content-Type: application/json; charset=UTF-8
Accept: */*
Origin: https://www.baidu.com
Sec-Fetch-Site: cross-site
Sec-Fetch-Mode: cors
Sec-Fetch-Dest: empty
Referer: https://www.baidu.com/
Accept-Encoding: gzip, deflate, br
Accept-Language: zh-CN,zh;q=0.9,en;q=0.8

[{"events":[{"event":"onload","params":"{\"app_id\":4453,\"app_name\":\"\",\"sdk_version\":\"4.1.25\"}","local_time_ms":1620369519039}],"user":{"user_unique_id":"6919401372941993480"},"header":{}}]
`
	buf := newBuffer(req)
	dec := newDecoder()
	frame, err := dec.Decode(buf)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(frame)
		t.Log(frame.Payload().String())
	}
}

func TestEncode(t *testing.T) {
	ident := &netx.Identifier{
		IsResponse: true,
		Codec:      uint32(netx.CodecTypeJson),
	}
	header := netx.NewHeader()
	data, _ := json.Marshal(map[string]string{"aa": "aa"})
	buf := bytex.NewBuffer()
	_ = buf.Append(data)
	_, _ = buf.Seek(0, io.SeekStart)
	f := netx.NewFrame(netx.FrameTypeHeader, true, 0, ident, header, buf)
	enc := encoder{}
	b, err := enc.Encode(f)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println()
		fmt.Printf("%s\n", b.String())
	}
}
