# sqlx(WIP)

## 简介

sqlx主要用于扩展标准sql,统一不同平台下差异，区别于gorm等orm库，这里只会提供简单单表CRUD。

支持的特性有:
- 参数绑定:
    - mysql: 使用?
    - postgres: 使用$1, $2
    - oracle: 使用:arg1, :arg2
    - sqlserver: 使用 @p1, @p2
- IN数组参数绑定:
- 查询结果绑定struct,map,slice

## 使用方法

```go

```

## 其他

- [sqlx](https://github.com/jmoiron/sqlx)
- [Illustrated guide to SQLX](https://jmoiron.github.io/sqlx/)
- [sqlx库使用指南](https://www.liwenzhou.com/posts/Go/sqlx/)
- [gorm](https://github.com/go-gorm/gorm)
- [xorm](https://gitea.com/xorm/xorm)
- [scan](https://github.com/blockloop/scan)
- [scany](https://github.com/georgysavva/scany)
- [db](https://github.com/upper/db)
- [mongo upsert gnore some fields](https://stackoverflow.com/questions/46886110/mongo-update-to-ignore-some-fields-or-upsert)
- [upsert](https://wiki.postgresql.org/wiki/UPSERT#.22UPSERT.22_definition)
- [ramsql for Test](https://github.com/proullon/ramsql)
- [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)
- [vitess](https://github.com/vitessio/vitess)
- [squirrel](https://github.com/Masterminds/squirrel)