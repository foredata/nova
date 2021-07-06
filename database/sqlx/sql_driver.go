package sqlx

type sqlDriver struct {
	rawDriverName string // 原生driver name
}

func (d *sqlDriver) Open(dataSourceName string, opts *OpenOptions) (Conn, error) {
	return newSqlConn(d.rawDriverName, dataSourceName, opts)
}
