import { defineConfig } from "astro/config";

// The build lands in Source/Web/static, which the Go binary embeds.
// `npm run dev` proxies /api to a running Silicon on 8420.
export default defineConfig({
  outDir: "../Source/Web/static",
  build: { assets: "assets", inlineStylesheets: "never" },
  vite: {
    build: { assetsInlineLimit: 0, emptyOutDir: false },
    server: { proxy: { "/api": "http://127.0.0.1:8420" } },
  },
});
