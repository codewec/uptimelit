let signedInUser = null;
let authConfiguration = null;
let userList = [];
const nativeFetch = window.fetch.bind(window);
window.fetch = async (input, options = {}) => {
  const headers = new Headers(options.headers);
  headers.set('X-UptimeLit-Request', '1');
  const response = await nativeFetch(input, {...options, headers});
  if (response.status === 401 && signedInUser) {
    signedInUser = null;
    clearInterval(refreshInterval);
    document.getElementById('app').hidden = true;
    // A fresh document also clears modal fields, cached secrets and pending UI work.
    location.replace('/');
    throw new Error('Session expired');
  }
  return response;
};
async function accountRequest(path, method = 'GET', body) {
  const response = await fetch(path, {method, headers: {'Content-Type':'application/json'}, ...(body === undefined ? {} : {body: JSON.stringify(body)})});
  if (response.status === 204) return null;
  const result = await response.json();
  if (!response.ok) throw new Error(result.error || 'Request failed');
  return result;
}
function accountError(error) { document.getElementById('account-error').textContent = error.message; }
async function initAuth() {
  const error = document.getElementById('login-error');
  try {
    authConfiguration = await accountRequest('/api/auth/config');
    const response = await fetch('/api/auth/me');
    if (response.ok) {
      signedInUser = await response.json();
      document.getElementById('account-name').textContent = `${signedInUser.username} · ${signedInUser.role}`;
      document.getElementById('users-button').classList.toggle('hidden', signedInUser.role !== 'admin');
      document.getElementById('app').hidden = false;
      startMonitoring();
      return;
    }
    document.getElementById('login-view').hidden = false;
    document.getElementById('login-form').hidden = !authConfiguration.passwordLoginEnabled;
    const oidc = document.getElementById('oidc-button');
    oidc.hidden = !authConfiguration.oidcEnabled;
    oidc.textContent = authConfiguration.oidcButtonTitle;
    if (authConfiguration.bootstrap) {
      document.getElementById('login-description').textContent = 'First launch: sign in with a new username and password (8–72 bytes), or use OIDC. This creates the first administrator. Later accounts can only be created by an administrator.';
      document.querySelector('#login-form button').textContent = 'Create administrator and sign in';
      document.querySelector('#login-form [name=password]').autocomplete = 'new-password';
    }
    error.textContent = new URLSearchParams(location.search).get('oidc_error') || '';
    history.replaceState(null, '', '/');
  } catch (e) { document.getElementById('login-view').hidden = false; error.textContent = e.message; }
}
document.getElementById('login-form').addEventListener('submit', async event => {
  event.preventDefault();
  const button = event.target.querySelector('button');
  button.disabled = true;
  document.getElementById('login-error').textContent = '';
  try {
    await accountRequest('/api/auth/login', 'POST', Object.fromEntries(new FormData(event.target)));
    location.replace('/');
  } catch (e) { document.getElementById('login-error').textContent = e.message; button.disabled = false; }
});
async function logoutAccount() {
  try { await accountRequest('/api/auth/logout', 'POST'); location.replace('/'); }
  catch (e) { alert(e.message); }
}
function accountDialog(title, html) {
  document.getElementById('account-title').textContent = title;
  document.getElementById('account-content').innerHTML = html;
  document.getElementById('account-error').textContent = '';
  const dialog = document.getElementById('account-dialog');
  if (!dialog.open) dialog.showModal();
}
function openProfile() {
  accountDialog('Profile', `<form id="profile-form" class="account-form">
    <label>Username<input name="username" required maxlength="64" pattern="[a-zA-Z0-9._-]+" value="${escapeHtml(signedInUser.username)}" ${signedInUser.authMethod === 'oidc' ? 'readonly' : ''} autocomplete="username"></label>
    ${authConfiguration.passwordLoginEnabled ? `<label>Current password<input name="currentPassword" type="password" autocomplete="current-password"></label><label>New password (leave empty to keep)<input name="newPassword" type="password" autocomplete="new-password"></label>` : ''}
    <p class="text-sm text-gray-500">Saving signs out your sessions. Sign in again to continue.</p><button type="submit">Save profile</button></form>`);
  document.getElementById('profile-form').onsubmit = async event => {
    event.preventDefault();
    try { await accountRequest('/api/auth/profile', 'PUT', Object.fromEntries(new FormData(event.target))); location.replace('/'); }
    catch (e) { accountError(e); }
  };
}
async function openUsers() {
  try {
    userList = await accountRequest('/api/admin/users');
    accountDialog('Users', `<button id="new-user" class="account-button mb-4">Create user</button><div id="user-rows"></div>`);
    document.getElementById('new-user').onclick = () => editAccountUser();
    const rows = document.getElementById('user-rows');
    for (const user of userList) {
      const row = document.createElement('div'); row.className = 'account-user';
      const name = document.createElement('span'); name.className = 'flex-1';
      name.textContent = `${user.username} · ${user.role} · ${user.hasPassword ? 'Password' : 'No password'}${user.hasOIDC ? ' + OIDC' : ''}`;
      const edit = document.createElement('button'); edit.textContent = 'Edit'; edit.onclick = () => editAccountUser(user);
      const remove = document.createElement('button'); remove.textContent = 'Delete'; remove.className = 'text-red-600'; remove.disabled = user.id === signedInUser.id;
      remove.onclick = async () => {
        if (!confirm(`Delete ${user.username} and all their monitors, history and notifiers?`)) return;
        try { await accountRequest(`/api/admin/users/${encodeURIComponent(user.id)}`, 'DELETE'); await openUsers(); }
        catch (e) { accountError(e); }
      };
      row.append(name, edit, remove); rows.append(row);
    }
  } catch (e) { accountError(e); }
}
function editAccountUser(user = null) {
  accountDialog(user ? 'Edit user' : 'Create user', `<form id="user-form" class="account-form">
    <label>Username<input name="username" required maxlength="64" pattern="[a-zA-Z0-9._-]+" value="${escapeHtml(user?.username || '')}" autocomplete="off"></label>
    <label>Role<select name="role"><option value="user">User</option><option value="admin">Administrator</option></select></label>
    <label>${user ? 'New password (leave empty to keep)' : 'Password (optional for OIDC)'}<input name="${user ? 'newPassword' : 'password'}" type="password" autocomplete="new-password"></label>
    ${user?.hasOIDC ? '<label><span><input name="resetOIDC" type="checkbox"> Reset OIDC binding (requires an available password)</span></label>' : ''}
    <p class="text-sm text-gray-500">OIDC links to a pre-created account with the same normalized username. Saving changes revokes this user's sessions.</p>
    <button type="submit">Save user</button><button type="button" id="users-back">Back to users</button></form>`);
  document.querySelector('#user-form [name=role]').value = user?.role || 'user';
  document.getElementById('users-back').onclick = openUsers;
  document.getElementById('user-form').onsubmit = async event => {
    event.preventDefault(); const body = Object.fromEntries(new FormData(event.target)); body.resetOIDC = body.resetOIDC === 'on';
    try {
      await accountRequest(user ? `/api/admin/users/${encodeURIComponent(user.id)}` : '/api/admin/users', user ? 'PUT' : 'POST', body);
      if (user?.id === signedInUser.id) { location.replace('/'); return; }
      await openUsers();
    } catch (e) { accountError(e); }
  };
}
