# Models page leads with video — design

**Date:** 2026-10-18
**Status:** Pending review
**Scope:** The Models screen lists models and offers two ways to add one. **Create from video** is the main action and opens the video tab. **Upload file** is the quieter second action and opens the file tab. The new-model page uses the same order and opens on video when no tab is selected. No new `REQ-*`. The CSV file rules, the video form, and the way a listed model opens stay as they are.

## Context

`/models` is a list of models. Each row shows the name, light count, and created time, with **View** and **Delete**. The name and **View** already open the model. The only add action is a button labelled **Upload CSV**, and the subtitle and empty list talk only about a spreadsheet. That button opens `/models/new`, which defaults to the **CSV upload** tab. **Create from video** is the second tab.

Filming is the usual way to add a model. The user guide already tells people to choose **Create from video** on the Models screen. The screen should say that.

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Page job | `/models` lists the models you have. Adding one is two header actions |
| Main action | **Create from video** goes to `/models/new?tab=video` |
| Second action | **Upload file** goes to `/models/new?tab=csv`. It is visually quieter than the main button |
| Narrow screens | The two buttons stack full width, video on top |
| List rows | Unchanged. Name link, **View**, and **Delete** stay. A later edit-from-the-list pass is out of scope |
| New-model tabs | **Create from video** first, **Upload file** second. The file tab's internal name stays `csv` |
| Default tab | A visit with no `tab`, an empty `tab`, or any value other than `csv` opens **Create from video**. Only `tab=csv` opens the file form |
| File form | Unchanged. Label **CSV file**, header `id,x,y,z`, error `Choose a CSV file.`, submit **Upload** |
| Video form | Unchanged |
| Query names | `tab=video` and `tab=csv` stay, including while a capture `jobId` is in the URL |
| Requirements | No new `REQ-*`. This is wording and entry order for the list (REQ-002) and the two create paths (REQ-006, REQ-049) |

## Models screen

Title stays **Light models**.

Subtitle:

> Film your lights and the app works out where they are. Or upload a file with columns `id,x,y,z`. Up to 1000 lights per model.

**Create from video** uses the primary button style and the Font Awesome `faVideo` icon. **Upload file** uses the quieter slate style already used by **View** (`bg-slate-600`), with the existing `faUpload` icon.

Empty list:

> No models yet. Create from video, or upload a file.

**Create from video** in that sentence links to `/models/new?tab=video`. **upload a file** links to `/models/new?tab=csv`. When models exist, the header buttons are still both shown.

## New-model screen

Title stays **New model**.

Intro:

> Film your lights and the app works out where they are. Or upload a file of light coordinates.

The tab list is labelled `Model creation method`, as it is today. The first tab is **Create from video**. The second is **Upload file**. Selecting a tab still uses `buildNewModelUrl`, so an in-progress `jobId` is kept.

`web/lib/newModelRoute.ts` gains `activeNewModelTab(rawTab: string | null)`, which returns `"csv"` only when `rawTab === "csv"` and `"video"` otherwise. `NewModelClient` uses that helper instead of treating every non-video value as CSV.

## Docs

**`docs/requirements/requirements.md`**, under "What you can do with models", replace the add bullet with:

> **Add a new model** by filming your lights (**Create from video**), or by uploading a file.

In section 11, replace step 3's first sentence (`On the model screen there's a "create from video" option alongside the CSV upload.`) with:

> On the Models screen, **Create from video** is the main button. **Upload file** is the second button, for a spreadsheet of coordinates.

**`docs/requirements/acceptance-criteria.md`**, under "Building a model from video", add this try/see before the existing "You can build a model from uploaded videos" item:

> **The Models screen leads with filming.**
> - Try: open the Models screen, including when it has models and when the list is empty.
> - You should see: **Create from video** as the main button and a quieter **Upload file** button. Create from video opens the new-model page on the video tab. Upload file opens it on the file tab. The page text mentions filming first, then a file whose columns are `id,x,y,z`. With no models, the empty sentence offers the same two links.

Leave the existing "choose create from video" try/see as it is.

**`docs/userguide/using-the-app.md`**, in "Uploading a file (CSV)", replace `On the Models screen you can upload one to create a model.` with:

> On the Models screen, **Upload file** is the second button, next to **Create from video**.

**`docs/userguide/build-model-from-video.md`**, after the sentence that begins `On the Models screen, choose Create from video`, add:

> **Upload file** is the other button on that screen, for when you already have a spreadsheet of positions.

**`docs/design/frontend.md`** §4.6, replace the plain-terms sentence with:

> The Models area lists models (a model = a set of 3D light positions). **Create from video** is the main way to add one and **Upload file** is the second. The detail page shows a table of lights plus a 3D view and controls to change light state.

In that same section, after the Routes bullet, add:

> **Add actions:** The list header has **Create from video** (`/models/new?tab=video`) and a quieter **Upload file** (`/models/new?tab=csv`). On a narrow screen they stack, video on top. The subtitle and the empty list lead with filming and mention a file with columns `id,x,y,z`.

§4.17, replace the plain-terms sentence and the **Entry** bullet's opening through `Create from video (REQ-049).` with:

> You add a model from videos of the lights, or from a coordinate file. From the Models list, **Create from video** is the main button and **Upload file** is the second. On `/models/new` the same two paths are tabs, video first. With no `tab` selected, the page opens on video. The video path uploads 2 or more videos, the server processes them into candidate 3D positions, you review the result (optionally in a read-only 3D preview), then confirm to create the model or cancel to discard.
>
> - **Entry:** **Create from video** (REQ-049) is the first tab and the default. **Upload file** is the second tab and keeps the CSV form (REQ-006 / REQ-007, §4.6).

Leave the rest of the Entry bullet, from "The video path is a multi-file uploader", unchanged.

Do not edit the API notes that describe `POST /api/v1/models` as a CSV upload. Do not renumber a `§` section. The device page sentence "upload the recorded videos later via Models → Create from video" already matches and stays.

## Tests

In `web/lib/newModelRoute.test.ts`, cover `activeNewModelTab`:

- `null`, `""`, `"video"`, and any other value return `"video"`.
- `"csv"` returns `"csv"`.

`buildNewModelUrl` expectations stay as they are. The Models buttons and the empty-list links call that helper, so their addresses are the ones already tested.

There is no page-render test runner in `web/`. Do not add one for these labels.

## Out of scope

- A new way to edit a model from the list.
- The CSV checks, the video uploader, review, and confirm.
- Renaming `tab=csv` or the file field name `file`.
