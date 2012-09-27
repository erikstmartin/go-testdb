package testdb

import (
	"bytes"
	"database/sql/driver"
	"encoding/csv"
	"errors"
	"io"
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

func (c *Conn) StubQuery(query string, result driver.Rows) {
	c.queries[query] = Query{
		result: result,
	}
}

func (c *Conn) StubQueryError(query string, err error) {
	c.queries[query] = Query{
		err: err,
	}
}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	if q, ok := c.queries[query]; ok {
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
	if rs.pos >= len(rs.rows) {
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

func (rs *Rows) Scan(dest ...interface{}) error {
	return nil
}

func (rs *Rows) Close() error {
	return nil
}

type Value struct {
}

type Query struct {
	result driver.Rows
	err    error
}

func RowsFromCSVString(columns []string, s string) driver.Rows {
	rows := &Rows{
		columns: columns,
	}

	r := strings.NewReader(strings.TrimSpace(s))
	csvReader := csv.NewReader(r)

	rows.rows, _ = csvReader.ReadAll()

	return rows
}
