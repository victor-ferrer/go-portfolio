(function(){const t=document.createElement("link").relList;if(t&&t.supports&&t.supports("modulepreload"))return;for(const n of document.querySelectorAll('link[rel="modulepreload"]'))a(n);new MutationObserver(n=>{for(const r of n)if(r.type==="childList")for(const s of r.addedNodes)s.tagName==="LINK"&&s.rel==="modulepreload"&&a(s)}).observe(document,{childList:!0,subtree:!0});function o(n){const r={};return n.integrity&&(r.integrity=n.integrity),n.referrerPolicy&&(r.referrerPolicy=n.referrerPolicy),n.crossOrigin==="use-credentials"?r.credentials="include":n.crossOrigin==="anonymous"?r.credentials="omit":r.credentials="same-origin",r}function a(n){if(n.ep)return;n.ep=!0;const r=o(n);fetch(n.href,r)}})();const m="/transactions";let l=[];function p(e){const t=new URLSearchParams;e.instrument&&t.set("instrument",e.instrument),e.broker&&t.set("broker",e.broker),e.type&&t.set("type",e.type),e.from&&t.set("from",e.from),e.to&&t.set("to",e.to);const o=t.toString();return o?`?${o}`:""}async function f(e){const t=`${m}${p(e)}`,o=await fetch(t);if(!o.ok){const a=await o.json().catch(()=>({error:o.statusText}));throw new Error(a.error||`HTTP ${o.status}`)}return o.json()}function y(e){if(!e)return"";const t=new Date(e),o=String(t.getUTCDate()).padStart(2,"0"),a=String(t.getUTCMonth()+1).padStart(2,"0"),n=t.getUTCFullYear();return`${o}/${a}/${n}`}function b(e,t){return new Intl.NumberFormat(void 0,{style:"currency",currency:t||"USD",minimumFractionDigits:2}).format(e)}function h(e){return e.length===0?'<tr><td colspan="9" class="empty">No transactions found.</td></tr>':e.map((t,o)=>`
      <tr>
        <td>${y(t.created_at)}</td>
        <td><span class="instrument" title="${t.isin||""}">${t.instrument}</span></td>
        <td>${t.isin||"—"}</td>
        <td><span class="badge badge-${t.type}">${t.type}</span></td>
        <td>${t.category}</td>
        <td class="number">${t.quantity}</td>
        <td class="number">${b(t.amount,t.currency)}</td>
        <td>${t.description||"—"}</td>
        <td><button class="btn btn-json" data-tx-idx="${o}">JSON</button></td>
      </tr>`).join("")}function g(){const e=document.getElementById("app");e.innerHTML=`
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
                <th>Actions</th>
              </tr>
            </thead>
            <tbody id="table-body">
              <tr><td colspan="9" class="empty">Use the filters above and click Search.</td></tr>
            </tbody>
          </table>
        </div>
      </section>
    </main>

    <div id="json-modal" class="modal-overlay" role="dialog" aria-modal="true" aria-labelledby="json-modal-title">
      <div class="modal">
        <div class="modal-header">
          <h3 id="json-modal-title">Transaction JSON</h3>
          <button id="json-modal-close" class="modal-close" aria-label="Close">&times;</button>
        </div>
        <pre id="json-modal-content" class="json-content"></pre>
      </div>
    </div>
  `;const t=document.getElementById("filter-form"),o=document.getElementById("table-body"),a=document.getElementById("status"),n=document.getElementById("count");t.addEventListener("submit",async r=>{r.preventDefault();const s=new FormData(t),d={instrument:s.get("instrument")||"",broker:s.get("broker")||"",type:s.get("type")||"",from:s.get("from")||"",to:s.get("to")||""};a.textContent="Loading…",n.textContent="",o.innerHTML="";try{const i=await f(d);l=i,a.textContent="",n.textContent=`${i.length} result${i.length!==1?"s":""}`,o.innerHTML=h(i)}catch(i){a.textContent=`Error: ${i.message}`,o.innerHTML="",n.textContent=""}}),t.addEventListener("reset",()=>{a.textContent="",n.textContent="",l=[],o.innerHTML='<tr><td colspan="9" class="empty">Use the filters above and click Search.</td></tr>'}),o.addEventListener("click",r=>{const s=r.target.closest(".btn-json");if(!s)return;const d=parseInt(s.getAttribute("data-tx-idx")||"-1",10),i=l[d];if(!i)return;const c=document.getElementById("json-modal"),u=document.getElementById("json-modal-content");u.textContent=JSON.stringify(i,null,2),c.classList.add("open")}),document.getElementById("json-modal-close").addEventListener("click",()=>{document.getElementById("json-modal").classList.remove("open")}),document.getElementById("json-modal").addEventListener("click",r=>{r.target===r.currentTarget&&r.currentTarget.classList.remove("open")})}g();
