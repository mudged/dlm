# Run as a service on Raspberry Pi (starts on boot)

This uses **systemd**, which is standard on Raspberry Pi OS. Adjust paths and user if you prefer.

1. Extract the release archive into a folder for the app (example):

   ```bash
   sudo mkdir -p /opt/dlm
   sudo tar -xzf /path/to/dlm_linux_arm64.tar.gz -C /opt/dlm
   sudo chmod +x /opt/dlm/dlm_linux_arm64
   sudo mkdir -p /opt/dlm/data
   sudo chown -R pi:pi /opt/dlm
   ```

   The **`runtime/cv/`** folder must stay next to the binary (for example `/opt/dlm/runtime/cv/`).

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
   ExecStart=/opt/dlm/dlm_linux_arm64
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

4. Check status: `systemctl status dlm` — then open **[http://127.0.0.1:8080/](http://127.0.0.1:8080/)**
   from another device on the same network (use the Pi’s IP address instead of `127.0.0.1` if needed).

The server listens on **port 8080** by default. To use another port, set **`HTTP_LISTEN`** (for
example add `Environment=HTTP_LISTEN=:80` in the `[Service]` section).

## Updating after a new release

1. Download the new release for your platform from
   **[Releases](https://github.com/mudged/dlm/releases)**.
2. If you use **systemd**: `sudo systemctl stop dlm`
3. Replace the old files with the new ones — on Linux, extract the new `.tar.gz` over `/opt/dlm` (or
   replace the binary **and** the `runtime/cv/` folder together). Keep the same layout: binary and
   `runtime/cv/` as siblings.
4. If you use **systemd**: `sudo systemctl start dlm`

Keep your **`data`** directory (or whatever you set **`DLM_DATA_DIR`** / **`DLM_DB_PATH`** to) so
models and settings stay unless release notes say otherwise.
