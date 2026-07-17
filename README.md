# Domestic Light & Magic

**Domestic Light & Magic** (also called **dlm**) is a small web application for working with **3D light models**, **scenes**, and **routines**—for example, arranging many lights on a shape and seeing them in the browser. You use it through a normal website that runs on your own computer; everything is stored locally in a simple database file (no cloud account required).

The interface is built to work on a **phone, tablet, or desktop** so you can use whatever device you have handy.

## Just want to use it?

You don't need any programming tools to run a released version—download it, start it, and open it in your browser.

**→ Start with the [user guide](docs/userguide/).** It walks you through everything in plain language:

- **[Getting started](docs/userguide/getting-started.md)** — download the right file for your Raspberry Pi, Linux PC, or Windows PC and run it.
- **[Using the app](docs/userguide/using-the-app.md)** — models, scenes, routines, devices, options, and the sample data.
- **[Build a model from video](docs/userguide/build-model-from-video.md)** — let the app work out where your lights are by filming them.
- **[Run it as a service](docs/userguide/running-as-a-service.md)** — start automatically on boot on a Raspberry Pi, and update to a new release.

Quick reference: releases live on **[GitHub Releases](https://github.com/mudged/dlm/releases)**, and once running the app opens at **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)**.

---

## Build from source (developers)

Use this path when you are working on the project or building it yourself. If you plan to contribute,
start with **[CONTRIBUTING.md](CONTRIBUTING.md)** for one-time setup (including the Superpowers plugin).

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

## More for developers and advanced users

Scripts, optional environment variables, developing with two terminals (live-reload UI + API), continuous integration, cutting releases, API performance tips, and troubleshooting live light updates (Server-Sent Events) are in **[docs/engineering/](docs/engineering/)** — start with [build and run](docs/engineering/build-and-run.md), [environment and API](docs/engineering/environment-and-api.md), and [CI and release](docs/engineering/ci-and-release.md). The design lives in **[docs/design/architecture.md](docs/design/architecture.md)**.

## License

See [LICENSE](LICENSE).
