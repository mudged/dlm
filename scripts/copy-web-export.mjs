import { cpSync, existsSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const src = join(root, "web", "out");
const dest = join(root, "backend", "internal", "webdist", "dist");
const placeholder = join(dest, "placeholder.txt");
const placeholderBody =
  "Keeps dist/ non-empty for //go:embed before the first UI sync.\n";

if (!existsSync(src)) {
  console.error("Missing web/out. Run: cd web && npm run build");
  process.exit(1);
}

rmSync(dest, { recursive: true, force: true });
mkdirSync(dest, { recursive: true });
cpSync(src, dest, { recursive: true });
// Restore the tracked placeholder so a sync does not delete it from the tree.
writeFileSync(placeholder, placeholderBody);
console.log("copied web/out -> backend/internal/webdist/dist");
