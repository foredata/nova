package sqlx

import (
	"context"
	"testing"
)

func TestDriver(t *testing.T) {
	d, err := Open("ramsql", "xx")
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	db, err := d.Database(ctx, "demo")
	if err != nil {
		t.Error(err)
	}

	type User struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}

	userTable := "user"

	if _, err := db.InsertOne(ctx, userTable, &User{ID: "id1", Name: "name1"}); err != nil {
		t.Error(err)
	}

	// NewFilter("name = ? AND age = ?", "id1")
	// ID = ? sqlx.D{"ID"}
	// ID = ? OR (age > ? AND age < ?)
	if _, err := db.UpdateOne(ctx, userTable, M{"ID": "id1"}, M{"name1": "name_updated"}); err != nil {
		t.Error(err)
	}

	// 查询单条数据
	var user User
	if err := db.FindOne(ctx, userTable, M{"ID": "id1"}).Decode(&user); err != nil {
		t.Error(err)
	}

	// 查询多条数据
	var users []*User
	if err := db.Find(ctx, userTable, M{"ID": "id1"}).All(&users); err != nil {
		t.Error(err)
	}

	// 或者使用map
	var userObject []M
	if err := db.Find(ctx, userTable, M{"name": "aa"}).All(&userObject); err != nil {
		t.Error(err)
	}

	if _, err := db.DeleteOne(ctx, userTable, M{"ID": "id1"}); err != nil {
		t.Error(err)
	}

	// db.ExecTx(func(tx Transaction) error {
	// 	if err := db.UpdateOne(ctx, userTable, Eq("ID", "id1"), M{"name1": "name_tx"}); err != nil {
	// 		return err
	// 	}

	// 	return nil
	// })

	// query, err := d.Complie("SELECT * FROM user WHERE ID=? LIMIT 1")
	// if err != nil {
	// 	t.Error(err)
	// }

	// db.Query(ctx, query, "id1")

	// sdb := sql.DB{}
	// tx, _ := sdb.Begin()
	// tx.Exec()
}
