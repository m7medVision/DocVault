# mobile/AGENTS.md — Expo app

Part of the docvault monorepo. Read root **[../AGENTS.md](../AGENTS.md)** first — the
**commit-identity rule (never attribute commits to Claude/AI)** and conventions apply here.

## ⚠️ Expo HAS CHANGED

Read the exact versioned docs at **https://docs.expo.dev/versions/v55.0.0/** before writing
any code. APIs differ across SDKs — do not rely on memory or older examples.

## Stack

**Expo SDK 55** · **expo-router** (file-based routing) · React Native · `@react-navigation`
(bottom tabs) · **TanStack Query** · `expo-camera` (scan) · `expo-secure-store` (auth tokens)
· `expo-localization` (bilingual AR/EN) · `expo-document-picker`/`expo-file-system` (upload)
· `react-native-reanimated`/`gesture-handler`. Talks to the Go backend API.

## Layout

```
src/        screens/routes (expo-router), components, api clients
assets/     fonts, images
```

## Commands

```sh
just dev-mobile-install      # install deps
just dev-mobile             # start Expo
```

## Conventions

- **File-based routing** via expo-router; check versioned docs for the current router API.
- **Auth tokens in `expo-secure-store`** — never in async/local storage.
- **Bilingual** via `expo-localization`; mirror the web's AR/EN parity and RTL handling.
- Same backend contracts as web — reuse the API shapes, don't fork them.
