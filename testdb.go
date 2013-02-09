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

var d *Driver

func init() {
	d = &Driver {
    conn: newConn(),
  }

	sql.Register("testdb", d)
}

type Driver struct {
	open opener
	conn *conn
}

func (d *Driver) Open(dsn string) (driver.Conn, error) {
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
	queries   map[string]Query
	queryFunc func(query string) (result driver.Rows, err error)
}

func newConn() *conn {
	return &conn{
		queries: make(map[string]Query),
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

		return &Stmt{
			result: result,
			err:    err,
		}, nil
	}

	if q, ok := d.conn.queries[getQueryHash(query)]; ok {
		return &Stmt{
			result: q.result,
			err:    q.err,
		}, nil
	}

	return &Stmt{}, errors.New("Query not stubbed: " + query)
}

func (*conn) Close() error {
	return nil
}

func (*conn) Begin() (driver.Tx, error) {
	return &Tx{}, nil
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, nil
}

type Stmt struct {
	result driver.Rows
	err    error
}

func (*Stmt) Close() error {
	return nil
}

func (*Stmt) NumInput() int {
	return 0
}

func (*Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.result, s.err
}

type Tx struct {
}

func (*Tx) Commit() error {
	return nil
}

func (*Tx) Rollback() error {
	return nil
}

type Rows struct {
	closed  bool
	columns []string
	rows    [][]driver.Value
	pos     int
}

func (rs *Rows) Next(dest []driver.Value) error {
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

func (rs *Rows) Err() error {
	return nil
}

func (rs *Rows) Columns() []string {
	return rs.columns
}

func (rs *Rows) Close() error {
	return nil
}

type Query struct {
	result driver.Rows
	err    error
}

func SetQueryFunc(f func(query string) (result driver.Rows, err error)) {
	d.conn.queryFunc = f
}

func StubQuery(query string, result driver.Rows) {
	d.conn.queries[getQueryHash(query)] = Query{
		result: result,
	}
}

func StubQueryError(query string, err error) {
	d.conn.queries[getQueryHash(query)] = Query{
		err: err,
	}
}

func SetOpenFunc(f opener) {
	d.open = f
}

func Reset(){
  d.conn = newConn()
}

func Conn()(driver.Conn){
  return d.conn
}

var timeRegex, _ = regexp.Compile(`^\d{4}-\d{2}-\d{2}(\s\d{2}:\d{2}:\d{2})?$`)

func RowsFromCSVString(columns []string, s string) driver.Rows {
	rows := &Rows{
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

		rows.rows = append(rows.rows, row)
	}

	return rows
}
