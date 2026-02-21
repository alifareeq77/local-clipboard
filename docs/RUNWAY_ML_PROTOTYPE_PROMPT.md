# Runway ML: Developer-Focused Visual Prototype Prompts

Use these prompts in **Runway** (video or image) to generate **visual prototypes** for an app or video idea. The outputs give developers and creators clear reference for flow, shots, and implementation.

---

## 1. Video idea (Runway Gen-3 / video) — primary

Use when the **deliverable is a short video**: concept demo, explainer, product teaser, or narrative clip. Tuned for Runway’s video model (clear actions, simple scenes, consistent style).

### Video idea template

```
Create a [5 / 8 / 10] second video for Runway.

**Video idea (one sentence):** [e.g. "A developer opens a clipboard sync app on their phone, pastes a link, taps Send, and sees it appear in the history with a copy button."]

**Story beats (in order):**
1. [Opening shot: what we see first, e.g. "Phone in hand, dark app screen with header and empty form."]
2. [Action: what happens, e.g. "Finger taps text area; cursor appears; short typed or pasted URL."]
3. [Action: e.g. "Finger taps Send button."]
4. [Result: e.g. "Latest section updates; new card slides in at top with Copy button."]
5. [Optional end: e.g. "Brief hold on updated screen."]

**Visual spec for developers:**
- Shot type: [single continuous take / 2–3 cuts between beats].
- Camera: [static / slight push-in / phone POV / screen-recording style].
- Frame: [phone in frame / full-screen UI only / device mockup].
- Style: [minimal UI, dark theme, teal accent / clean product shot / no branding].
- Motion: [subtle only / clear tap and transition / no flashy effects].
- Aspect ratio: [16:9 / 9:16 vertical for phone / 1:1].

**Output use:** This video will be used as [reference for implementation / pitch / storyboard for real build]. Keep [UI elements / actions / timing] clear and readable.
```

### Example (clipboard app — video idea)

```
Create a 8 second video for Runway.

**Video idea (one sentence):** A developer opens a clipboard sync app on their phone, pastes a link, taps Send, and sees it appear in the history with a copy button.

**Story beats (in order):**
1. Phone in hand or on surface; screen shows minimal dark app: "Local Clipboard" header, "Paste or type to send" text area, "Send" button, "Latest" section empty or with one line.
2. Finger taps the text area; cursor appears; a short URL or code snippet appears (typed or pasted).
3. Finger taps "Send" button.
4. "Latest" section updates to show that text; a new card appears at top of history list with "Copy" and "Pin"; smooth, minimal transition.
5. Hold 1 second on the updated screen.

**Visual spec for developers:**
- Shot type: single continuous take, or 2 cuts (before send / after send).
- Camera: static, focus on screen; or slight push-in on the phone.
- Frame: phone screen fills most of frame, or phone in hand so we see it’s mobile.
- Style: minimal UI, dark theme, teal accent, readable sans-serif, no logos.
- Motion: clear tap and content change only; no fancy animations or particles.
- Aspect ratio: 9:16 (vertical) for phone, or 16:9 if showing phone in environment.

**Output use:** Reference for implementing the real app flow; developer should be able to infer: form → POST → updated latest + new history card.
```

### Runway video tips (developers)

- **One clear action per beat:** Runway works best with simple, explicit actions (tap, type, scroll, cut).
- **Short duration:** 5–10 s is reliable; describe the exact beats so the model doesn’t drift.
- **Name the UI:** Use concrete labels ("Send button", "Latest section", "history card") so the video doubles as a spec.
- **Consistent style:** Repeat "dark theme, teal accent, minimal" (or your stack) so all generations match.
- **Storyboard first:** Write the beats in a doc, then paste into Runway; use the same beats for image-to-video if you have a key frame.

### Non-UI video idea (narrative / explainer / ad)

When the video is **not** an app screen (e.g. short narrative, explainer, or ad), use the same structure with story beats and a clear visual spec:

```
Create a [5 / 8 / 10] second video for Runway.

**Video idea (one sentence):** [e.g. "A developer at a desk copies code on their laptop; their phone buzzes; they open an app and the same snippet is there, ready to paste."]

**Story beats (in order):**
1. [Establishing shot.]
2. [Action / cause.]
3. [Reaction / result.]
4. [Optional payoff or hold.]

**Visual spec:**
- Shot type: [continuous / 2–3 cuts].
- Camera: [static / slow push / POV].
- Style: [realistic / minimal / product-style / mood].
- Aspect ratio: [16:9 / 9:16 / 1:1].

**Output use:** [Pitch / storyboard / reference for edit.] Keep [actions / timing] clear.
```

---

## 2. Image / UI mockup (Runway Image or similar)

Use for **single-screen or multi-screen UI mockups** that a developer can implement from.

### Template

```
Generate a [wireframe / high-fidelity UI mockup] for a [platform: web app / mobile app / desktop] screen.

App concept: [One sentence, e.g. "Local network clipboard sync between laptop and phone."]

This screen: [e.g. "Main dashboard with a send-clipboard form, latest clipboard preview, and a scrollable history list with copy and pin buttons."]

Developer-friendly spec:
- Clear visual hierarchy: [header / main content / sidebar / footer].
- Label or show: [primary CTA], [form fields], [list/cards], [navigation if any].
- [Dark theme / light theme], [single accent color, e.g. teal].
- [Desktop only / mobile-only / responsive layout].
- No Lorem ipsum: use short placeholder text that suggests real content (e.g. "Paste or type…", "Latest: https://…", "Copy").
- Style: [minimal, clean, dev-tool aesthetic / product UI / etc.].
```

### Example (local clipboard app)

```
Generate a high-fidelity UI mockup for a web app screen, mobile-first.

App concept: Local network clipboard sync between laptop and phone.

This screen: Main view with (1) header showing "Local Clipboard" and "Open from phone: http://192.168.1.5:8080", (2) a text area "Paste or type to send" with a "Send" button, (3) a "Latest" card showing the most recent clipboard text, (4) a searchable history list of cards with "Copy" and "Pin" on each, pinned items at top.

Developer-friendly spec:
- Clear hierarchy: header, then send form, then latest, then history.
- Show: primary button (Send), input area, copy/pin controls on each history card.
- Dark theme, teal accent, readable sans-serif.
- Mobile viewport (narrow), touch-friendly buttons.
- Placeholder text like "Paste or type…", "Latest: (empty or sample URL)", "Copy", "Pin".
- Style: minimal, clean, dev-tool aesthetic.
```

---

## 3. Multi-screen set (use Image multiple times)

For **several key screens** that together define the app for a developer:

| Screen        | Runway prompt focus |
|---------------|----------------------|
| **Login / home** | Layout, nav, primary CTA. |
| **Main feature** | Form, list, or main content area; labels for buttons/inputs. |
| **Detail / secondary** | One item expanded or one extra view; back/side navigation. |
| **Empty / error** | Empty state or error message so dev knows what to build. |

Use the **same app concept and style** in each prompt (e.g. same theme, accent, and "developer-friendly spec" line) so the set feels like one app.

---

## Tips for developers

- **One concept per prompt**: Runway works best with one clear scene or screen; split "dashboard + settings + history" into separate prompts.
- **Name components**: Use words a dev would use ("Send button", "history card", "latest clipboard block") so the image doubles as a light spec.
- **Reference your stack**: Add "Vue/React-style SPA" or "mobile web" so layout and density match what you’ll build.
- **Export and annotate**: After generating, add sticky notes or arrows in Figma/paper with component names and API hooks (e.g. "GET /api/clipboard", "POST /api/clipboard") for implementation.
