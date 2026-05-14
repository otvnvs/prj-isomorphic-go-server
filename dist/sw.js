// Service Worker — loads the Go WASM binary and handles SQLite via the JS OO1 API.
//
// The database is currently in-memory (sqlite3.oo1.DB). Data is lost when the
// service worker is terminated.
// TODO: swap sqlite3.oo1.DB() for OpfsSAHPoolDb for OPFS-backed persistence.
//
// wasm_exec.js version must match the Go toolchain used to compile a.wasm.

const SQLITE_BASE_DIRECTORY = './sqlite-wasm-3530100';
const DEBUG = false;

importScripts(`${SQLITE_BASE_DIRECTORY}/jswasm/sqlite3-worker1-promiser.js`);
importScripts(`${SQLITE_BASE_DIRECTORY}/jswasm/sqlite3.js`);
importScripts(`./cdn.jsdelivr.net/gh/golang/go@go1.23.0/misc/wasm/wasm_exec.js`);
importScripts(`./cdn.jsdelivr.net/gh/nlepage/go-wasm-http-server@v2.2.1/sw.js`);

// ---------------------------------------------------------------------------
// SQLite-WASM setup
// ---------------------------------------------------------------------------

let sqliteReady = false;
let db = null;

const sqliteReadyPromise = sqlite3InitModule({
    locateFile: (path) => `${SQLITE_BASE_DIRECTORY}/jswasm/${path}`,
}).then(async (sqlite3) => {
    db = new sqlite3.oo1.DB();
    sqliteReady = true;
    console.log('Service Worker: SQLite ready');
}).catch((err) => {
    console.error('Service Worker: SQLite init failed', err);
});

// ---------------------------------------------------------------------------
// JS helper called by Go via syscall/js
//
// self.sqlExec(sql, jsonParams) → Promise<jsonEnvelope>
//
//   sql:        SQL string, may contain ? placeholders
//   jsonParams: JSON array of bind values, e.g. '["alice","secret"]'
//
// Resolves to a JSON string: { rows, rowsAffected, lastInsertId }
// ---------------------------------------------------------------------------

self.sqlExec = async (sql, jsonParams = '[]') => {
    if (!sqliteReady) await sqliteReadyPromise;
    if (!db) throw new Error('SQLite DB not available');

    const params = JSON.parse(jsonParams);
    if (DEBUG) {
        console.log('sqlExec sql:', sql);
        console.log('sqlExec params:', params);
    }

    const rows = [];
    db.exec({
        sql,
        bind: params,
        rowMode: 'object',
        callback: (row) => rows.push(row),
    });

    if (DEBUG) console.log('sqlExec rows:', rows);

    return JSON.stringify({
        rows,
        rowsAffected: db.changes(),
        lastInsertId: db.last_insert_rowid ? db.last_insert_rowid() : 0,
    });
};

// ---------------------------------------------------------------------------
// Service Worker lifecycle — take control immediately
// ---------------------------------------------------------------------------

registerWasmHTTPListener('a.wasm');

addEventListener('install', (event) => {
    event.waitUntil(skipWaiting());
});

addEventListener('activate', (event) => {
    event.waitUntil(clients.claim());
});
