# Models page leads with video Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the Models screen list models and offer **Create from video** as the main way to add one, with **Upload file** as the quieter second way, and open the new-model page on the video tab unless `tab=csv`.

**Architecture:** A pure helper in `web/lib/newModelRoute.ts` decides the tab. The new-model page uses it, puts the video tab first, and renames the file tab. The Models list grows a second header button and new sentences. Docs match those words. The CSV form, the video form, and each model row stay as they are.

**Tech Stack:** Next.js, React, TypeScript, Vitest, Font Awesome (`faVideo`, `faUpload`). No new dependencies. No Go changes.

## Global Constraints

- No new `REQ-*`. This is wording and entry order for the list (REQ-002) and the two create paths (REQ-006, REQ-049).
- `/models` lists the models you have. Adding one is two header actions.
- **Create from video** goes to `/models/new?tab=video`. **Upload file** goes to `/models/new?tab=csv` and is visually quieter.
- On a narrow screen the two buttons stack full width, video on top.
- List rows stay as they are: name link, **View**, and **Delete**. Do not add edit-from-the-list.
- New-model tabs are **Create from video** first and **Upload file** second. The file tab's internal name stays `csv`.
- A visit with no `tab`, an empty `tab`, or any value other than `csv` opens **Create from video**. Only `tab=csv` opens the file form.
- The file form stays unchanged: label **CSV file**, header `id,x,y,z`, error `Choose a CSV file.`, submit **Upload**.
- The video form stays unchanged.
- `tab=video` and `tab=csv` stay, including while a capture `jobId` is in the URL.
- Do not edit the API notes that describe `POST /api/v1/models` as a CSV upload. Do not renumber a `§` section.
- The device page sentence "upload the recorded videos later via Models → Create from video" stays.
- There is no page-render test runner in `web/`. Do not add one for these labels.
- Tests run from `web/`:

```bash
cd /workspaces/dlm/web && npx vitest run lib/newModelRoute.test.ts
```

## File map

- `web/lib/newModelRoute.ts` — add `activeNewModelTab`. `buildNewModelUrl` stays.
- `web/lib/newModelRoute.test.ts` — cases for the helper.
- `web/app/models/new/NewModelClient.tsx` — default tab, tab order, tab label, intro sentence.
- `web/app/models/page.tsx` — subtitle, two header buttons, empty-list links.
- `docs/requirements/requirements.md` — the add-a-model bullet and section 11 step 3.
- `docs/requirements/acceptance-criteria.md` — one new try/see.
- `docs/userguide/using-the-app.md` — the CSV section's Models sentence.
- `docs/userguide/build-model-from-video.md` — one sentence after the existing Models sentence.
- `docs/design/frontend.md` — §4.6 and the opening of §4.17.

---

### Task 1: Default the new-model tab to video

**Files:**
- Modify: `web/lib/newModelRoute.ts`
- Test: `web/lib/newModelRoute.test.ts`

**Interfaces:**
- Consumes: `export type NewModelTab = "csv" | "video"` already in `web/lib/newModelRoute.ts`
- Produces: `activeNewModelTab(rawTab: string | null): NewModelTab`. Returns `"csv"` only when `rawTab === "csv"`. Every other value, including `null` and `""`, returns `"video"`.

- [ ] **Step 1: Write the failing test**

In `web/lib/newModelRoute.test.ts`, add `activeNewModelTab` to the existing import:

```ts
import {
  activeNewModelTab,
  buildNewModelUrl,
  initialVideoPhaseFromJobId,
} from "./newModelRoute";
```

Add this describe block at the end of the file:

```ts
describe("activeNewModelTab", () => {
  it("opens the file tab only for csv", () => {
    expect(activeNewModelTab("csv")).toBe("csv");
  });

  it("opens the video tab when tab is missing, video, or anything else", () => {
    expect(activeNewModelTab(null)).toBe("video");
    expect(activeNewModelTab("")).toBe("video");
    expect(activeNewModelTab("video")).toBe("video");
    expect(activeNewModelTab("CSV")).toBe("video");
    expect(activeNewModelTab("nope")).toBe("video");
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd /workspaces/dlm/web && npx vitest run lib/newModelRoute.test.ts`

Expected: FAIL because `activeNewModelTab` is not exported from `./newModelRoute`.

- [ ] **Step 3: Write the minimal implementation**

In `web/lib/newModelRoute.ts`, add this function after `buildNewModelUrl`:

```ts
/** Video is the usual path. Only an explicit csv tab opens the file form. */
export function activeNewModelTab(rawTab: string | null): NewModelTab {
  return rawTab === "csv" ? "csv" : "video";
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd /workspaces/dlm/web && npx vitest run lib/newModelRoute.test.ts`

Expected: PASS (the new cases and the existing `buildNewModelUrl` / `initialVideoPhaseFromJobId` cases).

- [ ] **Step 5: Commit**

```bash
git add web/lib/newModelRoute.ts web/lib/newModelRoute.test.ts
git commit -m "$(cat <<'EOF'
feat(web): default the new-model page to the video tab

Filming is the usual way to add a model, so only an explicit csv tab opens the file form.
EOF
)"
```

---

### Task 2: Lead the new-model page with video

**Files:**
- Modify: `web/app/models/new/NewModelClient.tsx` (import block, `TabBar` around lines 717–783, `NewModelClient` around lines 787–808)
- Do not modify `CsvPanel` or `VideoPanel`.

**Interfaces:**
- Consumes: `activeNewModelTab(rawTab: string | null): NewModelTab` and `buildNewModelUrl(tab: NewModelTab, jobId?: string | null): string` from `web/lib/newModelRoute.ts`
- Produces: the page heading **New model**, the intro sentence below, tabs in order **Create from video** then **Upload file**, and a selected tab from `activeNewModelTab`.

- [ ] **Step 1: Use the helper and the shared tab type**

Replace the import from `@/lib/newModelRoute` with:

```ts
import {
  activeNewModelTab,
  buildNewModelUrl,
  initialVideoPhaseFromJobId,
  type NewModelTab,
} from "@/lib/newModelRoute";
```

Delete the local line `type Tab = "csv" | "video";`.

Change `TabBar`'s props from `Tab` to `NewModelTab`:

```ts
function TabBar({
  active,
  onSelect,
}: {
  active: NewModelTab;
  onSelect: (t: NewModelTab) => void;
}) {
```

In `NewModelClient`, replace:

```ts
  const rawTab = params.get("tab");
  const activeTab: Tab = rawTab === "video" ? "video" : "csv";
  const jobId = params.get("jobId");

  function switchTab(t: Tab) {
    router.push(buildNewModelUrl(t, jobId));
  }
```

with:

```ts
  const rawTab = params.get("tab");
  const activeTab = activeNewModelTab(rawTab);
  const jobId = params.get("jobId");

  function switchTab(t: NewModelTab) {
    router.push(buildNewModelUrl(t, jobId));
  }
```

- [ ] **Step 2: Put the video tab first and rename the file tab**

Replace the two `<button role="tab">` elements inside `TabBar`, keeping `base`, `activeClass`, and `inactiveClass`. The video button comes first. The file button's visible label becomes `Upload file`. The `aria-selected` checks stay `active === "video"` and `active === "csv"`.

```tsx
      <button
        role="tab"
        aria-selected={active === "video"}
        type="button"
        className={`${base} ${active === "video" ? activeClass : inactiveClass}`}
        onClick={() => onSelect("video")}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4 shrink-0"
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden
        >
          <path d="M2 6a2 2 0 012-2h6a2 2 0 012 2v8a2 2 0 01-2 2H4a2 2 0 01-2-2V6zM14.553 7.106A1 1 0 0014 8v4a1 1 0 00.553.894l2 1A1 1 0 0018 13V7a1 1 0 00-1.447-.894l-2 1z" />
        </svg>
        Create from video
      </button>
      <button
        role="tab"
        aria-selected={active === "csv"}
        type="button"
        className={`${base} ${active === "csv" ? activeClass : inactiveClass}`}
        onClick={() => onSelect("csv")}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4 shrink-0"
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden
        >
          <path
            fillRule="evenodd"
            d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z"
            clipRule="evenodd"
          />
        </svg>
        Upload file
      </button>
```

- [ ] **Step 3: Lead the intro with filming**

Replace the paragraph under the **New model** heading with:

```tsx
        <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">
          Film your lights and the app works out where they are. Or upload a
          file of light coordinates.
        </p>
```

- [ ] **Step 4: Check the old words are gone and the helper tests still pass**

Run:

```bash
cd /workspaces/dlm/web && npx vitest run lib/newModelRoute.test.ts
rg -n "CSV upload|Upload a CSV of light coordinates|rawTab === \"video\"" web/app/models/new/NewModelClient.tsx
```

Expected: tests PASS. `rg` prints no matches. `CsvPanel` still contains the label `CSV file` and the error `Choose a CSV file.`

- [ ] **Step 5: Commit**

```bash
git add web/app/models/new/NewModelClient.tsx
git commit -m "$(cat <<'EOF'
feat(web): lead the new-model page with create from video

The video tab is first and selected unless the address asks for the file tab.
EOF
)"
```

---

### Task 3: Lead the Models list with video

**Files:**
- Modify: `web/app/models/page.tsx` (imports, the header around lines 103–124, the empty state around lines 147–156)
- Do not change the model `<li>` markup (name link, **View**, **Delete**).

**Interfaces:**
- Consumes: `buildNewModelUrl(tab: NewModelTab, jobId?: string | null): string` from `web/lib/newModelRoute.ts`
- Produces: header buttons **Create from video** then **Upload file**, the subtitle and empty sentence below.

- [ ] **Step 1: Import the video icon and the URL helper**

Replace the Font Awesome import with:

```ts
import {
  faEye,
  faTrash,
  faUpload,
  faVideo,
} from "@fortawesome/free-solid-svg-icons";
```

Add this import next to the other `@/lib` imports:

```ts
import { buildNewModelUrl } from "@/lib/newModelRoute";
```

- [ ] **Step 2: Replace the subtitle and the single Upload CSV button**

Replace from the subtitle `<p` through the Upload CSV `</Button>`, including the `</div>` that closes the title column, with:

```tsx
          <p className="mt-1 max-w-xl text-sm text-slate-600 dark:text-slate-400">
            Film your lights and the app works out where they are. Or upload a
            file with columns{" "}
            <code className="rounded bg-slate-200 px-1 dark:bg-slate-800">
              id,x,y,z
            </code>
            . Up to 1000 lights per model.
          </p>
        </div>
        <div className="flex w-full flex-col gap-2 sm:w-auto">
          <Button
            type="button"
            icon={faVideo}
            className="w-full sm:w-auto"
            onClick={() => router.push(buildNewModelUrl("video"))}
          >
            Create from video
          </Button>
          <Button
            type="button"
            icon={faUpload}
            className="w-full bg-slate-600 dark:bg-slate-600 sm:w-auto"
            onClick={() => router.push(buildNewModelUrl("csv"))}
          >
            Upload file
          </Button>
        </div>
```

The header's outer `flex flex-col ... sm:flex-row` wrapper stays. Keep the closing `</header>` that already follows the button.

- [ ] **Step 3: Lead the empty list with both links**

Replace the empty-state paragraph with:

```tsx
        <p className="text-sm text-slate-600 dark:text-slate-400">
          No models yet.{" "}
          <Link
            href={buildNewModelUrl("video")}
            className="font-medium text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
          >
            Create from video
          </Link>
          , or{" "}
          <Link
            href={buildNewModelUrl("csv")}
            className="font-medium text-sky-700 underline-offset-4 hover:underline dark:text-sky-400"
          >
            upload a file
          </Link>
          .
        </p>
```

- [ ] **Step 4: Check the old words are gone**

Run:

```bash
cd /workspaces/dlm && rg -n "Upload CSV|Upload a CSV" web/app/models/page.tsx
cd /workspaces/dlm/web && npx vitest run lib/newModelRoute.test.ts
```

Expected: `rg` prints no matches. Tests PASS. The file still contains `View` and `Delete`.

- [ ] **Step 5: Commit**

```bash
git add web/app/models/page.tsx
git commit -m "$(cat <<'EOF'
feat(web): lead the models list with create from video

The list keeps its rows, and uploading a coordinate file is the second action.
EOF
)"
```

---

### Task 4: Match the guides to the two buttons

**Files:**
- Modify: `docs/requirements/requirements.md`
- Modify: `docs/requirements/acceptance-criteria.md`
- Modify: `docs/userguide/using-the-app.md`
- Modify: `docs/userguide/build-model-from-video.md`
- Modify: `docs/design/frontend.md`
- Modify: `docs/superpowers/specs/2026-10-18-models-page-video-first-design.md` (status line only, if it still says `Pending review`)
- Do not modify `docs/engineering/`, `web/app/devices/detail/DeviceDetailClient.tsx`, or any `§` number.

**Interfaces:**
- Consumes: the visible labels from Tasks 2 and 3: **Create from video**, **Upload file**, `/models/new?tab=video`, `/models/new?tab=csv`
- Produces: the same story in the requirements, the acceptance checklist, the user guide, and §4.6 / §4.17.

- [ ] **Step 1: Update what you can do with a model**

In `docs/requirements/requirements.md`, under "What you can do with models", replace:

```markdown
- **Add a new model** by uploading a CSV file.
```

with:

```markdown
- **Add a new model** by filming your lights (**Create from video**), or by uploading a file.
```

In section 11, replace this sentence:

```markdown
On the model screen there's a **"create from video"** option alongside the CSV
   upload.
```

with:

```markdown
On the Models screen, **Create from video** is the main button. **Upload file** is the second button, for a spreadsheet of coordinates.
```

Leave the bold lead-in `3. **Upload the videos.**` and the following sentence `You upload your clips.`

- [ ] **Step 2: Add the Models-screen try/see**

In `docs/requirements/acceptance-criteria.md`, under "Building a model from video", insert this block immediately before `**You can build a model from uploaded videos.**`:

```markdown
**The Models screen leads with filming.**
- Try: open the Models screen, including when it has models and when the list is empty.
- You should see: **Create from video** as the main button and a quieter **Upload file** button. Create from video opens the new-model page on the video tab. Upload file opens it on the file tab. The page text mentions filming first, then a file whose columns are `id,x,y,z`. With no models, the empty sentence offers the same two links.

```

Leave the existing "choose create from video" try/see as it is.

- [ ] **Step 3: Update the user guide**

In `docs/userguide/using-the-app.md`, in "Uploading a file (CSV)", replace:

```markdown
On the **Models** screen you can upload one to create a model.
```

with:

```markdown
On the **Models** screen, **Upload file** is the second button, next to **Create from video**.
```

In `docs/userguide/build-model-from-video.md`, after the sentence that begins `On the **Models** screen, choose **Create from video** and upload your recordings.` add this sentence before the supported-video sentence:

```markdown
**Upload file** is the other button on that screen, for when you already have a spreadsheet of positions.
```

- [ ] **Step 4: Update the design notes**

In `docs/design/frontend.md` §4.6, replace:

```markdown
**In plain terms:** The Models area lists models (a model = a set of 3D light positions), lets you upload one, and shows a detail page with a table of lights plus a 3D view and controls to change light state.
```

with:

```markdown
**In plain terms:** The Models area lists models (a model = a set of 3D light positions). **Create from video** is the main way to add one and **Upload file** is the second. The detail page shows a table of lights plus a 3D view and controls to change light state.
```

Immediately after the Routes bullet in that same section, add:

```markdown
- **Add actions:** The list header has **Create from video** (`/models/new?tab=video`) and a quieter **Upload file** (`/models/new?tab=csv`). On a narrow screen they stack, video on top. The subtitle and the empty list lead with filming and mention a file with columns `id,x,y,z`.
```

In §4.17, replace the plain-terms sentence and the start of the Entry bullet through `Create from video (REQ-049).` so the opening reads:

```markdown
**In plain terms:** You add a model from videos of the lights, or from a coordinate file. From the Models list, **Create from video** is the main button and **Upload file** is the second. On `/models/new` the same two paths are tabs, video first. With no `tab` selected, the page opens on video. The video path uploads 2 or more videos, the server processes them into candidate 3D positions, you review the result (optionally in a read-only 3D preview), then confirm to create the model or cancel to discard.

- **Entry:** **Create from video** (REQ-049) is the first tab and the default. **Upload file** is the second tab and keeps the CSV form (REQ-006 / REQ-007, §4.6). The video path is a multi-file uploader requiring ≥ 2 videos
```

The words `The video path is a multi-file uploader requiring ≥ 2 videos` are the start of the existing Entry sentence that must stay. Leave everything after that phrase, including the marker, submit, review, and confirm bullets, unchanged. The section number stays `### 4.17`.

- [ ] **Step 5: Check the old entry wording is gone**

Run:

```bash
cd /workspaces/dlm && rg -n "alongside the CSV|by uploading a CSV file|you can upload one to create a model|Besides CSV upload" docs/requirements docs/userguide docs/design/frontend.md
```

Expected: no matches. `docs/design/frontend.md` still contains `### 4.6` and `### 4.17`. `web/app/devices/detail/DeviceDetailClient.tsx` still contains `Models → Create from video` (the source uses `&rarr;`).

- [ ] **Step 6: Mark the spec approved**

In `docs/superpowers/specs/2026-10-18-models-page-video-first-design.md`, the status line is `**Status:** Approved for planning`. If it still says `Pending review`, change it to that. Do not edit any other line of the spec.

- [ ] **Step 7: Commit**

```bash
git add docs/requirements/requirements.md docs/requirements/acceptance-criteria.md docs/userguide/using-the-app.md docs/userguide/build-model-from-video.md docs/design/frontend.md docs/superpowers/specs/2026-10-18-models-page-video-first-design.md
git commit -m "$(cat <<'EOF'
docs: describe the video-first models screen

The guides now name Create from video as the main action and Upload file as the second.
EOF
)"
```
