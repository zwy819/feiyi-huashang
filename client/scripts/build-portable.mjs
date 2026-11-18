import { cpSync } from "node:fs";
import { build } from "vite";

const output = "../server/internal/portable/assets/web";

await build({
  build: { outDir: output, emptyOutDir: true, copyPublicDir: false },
});

cpSync("public/logo.png", `${output}/logo.png`);
cpSync("public/geojson", `${output}/geojson`, { recursive: true });
