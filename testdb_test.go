package testdb

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"
)

func ExampleSetOpenFunc() {
	defer Reset()

	SetOpenFunc(func(dsn string) (driver.Conn, error) {
		// Conn() will return the same internal driver.Conn being used by the driver
		return Conn(), errors.New("test error")
	})

	_, err := sql.Open("testdb", "foo")

	if err != nil {
		fmt.Println("Stubbed error returned as expected: " + err.Error())
	}
}

func TestSetOpenFunc(t *testing.T) {
	SetOpenFunc(func(dsn string) (driver.Conn, error) {
		return Conn(), errors.New("test error")
	})
	defer SetOpenFunc(nil)

	db, _ := sql.Open("testdb", "foo")
	conn, err := db.Driver().Open("foo")

	if db == nil {
		t.Fatal("driver.Open not properly set: db was nil")
	}

	if conn == nil {
		t.Fatal("driver.Open not properly set: didn't connection")
	}

	if err.Error() != "test error" {
		t.Fatal("driver.Open not properly set: err was not returned properly")
	}
}

func TestStubQuery(t *testing.T) {
	db, _ := sql.Open("testdb", "")

	sql := "select count(*) from foo"
	columns := []string{"count"}
	result := `
  5
  `
	StubQuery(sql, RowsFromCSVString(columns, result))

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

func TestStubQueryAdditionalWhitespace(t *testing.T) {
	db, _ := sql.Open("testdb", "")

	sqlWhitespace := "select count(*) from              foo"
	sql := "select count(*) from foo"
	columns := []string{"count"}
	result := `
  5
  `
	StubQuery(sqlWhitespace, RowsFromCSVString(columns, result))

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

func TestStubQueryChangeCase(t *testing.T) {
	db, _ := sql.Open("testdb", "")

	sqlCase := "SELECT COUNT(*) FROM foo"
	sql := "select count(*) from foo"
	columns := []string{"count"}
	result := `
  5
  `
	StubQuery(sqlCase, RowsFromCSVString(columns, result))

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
	db, _ := sql.Open("testdb", "")

	sql := "select count(*) from foobar"
	_, err := db.Query(sql)

	if err == nil {
		t.Fatal("Unknown queries should fail")
	}
}

func TestStubQueryError(t *testing.T) {
	db, _ := sql.Open("testdb", "")

	sql := "select count(*) from error"

	StubQueryError(sql, errors.New("test error"))

	res, err := db.Query(sql)

	if err == nil {
		t.Fatal("failed to return error from stubbed query")
	}

	if res != nil {
		t.Fatal("result should be nil on error")
	}
}

type user struct {
	id      int64
	name    string
	age     int64
	created time.Time
}

func TestStubQueryMultipleResult(t *testing.T) {
	db, _ := sql.Open("testdb", "")

	sql := "select id, name, age from users"
	columns := []string{"id", "name", "age", "created"}
	result := `
  1,tim,20,2012-10-01 01:00:01
  2,joe,25,2012-10-02 02:00:02
  3,bob,30,2012-10-03 03:00:03
  `
	StubQuery(sql, RowsFromCSVString(columns, result))

	res, err := db.Query(sql)

	if err != nil {
		t.Fatal("stubbed query should not return error")
	}

	i := 0

	for res.Next() {
		var u = user{}
		err = res.Scan(&u.id, &u.name, &u.age, &u.created)

		if err != nil {
			t.Fatal(err)
		}

		ti := time.Date(2012, 10, i+1, i+1, 0, i+1, 0, time.UTC)

		if u.id == 0 || u.name == "" || u.age == 0 || u.created.Unix() != ti.Unix() {
			t.Fatal("failed to populate object with result")
		}
		i++
	}

	if i != 3 {
		t.Fatal("failed to return proper number of results")
	}
}

func TestStubQueryMultipleResultNewline(t *testing.T) {
	db, _ := sql.Open("testdb", "")

	sql := "select id, name, age from users"
	columns := []string{"id", "name", "age", "created"}
	result := "1,tim,20,2012-10-01 01:00:01\n2,joe,25,2012-10-02 02:00:02\n3,bob,30,2012-10-03 03:00:03"

	StubQuery(sql, RowsFromCSVString(columns, result))

	res, err := db.Query(sql)

	if err != nil {
		t.Fatal("stubbed query should not return error")
	}

	i := 0

	for res.Next() {
		var u = user{}
		err = res.Scan(&u.id, &u.name, &u.age, &u.created)

		if err != nil {
			t.Fatal(err)
		}

		ti := time.Date(2012, 10, i+1, i+1, 0, i+1, 0, time.UTC)

		if u.id == 0 || u.name == "" || u.age == 0 || u.created.Unix() != ti.Unix() {
			t.Fatal("failed to populate object with result")
		}
		i++
	}

	if i != 3 {
		t.Fatal("failed to return proper number of results")
	}
}

func TestSetQueryFunc(t *testing.T) {
	columns := []string{"id", "name", "age", "created"}
	rows := "1,tim,20,2012-10-01 01:00:01\n2,joe,25,2012-10-02 02:00:02\n3,bob,30,2012-10-03 03:00:03"

	SetQueryFunc(func(query string) (result driver.Rows, err error) {
		return RowsFromCSVString(columns, rows), nil
	})

	db, _ := sql.Open("testdb", "")

	res, err := db.Query("SELECT foo FROM bar")

	i := 0

	for res.Next() {
		var u = user{}
		err = res.Scan(&u.id, &u.name, &u.age, &u.created)

		if err != nil {
			t.Fatal(err)
		}

		ti := time.Date(2012, 10, i+1, i+1, 0, i+1, 0, time.UTC)

		if u.id == 0 || u.name == "" || u.age == 0 || u.created.Unix() != ti.Unix() {
			t.Fatal("failed to populate object with result")
		}
		i++
	}

	if i != 3 {
		t.Fatal("failed to return proper number of results")
	}
}

func TestSetQueryFuncError(t *testing.T) {
	SetQueryFunc(func(query string) (result driver.Rows, err error) {
		return nil, errors.New("stubbed error")
	})

	db, _ := sql.Open("testdb", "")

	_, err := db.Query("SELECT foo FROM bar")

	if err == nil {
		t.Fatal("failed to return error from QueryFunc")
	}
}

func TestReset(t *testing.T) {
	sql.Open("testdb", "")

	sql := "select count(*) from error"
	StubQueryError(sql, errors.New("test error"))

	Reset()

	if len(d.conn.queries) > 0 {
		t.Fatal("failed to reset connection")
	}
}
