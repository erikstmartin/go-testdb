go-testdb
=========

Framework for stubbing responses from go's driver.Driver interface.

This can be used to sit in place of your sql.Db so that you can stub responses for sql calls, and remove database dependencies for your test suite.

This project is in it's infancy, but has worked well for all the use cases i've had so far, and continues to evolve as new scenarios are uncovered. Please feel free to send pull-requests, or submit feature requests if you have scenarios that are not accounted for yet.

## Setup
The first thing that needs to be done is to make go's sql package aware of the new driver. We make sure that Conn is accessible by the rest of program, so that we can use it to stub our queries.

The reason this is not performed automatically by the package, is so that you can stub other conditions like failing to connect to the database.
<pre>
import (
	"database/sql"
	"github.com/erikstmartin/go-testdb"
)

var c *Conn

func init() {
	c := testdb.NewConn()
	d := &testdb.Driver{}
	d.SetConnection(conn)
	
	sql.Register("testdb", d)
}
</pre>

## Stubbing connect failure
You're able to set your own function to execute when the sql library calls sql.Open
<pre>
d.SetOpen(func(dsn string) (driver.Conn, error) {
	return c, errors.New("failed to connect")
})
</pre>

## Stubbing queries
You're able to stub responses to known queries, unknown queries will trigger log errors so that you can see that queries were executed that were not stubbed.

Differences in whitespace, and case are ignored.

For convenience a method has been created for you to take a CSV string and turn it into a database result object (RowsFromCSVString).

<pre>
db, _ := sql.Open("testdb", "")

sql := "select id, name, age from users"
columns := []string{"id", "name", "age", "created"}
result := `
1,tim,20,2012-10-01 01:00:01
2,joe,25,2012-10-02 02:00:02
3,bob,30,2012-10-03 03:00:03
`
conn.StubQuery(sql, RowsFromCSVString(columns, result))

res, err := db.Query(sql)
</pre>

## Stubbing Query function
Some times you need more control over Query being run, maybe you need to assert whether or not a particular query is run.

You can return either a driver.Rows for response (your own or utilize RowsFromCSV) or an error to be returned
<pre>
conn.SetQueryFunc(func(query string) (result driver.Rows, err error) {
	columns := []string{"id", "name", "age", "created"}
	result := `
1,tim,20,2012-10-01 01:00:01
2,joe,25,2012-10-02 02:00:02
3,bob,30,2012-10-03 03:00:03

	// inspect query to ensure it matches a pattern, or anything else you want to do first
	return RowsFromCSVString(columns, rows), nil
})

db, _ := sql.Open("testdb", "")

res, err := db.Query("SELECT foo FROM bar")
</pre>

## Stubbing errors returned from queries
In case you need to stub errors returned from queries to ensure your code handles them properly

<pre>
db, _ := sql.Open("testdb", "")

sql := "select count(*) from error"
conn.StubQueryError(sql, errors.New("test error"))

res, err := db.Query(sql)
</pre>
