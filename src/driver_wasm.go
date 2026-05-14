//go:build js && wasm

package main

import (
	"database/sql/driver"
	"fmt"
	"io"
)

// wasmDriver — driver.Driver

type wasmDriver struct{}

func (d *wasmDriver) Open(_ string) (driver.Conn, error) {
	return &wasmConn{}, nil
}

// wasmConn — driver.Conn

type wasmConn struct{}

func (c *wasmConn) Prepare(query string) (driver.Stmt, error) {
	return &wasmStmt{query: query}, nil
}

func (c *wasmConn) Close() error { return nil }

// Begin is not supported; the JS SQLite layer has no transaction API.
func (c *wasmConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("wasmDriver: transactions not supported")
}

// wasmStmt — driver.Stmt

type wasmStmt struct {
	query string
}

func (s *wasmStmt) Close() error { return nil }

// NumInput returns -1 to tell database/sql to skip argument count checks.
func (s *wasmStmt) NumInput() int { return -1 }

func (s *wasmStmt) Exec(args []driver.Value) (driver.Result, error) {
	env, err := callSQLExec(s.query, args)
	if err != nil {
		return nil, err
	}
	return sqlResult{lastID: env.LastInsertId, rowsAffected: env.RowsAffected}, nil
}

func (s *wasmStmt) Query(args []driver.Value) (driver.Rows, error) {
	env, err := callSQLExec(s.query, args)
	if err != nil {
		return nil, err
	}
	return newWasmRows(env.Rows), nil
}

// sqlResult — driver.Result

type sqlResult struct {
	lastID       int64
	rowsAffected int64
}

func (r sqlResult) LastInsertId() (int64, error) { return r.lastID, nil }
func (r sqlResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

// wasmRows — driver.Rows

type wasmRows struct {
	cols []string
	rows []map[string]any
	pos  int
}

func newWasmRows(rows []map[string]any) *wasmRows {
	var cols []string
	if len(rows) > 0 {
		for k := range rows[0] {
			cols = append(cols, k)
		}
	}
	return &wasmRows{cols: cols, rows: rows}
}

func (r *wasmRows) Columns() []string { return r.cols }
func (r *wasmRows) Close() error      { return nil }

func (r *wasmRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.pos]
	r.pos++
	for i, col := range r.cols {
		dest[i] = row[col]
	}
	return nil
}

// callSQLExec bridges driver.Value args into sqlExec (serve_wasm.go).

func callSQLExec(query string, args []driver.Value) (sqlExecEnvelope, error) {
	anyArgs := make([]any, len(args))
	for i, a := range args {
		anyArgs[i] = a
	}
	env, err := sqlExec(query, anyArgs)
	if err != nil {
		return sqlExecEnvelope{}, fmt.Errorf("wasmDriver: %w", err)
	}
	return env, nil
}
