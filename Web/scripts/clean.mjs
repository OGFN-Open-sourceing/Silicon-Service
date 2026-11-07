import { rmSync, mkdirSync, writeFileSync } from "node:fs";

const out = new URL("../../Source/Web/static/", import.meta.url);
rmSync(out, { recursive: true, force: true });
mkdirSync(out, { recursive: true });
writeFileSync(new URL(".gitkeep", out), "");
