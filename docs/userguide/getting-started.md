# Getting started

**Domestic Light & Magic** (also called **dlm**) is a small web application for working with **3D
light models**, **scenes**, and **routines**. You use it through a normal website that runs on your
own computer; everything is stored locally in a simple database file (no cloud account required). The
interface works on a **phone, tablet, or desktop**.

The simplest way to run it is from an official release — **no Node.js or Go install required**.

## Run from a release

Official builds are published on **[GitHub Releases](https://github.com/mudged/dlm/releases)**. Pick
the file that matches your system:

| System | File to download |
|--------|------------------|
| **Raspberry Pi** (64-bit Raspberry Pi OS or other 64-bit Linux on ARM) | `dlm_linux_arm64.tar.gz` |
| **Linux** PC or server (64-bit x86) | `dlm_linux_amd64.tar.gz` |
| **Windows** 10 or 11 (64-bit) | `dlm_windows_amd64.exe` |

Each **Linux** release is a `.tar.gz` archive containing the server binary **and** a bundled OpenCV
runtime folder (`runtime/cv/`). Extract the archive and run the binary from the extracted directory
so it can find that folder. The **Windows** release is a single executable; camera capture from video
is not available on Windows yet.

### Linux and Raspberry Pi (quick start)

1. Download `dlm_linux_arm64.tar.gz` (Pi) or `dlm_linux_amd64.tar.gz` (PC).
2. Extract and run:

   ```bash
   mkdir -p ~/dlm && tar -xzf dlm_linux_arm64.tar.gz -C ~/dlm
   cd ~/dlm
   chmod +x dlm_linux_arm64
   ./dlm_linux_arm64
   ```

   (Use `dlm_linux_amd64` / `dlm_linux_amd64.tar.gz` instead if you are on a 64-bit PC.)

   Keep the **`runtime/cv/`** folder next to the binary. If you move the binary, move the whole
   directory tree together.

3. Open **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)** in your browser.

Data is stored under a `data` folder in the **current working directory** by default.

### Windows (quick start)

1. Download `dlm_windows_amd64.exe`.
2. Run it (double-click or from PowerShell / Command Prompt).
3. Open **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)**.

## Python (only for custom Python routines)

The app can run **automated Python routines** that change lights through a built-in editor. That
feature needs **Python 3** on the **same machine as the server**, with `python3` available on your
PATH. If Python is not installed, the rest of the app still works; only starting a Python routine will
fail until you install Python. You can point the server at a specific interpreter by setting
**`DLM_PYTHON3`**.

## Camera capture (no extra install)

**Building a model from video** uses a bundled OpenCV runtime that ships inside the Linux release
archive (`runtime/cv/`). You do **not** need to install Python or OpenCV on the server for that
feature — it is separate from the optional Python routines above. See
[`build-model-from-video.md`](build-model-from-video.md).

## Build from source

If you are developing the project or want to build it yourself, see the developer guide in
[`../engineering/build-and-run.md`](../engineering/build-and-run.md) (one command: `./scripts/run.sh`).
