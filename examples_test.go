package testdb

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type user struct {
	id      int64
	name    string
	age     int64
	created time.Time
}

func ExampleSetOpenFunc() {
	defer Reset()

	SetOpenFunc(func(dsn string) (driver.Conn, error) {
		// Conn() will return the same internal driver.Conn being used by the driver
		return Conn(), errors.New("test error")
	})

	// err only returns from this if it's an unknown driver, we are stubbing opening a connection
	db, _ := sql.Open("testdb", "foo")
	_, err := db.Driver().Open("foo")

	if err != nil {
		fmt.Println("Stubbed error returned as expected: " + err.Error())
	}

	// Output:
	// Stubbed error returned as expected: test error
}

func ExampleRowsFromCSVString() {
	columns := []string{"id", "name", "age", "created"}
	result := `
  1,tim,20,2012-10-01 01:00:01
  2,joe,25,2012-10-02 02:00:02
  3,bob,30,2012-10-03 03:00:03
  `
	rows := RowsFromCSVString(columns, result)

	fmt.Println(rows.Columns())

	// Output:
	// [id name age created]
}

func ExampleStubQuery() {
	defer Reset()

	db, _ := sql.Open("testdb", "")

	sql := "select id, name, age from users"
	columns := []string{"id", "name", "age", "created"}
	result := `
  1,tim,20,2012-10-01 01:00:01
  2,joe,25,2012-10-02 02:00:02
  3,bob,30,2012-10-03 03:00:03
  `
	StubQuery(sql, RowsFromCSVString(columns, result))

	res, _ := db.Query(sql)

	for res.Next() {
		var u = new(user)
		res.Scan(&u.id, &u.name, &u.age, &u.created)

		fmt.Println(u.name + " - " + strconv.FormatInt(u.age, 10))
	}

	// Output:
	// tim - 20
	// joe - 25
	// bob - 30
}

func ExampleStubQuery_queryRow() {
	defer Reset()

	db, _ := sql.Open("testdb", "")

	sql := "select id, name, age from users"
	columns := []string{"id", "name", "age", "created"}
	result := `
  1,tim,20,2012-10-01 01:00:01
  `
	StubQuery(sql, RowsFromCSVString(columns, result))

	row := db.QueryRow(sql)

	u := new(user)
	row.Scan(&u.id, &u.name, &u.age, &u.created)

	fmt.Println(u.name + " - " + strconv.FormatInt(u.age, 10))

	// Output:
	// tim - 20
}

func ExampleStubQueryError() {
	defer Reset()

	db, _ := sql.Open("testdb", "")

	sql := "select count(*) from error"

	StubQueryError(sql, errors.New("test error"))

	_, err := db.Query(sql)

	if err != nil {
		fmt.Println("Error returned: " + err.Error())
	}

	// Output:
	// Error returned: test error
}

func ExampleSetQueryFunc() {
	defer Reset()

	columns := []string{"id", "name", "age", "created"}
	rows := "1,tim,20,2012-10-01 01:00:01\n2,joe,25,2012-10-02 02:00:02\n3,bob,30,2012-10-03 03:00:03"

	SetQueryFunc(func(query string) (result driver.Rows, err error) {
		return RowsFromCSVString(columns, rows), nil
	})

	db, _ := sql.Open("testdb", "")

	res, _ := db.Query("SELECT foo FROM bar")

	for res.Next() {
		var u = new(user)
		res.Scan(&u.id, &u.name, &u.age, &u.created)

		fmt.Println(u.name + " - " + strconv.FormatInt(u.age, 10))
	}

	// Output:
	// tim - 20
	// joe - 25
	// bob - 30
}

func ExampleSetQueryFunc_queryRow() {
	defer Reset()

	columns := []string{"id", "name", "age", "created"}
	rows := "1,tim,20,2012-10-01 01:00:01"

	SetQueryFunc(func(query string) (result driver.Rows, err error) {
		return RowsFromCSVString(columns, rows), nil
	})

	db, _ := sql.Open("testdb", "")

	row := db.QueryRow("SELECT foo FROM bar")

	var u = new(user)
	row.Scan(&u.id, &u.name, &u.age, &u.created)

	fmt.Println(u.name + " - " + strconv.FormatInt(u.age, 10))

	// Output:
	// tim - 20
}
