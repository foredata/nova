package slice

// Distinct 字符串去重
func Distinct(str []string) []string {
	if len(str) < 1 {
		return str
	}

	result := make([]string, 0, len(str))
	dict := make(map[string]struct{}, len(str))
	for _, v := range str {
		if _, ok := dict[v]; !ok {
			result = append(result, v)
			dict[v] = struct{}{}
		}
	}

	return result
}

// Paging 分页
func Paging(max int, step int, cb func(start, end int) error) {
	for i := 0; i < max; i += step {
		end := i + step
		if end > max {
			end = max
		}

		if err := cb(i, end); err != nil {
			break
		}
	}
}
