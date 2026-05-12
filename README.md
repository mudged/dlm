# Domestic Light & Magic

**Domestic Light & Magic** (also called **dlm**) is a small web application for working with **3D light models**, **scenes**, and **routines**—for example, arranging many lights on a shape and seeing them in the browser. You use it through a normal website that runs on your own computer; everything is stored locally in a simple database file (no cloud account required).

The interface is built to work on a **phone, tablet, or desktop** so you can use whatever device you have handy.

## Run from a release (simplest for Raspberry Pi or a server)

Official builds are published on **[GitHub Releases](https://github.com/mudged/dlm/releases)**. Each release includes one executable per platform—**no Node.js or Go install required** on the machine where you run the app.

Pick the file that matches your system:

| System | File to download |
|--------|------------------|
| **Raspberry Pi** (64-bit Raspberry Pi OS or other 64-bit Linux on ARM) | `dlm_linux_arm64` |
| **Linux** PC or server (64-bit x86) | `dlm_linux_amd64` |
| **Windows** 10 or 11 (64-bit) | `dlm_windows_amd64.exe` |

### Linux and Raspberry Pi (quick start)

1. Download `dlm_linux_arm64` (Pi) or `dlm_linux_amd64` (PC).
2. In a terminal, go to the folder containing the file and run:

   ```bash
   chmod +x dlm_linux_arm64
   ./dlm_linux_arm64
   ```

   (Use `dlm_linux_amd64` instead if you are on a 64-bit PC.)

3. Open **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)** in your browser.

Data is stored under a `data` folder in the **current working directory** by default (see environment variables in [docs/advanced-setup.md](docs/advanced-setup.md)).

### Windows (quick start)

1. Download `dlm_windows_amd64.exe`.
2. Run it (double-click or from PowerShell / Command Prompt).
3. Open **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)**.

### Python (only for Python routines)

The app can run **automated Python routines** that change lights through a built-in editor. That feature needs **Python 3** on the **same machine as the server**, with `python3` available on your PATH. If Python is not installed, the rest of the app still works; only starting a Python routine will fail until you install Python.

You can point the server at a specific interpreter by setting **`DLM_PYTHON3`** (see [docs/advanced-setup.md](docs/advanced-setup.md)).

### Run as a service on Raspberry Pi (starts on boot)

This uses **systemd**, which is standard on Raspberry Pi OS. Adjust paths and user if you prefer.

1. Create a folder for the app and data (example):

   ```bash
   sudo mkdir -p /opt/dlm/data
   sudo cp /path/to/your/dlm_linux_arm64 /opt/dlm/dlm
   sudo chmod +x /opt/dlm/dlm
   sudo chown -R pi:pi /opt/dlm
   ```

2. Create **`/etc/systemd/system/dlm.service`** (use `nano` or another editor **with sudo**):

   ```ini
   [Unit]
   Description=Domestic Light & Magic
   After=network.target

   [Service]
   Type=simple
   User=pi
   WorkingDirectory=/opt/dlm
   Environment=DLM_DATA_DIR=/opt/dlm/data
   ExecStart=/opt/dlm/dlm
   Restart=on-failure

   [Install]
   WantedBy=multi-user.target
   ```

3. Enable and start:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable dlm
   sudo systemctl start dlm
   ```

4. Check status: `systemctl status dlm` — then open **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)** from another device on the same network (use the Pi’s IP address instead of `127.0.0.1` if needed).

The server listens on **port 8080** by default. To use another port, set **`HTTP_LISTEN`** (for example `HTTP_LISTEN=:80` in the `[Service]` section as `Environment=HTTP_LISTEN=:80`).

### Updating after a new release

1. Download the new binary for your platform from **[Releases](https://github.com/mudged/dlm/releases)**.
2. If you use **systemd**: `sudo systemctl stop dlm`
3. Replace the old binary with the new one (same path as before, e.g. `/opt/dlm/dlm`).
4. If you use **systemd**: `sudo systemctl start dlm`

Keep your **`data`** directory (or whatever you set **`DLM_DATA_DIR`** / **`DLM_DB_PATH`** to) so models and settings stay unless release notes say otherwise.

---

## Build from source (developers)

Use this path when you are working on the project or building yourself.

### What you need installed

1. **Go ≥ 1.25** — the language the server is written in ([https://go.dev/dl/](https://go.dev/dl/)). `backend/go.mod` declares `go 1.25.0`.
2. **Node.js** — Active LTS is fine ([https://nodejs.org/](https://nodejs.org/)). It is only used **while building** the web interface.

If you are on **Windows**, use **Git Bash**, **WSL**, or similar so you can run the same commands as on Mac or Linux.

### One command from a clone

1. Clone this repository and open a terminal in the **top folder** (the one that contains `scripts/` and `README.md`).

2. Run:

   ```bash
   ./scripts/run.sh
   ```

   The first time, this may take a while because it downloads web dependencies and builds the interface. Later runs are usually faster.

3. Open **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)**

### If something feels slow

The slow part is usually building the web interface (`npm` and `next build`). That is normal on the first run. Repeat runs reuse caches and skip reinstalling dependencies when possible.

---

## What you can do

- **Models** — Create or import light layouts (for example from a CSV file) and view them in 3D.
- **Scenes** — Combine models and control how they are used together.
- **Routines** — Run automated sequences (including optional Python-based scripts if you install Python on the machine that runs the app).
- **Devices and options** — Configure how the app talks to your setup.

On a **fresh install**, the app adds a few **sample models** (for example a sphere, cube, and cone) so you can explore immediately. If you delete every model and restart, those samples come back.

## Live light updates for integrators

The server pushes light changes with **Server-Sent Events** so clients do not need to poll rapidly. For batch HTTP updates when changing many lights, prefer the batch routes documented in **[docs/architecture.md](docs/architecture.md)** instead of one request per light—this keeps performance reasonable on small hardware when many lights update often.

- **Model stream:** `GET /api/v1/models/{id}/lights/events` (`text/event-stream`) sends JSON `data:` lines shaped like `{ "seq": <uint64>, "deltas": [ { "light_id", "on"?, "color"?, "brightness_pct"? }, ... ] }`; `model_id` is implicit from the URL.
- **Scene stream:** `GET /api/v1/scenes/{id}/lights/events` sends `{ "seq": <uint64>, "deltas": [ { "model_id", "light_id", "on"?, "color"?, "brightness_pct"? }, ... ] }`.
- Clients merge `deltas[]` incrementally per **REQ-041** in **[docs/requirements.md](docs/requirements.md)** and **[docs/architecture.md](docs/architecture.md)** §3.18; resync with `GET .../lights/state`, `GET /api/v1/models/{id}`, or `GET /api/v1/scenes/{id}` only on a sequence gap or `EventSource.onerror`.

## More for developers and advanced users

Scripts, optional environment variables, developing with two terminals (live-reload UI + API), continuous integration, cutting releases, API performance tips, and troubleshooting live updates are in **[docs/advanced-setup.md](docs/advanced-setup.md)**.

## License

See [LICENSE](LICENSE).

<!-- req-029: throughput guidance references architecture and SSE endpoints above. -->
