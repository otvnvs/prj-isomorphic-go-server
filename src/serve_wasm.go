//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"
	"time"

	promise "github.com/nlepage/go-js-promise"
	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

func startServer() {
	waitForSQL()
	openDB()
	initDB()
	wasmhttp.Serve(nil)
	select {}
}

// waitForSQL polls until self.sqlExec is defined in the SW global scope.
func waitForSQL() {
	for js.Global().Get("sqlExec").IsUndefined() {
		time.Sleep(50 * time.Millisecond)
	}
}

// sqlExecEnvelope is the JSON envelope returned by self.sqlExec in sw.js.
type sqlExecEnvelope struct {
	Rows         []map[string]any `json:"rows"`
	RowsAffected int64            `json:"rowsAffected"`
	LastInsertId int64            `json:"lastInsertId"`
}

// sqlExec calls self.sqlExec(sql, jsonParams) in the Service Worker and blocks
// until the Promise resolves. Used by driver_wasm.go via callSQLExec.
func sqlExec(query string, params []any) (sqlExecEnvelope, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return sqlExecEnvelope{}, err
	}
	p := js.Global().Call("sqlExec", query, string(paramsJSON))
	result, err := promise.Await(p)
	if err != nil {
		return sqlExecEnvelope{}, err
	}
	var env sqlExecEnvelope
	if err := json.Unmarshal([]byte(result.String()), &env); err != nil {
		return sqlExecEnvelope{}, err
	}
	return env, nil
}
