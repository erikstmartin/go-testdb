package testdb

import (
	"crypto/sha1"
	"database/sql"
	"database/sql/driver"
	"encoding/csv"
	"errors"
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

type conn struct {
	queries   map[string]query
	queryFunc func(query string) (result driver.Rows, err error)
}

func newConn() *conn {
	return &conn{
		queries: make(map[string]query),
	}
}

var whitespaceRegexp = regexp.MustCompile("\\s")

func getQueryHash(query string) string {
	// Remove whitespace and lowercase to make stubbing less brittle
	query = strings.ToLower(whitespaceRegexp.ReplaceAllString(query, ""))

	h := sha1.New()
	io.WriteString(h, query)

	return string(h.Sum(nil))
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	if c.queryFunc != nil {
		result, err := c.queryFunc(query)

		return &stmt{
			result: result,
			err:    err,
		}, nil
	}

	if q, ok := d.conn.queries[getQueryHash(query)]; ok {
		return &stmt{
			result: q.result,
			err:    q.err,
		}, nil
	}

	return new(stmt), errors.New("Query not stubbed: " + query)
}

func (*conn) Close() error {
	return nil
}

func (*conn) Begin() (driver.Tx, error) {
	return &tx{}, nil
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, nil
}

type stmt struct {
	result driver.Rows
	err    error
}

func (*stmt) Close() error {
	return nil
}

func (*stmt) NumInput() int {
	return 0
}

func (*stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.result, s.err
}

type tx struct {
}

func (*tx) Commit() error {
	return nil
}

func (*tx) Rollback() error {
	return nil
}

type rows struct {
	closed  bool
	columns []string
	rows    [][]driver.Value
	pos     int
}

func (rs *rows) Next(dest []driver.Value) error {
	rs.pos++
	if rs.pos > len(rs.rows) {
		rs.closed = true

		return io.EOF // per interface spec
	}

	for i, col := range rs.rows[rs.pos-1] {
		dest[i] = col
	}

	return nil
}

func (rs *rows) Err() error {
	return nil
}

func (rs *rows) Columns() []string {
	return rs.columns
}

func (rs *rows) Close() error {
	return nil
}

type query struct {
	result driver.Rows
	err    error
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
