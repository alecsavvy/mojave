# Mojave Design Direction

This document describes the intended aesthetic and product character for Mojave’s reference client and showcase frontends. The protocol is infrastructure; the default experience should feel like a place you *live* with your music, not a generic streaming dashboard.

## Aesthetic: vinyl listening room meets late-2000s desktop

**Chill, wood, bohemian vinyl listening room.** Warm and low-key. Think wood (walnut, teak, light oak), soft fabrics, plants, warm lighting. Not sterile or “tech gray.” The app should feel like a room where you sit and listen — a place you own your collection.

**Callback to late-2000s desktop apps.** Brushed metal, subtle gradients, soft shadows, a bit of skeuomorphism. An echo of iTunes 7–9, Winamp, early Spotify — “this is a thing you use on a desk to play your music,” not a flat SaaS panel. Can be literal (metal textures, bevelled controls) or a modern take (same warmth and tactility, less literal chrome).

**Character.** Typography with personality: a serif or distinctive sans for headings; readable body. Muted, warm palette — creams, warm grays, wood tones, one or two accent colors. Enough whitespace so it feels calm. Subtle grain or texture so it doesn’t look like every other app.

## Why this fits the product

“Your library” and “you own it” align with “a place you live with your music” — room, not spreadsheet. The metallic / desktop nod says “this is for people who care about playback and ownership,” not another streaming skin. The reference client (Tauri app) and main mojave.audio frontend should embody this; other showcase frontends (Spotify-style, Bandcamp-style, etc.) can riff on it or contrast it.

## Showcase frontends

The main site and subdomains can implement this direction to varying degrees:

| Frontend | Aesthetic note |
|----------|----------------|
| **mojave.audio** | Default: vinyl room + metallic desktop (Horizon theme can implement this). |
| **itunes.mojave.audio** | iTunes-style library + store; strong fit for metallic / ownership vibe. |
| **bandcamp.mojave.audio** | Artist-first; can lean warm and tactile. |
| **spotify.mojave.audio** | Browse/playlists; can be brighter but still warm. |
| **audius.mojave.audio** | Community/social; character over generic. |
| **myspace.mojave.audio** | Early social music; personality and custom pages. |

Same protocol; different UX paradigms. The default Mojave character is warm, owned, and a bit nostalgic — not cold or corporate.
