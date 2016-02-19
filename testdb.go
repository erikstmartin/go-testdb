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
	openFunc          func(dsn string) (driver.Conn, error)
	connections       map[string]*conn
	enableTimeParsing bool
}

type query struct {
	rows   driver.Rows
	result *Result
	err    error
}

func newDriver() *testDriver {
	return &testDriver{
		connections: make(map[string]*conn),
	}
}

func EnableTimeParsing(flag bool) {
	d.enableTimeParsing = flag
}

func (d *testDriver) Open(dsn string) (driver.Conn, error) {
	if d.openFunc != nil {
		conn, err := d.openFunc(dsn)
		return conn, err
	}

	if _, ok := d.connections[dsn]; !ok {
		d.connections[dsn] = newConn()
	}

	return d.connections[dsn], nil
}

// Clears all stubbed queries, and replaced functions.
func Reset() {
	d.connections = make(map[string]*conn)
	d.openFunc = nil
}

// Returns a pointer to the conn object associated with this dsn
func Connection(dsn string) *conn {
	if _, ok := d.connections[dsn]; !ok {
		d.connections[dsn] = newConn()
	}

	return d.connections[dsn]
}

func DefaultConnection() *conn {
	return Connection("")
}

// Left for backwards compatibility
func Conn() driver.Conn {
	return DefaultConnection()
}

var whitespaceRegexp = regexp.MustCompile("\\s")

func getQueryHash(query string) string {
	// Remove whitespace and lowercase to make stubbing less brittle
	query = strings.ToLower(whitespaceRegexp.ReplaceAllString(query, ""))

	h := sha1.New()
	io.WriteString(h, query)

	return string(h.Sum(nil))
}

// Set your own function to be executed when db.Open() is called. You can either hand back a valid connection, or an error. Conn() can be used to grab the global Conn object containing stubbed queries.
func SetOpenFunc(f func(dsn string) (driver.Conn, error)) {
	d.openFunc = f
}

// These are here for backwards compatibility
// Set your own function to be executed when db.Query() is called. As with StubQuery() you can use the RowsFromCSVString() method to easily generate the driver.Rows, or you can return your own.
func SetQueryFunc(f func(query string) (result driver.Rows, err error)) {
	DefaultConnection().SetQueryFunc(f)
}

// Set your own function to be executed when db.Query() is called. As with StubQuery() you can use the RowsFromCSVString() method to easily generate the driver.Rows, or you can return your own.
func SetQueryWithArgsFunc(f func(query string, args []driver.Value) (result driver.Rows, err error)) {
	DefaultConnection().SetQueryWithArgsFunc(f)
}

// Stubs the global driver.Conn to return the supplied driver.Rows when db.Query() is called, query stubbing is case insensitive, and whitespace is also ignored.
func StubQuery(q string, rows driver.Rows) {
	DefaultConnection().StubQuery(q, rows)
}

// Stubs the global driver.Conn to return the supplied error when db.Query() is called, query stubbing is case insensitive, and whitespace is also ignored.
func StubQueryError(q string, err error) {
	DefaultConnection().StubQueryError(q, err)
}

// Set your own function to be executed when db.Exec is called. You can return an error or a Result object with the LastInsertId and RowsAffected
func SetExecFunc(f func(query string) (driver.Result, error)) {
	DefaultConnection().SetExecFunc(f)
}

// Set your own function to be executed when db.Exec is called. You can return an error or a Result object with the LastInsertId and RowsAffected
func SetExecWithArgsFunc(f func(query string, args []driver.Value) (driver.Result, error)) {
	DefaultConnection().SetExecWithArgsFunc(f)
}

// Stubs the global driver.Conn to return the supplied Result when db.Exec is called, query stubbing is case insensitive, and whitespace is also ignored.
func StubExec(q string, r *Result) {
	DefaultConnection().StubExec(q, r)
}

// Stubs the global driver.Conn to return the supplied error when db.Exec() is called, query stubbing is case insensitive, and whitespace is also ignored.
func StubExecError(q string, err error) {
	DefaultConnection().StubExecError(q, err)
}

// Set your own function to be executed when db.Begin() is called. You can either hand back a valid transaction, or an error. Conn() can be used to grab the global Conn object containing stubbed queries.
func SetBeginFunc(f func() (driver.Tx, error)) {
	DefaultConnection().SetBeginFunc(f)
}

// Stubs the global driver.Conn to return the supplied tx and error when db.Begin() is called.
func StubBegin(tx driver.Tx, err error) {
	DefaultConnection().StubBegin(tx, err)
}

// Set your own function to be executed when tx.Commit() is called on the default transcation. Conn() can be used to grab the global Conn object containing stubbed queries.
func SetCommitFunc(f func() error) {
	DefaultConnection().SetCommitFunc(f)
}

// Stubs the default transaction to return the supplied error when tx.Commit() is called.
func StubCommitError(err error) {
	DefaultConnection().StubCommitError(err)
}

// Set your own function to be executed when tx.Rollback() is called on the default transcation. Conn() can be used to grab the global Conn object containing stubbed queries.
func SetRollbackFunc(f func() error) {
	DefaultConnection().SetRollbackFunc(f)
}

// Stubs the default transaction to return the supplied error when tx.Rollback() is called.
func StubRollbackError(err error) {
	DefaultConnection().StubRollbackError(err)
}

func RowsFromCSVString(columns []string, s string, c ...rune) driver.Rows {
	r := strings.NewReader(strings.TrimSpace(s))
	csvReader := csv.NewReader(r)
	if len(c) > 0 {
		csvReader.Comma = c[0]
	}

	rows := [][]driver.Value{}
	for {
		r, err := csvReader.Read()

		if err != nil || r == nil {
			break
		}

		row := make([]driver.Value, len(columns))

		for i, v := range r {
			v := strings.TrimSpace(v)

			// If enableTimeParsing is on, check to see if this is a
			// time in RFC33339 format
			if d.enableTimeParsing {
				if time, err := time.Parse(time.RFC3339, v); err == nil {
					row[i] = time
				} else {
					row[i] = v
				}
			} else {
				row[i] = v
			}
		}

		rows = append(rows, row)
	}

	return RowsFromSlice(columns, rows)
}

func RowsFromSlice(columns []string, data [][]driver.Value) driver.Rows {
	return &rows{
		closed:  false,
		columns: columns,
		rows:    data,
		pos:     0,
	}
}
