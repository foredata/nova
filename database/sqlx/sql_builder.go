package sqlx

import (
	"strconv"
	"strings"
)

type sqlBuilder struct {
	buf strings.Builder
}

func (b *sqlBuilder) Write(datas ...string) {
	for _, s := range datas {
		b.buf.WriteString(s)
	}
}

func (b *sqlBuilder) WriteInt(d int) {
	v := strconv.Itoa(d)
	b.buf.WriteString(v)
}

func (b *sqlBuilder) String() string {
	return b.buf.String()
}
