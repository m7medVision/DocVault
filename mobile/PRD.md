# DocVault Mobile PRD

## Goal

Build the DocVault mobile app with Expo and HeroUI Native. The mobile app must copy the core web functionality and add a native camera-first document capture workflow.

## Product Requirements

- Users can register, log in, log out, and restore sessions after app restart.
- Users can scan documents with the device camera.
- Users can pick existing files from the device.
- Users can upload documents and track upload/OCR/processing status.
- Users can browse, filter, and open documents.
- Users can view original files, OCR text, translated text, metadata, extracted dates, and reminders.
- Users can edit document title and corrected metadata values.
- Users can search documents with filters.
- Users can view and manage reminders.
- Users can browse and manage folders without drag-and-drop.
- Admin users can access members and audit information from mobile.
- The app supports English and Arabic content where backend/web already support it.

## Mobile Navigation

Bottom tabs:

- Home
- Documents
- Scan
- Search
- Reminders

Secondary screens:

- Login
- Register
- Upload
- Document detail
- Folder browser
- Metadata editor
- Document chat
- Admin members
- Admin audit log
- Settings/profile

## Camera Requirements

- Ask for camera permission before opening the scanner.
- Show clear states for granted, denied, and unavailable permissions.
- Capture one or more pages.
- Let users retake, remove, or add pages before upload.
- Upload captured images as supported document files.
- Validate file size and supported MIME type before upload.
- Show upload and backend processing status after submit.
- Include multi-page PDF generation, edge detection, crop, and deskew inside the same product scope as enhancements.

## Web Functionality To Copy

- Auth validation and protected access.
- Documents list filters by type and status.
- Document detail tabs for file, OCR text, and translation.
- Metadata display and correction.
- Extracted dates and reminders.
- Search query and filters.
- Reminder all/pending/sent filters.
- Folder create, rename, delete, and move document.
- Document chat.
- Admin members and audit log.

## Engineering Requirements

- Use Expo Router file-based routes.
- Use HeroUI Native for user-facing controls and cards.
- Keep route files thin.
- Keep API code in `src/features/*/api.ts`.
- Keep stateful feature logic in `src/features/*/use-*.ts`.
- Keep reusable UI in `src/components/*`.
- Keep shared constants in `src/constants/*`.
- Store tokens in native secure storage, not cookies.
- Avoid browser-only APIs in mobile code.
- Use direct backend API calls from mobile.
- Preserve strict TypeScript.

## Acceptance Criteria

- App starts with DocVault UI instead of Expo starter UI.
- HeroUI Native provider wraps the app.
- Main tabs are DocVault-specific.
- Scan screen handles camera permission and capture entry points.
- Upload/document/search/reminder screens have mobile-first layouts and API-ready hooks.
- Code structure is ready for backend integration without mixing UI and API concerns.
- Mobile lint passes.
