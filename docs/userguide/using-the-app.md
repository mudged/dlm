# Using the app

Once the server is running, open **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)**. The interface
adapts to phone, tablet, and desktop.

## What you can do

- **Models** — Create or import light layouts (from a CSV file or from video) and view them in 3D.
- **Scenes** — Combine models and control how they are used together, arranged in a shared 3D space.
- **Routines** — Run automated sequences, including optional Python-based scripts (if you install
  Python on the machine that runs the app) and declarative shape animations.
- **Devices** — Configure physical lighting controllers (ESP32 / WLED first), assign a device to a
  model, and optionally discover devices on your network.
- **Options** — Adjust application settings, including a factory reset (with confirmation).

## Sample data

On a **fresh install**, the app adds a few **sample models** (a sphere, cube, and cone) and three
sample Python routines so you can explore immediately. If you delete every model and restart, those
samples come back.

## Importing a model from CSV

You can upload a CSV of light positions on the **Models** screen. The CSV header must be exactly
`id,x,y,z`, and light IDs must be **0-based sequential** integers (`0`, `1`, `2`, …).

## Building a model from video

Instead of typing coordinates, you can estimate light positions from short video recordings. See
[`build-model-from-video.md`](build-model-from-video.md).

## Live light updates for integrators

The server pushes light changes with **Server-Sent Events** so external tools do not need to poll
rapidly. For technical details (SSE URLs, batch update routes, and reconnect/resync behavior) see
[`../engineering/environment-and-api.md`](../engineering/environment-and-api.md).
