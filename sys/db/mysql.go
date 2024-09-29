package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lircstar/nemo/sys/log"
)

type MysqlConn struct {
	db             *sql.DB
	driverName     string
	dataSourceName string
}

// New creates a new connection pool.
func newPool(driverName, dataSourceName string) *MysqlConn {
	return &MysqlConn{
		driverName:     driverName,
		dataSourceName: dataSourceName,
	}
}

// Open initializes the database connection.
func (p *MysqlConn) Open() error {
	var err error
	if p.db, err = sql.Open(p.driverName, p.dataSourceName); err != nil {
		return err
	}
	if err = p.db.Ping(); err != nil {
		return err
	}
	return nil
}

// Close closes the database connection.
func (p *MysqlConn) Close() {
	if p.db != nil {
		p.db.Close()
	}
}

// Query executes a query that returns multiple rows.
func (p *MysqlConn) Query(query string, args ...any) (*sql.Rows, error) {
	rows, err := p.db.Query(query, args...)
	if err != nil {
		p.fail("QUERY", query, err, args...)
		return nil, err
	}
	return rows, nil
}

// QueryRow executes a query that returns a single row.
func (p *MysqlConn) QueryRow(query string, args ...any) *sql.Row {
	return p.db.QueryRow(query, args...)
}

// Exec executes a query without returning any rows.
func (p *MysqlConn) Exec(query string, args ...any) (sql.Result, error) {
	res, err := p.db.Exec(query, args...)
	if err != nil {
		p.fail("EXEC", query, err, args...)
		return nil, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		p.fail("EXEC", query, errors.New(fmt.Sprintf("Affected rows:%s", err.Error())), args...)
		return nil, err
	}
	if count != 1 {
		p.fail("EXEC", query, errors.New(fmt.Sprintf("Affected rows:%d", count)), args...)
		return nil, errors.New(fmt.Sprintf("Affected rows:%d", count))
	}
	return res, nil
}

// fail logs the error details.
func (p *MysqlConn) fail(method, query string, err error, args ...any) {
	log.Errorf("Failed to execute SQL [%s][%s] [Parameter]:%v [Error]:%s\n", method, query, args, err.Error())
}

// NewMysql initializes a MySQL connection pool.
func NewMysql(DSN string) *MysqlConn {
	pool := newPool("mysql", DSN)
	if err := pool.Open(); err != nil {
		log.Fatalf("Failed to initialize database, Error:%v", err)
	}
	return pool
}
