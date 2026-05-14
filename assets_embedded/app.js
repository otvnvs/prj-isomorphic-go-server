// app.js — shared API client and auth helpers
const getToken  = () => localStorage.getItem('wasm_token');
const setToken  = (t) => localStorage.setItem('wasm_token', t);
const clearToken = ()  => localStorage.removeItem('wasm_token');

async function api(path, method = 'GET', body = null) {
    const headers = { 'Content-Type': 'application/json' };
    const token = getToken();
    if (token) headers['Authorization'] = 'Bearer ' + token;
    return fetch(path, { method, headers, body: body ? JSON.stringify(body) : null });
}

async function requireLogin() {
    const res = await api('./api/me');
    const data = await res.json();
    if (!data.logged_in) window.location.href = './index.html';
    return data;
}
