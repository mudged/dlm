import { cpSync, existsSync, mkdirSync, rmSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const src = join(root, "web", "out");
const dest = join(root, "backend", "internal", "webdist", "dist");

if (!existsSync(src)) {
  console.error("Missing web/out. Run: cd web && npm run build");
  process.exit(1);
}

rmSync(dest, { recursive: true, force: true });
mkdirSync(dest, { recursive: true });
cpSync(src, dest, { recursive: true });
console.log("copied web/out -> backend/internal/webdist/dist");
