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

	if err != nil {
		t.Fatal("stubbed query should not return error")
	}

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

func TestStubQueryError(t *testing.T) {
	conn := NewConn()

	d.SetConnection(conn)
	db, _ := sql.Open("testdb", "")

	sql := "select count(*) from error"

	conn.StubQueryError(sql, errors.New("test error"))

	res, err := db.Query(sql)

	if err == nil {
		t.Fatal("failed to return error from stubbed query")
	}

	if res != nil {
		t.Fatal("result should be nil on error")
	}
}

type user struct {
	id   int64
	name string
	age  int64
}

func TestStubQueryMultipleResult(t *testing.T) {
	conn := NewConn()

	d.SetConnection(conn)
	db, _ := sql.Open("testdb", "")

	sql := "select id, name, age from users"
	columns := []string{"id", "name", "age"}
	result := `
  1,tim,20
  2,joe,25
  3,bob,30
  `
	conn.StubQuery(sql, RowsFromCSVString(columns, result))

	res, err := db.Query(sql)

	if err != nil {
		t.Fatal("stubbed query should not return error")
	}

	i := 0

	for res.Next() {
		var u = user{}
		err = res.Scan(&u.id, &u.name, &u.age)

		if err != nil {
			t.Fatal(err)
		}

		if u.id == 0 || u.name == "" || u.age == 0 {
			t.Fatal("failed to populate object with result")
		}
		i++
	}

	if i != 3 {
		t.Fatal("failed to return proper number of results")
	}
}
