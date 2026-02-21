(function(){const t=document.createElement("link").relList;if(t&&t.supports&&t.supports("modulepreload"))return;for(const n of document.querySelectorAll('link[rel="modulepreload"]'))s(n);new MutationObserver(n=>{for(const o of n)if(o.type==="childList")for(const a of o.addedNodes)a.tagName==="LINK"&&a.rel==="modulepreload"&&s(a)}).observe(document,{childList:!0,subtree:!0});function r(n){const o={};return n.integrity&&(o.integrity=n.integrity),n.referrerPolicy&&(o.referrerPolicy=n.referrerPolicy),n.crossOrigin==="use-credentials"?o.credentials="include":n.crossOrigin==="anonymous"?o.credentials="omit":o.credentials="same-origin",o}function s(n){if(n.ep)return;n.ep=!0;const o=r(n);fetch(n.href,o)}})();const d="/transactions";function l(e){const t=new URLSearchParams;e.instrument&&t.set("instrument",e.instrument),e.broker&&t.set("broker",e.broker),e.type&&t.set("type",e.type),e.from&&t.set("from",e.from),e.to&&t.set("to",e.to);const r=t.toString();return r?`?${r}`:""}async function u(e){const t=`${d}${l(e)}`,r=await fetch(t);if(!r.ok){const s=await r.json().catch(()=>({error:r.statusText}));throw new Error(s.error||`HTTP ${r.status}`)}return r.json()}function m(e){return e?new Date(e).toLocaleDateString(void 0,{year:"numeric",month:"short",day:"2-digit"}):""}function p(e,t){return new Intl.NumberFormat(void 0,{style:"currency",currency:t||"USD",minimumFractionDigits:2}).format(e)}function f(e){return e.length===0?'<tr><td colspan="8" class="empty">No transactions found.</td></tr>':e.map(t=>`
      <tr>
        <td>${m(t.created_at)}</td>
        <td><span class="instrument">${t.instrument}</span></td>
        <td>${t.isin||"—"}</td>
        <td><span class="badge badge-${t.type}">${t.type}</span></td>
        <td>${t.category}</td>
        <td class="number">${t.quantity}</td>
        <td class="number">${p(t.amount,t.currency)}</td>
        <td>${t.description||"—"}</td>
      </tr>`).join("")}function y(){const e=document.getElementById("app");e.innerHTML=`
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
  `;const t=document.getElementById("filter-form"),r=document.getElementById("table-body"),s=document.getElementById("status"),n=document.getElementById("count");t.addEventListener("submit",async o=>{o.preventDefault();const a=new FormData(t),c={instrument:a.get("instrument")||"",broker:a.get("broker")||"",type:a.get("type")||"",from:a.get("from")||"",to:a.get("to")||""};s.textContent="Loading…",n.textContent="",r.innerHTML="";try{const i=await u(c);s.textContent="",n.textContent=`${i.length} result${i.length!==1?"s":""}`,r.innerHTML=f(i)}catch(i){s.textContent=`Error: ${i.message}`,r.innerHTML="",n.textContent=""}}),t.addEventListener("reset",()=>{s.textContent="",n.textContent="",r.innerHTML='<tr><td colspan="8" class="empty">Use the filters above and click Search.</td></tr>'})}y();
