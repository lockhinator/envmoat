# Progress

## Status
In Progress

## Tasks
- [x] Read PLAN.md lines 1-100 for encryption details
- [x] Append implementation plan to SM-gtv.7 (encryption + storage)
- [x] Append implementation plan to SM-gtv.4 (directory walk)
- [x] Append implementation plan to SM-gtv.1 (platform abstraction)

## Files Changed
- None (backlog database updates only)

## Notes
All Phase 1 infra tasks have detailed implementation plans appended via `bd update`. Plans include package structure, files, dependencies, and acceptance criteria aligned with PLAN.md encryption model (scrypt N=262144, HKDF-SHA256, AES-256-GCM, file format [0x01][12B nonce][ciphertext][32B auth tag]).

## Phase 6 (Backup/Rotate) Implementation Plans Added

- **SM-1i3.1** (backup + restore): Implementation plan appended covering cmd/ (backup.go, restore.go), EMBT binary format, HKDF backup-key derivation, conflict prompts (skip/overwrite/cancel), export/import aliases
- **SM-1i3.2** (edit): Implementation plan appended covering cmd/ (edit.go), temp file in ~/.envmoat with mode 0600, $EDITOR exec, atomic write on success, cleanup via defer
- **SM-1i3.3** (clipboard): Implementation plan appended covering cmd/ (get.go --clip flag), clipboard backend, stderr-only confirmation, no stdout secret leak
- **SM-1i3.4** (rotate): Implementation plan appended covering cmd/ (rotate.go), two-phase decrypt/re-encrypt, .bak rollback, keychain SecItemDelete, progress messages

## Phase 7 (Polish) Implementation Plans Added

- **SM-rgt.1** (CI pipeline): Implementation plan appended covering .github/workflows/ci.yml, go mod verify, govulncheck, reproducible builds, SHA256SUMS, GPG signatures, pinned Actions versions
- **SM-rgt.2** (shell completions): Implementation plan appended covering cmd/ (completions.go), cobra GenZsh/GenBashCompletion, dynamic profile/key completions, install hints
- **SM-rgt.3** (asdf plugin): Implementation plan appended covering asdf plugin repo structure, bin/install/list-all-versions/latest-version, GitHub release downloads, go build fallback
- **SM-rgt.4** (mise plugin): Implementation plan appended covering mise.toml plugin spec, GitHub releases with OS/arch detection, build-from-source fallback
- **SM-rgt.5** (codesign + notarization): Implementation plan appended covering .github/workflows/release.yml, codesign with Developer ID, notarytool submission, stapler staple

All nine `bd update` commands succeeded. Phase 6 and Phase 7 tasks now have detailed implementation plans with package structure, file layout, dependencies, and acceptance criteria.

## Phase 1 CLI Implementation Plans Added

- **SM-gtv.5** (setup + init): Implementation plan appended covering cmd/ package structure, password setup via golang.org/x/term, salt generation, shell hook install (zsh chpwd + bash PROMPT_COMMAND), --reset flag, init marker/bundle creation, auto-bundle naming, .gitignore append
- **SM-gtv.2** (set + get + list): Implementation plan appended covering set/get/list commands, interactive prompts, --stdin/--file flags, key validation regex, 1MB size warning, context resolution via resolver
- **SM-gtv.3** (load + remove + deinit + verify): Implementation plan appended covering shell-safe output (single-quote escaping), bundle hash line, remove/deinit with confirmation, verify orphan detection, -y/--yes flag

All three `bd update` commands succeeded. PLAN.md lines 100-200 reviewed for shell hook details (zsh chpwd hook, bash PROMPT_COMMAND, shell-safe output format with single-quote escaping, bundle hash change detection).

## Phase 2b Linux Implementation Plans Added

- **SM-coy.2** (Linux GNOME Keyring): Implementation plan appended covering internal/backend/linux_keyring.go, godbus/dbus/v5, Secret Service API (org.freedesktop.secrets), OpenSession/CreateCollection/SearchItems/SetAttributes/Clear methods, two-item pattern, item attributes
- **SM-coy.3** (Linux KDE KWallet): Implementation plan appended covering KDE detection via XDG_CURRENT_DESKTOP, KWallet DBus (org.kde.KWallet), open/folderList/writeEntry methods, fallback to Secret Service
- **SM-coy.1** (Linux clipboard): Implementation plan appended covering internal/backend/linux_clipboard.go, Wayland (wl-clipboard) vs X11 (xclip) detection, no-op fallback

## Phase 3 Profiles Implementation Plans Added

- **SM-idf.3** (profiles list+create): Implementation plan appended covering cmd/ profiles.go, Cobra subcommands, name validation regex, atomic index.json writes, bundle naming
- **SM-idf.1** (profiles delete): Implementation plan appended covering deletion from index.json + bundle file, confirmation prompt, -y flag
- **SM-idf.2** (profiles link+unlink): Implementation plan appended covering .envmoat marker creation, .gitignore append, --force flag, unlink marker-only removal

## Phase 4 Inheritance Implementation Plans Added

- **SM-1pz.2** (disabled marker): Implementation plan appended covering "disabled" marker content handling, nil bundle return, load command no-output behavior
- **SM-1pz.4** (profile override marker): Implementation plan appended covering "profile: <name>" parsing, index.json lookup, error on missing profile
- **SM-1pz.1** (symlink canonicalization): Implementation plan appended covering filepath.EvalSymlinks, canonical path storage, duplicate prevention
- **SM-1pz.3** (walk root boundary): Implementation plan appended covering ENVMOAT_WALK_ROOT env var, absolute path validation, walk boundary enforcement

All 10 `bd update` commands succeeded for Phase 2b (Linux), Phase 3 (Profiles), and Phase 4 (Inheritance).
