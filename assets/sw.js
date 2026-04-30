// Service Worker — loads the WASM binary and handles IndexedDB persistence.
//
// wasm_exec.js version must match the Go toolchain used to compile a.wasm.
// Update the tag in the URL below if you upgrade Go.
importScripts('https://cdn.jsdelivr.net/gh/golang/go@go1.23.0/misc/wasm/wasm_exec.js')
importScripts('https://cdn.jsdelivr.net/gh/nlepage/go-wasm-http-server@v2.2.1/sw.js')

// ---------------------------------------------------------------------------
// IndexedDB setup
// ---------------------------------------------------------------------------

let db;
const dbReq = indexedDB.open("WasmAuthDB", 1);

dbReq.onupgradeneeded = (e) => {
    const database = e.target.result;
    if (!database.objectStoreNames.contains("users")) {
        database.createObjectStore("users", { keyPath: "username" });
    }
};

dbReq.onsuccess = (e) => {
    db = e.target.result;
    console.log("📦 Service Worker: IndexedDB ready");
};

dbReq.onerror = (e) => {
    console.error("📦 Service Worker: IndexedDB error", e.target.error);
};

// ---------------------------------------------------------------------------
// JS helpers called by Go via syscall/js
// ---------------------------------------------------------------------------

/** Persist a single user record. Called by Go after a successful signup. */
self.saveUserToDB = (userObj) => {
    if (!db) { console.error("saveUserToDB: DB not ready"); return; }
    const tx = db.transaction("users", "readwrite");
    tx.objectStore("users").put(JSON.parse(userObj));
};

/** Return all stored users as a JSON string. Called by Go on startup. */
self.loadUsersFromDB = () => {
    return new Promise((resolve) => {
        if (!db) { resolve("[]"); return; }
        const tx = db.transaction("users", "readonly");
        const request = tx.objectStore("users").getAll();
        request.onsuccess = () => {
            const data = JSON.stringify(request.result);
            console.log("📦 DB sending to Go:", data);
            resolve(data);
        };
        request.onerror = () => resolve("[]");
    });
};

// ---------------------------------------------------------------------------
// Service Worker lifecycle — take control immediately
// ---------------------------------------------------------------------------

registerWasmHTTPListener('a.wasm')

addEventListener('install', (event) => {
    event.waitUntil(skipWaiting());
});

addEventListener('activate', (event) => {
    event.waitUntil(clients.claim());
});
