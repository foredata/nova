package metadata

import "testing"

func TestMetadata(t *testing.T) {
	md := New()
	md.Add("c", "c")
	md.Add("b", "b")
	md.Add("d", "d")
	md.Add("a", "a")
	t.Log(md.Len())
	md.Walk(func(key string, values []string) bool {
		t.Logf("%s:%+v", key, values)
		return true
	})

	t.Log("---------------")
	md.Del("c")
	md.Walk(func(key string, values []string) bool {
		t.Logf("%s:%+v", key, values)
		return true
	})

	t.Log("-------------")
	md.Add("a", "aa")
	t.Log(md.Values("a"))

	var md1 Metadata
	t.Log(md1.Len())
	md1.Add("a", "a")
	t.Log(md1)
}
