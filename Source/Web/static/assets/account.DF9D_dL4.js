import{a as e,b as t,c as n,d as r,h as i,l as a,n as o,o as s,p as c,r as l,s as u,x as d}from"./core.BJARhmdM.js";var f={lobby:`In the lobby`,queue:`In queue`,match:`In a match`,offline:`Offline`};function p(e){return[e.banned?`<span class="pill bad">Banned</span>`:``,e.online?`<span class="pill ok">Online</span>`:``,e.founder?`<span class="pill info plain">Founder</span>`:``,e.host?`<span class="pill warn plain">Host account</span>`:``].join(` `)}async function m(h,g){a(`<div class="drawer-body"><div class="skeleton" style="height:48px"></div><div class="skeleton"></div><div class="skeleton"></div></div>`);let _;try{_=await s(`GET`,`/accounts/${encodeURIComponent(h)}`)}catch(e){a(`<div class="drawer-body">${c(e)}</div>`);return}let v=(await i()).role===`owner`,y=async()=>{await g?.(),await m(h,g)},b=a(`
    <div class="drawer-head">
      <div class="avatar">${r(_.displayName.charAt(0))}</div>
      <div style="flex:1;min-width:0"><h3>${r(_.displayName)}</h3><div class="sub mono">${r(_.accountId)}</div></div>
      <button class="ghost icon" data-close title="Close">${d(`x`)}</button>
    </div>
    <div class="drawer-body">
      <div class="chips">${p(_)||`<span class="pill plain">Active</span>`}</div>
      ${_.banned&&_.banReason?`<div class="banner">${d(`ban`)}<div><strong>Ban reason</strong>${r(_.banReason)}</div></div>`:``}
      <div class="facts">
        <div class="fact"><div class="stat-label">V-Bucks</div><b>${_.vbucks.toLocaleString()}</b></div>
        <div class="fact"><div class="stat-label">Level</div><b>${_.level}</b><div class="sub">${_.xp.toLocaleString()} XP</div></div>
        <div class="fact"><div class="stat-label">Battle pass</div><b>Tier ${_.bookLevel}</b><div class="sub">${_.battlePass?`purchased`:`free pass`}</div></div>
        <div class="fact"><div class="stat-label">Status</div><b style="font-size:15px">${r(f[_.state]??_.state)}</b></div>
      </div>
      <div>
        <div class="section-title">Account</div>
        <div class="facts">
          <div class="fact"><div class="stat-label">Email</div><div class="mono" style="overflow-wrap:anywhere">${r(_.email||`-`)}</div></div>
          <div class="fact"><div class="stat-label">Discord</div><div>${_.discordLinked?`linked`:`not linked`}</div></div>
          <div class="fact"><div class="stat-label">Created</div><div>${r(t(_.created))}</div></div>
          <div class="fact"><div class="stat-label">Last login</div><div title="${r(t(_.lastLogin))}">${r(l(_.lastLogin))}</div></div>
        </div>
      </div>
      <div>
        <div class="section-title">Actions</div>
        <div class="row">
          ${_.online?u(`Kick`,{act:`kick`},`ghost`,`kick`):``}
          ${_.banned?u(`Unban`,{act:`unban`},`ghost`,`check`):u(`Ban`,{act:`ban`},`danger`,`ban`)}
          ${v?u(`V-Bucks`,{act:`vbucks`},`ghost`,`coin`):``}
          ${v?u(_.founder?`Remove founder`:`Make founder`,{act:`founder`},`ghost`,`shield`):``}
        </div>
      </div>
    </div>`),x=`/accounts/${encodeURIComponent(_.accountId)}`;b.querySelector(`[data-close]`)?.addEventListener(`click`,()=>b.closest(`dialog`).close()),b.querySelectorAll(`[data-act]`).forEach(t=>t.addEventListener(`click`,async()=>{switch(t.dataset.act){case`kick`:{let t=await e(`Kick ${_.displayName}?`,{label:`Reason (optional)`,ok:`Kick`,danger:!0});t!==null&&o(()=>s(`POST`,`/players/${encodeURIComponent(_.accountId)}/kick`,{reason:t}),`Kicked ${_.displayName}`,y);break}case`ban`:{let t=await e(`Ban ${_.displayName}?`,{text:`They are kicked and can't log in until unbanned.`,label:`Reason`,ok:`Ban`,danger:!0});t!==null&&o(()=>s(`POST`,`${x}/ban`,{reason:t}),`Banned ${_.displayName}`,y);break}case`unban`:await n(`Unban ${_.displayName}?`,{ok:`Unban`})&&o(()=>s(`POST`,`${x}/unban`,{}),`Unbanned ${_.displayName}`,y);break;case`vbucks`:{let t=await e(`V-Bucks for ${_.displayName}`,{text:`Balance now: ${_.vbucks.toLocaleString()}`,label:`Amount to add (negative removes)`,value:`1000`,type:`number`,ok:`Apply`});t!==null&&Number.isFinite(Number(t))&&o(()=>s(`POST`,`${x}/vbucks`,{amount:Number(t)}),`Updated ${_.displayName}'s V-Bucks`,y);break}case`founder`:o(()=>s(`POST`,`${x}/founder`,{founder:!_.founder}),_.founder?`Founder removed`:`Founder granted`,y)}}))}export{m as n,p as t};