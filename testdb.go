package testdb

import (
	"bytes"
	"crypto/sha1"
	"database/sql/driver"
	"encoding/csv"
	"errors"
	"io"
	"regexp"
	"strings"
)

type opener func(dsn string) (driver.Conn, error)

type Driver struct {
	open opener
	conn *Conn
}

func (d *Driver) Open(dsn string) (driver.Conn, error) {
	if d.open != nil {
		conn, err := d.open(dsn)
		return conn, err
	}

	if d.conn == nil {
		d.conn = NewConn()
	}

	return d.conn, nil
}

func (d *Driver) SetOpen(f opener) {
	d.open = f
}

func (d *Driver) SetConnection(conn *Conn) {
	d.conn = conn
}

type Conn struct {
	queries map[string]Query
}

func NewConn() *Conn {
	return &Conn{
		queries: make(map[string]Query),
	}
}

var whitespaceRegexp = regexp.MustCompile("\\s")

func (c *Conn) getQueryHash(query string) string {
	// Remove whitespace and lowercase to make stubbing less brittle
	query = strings.ToLower(whitespaceRegexp.ReplaceAllString(query, ""))

	h := sha1.New()
	io.WriteString(h, query)

	return string(h.Sum(nil))
}

func (c *Conn) StubQuery(query string, result driver.Rows) {
	c.queries[c.getQueryHash(query)] = Query{
		result: result,
	}
}

func (c *Conn) StubQueryError(query string, err error) {
	c.queries[c.getQueryHash(query)] = Query{
		err: err,
	}
}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	if q, ok := c.queries[c.getQueryHash(query)]; ok {
		return &Stmt{
			result: q.result,
			err:    q.err,
		}, nil
	}

	return &Stmt{}, errors.New("Query not stubbed: " + query)
}

func (*Conn) Close() error {
	return nil
}

func (*Conn) Begin() (driver.Tx, error) {
	return &Tx{}, nil
}

func (c *Conn) Exec(query string, args []driver.Value) (driver.Result, error) {
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
	rows    [][]string
	pos     int
}

func (rs *Rows) Next(dest []driver.Value) error {
	rs.pos++
	if rs.pos > len(rs.rows) {
		rs.closed = true

		return io.EOF // per interface spec
	}

	for i, col := range rs.rows[rs.pos-1] {
		b := bytes.NewBufferString(col)

		dest[i] = b.Bytes()
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

		for i, v := range r {
			r[i] = strings.TrimSpace(v)
		}

		rows.rows = append(rows.rows, r)
	}

	return rows
}
