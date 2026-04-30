// Use 'const' for elements to improve performance
const statusEl = document.getElementById('status');
const authFields = document.getElementById('auth-fields');
const userFields = document.getElementById('user-fields');
const toastEl = document.getElementById('toast');

async function fetchVersion() {
    try {
	const response = await fetch('./api/version');
	const data = await response.json();
	const date = new Date(data.build_time).toLocaleDateString(undefined, {
	    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
	});

	document.getElementById('v-num').textContent = data.version;
	document.getElementById('v-time').textContent = date;
    } catch (error) {
	document.getElementById('version-info').style.opacity = '0.5';
	document.getElementById('v-num').textContent = 'v0.0.0';
    }
}

const getToken = () => localStorage.getItem('wasm_token');
const setToken = (t) => localStorage.setItem('wasm_token', t);
const clearToken = () => localStorage.removeItem('wasm_token');

function notify(msg, isError = false) {
    toastEl.innerText = msg;
    toastEl.style.background = isError ? "var(--danger)" : "#2ed573";
    toastEl.classList.add('show-toast');
    // Longer duration for mobile users to read
    setTimeout(() => toastEl.classList.remove('show-toast'), 3500);
}

function setLoading(isLoading) {
    const btns = document.querySelectorAll('button');
    btns.forEach(b => b.disabled = isLoading);
    document.querySelector('.btn-text').style.opacity = isLoading ? '0' : '1';
    document.querySelector('.spinner').style.display = isLoading ? 'block' : 'none';
}

async function api(path, method = 'GET', body = null) {
    const headers = { 'Content-Type': 'application/json' };
    const token = getToken();
    if (token) headers['Authorization'] = `Bearer ${token}`;
    try {
	return await fetch(path, { method, headers, body: body ? JSON.stringify(body) : null });
    } catch (e) {
	return { ok: false, json: () => ({ error: "Connection failed" }) };
    }
}

async function updateUI() {
    const res = await api('./api/me');
    const data = await res.json();

    if (data.logged_in) {
	statusEl.innerText = `Hi, ${data.user}`;
	statusEl.style.background = "rgba(56, 189, 248, 0.1)";
	statusEl.style.color = "var(--primary)";
	authFields.classList.add('hidden');
	userFields.classList.remove('hidden');
    } else {
	statusEl.innerText = "Welcome Back";
	statusEl.style.background = "transparent";
	statusEl.style.color = "var(--text-muted)";
	authFields.classList.remove('hidden');
	userFields.classList.add('hidden');
    }
}

async function doAuth(path) {
    const u = document.getElementById('u').value.trim();
    const p = document.getElementById('p').value;
    if (!u || !p) return notify("Fields required", true);

    setLoading(true);
    const res = await api(path, 'POST', { username: u, password: p });
    const data = await res.json();
    setLoading(false);

    if (res.ok) {
	if (data.token) setToken(data.token);
	notify(data.message || "Welcome!");
	updateUI();
    } else {
	notify(data.error || "Access denied", true);
    }
}

function doLogout() {
    clearToken();
    notify("Safe travels!");
    updateUI();
}

// Init
fetchVersion();
updateUI();
