# How we know each part works

This page is the "does it actually work?" checklist for Domestic Light & Magic. For each feature it
gives you something to **try** and what you **should see** if the app is behaving. If you can walk
through these and everything matches, the app is doing its job.

It's written to pair with [`requirements.md`](requirements.md) (the plain-English tour of what the app
does). No coding knowledge needed — just follow the steps.

> The `REQ-NNN` codes from the tour aren't repeated here. If you want to connect a check back to a
> feature code, use the lookup table at the bottom of [`requirements.md`](requirements.md).

---

## Models

**Uploading a good CSV creates a model.**
- Try: upload a file whose first line is exactly `id,x,y,z`, followed by lights numbered `0, 1, 2, …`
  with three numbers each.
- You should see: a new model appears in your list, showing every light you uploaded.

**A broken CSV is rejected — and nothing is half-saved.**
- Try: upload a file with the wrong heading (like `idx,x,y,z`), or with gaps in the numbering (`0` then
  `2`), or with a word where a number should be, or with more than 1000 lights.
- You should see: a clear message explaining what's wrong. No partial model is created.

**You can browse, open, and delete models.**
- Try: view the list, open a model to see its lights and details, then delete one you don't need.
- You should see: the list of all models; a detail view for the one you open; and the deleted model
  disappears.

**A model in use can't be silently deleted.**
- Try: delete a model that's part of a scene.
- You should see: the app refuses and tells you it's in use, and what to do about it (remove it from the
  scene first).

**A fresh install comes with three examples.**
- Try: start the app with nothing saved yet.
- You should see: three ready-made models named for a **sphere**, a **cube**, and a **cone**, each about
  2 m across with lights spread over the whole surface (not just the edges).

---

## Seeing your lights in 3D

**Opening a model shows a 3D picture.**
- Try: open any model.
- You should see: every light drawn as a small white ball, with faint grey lines linking each light to
  the next, on a dark-grey background.

**Pointing at a light reveals its details.**
- Try: hover a light with the mouse, or tap it on a touchscreen.
- You should see: that light's number and its x/y/z position.

**The picture matches each light's state.**
- Try: turn some lights on with a colour and brightness; leave others off.
- You should see: "on" lights glow in their colour (brighter ones glow more strongly), and "off" lights
  fade to the faint grey of the wires. Changing a light updates the picture without reloading the page.

**Reset camera puts the view back.**
- Try: spin and zoom the view, then press **Reset camera**.
- You should see: the camera returns to its starting position. Your saved lights are unchanged.

**Big models stay manageable.**
- Try: open a model with many lights.
- You should see: the light list split into pages, a way to change how many show per page, and a box to
  jump to a light by number (with a friendly error for a number that doesn't exist).

**Multi-select changes many lights at once.**
- Try: tick several lights and apply one colour/brightness/on-off setting.
- You should see: all the selected lights change together, and the picture and list update promptly.

**Reset returns lights to the starting state.**
- Try: press the model's **Reset** button.
- You should see: every light in that model goes off, white, at 100% brightness.

---

## Controlling the lights

**Each light can be controlled individually.**
- Try: turn a single light on, give it a hex colour like `#FF8800`, and set brightness to 50%.
- You should see: that light updates to match, everywhere it's shown.

**Bad values are refused.**
- Try: send an invalid colour or a brightness outside 0–100.
- You should see: a clear error, and the light is left as it was.

**Many changes stay smooth.**
- Try: run something that changes lots of lights quickly.
- You should see: updates arrive without the view stuttering, and lights that didn't actually change
  aren't needlessly redrawn.

---

## Scenes

**Creating a scene needs a name and at least one model.**
- Try: create a scene, give it a name, and add a model.
- You should see: the scene is created with the model placed automatically — you don't type any
  position numbers.

**Adding models drops them in beside each other.**
- Try: add a second model to a scene.
- You should see: it appears to the right of the first, without overlapping. The models' own saved
  coordinates never change.

**Removing the last model asks first.**
- Try: remove the only remaining model from a scene.
- You should see: a warning that this deletes the whole scene, and a chance to confirm or cancel.

**The scene 3D view shows everything.**
- Try: open a scene with several models.
- You should see: all the lights of all the models, each model's wires linking only its own lights, with
  a faint outline box around the whole thing.

---

## Routines (light shows)

**You can only pick the two kinds.**
- Try: create a new routine.
- You should see: a choice between a **Python** routine and a **shape animation** routine — and nothing
  else.

**Starting and stopping a routine works from one place.**
- Try: on a routine's page, choose a scene, press **Start**, watch the 3D view, then press **Stop**.
- You should see: the same chosen scene both running and shown live; the lights change while it runs;
  and it stops within about **two seconds** of pressing Stop.

**A scene runs one show at a time.**
- Try: start a routine on a scene that already has one running.
- You should see: the app refuses and tells you the scene is busy.

**A new Python routine isn't blank.**
- Try: create a new Python routine.
- You should see: starter code already in the editor (it changes the colour of lights inside a sphere),
  ready to run.

**The Python editor helps you.**
- Try: type some Python.
- You should see: colour-coded text, mistakes flagged, word suggestions, and a reference guide below the
  editor whose examples you can insert into your code with a button.

**The three sample routines are there.**
- Try: look at your routine list on a fresh install (or after a factory reset).
- You should see: **growing sphere**, **sweeping cuboid**, and **random colour cycle** — real routines
  you can open and edit.

**Shape animations move shapes around.**
- Try: build a shape animation with one or two moving shapes and run it.
- You should see: lights lighting up in each shape's colour as it moves, with the rest as background or
  off, and edge behaviour (wrap, stop, or bounce) as you chose.

**Coming back to a running scene catches up.**
- Try: start a routine, navigate away, then return to the scene.
- You should see: the app correctly shows it's still running and displays the current light colours (not
  a frozen or "stopped" view).

---

## Real lights (hardware)

**You can add and name a device.**
- Try: open the **Devices** area and add a device with its connection details and a name.
- You should see: the device listed. A clearly-invalid address is refused with a helpful message.

**One device links to one model.**
- Try: assign a device to a model, then try to assign the same device to a second model.
- You should see: the app keeps it one-to-one and won't silently steal an existing link.

**Removing a device tidies up.**
- Try: delete a device that's linked to a model.
- You should see: the link is removed at the same time — no leftover, dangling assignment.

**Real lights follow the show.**
- Try: run a routine on a scene whose model has a connected device.
- You should see: the physical lights change in step with the on-screen view.

---

## The look and feel

**Light and dark themes both work.**
- Try: switch between light and dark mode.
- You should see: light mode is white with dark text; dark mode is dark grey with light text; the 3D
  view stays dark grey in both. Your choice is remembered next time.

**The menu opens and closes.**
- Try: tap the burger (three-lines) button, including on a touchscreen.
- You should see: the left-hand menu fold away and come back.

**Branding is correct.**
- Try: look at the app's title and logo.
- You should see: the name **Domestic Light & Magic** and a lightbulb icon; buttons carry icons too.

**Factory reset needs confirmation.**
- Try: use **Factory reset** in Options.
- You should see: a clear warning that everything will be wiped and it can't be undone. If you confirm,
  everything is cleared and the three sample models and three sample Python routines come back. If you
  cancel, nothing changes.

---

## Building a model from video

**The capture sweep blinks lights in order.**
- Try: start the capture sweep from the Devices screen.
- You should see: one light on at a time, in order (0, 1, 2, …), about a second each. Stopping (or
  finishing) turns them all off within about two seconds.

**You can build a model from uploaded videos.**
- Try: on the model screen, choose "create from video" and upload two or more clips of the same sweep
  from different angles.
- You should see: the app processing the videos, then a review showing how many lights it found (and any
  it wasn't sure about) before you confirm and save. A printable marker is offered but isn't required.

---

## Getting it and running it

**There's a download for each platform.**
- Try: look at the project's releases.
- You should see: one download each for **Windows**, **Linux**, and **Linux on ARM** (Raspberry Pi) —
  Windows may be a bare program file; Linux may be a `.tar.gz` with the program plus a `runtime/cv/`
  folder that stays next to it.

**Changes are built and tested automatically.**
- Try: propose a change to the project.
- You should see: automated build-and-test checks run, and they must pass before the change is accepted.

**Running it needs almost nothing installed.**
- Try: download the right asset and run it on a fresh machine (on Linux: unpack the archive and keep
  `runtime/cv/` beside the program).
- You should see: the app starts and the web page opens, with no extra installs — except that running
  your own **Python** routines needs Python on the machine.

**Setup is explained in plain words.**
- Try: read the main README, then follow its link to the user guide.
- You should see: a short README landing page, and in the user guide how to download the right file,
  run it, set it up on a Raspberry Pi so it starts on boot, and update it later — no jargon or
  internal codes.
