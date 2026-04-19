# Domestic Light & Magic

**Domestic Light & Magic** (also called **dlm**) is a small web application for working with **3D light models**, **scenes**, and **routines**—for example, arranging many lights on a shape and seeing them in the browser. You use it through a normal website that runs on your own computer; everything is stored locally in a simple database file (no cloud account required).

The interface is built to work on a **phone, tablet, or desktop** so you can use whatever device you have handy.

## What you can do

- **Models** — Create or import light layouts (for example from a CSV file) and view them in 3D.
- **Scenes** — Combine models and control how they are used together.
- **Routines** — Run automated sequences (including optional Python-based scripts if you install Python on the machine that runs the app).
- **Devices and options** — Configure how the app talks to your setup.

On a **fresh install**, the app adds a few **sample models** (for example a sphere, cube, and cone) so you can explore immediately. If you delete every model and restart, those samples come back.

## What you need installed

You will need two free tools on your computer:

1. **Go** — the language the server is written in ([https://go.dev/dl/](https://go.dev/dl/)).
2. **Node.js** — used only while building the web interface; it is not required to *run* a finished build on something like a Raspberry Pi the same way ([https://nodejs.org/](https://nodejs.org/)).

If you are on **Windows**, use **Git Bash**, **WSL**, or similar so you can run the same commands as on Mac or Linux.

## How to install and run (simplest path)

1. **Download** this repository (clone or ZIP) and open a terminal in the **top folder** of the project (the one that contains `scripts/` and `README.md`).

2. **Run one command:**

   ```bash
   ./scripts/run.sh
   ```

   The first time, this may take a while because it downloads web dependencies and builds the interface. Later runs are usually faster.

3. **Open your browser** to:

   **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)**

You should see the Domestic Light & Magic home page. Use the menu to open **Models**, **Scenes**, **Routines**, and **Options**.

### If something feels slow

The slow part is usually building the web interface (`npm` and `next build`). That is normal on the first run. Repeat runs reuse caches and skip reinstalling dependencies when possible.

---

## More for developers and advanced users

Scripts, optional environment variables, developing with two terminals (live-reload UI + API), building for a Raspberry Pi, API performance tips, and troubleshooting live updates are in **[docs/user/advanced-setup.md](docs/user/advanced-setup.md)**.

## License

See [LICENSE](LICENSE).
