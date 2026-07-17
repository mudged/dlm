# Build a model from video

You can estimate light positions automatically from short video recordings instead of typing
coordinates into a CSV.

1. **Set up your device** — On the **Devices** screen, add your light controller and set its **light
   count** to match the physical string.
2. **Record from two or more angles** — Point a phone or camera at the lights from different
   viewpoints. You need at least two recordings; more angles usually give better accuracy.
3. **Run the capture light sequence** — On the same device’s detail page, press **Start capture**. The
   app lights each bulb in turn for about one second, then turns them all off. Begin recording
   **before** you start the sequence, and keep the camera steady while each light is on.
4. **Upload and review** — On the **Models** screen, choose **Create from video**. Upload your
   recordings (MP4, MOV, MKV, or WebM). When processing finishes, review the detected light count, any
   missing or uncertain lights, and an optional 3D preview. Enter a name and **Confirm** to save the
   model, or **Cancel** to discard the result.
5. **Optional printable marker** — If your setup includes a known-size fiducial marker in the frame,
   you can download a printable marker from the upload screen to help with scale. This is optional and
   never required to start processing.

The OpenCV runtime used for reconstruction is bundled with the Linux release — no separate install on
the server. Camera capture from video is not available on Windows yet.
