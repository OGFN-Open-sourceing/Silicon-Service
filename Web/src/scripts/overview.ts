import { $, api, backend, esc, ago, act, ask, confirmBox, every, failed } from "../lib/core";
import { icon } from "../lib/icons";

type Status = { version?: string; online: number; inQueue: number; inMatch: number; servers: number; uptimeSeconds: number };
type Sample = { at: string; up: boolean; status?: Status; error?: string };
type Audit = { at: string; user: string; action: string; path?: string; status: number };

let hours = 6;

function uptime(s: number) {
  const d = Math.floor(s / 86400), h = Math.floor((s % 86400) / 3600), m = Math.floor((s % 3600) / 60);
  return d ? `${d}d ${h}h` : h ? `${h}h ${m}m` : `${m}m`;
}

async function stats() {
  try {
    const o = await api<{ latest?: Sample }>("GET", "/overview");
    const l = o.latest;
    const s = l?.status;
    $("#s-state").innerHTML = !l ? '<span class="pill warn">connecting</span>' : l.up ? '<span class="pill ok">Online</span>' : '<span class="pill bad">Offline</span>';
    $("#s-version").textContent = s ? `up ${uptime(s.uptimeSeconds)}${s.version ? " · v" + s.version : ""}` : " ";
    $("#s-online").textContent = s ? String(s.online) : "-";
    $("#s-queue").textContent = s ? String(s.inQueue) : "-";
    $("#s-match").textContent = s ? String(s.inMatch) : "-";
    $("#s-servers").textContent = s ? String(s.servers) : "-";
    $("#problem").innerHTML = l && !l.up
      ? `<div class="banner">${icon("alert")}<div><strong>Silicon can't reach the backend.</strong>${esc(l.error)}</div></div>`
      : "";
  } catch (err) {
    $("#problem").innerHTML = failed(err);
  }
}

async function chart() {
  const samples = await api<Sample[]>("GET", `/activity?hours=${hours}`).catch(() => []);
  const el = $("#chart");
  if (samples.length < 2) {
    el.innerHTML = '<div class="empty">Collecting samples. The chart fills in as Silicon polls the backend.</div>';
    $("#peak").textContent = "";
    return;
  }
  const pts = samples.map((x) => (x.up && x.status ? x.status.online : 0));
  const max = Math.max(1, ...pts);
  const w = 1000, h = 220, pad = 8;
  const xy = pts.map((v, i) => [(i / (pts.length - 1)) * w, h - pad - (v / max) * (h - pad * 2)] as const);
  const line = xy.map(([x, y], i) => `${i ? "L" : "M"}${x.toFixed(1)},${y.toFixed(1)}`).join("");
  const grid = [0.25, 0.5, 0.75].map((f) => `<line class="grid-line" x1="0" x2="${w}" y1="${h * f}" y2="${h * f}"/>`).join("");
  el.innerHTML = `<svg class="chart" viewBox="0 0 ${w} ${h}" preserveAspectRatio="none">
    <defs><linearGradient id="chart-fill" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#26bbff" stop-opacity=".35"/><stop offset="1" stop-color="#26bbff" stop-opacity="0"/></linearGradient></defs>
    ${grid}<path class="area" d="${line}L${w},${h}L0,${h}Z"/><path class="line" d="${line}"/></svg>`;
  $("#peak").textContent = `peak ${Math.max(...pts)}`;
}

async function feed() {
  const entries = await api<Audit[]>("GET", "/audit").catch(() => []);
  $("#feed").innerHTML = entries.length
    ? entries.slice(0, 6).map((e) => `<div class="feed-item"><span class="feed-dot${e.status >= 400 ? " bad" : ""}"></span><div style="flex:1"><strong>${esc(e.user)}</strong> <span class="muted">${esc(e.action)}</span> <span class="mono">${esc(e.path ?? "")}</span></div><span class="muted">${esc(ago(e.at))}</span></div>`).join("")
    : '<div class="empty">No admin actions yet.</div>';
}

$("#range").addEventListener("click", (e) => {
  const b = (e.target as HTMLElement).closest("button");
  if (!b) return;
  hours = Number(b.dataset.hours);
  document.querySelectorAll("#range button").forEach((x) => x.classList.toggle("on", x === b));
  chart();
});

$("#announce").addEventListener("click", async () => {
  const message = await ask("Send announcement", { label: "Message to everyone online", ok: "Send" });
  if (message?.trim()) act(() => backend("POST", "/announce", { message }), "Announcement sent", feed);
});

const restart = async (mode: "graceful" | "now") => {
  const ok = await confirmBox(mode === "now" ? "Restart right now?" : "Restart with a warning?", {
    text: mode === "now" ? "Everyone online is disconnected immediately." : "Players get a 75 second warning, then the backend restarts.",
    ok: "Restart",
    danger: true,
  });
  if (ok) act(() => backend("POST", "/restart", { mode }), "Restart requested", feed);
};
$("#restart-soft").addEventListener("click", () => restart("graceful"));
$("#restart-now").addEventListener("click", () => restart("now"));

stats();
chart();
feed();
every(10_000, stats);
every(30_000, chart);
every(15_000, feed);
