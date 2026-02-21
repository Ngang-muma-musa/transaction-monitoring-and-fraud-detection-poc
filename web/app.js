let ws = null;
let transactions = [];
let filter = 'all';
let totalVolume = 0;
let wsConnected = false;

// Sparkline data
let sparkData = Array(30).fill(0);
let sparkCtx = null;
let txPerMinute = 0;
let txCountThisMinute = 0;

window.onload = () => {
  sparkCtx = document.getElementById('spark').getContext('2d');
  drawSpark();
  setInterval(() => {
    sparkData.push(txCountThisMinute);
    sparkData.shift();
    txCountThisMinute = 0;
    drawSpark();
  }, 2000);
  setNewUUID();
  fetchAllTransactions();
};

function drawSpark() {
  const canvas = document.getElementById('spark');
  const ctx = sparkCtx;
  const w = canvas.offsetWidth || 288;
  const h = 60;
  canvas.width = w;
  canvas.height = h;
  ctx.clearRect(0,0,w,h);
  const max = Math.max(...sparkData, 1);
  const step = w / (sparkData.length - 1);

  ctx.beginPath();
  sparkData.forEach((v, i) => {
    const x = i * step;
    const y = h - (v / max) * (h - 8) - 4;
    i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y);
  });
  ctx.strokeStyle = 'rgba(0,245,196,0.7)';
  ctx.lineWidth = 2;
  ctx.stroke();

  // Fill
  ctx.lineTo(w, h); ctx.lineTo(0, h); ctx.closePath();
  ctx.fillStyle = 'rgba(0,245,196,0.08)';
  ctx.fill();
}

async function fetchAllTransactions() {
  const base = document.getElementById('cfg-api').value;
  const list = document.getElementById('tx-list');

  // Show loading state
  list.innerHTML = '<div class="empty-state"><div class="empty-icon" style="animation:pulse-hex 1s infinite">⬡</div><div>Loading transactions...</div></div>';

  try {
    const res = await fetch(`${base}/payments`);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();

    // data may be an array directly, or wrapped e.g. { transactions: [...] }
    const txList = Array.isArray(data) ? data : (data.transactions || data.data || []);

    if (txList.length === 0) {
      list.innerHTML = '<div class="empty-state"><div class="empty-icon">📡</div><div>No transactions yet</div><div style="font-size:11px;color:var(--text-dim)">Send a test transaction to get started</div></div>';
      return;
    }

    // Sort by created_at descending (newest first)
    txList.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

    // Load into state without triggering re-renders each time
    txList.forEach(raw => {
      const tx = {
        id:         raw.id || genId(),
        user_id:    raw.user_id || '—',
        amount:     raw.amount || 0,
        currency:   raw.currency || 'USD',
        status:     (raw.status || 'PENDING').toUpperCase(),
        created_at: raw.created_at ? new Date(raw.created_at) : new Date(),
        risk_score: null,
        reason:     null,
        raw
      };
      transactions.push(tx);
      totalVolume += tx.amount;
    });

    wsLog('sys', `✓ Loaded ${txList.length} transactions from GET /payments`);
    updateStats();
    renderFeed();
  } catch(e) {
    wsLog('err', `GET /payments failed: ${e.message}`);
    list.innerHTML = '<div class="empty-state"><div class="empty-icon">⚠️</div><div style="color:var(--danger)">Could not load transactions</div><div style="font-size:11px;color:var(--text-dim)">Is your server running?</div></div>';
  }
}

function connectWS() {
  const url = document.getElementById('cfg-ws').value;
  if (ws) { ws.close(); }
  wsLog('sys', `Connecting to ${url}...`);
  try {
    ws = new WebSocket(url);
    ws.onopen = () => {
      wsConnected = true;
      document.getElementById('ws-dot').className = 'ws-dot connected';
      document.getElementById('ws-label').textContent = 'CONNECTED';
      wsLog('sys', '✓ Connected');
    };
    ws.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data);
        // Log full raw payload so we can see exact field values
        wsLog('in', `← RAW: ${JSON.stringify(data)}`);
        handleIncoming(data);
      } catch(err) {
        wsLog('err', `WS parse error: ${err.message} | raw: ${e.data.substring(0,120)}`);
      }
    };
    ws.onclose = () => {
      wsConnected = false;
      document.getElementById('ws-dot').className = 'ws-dot';
      document.getElementById('ws-label').textContent = 'DISCONNECTED';
      wsLog('sys', '✗ Connection closed');
    };
    ws.onerror = () => {
      document.getElementById('ws-dot').className = 'ws-dot error';
      wsLog('err', 'WebSocket error');
    };
  } catch(e) {
    wsLog('err', 'Failed: ' + e.message);
  }
}

// handleTransaction maps a Transaction struct response from POST /payments
function handleTransaction(data) {
  const tx = {
    id:         data.id || genId(),
    user_id:    data.user_id || '—',
    amount:     data.amount || 0,
    currency:   data.currency || 'USD',
    status:     (data.status || 'PENDING').toUpperCase(),
    created_at: data.created_at ? new Date(data.created_at) : new Date(),
    // fraud alert fields populated when a FraudAlert arrives via WS
    risk_score: null,
    reason:     null,
    raw:        data
  };
  addTransaction(tx);
}

// handleFraudAlert maps a FraudAlert struct arriving over WebSocket:
// { transaction_id, risk_score, reason }
function handleFraudAlert(alert) {
  wsLog('in', `← FRAUD ALERT txn=${alert.transaction_id} score=${alert.risk_score}`);
  // Enrich existing transaction if already in list
  const existing = transactions.find(t => t.id === alert.transaction_id);
  if (existing) {
    existing.risk_score = alert.risk_score;
    existing.reason     = alert.reason;
    // Only override status if score actually signals fraud — don't blindly overwrite API status
    if (alert.risk_score > 0.9) {
      existing.status = 'DECLINED';
    } else if (alert.risk_score > 0.6) {
      existing.status = 'FLAGGED';
    }
    existing.raw = { ...existing.raw, ...alert };
    updateStats();
    updateRisk(alert.risk_score);
    renderFeed();
    addAlert(existing);
  } else {
    const tx = {
      id:         alert.transaction_id,
      user_id:    '—',
      amount:     0,
      currency:   '—',
      status:     alert.risk_score > 0.8 ? 'DECLINED' : alert.risk_score > 0.6 ? 'FLAGGED' : 'PENDING',
      created_at: new Date(),
      risk_score: alert.risk_score,
      reason:     alert.reason,
      raw:        alert
    };
    addTransaction(tx);
  }
}

// Route incoming WS message to the right handler
function handleIncoming(data) {
  if (data.transaction_id !== undefined && data.risk_score !== undefined && data.reason !== undefined) {
    handleFraudAlert(data);
  } else if (data.id !== undefined) {
    handleTransaction(data);
  } else {
    wsLog('sys', 'Unknown message shape: ' + JSON.stringify(data).substring(0, 80));
  }
}

function wsLog(type, msg) {
  const log = document.getElementById('ws-log');
  const line = document.createElement('div');
  line.className = `ws-log-line ${type}`;
  const t = new Date().toLocaleTimeString('en-US', {hour12:false});
  line.textContent = `[${t}] ${msg}`;
  log.appendChild(line);
  log.scrollTop = log.scrollHeight;
  while (log.children.length > 50) log.removeChild(log.firstChild);
}

const scenarios = {
  'l1-trigger':  { amount: 6500,   currency: 'USD', label: 'Foreign >6000', desc: 'Triggers Layer 1 (+0.5)' },
  'l1-safe':     { amount: 5999,   currency: 'USD', label: 'Foreign <6000', desc: 'Just under Layer 1 threshold' },
  'l1-xaf-safe': { amount: 50000,  currency: 'XAF', label: 'XAF Large',    desc: 'Large but XAF — no Layer 1' },
  'l2-velocity': { amount: 400,    currency: 'XAF', label: 'Velocity ×3',  desc: 'Send 3x to exceed 1000/hr (+0.3)' },
  'l1-l2-combo': { amount: 7000,   currency: 'EUR', label: 'L1 + Velocity', desc: 'High foreign + send multiple times' },
  'safe':        { amount: 500,    currency: 'XAF', label: 'Normal XAF',   desc: 'Low amount, local currency' },
};

function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
    const r = crypto.getRandomValues(new Uint8Array(1))[0] % 16;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

function setNewUUID() {
  const input = document.getElementById('f-userid');
  input.value = generateUUID();
  input.style.borderColor = 'var(--accent)';
  input.style.color = 'var(--accent)';
  setTimeout(() => { input.style.borderColor = ''; input.style.color = ''; }, 400);
}

function loadScenario(key) {
  const s = scenarios[key];
  if (!s) return;
  setNewUUID();
  document.getElementById('f-amount').value = s.amount;
  document.getElementById('f-currency').value = s.currency;
  if (s.desc) wsLog('sys', `ℹ ${s.label}: ${s.desc}`);
}

async function sendTransaction() {
  const btn = document.getElementById('send-btn');
  btn.disabled = true;
  btn.textContent = 'SENDING...';
  const payload = {
    user_id:  document.getElementById('f-userid').value,
    amount:   parseFloat(document.getElementById('f-amount').value),
    currency: document.getElementById('f-currency').value,
  };
  const base = document.getElementById('cfg-api').value;
  wsLog('out', `→ POST ${base}/payments ${JSON.stringify(payload).substring(0,60)}`);
  try {
    const res = await fetch(`${base}/payments`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload)
    });
    const data = await res.json();
    wsLog('in', `← ${res.status} ${JSON.stringify(data).substring(0,80)}`);
    showResponse(data);
    handleTransaction(data);
  } catch(e) {
    wsLog('err', 'Request failed: ' + e.message);
    showError(`Could not reach API at ${base}/payments — is your server running with CORS enabled?`);
  } finally {
    btn.disabled = false; btn.textContent = '▶ SEND TRANSACTION';
  }
}

function showError(msg) {
  const body = document.getElementById("resp-body");
  body.textContent = "⚠ " + msg;
  body.style.color = "var(--danger)";
  document.getElementById("resp-panel").classList.add("show");
  setTimeout(() => { document.getElementById("resp-panel").classList.remove("show"); body.style.color = ""; }, 6000);
}

function showResponse(data) {
  document.getElementById('resp-body').textContent = JSON.stringify(data, null, 2);
  document.getElementById('resp-panel').classList.add('show');
  setTimeout(() => document.getElementById('resp-panel').classList.remove('show'), 6000);
}

function addTransaction(tx) {
  transactions.unshift(tx);
  txCountThisMinute++;
  totalVolume += tx.amount;
  updateStats();
  if (tx.risk_score != null) updateRisk(tx.risk_score);
  renderFeed();
  if (tx.status === 'DECLINED' || tx.status === 'FLAGGED') addAlert(tx);
}

function updateStats() {
  const total = transactions.length;
  const approved = transactions.filter(t => t.status === 'APPROVED').length;
  const flagged = transactions.filter(t => t.status === 'FLAGGED').length;
  const declined = transactions.filter(t => t.status === 'DECLINED').length;
  document.getElementById('stat-total').textContent = total;
  document.getElementById('stat-approved').textContent = approved;
  document.getElementById('stat-flagged').textContent = flagged;
  document.getElementById('stat-declined').textContent = declined;
  const approvalRate = total ? Math.round(approved/total*100) : 0;
  const fraudRate = total ? Math.round(declined/total*100) : 0;
  document.getElementById('s-rate').textContent = approvalRate + '%';
  document.getElementById('s-fraud').textContent = fraudRate + '%';
  document.getElementById('s-volume').textContent = '$' + (totalVolume/1000).toFixed(1) + 'k';
  document.getElementById('s-avg').textContent = '$' + (total ? totalVolume/total : 0).toFixed(0);
}

let riskHistory = [];
function updateRisk(score) {
  const scaled = score * 100;
  riskHistory.unshift(scaled);
  if (riskHistory.length > 10) riskHistory.pop();
  const avg = riskHistory.reduce((a,b) => a+b, 0) / riskHistory.length;
  document.getElementById('gauge-num').textContent = avg.toFixed(1);
  const color = avg > 70 ? 'var(--danger)' : avg > 40 ? 'var(--warn)' : 'var(--accent)';
  document.getElementById('gauge-num').style.color = color;
  const arc = document.getElementById('gauge-arc');
  const offset = 220 - (avg / 100) * 220;
  arc.style.strokeDashoffset = offset;
  arc.style.stroke = color;
}

function renderFeed() {
  const list = document.getElementById('tx-list');
  const filtered = filter === 'all' ? transactions : transactions.filter(t => t.status === filter);
  if (filtered.length === 0) {
    list.innerHTML = '<div class="empty-state"><div class="empty-icon">📡</div><div>No transactions yet</div><div style="font-size:11px;color:var(--text-dim)">Send a test transaction to get started</div></div>';
    return;
  }
  list.innerHTML = '';
  filtered.slice(0, 100).forEach(tx => {
    const card = document.createElement('div');
    card.className = `tx-card ${tx.status}`;
    const time = tx.created_at ? tx.created_at.toLocaleTimeString('en-US', {hour12:false}) : '—';
    const riskBadge = tx.risk_score != null
      ? `<span style="font-size:10px;color:${tx.risk_score>=0.7?'var(--danger)':tx.risk_score>=0.4?'var(--warn)':'var(--accent)'};margin-left:6px;font-family:var(--mono)">⬡ ${(tx.risk_score*100).toFixed(0)}%</span>`
      : '';
    const reasonLine = tx.reason
      ? `<div style="font-size:10px;color:var(--text-dim);margin-top:2px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">${tx.reason}</div>`
      : '';
    card.innerHTML = `
      <div class="tx-icon">💳</div>
      <div class="tx-meta">
        <div class="tx-id">${tx.id}</div>
        <div class="tx-desc">user: ${tx.user_id}${riskBadge}</div>
        ${reasonLine}
      </div>
      <div class="tx-amount" style="color:${tx.status==='DECLINED'?'var(--danger)':tx.status==='FLAGGED'?'var(--warn)':'var(--text)'}">${formatCurrency(tx.amount, tx.currency)}</div>
      <div>
        <div class="tx-status status-${tx.status}">${tx.status}</div>
        <div class="tx-time" style="margin-top:4px;text-align:right">${time}</div>
      </div>`;
    card.onclick = () => showResponse(tx.raw || tx);
    list.appendChild(card);
  });
}

function formatCurrency(amount, currency) { if (!amount) return '$0.00'; const sym = {USD:'$',EUR:'€',GBP:'£',BTC:'₿',NGN:'₦'}[currency] || currency+' '; return sym + parseFloat(amount).toFixed(2) }

function addAlert(tx) {
  const list = document.getElementById('alert-list');
  const item = document.createElement('div');
  item.className = 'alert-item';
  const color = tx.status === 'DECLINED' ? 'var(--danger)' : 'var(--warn)';
  const score = tx.risk_score != null ? (tx.risk_score * 100).toFixed(0) + '%' : '—';
  const msg = tx.status === 'DECLINED'
    ? `${tx.id} declined — score ${score} — ${tx.reason || 'High risk'}`
    : `${tx.id} flagged — score ${score} — ${tx.reason || 'Suspicious pattern'}`;
  item.innerHTML = `
    <div class="alert-dot" style="background:${color};box-shadow:0 0 6px ${color}"></div>
    <div class="alert-content">
      <div style="color:${color};font-weight:700;font-size:11px">${tx.status}</div>
      <div>${msg}</div>
      <div class="alert-time">${tx.created_at ? tx.created_at.toLocaleTimeString() : new Date().toLocaleTimeString()}</div>
    </div>`;
  list.insertBefore(item, list.firstChild);
  while (list.children.length > 20) list.removeChild(list.lastChild);
}

function setFilter(f, btn) { filter = f; document.querySelectorAll('.tab').forEach(t => t.classList.remove('active')); btn.classList.add('active'); renderFeed(); }
function clearFeed() { transactions = []; totalVolume = 0; riskHistory = []; updateStats(); updateRisk(0); document.getElementById('alert-list').innerHTML = ''; renderFeed(); }
async function refreshFeed() { transactions = []; totalVolume = 0; riskHistory = []; document.getElementById('alert-list').innerHTML = ''; await fetchAllTransactions(); }
function genId() { return 'txn_' + Math.random().toString(36).substr(2, 8) }

// Auto-attempt WS connection on load
setTimeout(() => { connectWS(); }, 500);
