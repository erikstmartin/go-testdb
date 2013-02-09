package testdb

import (
	"crypto/sha1"
	"database/sql"
	"database/sql/driver"
	"encoding/csv"
	"io"
	"regexp"
	"strings"
	"time"
)

var d *testDriver

func init() {
	d = newDriver()
	sql.Register("testdb", d)
}

type testDriver struct {
	open func(dsn string) (driver.Conn, error)
	conn *conn
}

type query struct {
	result driver.Rows
	err    error
}

func newDriver() *testDriver {
	return &testDriver{
		conn: newConn(),
	}
}

func (d *testDriver) Open(dsn string) (driver.Conn, error) {
	if d.open != nil {
		conn, err := d.open(dsn)
		return conn, err
	}

	if d.conn == nil {
		d.conn = newConn()
	}

	return d.conn, nil
}

var whitespaceRegexp = regexp.MustCompile("\\s")

func getQueryHash(query string) string {
	// Remove whitespace and lowercase to make stubbing less brittle
	query = strings.ToLower(whitespaceRegexp.ReplaceAllString(query, ""))

	h := sha1.New()
	io.WriteString(h, query)

	return string(h.Sum(nil))
}

// Set your own function to be executed when db.Query() is called. As with StubQuery() you can use the RowsFromCSVString() method to easily generate the driver.Rows, or you can return your own.
func SetQueryFunc(f func(query string) (result driver.Rows, err error)) {
	d.conn.queryFunc = f
}

// Stubs the global driver.Conn to return the supplied driver.Rows when db.Query() is called, query stubbing is case insensitive, and whitespace is also ignored.
func StubQuery(q string, result driver.Rows) {
	d.conn.queries[getQueryHash(q)] = query{
		result: result,
	}
}

// Stubs the global driver.Conn to return the supplied error when db.Query() is called, query stubbing is case insensitive. and whitespace is also ignored.
func StubQueryError(q string, err error) {
	d.conn.queries[getQueryHash(q)] = query{
		err: err,
	}
}

// Set your own function to be executed when db.Open() is called. You can either hand back a valid connection, or an error. Conn() can be used to grab the global Conn object containing stubbed queries.
func SetOpenFunc(f func(dsn string) (driver.Conn, error)) {
	d.open = f
}

// Clears all stubbed queries, and replaced functions.
func Reset() {
	d = newDriver()
}

// Returns a pointer to the global conn object associated with this driver.
func Conn() driver.Conn {
	return d.conn
}

var timeRegex, _ = regexp.Compile(`^\d{4}-\d{2}-\d{2}(\s\d{2}:\d{2}:\d{2})?$`)

// Helper method to create a driver.Rows object to be used as part of stubbing a Query, the csv string can be any string supported by the csv package.
func RowsFromCSVString(columns []string, s string) driver.Rows {
	rs := &rows{
		columns: columns,
		closed:  false,
	}

	r := strings.NewReader(strings.TrimSpace(s))
	csvReader := csv.NewReader(r)

	for {
		r, err := csvReader.Read()

		if err != nil || r == nil {
			break
		}

		row := make([]driver.Value, len(columns))

		for i, v := range r {
			v := strings.TrimSpace(v)

			if timeRegex.MatchString(v) {
				t, _ := time.Parse("2006-01-02 15:04:05", v)
				row[i] = t
			} else {
				row[i] = v
			}
		}

		rs.rows = append(rs.rows, row)
	}

	return rs
}
