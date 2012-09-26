package testdb

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestQuery(t *testing.T) {
	d := &Driver{}
	sql.Register("testdb", d)

	db, err := sql.Open("testdb", "")

	sql := "select count(*) from foo"
	res, err := db.Query(sql)

	if err != nil {
		fmt.Println(res)
	}
}
