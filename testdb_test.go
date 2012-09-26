package testdb

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

var d *Driver
var c *Conn

func init() {
	if d == nil {
		c = NewConn()
		d = &Driver{}
		sql.Register("testdb", d)
	}
}

// Driver
func TestSetOpen(t *testing.T) {
	d.SetOpen(func(dsn string) (driver.Conn, error) {
		return c, errors.New("test error")
	})
	defer d.SetOpen(nil)

	db, _ := sql.Open("testdb", "foo")
	conn, err := db.Driver().Open("foo")

	if db == nil {
		t.Fatal("driver.Open not properly set: db was nil")
	}

	if conn != c {
		t.Fatal("driver.Open not properly set: db was not returned properly")
	}

	if err.Error() != "test error" {
		t.Fatal("driver.Open not properly set: err was not returned properly")
	}
}

func TestStubQuery(t *testing.T) {
	conn := NewConn()

	d.SetConnection(conn)
	db, _ := sql.Open("testdb", "")

	sql := "select count(*) from foo"
	columns := []string{"count"}
	result := `
  5
  `
	conn.StubQuery(sql, RowsFromCSVString(columns, result))

	res, err := db.Query(sql)

	if res.Next() {
		var count int64
		err = res.Scan(&count)

		if err != nil {
			t.Fatal(err)
		}

		if count != 5 {
			t.Fatal("failed to return count")
		}
	}
}

func TestUnknownQuery(t *testing.T) {
	conn := NewConn()
	d.SetConnection(conn)
	db, _ := sql.Open("testdb", "")

	sql := "select count(*) from foobar"
	_, err := db.Query(sql)

	if err == nil {
		t.Fatal("Unknown queries should fail")
	}

}
