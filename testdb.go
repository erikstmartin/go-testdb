package testdb

import (
	"database/sql/driver"
)

type Driver struct {
}

func (d *Driver) Open(dsn string) (driver.Conn, error) {
	return &Conn{}, nil
}

type Conn struct {
}

func (*Conn) Prepare(query string) (driver.Stmt, error) {
	return &Stmt{}, nil
}

func (*Conn) Close() error {
	return nil
}

func (*Conn) Begin() (driver.Tx, error) {
	return &Tx{}, nil
}

type Stmt struct {
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

func (*Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, nil
}

type Tx struct {
}

func (*Tx) Commit() error {
	return nil
}

func (*Tx) Rollback() error {
	return nil
}
