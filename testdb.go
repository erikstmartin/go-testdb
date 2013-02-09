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

type opener func(dsn string) (driver.Conn, error)

var d *testDriver

func init() {
	d = newDriver()
	sql.Register("testdb", d)
}

type testDriver struct {
	open opener
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

func SetQueryFunc(f func(query string) (result driver.Rows, err error)) {
	d.conn.queryFunc = f
}

func StubQuery(q string, result driver.Rows) {
	d.conn.queries[getQueryHash(q)] = query{
		result: result,
	}
}

func StubQueryError(q string, err error) {
	d.conn.queries[getQueryHash(q)] = query{
		err: err,
	}
}

func SetOpenFunc(f opener) {
	d.open = f
}

func Reset() {
	d = newDriver()
}

func Conn() driver.Conn {
	return d.conn
}

var timeRegex, _ = regexp.Compile(`^\d{4}-\d{2}-\d{2}(\s\d{2}:\d{2}:\d{2})?$`)

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
