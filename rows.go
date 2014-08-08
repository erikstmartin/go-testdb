package testdb

import (
	"database/sql/driver"
	"io"
)

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
		rs.pos = 0

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
