import { $, api, me } from "./core";

type Overview = { latest?: { up: boolean; error?: string; status?: { online: number } } };

async function pill() {
  const el = $("#backend-pill");
  try {
    const o = await api<Overview>("GET", "/overview");
    const up = !!o.latest?.up;
    el.className = "pill " + (o.latest ? (up ? "ok" : "bad") : "warn");
    el.textContent = !o.latest ? "connecting" : up ? `online · ${o.latest.status?.online ?? 0} players` : "backend offline";
    el.title = o.latest?.error ?? "";
  } catch {
    el.className = "pill bad";
    el.textContent = "Silicon unreachable";
  }
}

me().then((m) => {
  $("#backend-name").textContent = m.backend;
  $("#me-name").textContent = m.user;
  $("#me-role").textContent = m.role;
  $("#me-initial").textContent = m.user.charAt(0);
  document.querySelectorAll<HTMLElement>("[data-owner]").forEach((el) => el.classList.toggle("hidden", m.role !== "owner"));
});

$("#logout").addEventListener("click", async () => {
  await api("POST", "/logout").catch(() => {});
  location.href = "/login/";
});

pill();
setInterval(() => document.visibilityState === "visible" && pill(), 10_000);
