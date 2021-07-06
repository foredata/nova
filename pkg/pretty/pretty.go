package pretty

import (
	"bytes"
	"encoding/json"
	"sort"
	"strconv"
)

// Dump .
// [go-spew](https://github.com/davecgh/go-spew)
// TODO: impl
// func Dump() {

// }

// https://github.com/tidwall/pretty
// Indent .
func Indent(v interface{}) string {
	switch x := v.(type) {
	case string:
		d := prettyJson([]byte(x))
		return string(d)
	case []byte:
		d := prettyJson(x)
		return string(d)
	default:
		d, _ := json.MarshalIndent(v, "", "\t")
		if len(d) > 0 {
			d = append(d, '\n')
		}
		return string(d)
	}
}

// Flat .
func Flat(v interface{}) string {
	switch x := v.(type) {
	case string:
		d := uglyInPlace([]byte(x))
		return string(d)
	case []byte:
		d := uglyInPlace(x)
		return string(d)
	default:
		d, _ := json.Marshal(v)
		return string(d)
	}
}

func prettyJson(v []byte) []byte {
	buf := make([]byte, 0, len(v))
	buf, _, _, _ = appendPrettyAny(buf, v, 0, true, 80, "", "\t", false, 0, 0, -1)
	if len(buf) > 0 {
		buf = append(buf, '\n')
	}

	return buf
}

func uglyInPlace(v []byte) []byte {
	return ugly(v, v)
}

func ugly(dst, src []byte) []byte {
	dst = dst[:0]
	for i := 0; i < len(src); i++ {
		if src[i] > ' ' {
			dst = append(dst, src[i])
			if src[i] == '"' {
				for i = i + 1; i < len(src); i++ {
					dst = append(dst, src[i])
					if src[i] == '"' {
						j := i - 1
						for ; ; j-- {
							if src[j] != '\\' {
								break
							}
						}
						if (j-i)%2 != 0 {
							break
						}
					}
				}
			}
		}
	}
	return dst
}

func isNaNOrInf(src []byte) bool {
	return src[0] == 'i' || //Inf
		src[0] == 'I' || // inf
		src[0] == '+' || // +Inf
		src[0] == 'N' || // Nan
		(src[0] == 'n' && len(src) > 1 && src[1] != 'u') // nan
}

func appendPrettyAny(buf, json []byte, i int, pretty bool, width int, prefix, indent string, sortkeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == '"' {
			return appendPrettyString(buf, json, i, nl)
		}

		if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' || isNaNOrInf(json[i:]) {
			return appendPrettyNumber(buf, json, i, nl)
		}
		if json[i] == '{' {
			return appendPrettyObject(buf, json, i, '{', '}', pretty, width, prefix, indent, sortkeys, tabs, nl, max)
		}
		if json[i] == '[' {
			return appendPrettyObject(buf, json, i, '[', ']', pretty, width, prefix, indent, sortkeys, tabs, nl, max)
		}
		switch json[i] {
		case 't':
			return append(buf, 't', 'r', 'u', 'e'), i + 4, nl, true
		case 'f':
			return append(buf, 'f', 'a', 'l', 's', 'e'), i + 5, nl, true
		case 'n':
			return append(buf, 'n', 'u', 'l', 'l'), i + 4, nl, true
		}
	}
	return buf, i, nl, true
}

type pair struct {
	kstart, kend int
	vstart, vend int
}

type byKeyVal struct {
	sorted bool
	json   []byte
	buf    []byte
	pairs  []pair
}

func (arr *byKeyVal) Len() int {
	return len(arr.pairs)
}
func (arr *byKeyVal) Less(i, j int) bool {
	if arr.isLess(i, j, byKey) {
		return true
	}
	if arr.isLess(j, i, byKey) {
		return false
	}
	return arr.isLess(i, j, byVal)
}
func (arr *byKeyVal) Swap(i, j int) {
	arr.pairs[i], arr.pairs[j] = arr.pairs[j], arr.pairs[i]
	arr.sorted = true
}

type byKind int

const (
	byKey byKind = 0
	byVal byKind = 1
)

type jtype int

const (
	jnull jtype = iota
	jfalse
	jnumber
	jstring
	jtrue
	jjson
)

func getjtype(v []byte) jtype {
	if len(v) == 0 {
		return jnull
	}
	switch v[0] {
	case '"':
		return jstring
	case 'f':
		return jfalse
	case 't':
		return jtrue
	case 'n':
		return jnull
	case '[', '{':
		return jjson
	default:
		return jnumber
	}
}

func (arr *byKeyVal) isLess(i, j int, kind byKind) bool {
	k1 := arr.json[arr.pairs[i].kstart:arr.pairs[i].kend]
	k2 := arr.json[arr.pairs[j].kstart:arr.pairs[j].kend]
	var v1, v2 []byte
	if kind == byKey {
		v1 = k1
		v2 = k2
	} else {
		v1 = bytes.TrimSpace(arr.buf[arr.pairs[i].vstart:arr.pairs[i].vend])
		v2 = bytes.TrimSpace(arr.buf[arr.pairs[j].vstart:arr.pairs[j].vend])
		if len(v1) >= len(k1)+1 {
			v1 = bytes.TrimSpace(v1[len(k1)+1:])
		}
		if len(v2) >= len(k2)+1 {
			v2 = bytes.TrimSpace(v2[len(k2)+1:])
		}
	}
	t1 := getjtype(v1)
	t2 := getjtype(v2)
	if t1 < t2 {
		return true
	}
	if t1 > t2 {
		return false
	}
	if t1 == jstring {
		s1 := parsestr(v1)
		s2 := parsestr(v2)
		return string(s1) < string(s2)
	}
	if t1 == jnumber {
		n1, _ := strconv.ParseFloat(string(v1), 64)
		n2, _ := strconv.ParseFloat(string(v2), 64)
		return n1 < n2
	}
	return string(v1) < string(v2)

}

func parsestr(s []byte) []byte {
	for i := 1; i < len(s); i++ {
		if s[i] == '\\' {
			var str string
			json.Unmarshal(s, &str)
			return []byte(str)
		}
		if s[i] == '"' {
			return s[1:i]
		}
	}
	return nil
}

func appendPrettyObject(buf, json []byte, i int, open, close byte, pretty bool, width int, prefix, indent string, sortkeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	var ok bool
	if width > 0 {
		if pretty && open == '[' && max == -1 {
			// here we try to create a single line array
			max := width - (len(buf) - nl)
			if max > 3 {
				s1, s2 := len(buf), i
				buf, i, _, ok = appendPrettyObject(buf, json, i, '[', ']', false, width, prefix, "", sortkeys, 0, 0, max)
				if ok && len(buf)-s1 <= max {
					return buf, i, nl, true
				}
				buf = buf[:s1]
				i = s2
			}
		} else if max != -1 && open == '{' {
			return buf, i, nl, false
		}
	}
	buf = append(buf, open)
	i++
	var pairs []pair
	if open == '{' && sortkeys {
		pairs = make([]pair, 0, 8)
	}
	var n int
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == close {
			if pretty {
				if open == '{' && sortkeys {
					buf = sortPairs(json, buf, pairs)
				}
				if n > 0 {
					nl = len(buf)
					if buf[nl-1] == ' ' {
						buf[nl-1] = '\n'
					} else {
						buf = append(buf, '\n')
					}
				}
				if buf[len(buf)-1] != open {
					buf = appendTabs(buf, prefix, indent, tabs)
				}
			}
			buf = append(buf, close)
			return buf, i + 1, nl, open != '{'
		}
		if open == '[' || json[i] == '"' {
			if n > 0 {
				buf = append(buf, ',')
				if width != -1 && open == '[' {
					buf = append(buf, ' ')
				}
			}
			var p pair
			if pretty {
				nl = len(buf)
				if buf[nl-1] == ' ' {
					buf[nl-1] = '\n'
				} else {
					buf = append(buf, '\n')
				}
				if open == '{' && sortkeys {
					p.kstart = i
					p.vstart = len(buf)
				}
				buf = appendTabs(buf, prefix, indent, tabs+1)
			}
			if open == '{' {
				buf, i, nl, _ = appendPrettyString(buf, json, i, nl)
				if sortkeys {
					p.kend = i
				}
				buf = append(buf, ':')
				if pretty {
					buf = append(buf, ' ')
				}
			}
			buf, i, nl, ok = appendPrettyAny(buf, json, i, pretty, width, prefix, indent, sortkeys, tabs+1, nl, max)
			if max != -1 && !ok {
				return buf, i, nl, false
			}
			if pretty && open == '{' && sortkeys {
				p.vend = len(buf)
				if p.kstart > p.kend || p.vstart > p.vend {
					// bad data. disable sorting
					sortkeys = false
				} else {
					pairs = append(pairs, p)
				}
			}
			i--
			n++
		}
	}
	return buf, i, nl, open != '{'
}
func sortPairs(json, buf []byte, pairs []pair) []byte {
	if len(pairs) == 0 {
		return buf
	}
	vstart := pairs[0].vstart
	vend := pairs[len(pairs)-1].vend
	arr := byKeyVal{false, json, buf, pairs}
	sort.Stable(&arr)
	if !arr.sorted {
		return buf
	}
	nbuf := make([]byte, 0, vend-vstart)
	for i, p := range pairs {
		nbuf = append(nbuf, buf[p.vstart:p.vend]...)
		if i < len(pairs)-1 {
			nbuf = append(nbuf, ',')
			nbuf = append(nbuf, '\n')
		}
	}
	return append(buf[:vstart], nbuf...)
}

func appendPrettyString(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] == '"' {
			var sc int
			for j := i - 1; j > s; j-- {
				if json[j] == '\\' {
					sc++
				} else {
					break
				}
			}
			if sc%2 == 1 {
				continue
			}
			i++
			break
		}
	}
	return append(buf, json[s:i]...), i, nl, true
}

func appendPrettyNumber(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ':' || json[i] == ']' || json[i] == '}' {
			break
		}
	}
	return append(buf, json[s:i]...), i, nl, true
}

func appendTabs(buf []byte, prefix, indent string, tabs int) []byte {
	if len(prefix) != 0 {
		buf = append(buf, prefix...)
	}
	if len(indent) == 2 && indent[0] == ' ' && indent[1] == ' ' {
		for i := 0; i < tabs; i++ {
			buf = append(buf, ' ', ' ')
		}
	} else {
		for i := 0; i < tabs; i++ {
			buf = append(buf, indent...)
		}
	}
	return buf
}
