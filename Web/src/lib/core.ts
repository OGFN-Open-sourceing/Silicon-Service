import { icon, type IconName } from "./icons";

export type Me = { user: string; role: "owner" | "moderator"; backend: string };

export const $ = <T extends HTMLElement = HTMLElement>(sel: string, root: ParentNode = document) =>
  root.querySelector(sel) as T;

const entities: Record<string, string> = { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" };
export function esc(v: unknown): string {
  return String(v ?? "").replace(/[&<>"']/g, (c) => entities[c]);
}

export function when(t: number | string | undefined): string {
  if (!t) return "";
  return new Date(t).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
}

export function ago(t: number | string | undefined): string {
  if (!t) return "never";
  const s = Math.max(0, (Date.now() - new Date(t).getTime()) / 1000);
  if (s < 60) return "just now";
  if (s < 3600) return `${Math.floor(s / 60)} min ago`;
  if (s < 86400) return `${Math.floor(s / 3600)} h ago`;
  return `${Math.floor(s / 86400)} d ago`;
}

export async function api<T = any>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch("/api" + path, {
    method,
    headers: body !== undefined ? { "Content-Type": "application/json" } : {},
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (res.status === 401 && path !== "/login") {
    location.href = "/login/";
    throw new Error("not logged in");
  }
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || `Request failed (${res.status})`);
  return data as T;
}

export const backend = <T = any>(method: string, path: string, body?: unknown) => api<T>(method, "/b" + path, body);

let mePromise: Promise<Me> | null = null;
export const me = () => (mePromise ??= api<Me>("GET", "/me"));

export function toast(message: string, bad = false) {
  const el = document.createElement("div");
  el.className = "toast" + (bad ? " bad" : "");
  el.innerHTML = icon(bad ? "alert" : "check") + `<span>${esc(message)}</span>`;
  $("#toasts").append(el);
  setTimeout(() => el.remove(), bad ? 6000 : 3500);
}

type DialogOpts = { text?: string; label?: string; value?: string; placeholder?: string; ok?: string; danger?: boolean; type?: string };

function openDialog(title: string, opts: DialogOpts, withInput: boolean): Promise<string | null> {
  const dlg = $<HTMLDialogElement>("#dialog");
  const input = $<HTMLInputElement>("#dialog-input");
  const ok = $<HTMLButtonElement>("#dialog-ok");
  $("#dialog-title").textContent = title;
  $("#dialog-text").textContent = opts.text ?? "";
  $("#dialog-text").classList.toggle("hidden", !opts.text);
  $("#dialog-field").classList.toggle("hidden", !withInput);
  $("#dialog-label").textContent = opts.label ?? "";
  input.type = opts.type ?? "text";
  input.value = opts.value ?? "";
  input.placeholder = opts.placeholder ?? "";
  ok.textContent = opts.ok ?? "OK";
  ok.className = opts.danger ? "danger solid" : "";
  const form = dlg.querySelector("form")!;
  dlg.showModal();
  if (withInput) input.select();
  return new Promise((resolve) => {
    const cancel = $("#dialog-cancel");
    const finish = (value: string | null) => {
      form.removeEventListener("submit", onSubmit);
      dlg.removeEventListener("cancel", onCancel);
      cancel.removeEventListener("click", onCancel);
      if (dlg.open) dlg.close();
      resolve(value);
    };
    const onSubmit = (e: Event) => {
      e.preventDefault();
      finish(input.value);
    };
    const onCancel = (e: Event) => {
      e.preventDefault();
      finish(null);
    };
    form.addEventListener("submit", onSubmit);
    dlg.addEventListener("cancel", onCancel);
    cancel.addEventListener("click", onCancel);
  });
}

export function drawer(html: string): HTMLElement {
  const dlg = $<HTMLDialogElement>("#drawer");
  const box = $("#drawer-content");
  box.innerHTML = html;
  if (!dlg.open) dlg.showModal();
  dlg.onclick = (e) => {
    if (e.target === dlg) dlg.close();
  };
  return box;
}

export const closeDrawer = () => $<HTMLDialogElement>("#drawer").close();

export const confirmBox =async (title: string, opts: DialogOpts = {}) => (await openDialog(title, opts, false)) !== null;
export const ask = (title: string, opts: DialogOpts = {}) => openDialog(title, opts, true);

export async function act(fn: () => Promise<unknown>, done = "Done", after?: () => unknown) {
  try {
    await fn();
    toast(done);
    await after?.();
  } catch (err) {
    toast((err as Error).message, true);
  }
}

export function empty(iconName: IconName, text: string): string {
  return `<div class="empty">${icon(iconName)}<div>${esc(text)}</div></div>`;
}

export function failed(err: unknown): string {
  return `<div class="banner">${icon("alert")}<div><strong>Couldn't load this.</strong>${esc((err as Error).message)}</div></div>`;
}

export function table(headers: string[], rows: string[][], emptyHtml: string): string {
  if (!rows.length) return emptyHtml;
  const cls = (i: number) => (i === headers.length - 1 && headers[i] === "" ? ' class="actions"' : "");
  const head = headers.map((h, i) => `<th${cls(i)}>${esc(h)}</th>`).join("");
  const body = rows.map((r) => "<tr>" + r.map((c, i) => `<td${cls(i)}>${c}</td>`).join("") + "</tr>").join("");
  return `<div class="table-wrap"><table><thead><tr>${head}</tr></thead><tbody>${body}</tbody></table></div>`;
}

export function person(name: string, sub: string): string {
  const initial = esc((name || "?").trim().charAt(0) || "?");
  return `<div class="person"><div class="avatar">${initial}</div><div><strong>${esc(name)}</strong><div class="sub mono">${esc(sub)}</div></div></div>`;
}

export function onClick(root: ParentNode, selector: string, handler: (data: DOMStringMap, el: HTMLElement) => void) {
  root.querySelectorAll<HTMLElement>(selector).forEach((el) => el.addEventListener("click", () => handler(el.dataset, el)));
}

export function loading(rows = 4): string {
  return `<div class="card-body">${'<div class="skeleton" style="margin:14px 0"></div>'.repeat(rows)}</div>`;
}

export function every(ms: number, fn: () => unknown) {
  setInterval(() => document.visibilityState === "visible" && fn(), ms);
}

export function button(label: string, data: Record<string, string>, kind = "ghost", iconName?: IconName): string {
  const attrs = Object.entries(data).map(([k, v]) => `data-${k}="${esc(v)}"`).join(" ");
  return `<button class="${kind} small" ${attrs}>${iconName ? icon(iconName) : ""}${esc(label)}</button>`;
}
