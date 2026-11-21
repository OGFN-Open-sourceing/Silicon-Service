import { backend, esc, when, ago, act, ask, confirmBox, drawer, me, button, failed } from "./core";
import { icon } from "./icons";

export type Account = {
  accountId: string;
  displayName: string;
  email: string;
  created: number;
  lastLogin: number;
  banned: boolean;
  banReason: string;
  founder: boolean;
  host: boolean;
  discordLinked: boolean;
  online: boolean;
  state: string;
  vbucks: number;
  level: number;
  xp: number;
  bookLevel: number;
  battlePass: boolean;
  acActive: boolean;
  acDriver: boolean;
  acFlagged: string;
  acStrikes: number;
  acExempt: boolean;
  hwidBanned: boolean;
  machines: Machine[];
};

type Machine = {
  kind: string;
  id: string;
  firstSeen: number;
  lastSeen: number;
  banned: boolean;
  sharedWith: { accountId: string; displayName: string }[];
};

const machineLabel: Record<string, string> = { board: "Motherboard", disk: "Boot disk", mac: "Network card", os: "Windows install" };

function machinesHtml(machines: Machine[]): string {
  if (!machines.length) return '<div class="sub">No hardware ids yet. They are recorded the first time this account runs Aether.</div>';
  return `<div class="facts">${machines
    .map(
      (m) => `<div class="fact">
        <div class="stat-label">${esc(machineLabel[m.kind] ?? m.kind)}${m.banned ? ' <span class="pill bad">banned</span>' : ""}</div>
        <div class="mono">${esc(m.id)}</div>
        <div class="sub" title="${esc(when(m.lastSeen))}">seen ${esc(ago(m.lastSeen))}</div>
        ${
          m.sharedWith.length
            ? `<div class="sub">Also used by ${m.sharedWith.map((o) => `<a href="#" data-account="${esc(o.accountId)}">${esc(o.displayName || o.accountId)}</a>`).join(", ")}</div>`
            : ""
        }
      </div>`,
    )
    .join("")}</div>`;
}

const stateLabel: Record<string, string> = { lobby: "In the lobby", queue: "In queue", match: "In a match", offline: "Offline" };

export function flags(a: { banned: boolean; founder: boolean; host: boolean; online: boolean; hwidBanned?: boolean; acExempt?: boolean }): string {
  return [
    a.banned ? '<span class="pill bad">Banned</span>' : "",
    a.hwidBanned ? '<span class="pill bad">HWID banned</span>' : "",
    a.acExempt ? '<span class="pill info plain">AC exempt</span>' : "",
    a.online ? '<span class="pill ok">Online</span>' : "",
    a.founder ? '<span class="pill info plain">Founder</span>' : "",
    a.host ? '<span class="pill warn plain">Host account</span>' : "",
  ].join(" ");
}

export async function openAccount(id: string, changed?: () => unknown) {
  drawer(`<div class="drawer-body"><div class="skeleton" style="height:48px"></div><div class="skeleton"></div><div class="skeleton"></div></div>`);
  let a: Account;
  try {
    a = await backend<Account>("GET", `/accounts/${encodeURIComponent(id)}`);
  } catch (err) {
    drawer(`<div class="drawer-body">${failed(err)}</div>`);
    return;
  }
  const m = await me();
  const owner = m.role === "owner";
  const refresh = async () => {
    await changed?.();
    await openAccount(id, changed);
  };

  const box = drawer(`
    <div class="drawer-head">
      <div class="avatar">${esc(a.displayName.charAt(0))}</div>
      <div style="flex:1;min-width:0"><h3>${esc(a.displayName)}</h3><div class="sub mono">${esc(a.accountId)}</div></div>
      <button class="ghost icon" data-close title="Close">${icon("x")}</button>
    </div>
    <div class="drawer-body">
      <div class="chips">${flags(a) || '<span class="pill plain">Active</span>'}</div>
      ${a.banned && a.banReason ? `<div class="banner">${icon("ban")}<div><strong>Ban reason</strong>${esc(a.banReason)}</div></div>` : ""}
      <div class="facts">
        <div class="fact"><div class="stat-label">V-Bucks</div><b>${a.vbucks.toLocaleString()}</b></div>
        <div class="fact"><div class="stat-label">Level</div><b>${a.level}</b><div class="sub">${a.xp.toLocaleString()} XP</div></div>
        <div class="fact"><div class="stat-label">Battle pass</div><b>Tier ${a.bookLevel}</b><div class="sub">${a.battlePass ? "purchased" : "free pass"}</div></div>
        <div class="fact"><div class="stat-label">Status</div><b style="font-size:15px">${esc(stateLabel[a.state] ?? a.state)}</b></div>
      </div>
      <div>
        <div class="section-title">Account</div>
        <div class="facts">
          <div class="fact"><div class="stat-label">Email</div><div class="mono" style="overflow-wrap:anywhere">${esc(a.email || "-")}</div></div>
          <div class="fact"><div class="stat-label">Discord</div><div>${a.discordLinked ? "linked" : "not linked"}</div></div>
          <div class="fact"><div class="stat-label">Created</div><div>${esc(when(a.created))}</div></div>
          <div class="fact"><div class="stat-label">Last login</div><div title="${esc(when(a.lastLogin))}">${esc(ago(a.lastLogin))}</div></div>
        </div>
      </div>
      <div>
        <div class="section-title">Anti-cheat</div>
        <div class="facts">
          <div class="fact"><div class="stat-label">Aether</div><div>${a.acActive ? (a.acDriver ? "running (kernel)" : "running (user mode)") : "not running"}</div></div>
          <div class="fact"><div class="stat-label">Severe strikes</div><b>${a.acStrikes}</b></div>
          <div class="fact"><div class="stat-label">Blacklist</div><div>${a.acExempt ? "exempt (dev)" : "enforced"}</div></div>
          <div class="fact"><div class="stat-label">Hardware</div><div>${a.hwidBanned ? "banned" : a.machines.length ? "not banned" : "unknown"}</div></div>
          ${a.acFlagged ? `<div class="fact" style="grid-column:1/-1"><div class="stat-label">Flagged</div><div class="mono" style="overflow-wrap:anywhere">${esc(a.acFlagged)}</div></div>` : ""}
        </div>
      </div>
      <div>
        <div class="section-title">Hardware ids</div>
        ${machinesHtml(a.machines)}
      </div>
      <div>
        <div class="section-title">Actions</div>
        <div class="row">
          ${a.online ? button("Kick", { act: "kick" }, "ghost", "kick") : ""}
          ${a.banned ? button("Unban", { act: "unban" }, "ghost", "check") : button("Ban", { act: "ban" }, "danger", "ban")}
          ${owner ? button("V-Bucks", { act: "vbucks" }, "ghost", "coin") : ""}
          ${owner ? button("Grant item", { act: "grant" }, "ghost", "plus") : ""}
          ${owner ? button(a.founder ? "Remove founder" : "Make founder", { act: "founder" }, "ghost", "shield") : ""}
          ${owner ? button(a.acExempt ? "Remove AC exemption" : "Exempt from AC blacklist", { act: "exempt" }, "ghost", "shield") : ""}
          ${owner && a.machines.length ? (a.hwidBanned ? button("Remove HWID ban", { act: "hwid-unban" }, "ghost", "check") : button("HWID ban", { act: "hwid-ban" }, "danger", "ban")) : ""}
        </div>
      </div>
    </div>`);

  const path = `/accounts/${encodeURIComponent(a.accountId)}`;
  box.querySelector("[data-close]")?.addEventListener("click", () => (box.closest("dialog") as HTMLDialogElement).close());
  box.querySelectorAll<HTMLElement>("[data-account]").forEach((el) =>
    el.addEventListener("click", (e) => {
      e.preventDefault();
      openAccount(el.dataset.account!, changed);
    }),
  );
  box.querySelectorAll<HTMLElement>("[data-act]").forEach((el) =>
    el.addEventListener("click", async () => {
      switch (el.dataset.act) {
        case "kick": {
          const reason = await ask(`Kick ${a.displayName}?`, { label: "Reason (optional)", ok: "Kick", danger: true });
          if (reason !== null) act(() => backend("POST", `/players/${encodeURIComponent(a.accountId)}/kick`, { reason }), `Kicked ${a.displayName}`, refresh);
          break;
        }
        case "ban": {
          const reason = await ask(`Ban ${a.displayName}?`, { text: "They are kicked and can't log in until unbanned.", label: "Reason", ok: "Ban", danger: true });
          if (reason !== null) act(() => backend("POST", `${path}/ban`, { reason }), `Banned ${a.displayName}`, refresh);
          break;
        }
        case "unban":
          if (await confirmBox(`Unban ${a.displayName}?`, { ok: "Unban" })) act(() => backend("POST", `${path}/unban`, {}), `Unbanned ${a.displayName}`, refresh);
          break;
        case "vbucks": {
          const v = await ask(`V-Bucks for ${a.displayName}`, { text: `Balance now: ${a.vbucks.toLocaleString()}`, label: "Amount to add (negative removes)", value: "1000", type: "number", ok: "Apply" });
          if (v !== null && Number.isFinite(Number(v))) act(() => backend("POST", `${path}/vbucks`, { amount: Number(v) }), `Updated ${a.displayName}'s V-Bucks`, refresh);
          break;
        }
        case "grant": {
          const id = await ask(`Grant an item to ${a.displayName}`, { text: "A full template id. They see it after a relog.", label: "Template id", value: "AthenaCharacter:CID_028_Athena_Commando_F", ok: "Grant" });
          if (id) act(() => backend("POST", `${path}/grant`, { templateId: id.trim() }), `Granted ${id.trim()}`, refresh);
          break;
        }
        case "founder":
          act(() => backend("POST", `${path}/founder`, { founder: !a.founder }), a.founder ? "Founder removed" : "Founder granted", refresh);
          break;
        case "exempt":
          if (
            a.acExempt ||
            (await confirmBox(`Exempt ${a.displayName} from the Aether blacklist?`, {
              text: "Aether stops blocking and kicking them for open tools like IDEs and debuggers. Detections are still logged. Meant for developers.",
              ok: "Exempt",
            }))
          )
            act(() => backend("POST", `${path}/exempt`, { exempt: !a.acExempt }), a.acExempt ? "AC exemption removed" : `${a.displayName} is exempt`, refresh);
          break;
        case "hwid-ban": {
          const reason = await ask(`HWID ban ${a.displayName}?`, {
            text: "Every PC this account has run Aether on is banned. Any account on those PCs is refused when Aether starts.",
            label: "Reason",
            ok: "HWID ban",
            danger: true,
          });
          if (reason !== null) act(() => backend("POST", `${path}/hwid-ban`, { reason }), `HWID banned ${a.displayName}`, refresh);
          break;
        }
        case "hwid-unban":
          if (await confirmBox(`Remove the HWID ban for ${a.displayName}?`, { ok: "Remove" }))
            act(() => backend("POST", `${path}/hwid-unban`, {}), "HWID ban removed", refresh);
          break;
      }
    }),
  );
}
