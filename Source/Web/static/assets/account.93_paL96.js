import{a as e,b as t,c as n,d as r,h as i,l as a,n as o,o as s,p as c,r as l,s as u,x as d}from"./core.BJARhmdM.js";var f={board:`Motherboard`,disk:`Boot disk`,mac:`Network card`,os:`Windows install`};function p(e){return e.length?`<div class="facts">${e.map(e=>`<div class="fact">
        <div class="stat-label">${r(f[e.kind]??e.kind)}${e.banned?` <span class="pill bad">banned</span>`:``}</div>
        <div class="mono">${r(e.id)}</div>
        <div class="sub" title="${r(t(e.lastSeen))}">seen ${r(l(e.lastSeen))}</div>
        ${e.sharedWith.length?`<div class="sub">Also used by ${e.sharedWith.map(e=>`<a href="#" data-account="${r(e.accountId)}">${r(e.displayName||e.accountId)}</a>`).join(`, `)}</div>`:``}
      </div>`).join(``)}</div>`:`<div class="sub">No hardware ids yet. They are recorded the first time this account runs Aether.</div>`}var m={lobby:`In the lobby`,queue:`In queue`,match:`In a match`,offline:`Offline`};function h(e){return[e.banned?`<span class="pill bad">Banned</span>`:``,e.hwidBanned?`<span class="pill bad">HWID banned</span>`:``,e.acExempt?`<span class="pill info plain">AC exempt</span>`:``,e.online?`<span class="pill ok">Online</span>`:``,e.founder?`<span class="pill info plain">Founder</span>`:``,e.host?`<span class="pill warn plain">Host account</span>`:``].join(` `)}async function g(f,_){a(`<div class="drawer-body"><div class="skeleton" style="height:48px"></div><div class="skeleton"></div><div class="skeleton"></div></div>`);let v;try{v=await s(`GET`,`/accounts/${encodeURIComponent(f)}`)}catch(e){a(`<div class="drawer-body">${c(e)}</div>`);return}let y=(await i()).role===`owner`,b=async()=>{await _?.(),await g(f,_)},x=a(`
    <div class="drawer-head">
      <div class="avatar">${r(v.displayName.charAt(0))}</div>
      <div style="flex:1;min-width:0"><h3>${r(v.displayName)}</h3><div class="sub mono">${r(v.accountId)}</div></div>
      <button class="ghost icon" data-close title="Close">${d(`x`)}</button>
    </div>
    <div class="drawer-body">
      <div class="chips">${h(v)||`<span class="pill plain">Active</span>`}</div>
      ${v.banned&&v.banReason?`<div class="banner">${d(`ban`)}<div><strong>Ban reason</strong>${r(v.banReason)}</div></div>`:``}
      <div class="facts">
        <div class="fact"><div class="stat-label">V-Bucks</div><b>${v.vbucks.toLocaleString()}</b></div>
        <div class="fact"><div class="stat-label">Level</div><b>${v.level}</b><div class="sub">${v.xp.toLocaleString()} XP</div></div>
        <div class="fact"><div class="stat-label">Battle pass</div><b>Tier ${v.bookLevel}</b><div class="sub">${v.battlePass?`purchased`:`free pass`}</div></div>
        <div class="fact"><div class="stat-label">Status</div><b style="font-size:15px">${r(m[v.state]??v.state)}</b></div>
      </div>
      <div>
        <div class="section-title">Account</div>
        <div class="facts">
          <div class="fact"><div class="stat-label">Email</div><div class="mono" style="overflow-wrap:anywhere">${r(v.email||`-`)}</div></div>
          <div class="fact"><div class="stat-label">Discord</div><div>${v.discordLinked?`linked`:`not linked`}</div></div>
          <div class="fact"><div class="stat-label">Created</div><div>${r(t(v.created))}</div></div>
          <div class="fact"><div class="stat-label">Last login</div><div title="${r(t(v.lastLogin))}">${r(l(v.lastLogin))}</div></div>
        </div>
      </div>
      <div>
        <div class="section-title">Anti-cheat</div>
        <div class="facts">
          <div class="fact"><div class="stat-label">Aether</div><div>${v.acActive?v.acDriver?`running (kernel)`:`running (user mode)`:`not running`}</div></div>
          <div class="fact"><div class="stat-label">Severe strikes</div><b>${v.acStrikes}</b></div>
          <div class="fact"><div class="stat-label">Blacklist</div><div>${v.acExempt?`exempt (dev)`:`enforced`}</div></div>
          <div class="fact"><div class="stat-label">Hardware</div><div>${v.hwidBanned?`banned`:v.machines.length?`not banned`:`unknown`}</div></div>
          ${v.acFlagged?`<div class="fact" style="grid-column:1/-1"><div class="stat-label">Flagged</div><div class="mono" style="overflow-wrap:anywhere">${r(v.acFlagged)}</div></div>`:``}
        </div>
      </div>
      <div>
        <div class="section-title">Hardware ids</div>
        ${p(v.machines)}
      </div>
      <div>
        <div class="section-title">Actions</div>
        <div class="row">
          ${v.online?u(`Kick`,{act:`kick`},`ghost`,`kick`):``}
          ${v.banned?u(`Unban`,{act:`unban`},`ghost`,`check`):u(`Ban`,{act:`ban`},`danger`,`ban`)}
          ${y?u(`V-Bucks`,{act:`vbucks`},`ghost`,`coin`):``}
          ${y?u(`Grant item`,{act:`grant`},`ghost`,`plus`):``}
          ${y?u(v.founder?`Remove founder`:`Make founder`,{act:`founder`},`ghost`,`shield`):``}
          ${y?u(v.acExempt?`Remove AC exemption`:`Exempt from AC blacklist`,{act:`exempt`},`ghost`,`shield`):``}
          ${y&&v.machines.length?v.hwidBanned?u(`Remove HWID ban`,{act:`hwid-unban`},`ghost`,`check`):u(`HWID ban`,{act:`hwid-ban`},`danger`,`ban`):``}
        </div>
      </div>
    </div>`),S=`/accounts/${encodeURIComponent(v.accountId)}`;x.querySelector(`[data-close]`)?.addEventListener(`click`,()=>x.closest(`dialog`).close()),x.querySelectorAll(`[data-account]`).forEach(e=>e.addEventListener(`click`,t=>{t.preventDefault(),g(e.dataset.account,_)})),x.querySelectorAll(`[data-act]`).forEach(t=>t.addEventListener(`click`,async()=>{switch(t.dataset.act){case`kick`:{let t=await e(`Kick ${v.displayName}?`,{label:`Reason (optional)`,ok:`Kick`,danger:!0});t!==null&&o(()=>s(`POST`,`/players/${encodeURIComponent(v.accountId)}/kick`,{reason:t}),`Kicked ${v.displayName}`,b);break}case`ban`:{let t=await e(`Ban ${v.displayName}?`,{text:`They are kicked and can't log in until unbanned.`,label:`Reason`,ok:`Ban`,danger:!0});t!==null&&o(()=>s(`POST`,`${S}/ban`,{reason:t}),`Banned ${v.displayName}`,b);break}case`unban`:await n(`Unban ${v.displayName}?`,{ok:`Unban`})&&o(()=>s(`POST`,`${S}/unban`,{}),`Unbanned ${v.displayName}`,b);break;case`vbucks`:{let t=await e(`V-Bucks for ${v.displayName}`,{text:`Balance now: ${v.vbucks.toLocaleString()}`,label:`Amount to add (negative removes)`,value:`1000`,type:`number`,ok:`Apply`});t!==null&&Number.isFinite(Number(t))&&o(()=>s(`POST`,`${S}/vbucks`,{amount:Number(t)}),`Updated ${v.displayName}'s V-Bucks`,b);break}case`grant`:{let t=await e(`Grant an item to ${v.displayName}`,{text:`A full template id. They see it after a relog.`,label:`Template id`,value:`AthenaCharacter:CID_028_Athena_Commando_F`,ok:`Grant`});t&&o(()=>s(`POST`,`${S}/grant`,{templateId:t.trim()}),`Granted ${t.trim()}`,b);break}case`founder`:o(()=>s(`POST`,`${S}/founder`,{founder:!v.founder}),v.founder?`Founder removed`:`Founder granted`,b);break;case`exempt`:(v.acExempt||await n(`Exempt ${v.displayName} from the Aether blacklist?`,{text:`Aether stops blocking and kicking them for open tools like IDEs and debuggers. Detections are still logged. Meant for developers.`,ok:`Exempt`}))&&o(()=>s(`POST`,`${S}/exempt`,{exempt:!v.acExempt}),v.acExempt?`AC exemption removed`:`${v.displayName} is exempt`,b);break;case`hwid-ban`:{let t=await e(`HWID ban ${v.displayName}?`,{text:`Every PC this account has run Aether on is banned. Any account on those PCs is refused when Aether starts.`,label:`Reason`,ok:`HWID ban`,danger:!0});t!==null&&o(()=>s(`POST`,`${S}/hwid-ban`,{reason:t}),`HWID banned ${v.displayName}`,b);break}case`hwid-unban`:await n(`Remove the HWID ban for ${v.displayName}?`,{ok:`Remove`})&&o(()=>s(`POST`,`${S}/hwid-unban`,{}),`HWID ban removed`,b)}}))}export{g as n,h as t};