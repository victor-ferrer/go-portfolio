interface Transaction {
  id: string;
  instrument: string;
  isin: string;
  type: string;
  category: string;
  quantity: number;
  amount: number;
  currency: string;
  description: string;
  created_at: string;
}

interface FilterState {
  instrument: string;
  broker: string;
  type: string;
  from: string;
  to: string;
}

const API_BASE = '/transactions';

function buildQueryString(filters: FilterState): string {
  const params = new URLSearchParams();
  if (filters.instrument) params.set('instrument', filters.instrument);
  if (filters.broker) params.set('broker', filters.broker);
  if (filters.type) params.set('type', filters.type);
  if (filters.from) params.set('from', filters.from);
  if (filters.to) params.set('to', filters.to);
  const qs = params.toString();
  return qs ? `?${qs}` : '';
}

async function fetchTransactions(filters: FilterState): Promise<Transaction[]> {
  const url = `${API_BASE}${buildQueryString(filters)}`;
  const response = await fetch(url);
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(body.error || `HTTP ${response.status}`);
  }
  return response.json();
}

function formatDate(iso: string): string {
  if (!iso) return '';
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
  });
}

function formatAmount(amount: number, currency: string): string {
  return new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency: currency || 'USD',
    minimumFractionDigits: 2,
  }).format(amount);
}

function renderTable(transactions: Transaction[]): string {
  if (transactions.length === 0) {
    return `<tr><td colspan="8" class="empty">No transactions found.</td></tr>`;
  }
  return transactions
    .map(
      (tx) => `
      <tr>
        <td>${formatDate(tx.created_at)}</td>
        <td><span class="instrument">${tx.instrument}</span></td>
        <td>${tx.isin || '—'}</td>
        <td><span class="badge badge-${tx.type}">${tx.type}</span></td>
        <td>${tx.category}</td>
        <td class="number">${tx.quantity}</td>
        <td class="number">${formatAmount(tx.amount, tx.currency)}</td>
        <td>${tx.description || '—'}</td>
      </tr>`
    )
    .join('');
}

function render(): void {
  const appEl = document.getElementById('app')!;
  appEl.innerHTML = `
    <header>
      <h1>Portfolio Transactions</h1>
    </header>
    <main>
      <section class="filters card">
        <h2>Search &amp; Filter</h2>
        <form id="filter-form">
          <div class="form-grid">
            <div class="field">
              <label for="instrument">Instrument</label>
              <input id="instrument" name="instrument" type="text" placeholder="e.g. AAPL" />
            </div>
            <div class="field">
              <label for="broker">Broker</label>
              <input id="broker" name="broker" type="text" placeholder="e.g. click-trade" />
            </div>
            <div class="field">
              <label for="type">Type</label>
              <select id="type" name="type">
                <option value="">All</option>
                <option value="buy">Buy</option>
                <option value="sell">Sell</option>
              </select>
            </div>
            <div class="field">
              <label for="from">From</label>
              <input id="from" name="from" type="date" />
            </div>
            <div class="field">
              <label for="to">To</label>
              <input id="to" name="to" type="date" />
            </div>
          </div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary">Search</button>
            <button type="reset" class="btn btn-secondary">Clear</button>
          </div>
        </form>
      </section>

      <section class="results card">
        <div class="results-header">
          <h2>Results</h2>
          <span id="count" class="count"></span>
        </div>
        <div id="status" class="status" role="status" aria-live="polite"></div>
        <div class="table-wrapper">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Instrument</th>
                <th>ISIN</th>
                <th>Type</th>
                <th>Category</th>
                <th>Quantity</th>
                <th>Amount</th>
                <th>Description</th>
              </tr>
            </thead>
            <tbody id="table-body">
              <tr><td colspan="8" class="empty">Use the filters above and click Search.</td></tr>
            </tbody>
          </table>
        </div>
      </section>
    </main>
  `;

  const form = document.getElementById('filter-form') as HTMLFormElement;
  const tbody = document.getElementById('table-body')!;
  const statusEl = document.getElementById('status')!;
  const countEl = document.getElementById('count')!;

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const data = new FormData(form);
    const filters: FilterState = {
      instrument: (data.get('instrument') as string) || '',
      broker: (data.get('broker') as string) || '',
      type: (data.get('type') as string) || '',
      from: (data.get('from') as string) || '',
      to: (data.get('to') as string) || '',
    };

    statusEl.textContent = 'Loading…';
    countEl.textContent = '';
    tbody.innerHTML = '';

    try {
      const transactions = await fetchTransactions(filters);
      statusEl.textContent = '';
      countEl.textContent = `${transactions.length} result${transactions.length !== 1 ? 's' : ''}`;
      tbody.innerHTML = renderTable(transactions);
    } catch (err) {
      statusEl.textContent = `Error: ${(err as Error).message}`;
      tbody.innerHTML = '';
      countEl.textContent = '';
    }
  });

  form.addEventListener('reset', () => {
    statusEl.textContent = '';
    countEl.textContent = '';
    tbody.innerHTML = '<tr><td colspan="8" class="empty">Use the filters above and click Search.</td></tr>';
  });
}

render();
