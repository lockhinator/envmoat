# Secrets Manager — Plan

A macOS-first, terminal-targeted secrets manager for environment variable secrets in local development. Stores encrypted secret bundles centrally on your machine, injects them into terminal sessions when you `cd` into tracked directories.

## Core Requirements

- Secrets stored **outside** project directories (central global store)
- Values **encrypted** at rest
- **Biometric unlock** (Touch ID) with Mac login password fallback (macOS Keychain behavior)
- **Terminal injection** — env vars loaded into the current shell session
- **Hierarchical inheritance** — subdirectories inherit parent secrets unless explicitly disabled/overridden
- **Docker Compose integration** — generate `.env` files for container use
- macOS-targeted, installed via asdf/mise

## Architecture

### Storage Layout

```
~/.secrets-manager/
├── config.yaml              # global settings (TTL, defaults, global salt)
├── bundles/
│   ├── <bundle-id-1>.enc    # encrypted secret bundle
│   ├── <bundle-id-2>.enc
│   └── ...
└── index.json               # path → bundle mapping (schema below)
```

**index.json schema**:
```json
{
  "version": 1,
  "profiles": { "myapp-dev": "abc123.enc", "myapp-staging": "def456.enc" },
  "auto": { "/Users/cameron/projects/simple": "auto-simple.enc" }
}
```
- `profiles`: named profile → bundle filename (set by `profiles create`)
- `auto`: resolved directory path → bundle filename (set by `init`)

Session auth is stored in the **macOS Keychain** (two-item pattern), not on the filesystem.

### Encryption Model

**Key Derivation**

```
Master Password → scrypt(password, global_salt, N=262144, r=8, p=1) → 32-byte LUK (Local Unwrap Key)
DEK = HKDF-SHA256(LUK, bundle_filename, info="secrets-manager/v1/dek")   # 32-byte per-bundle DEK
```

- Global salt stored in `config.yaml` (one LUK for all bundles)
- `bundle_filename` is the `.enc` filename (e.g., `abc123.enc`) — deterministic per-bundle key derivation
- Compromising one bundle's DEK doesn't expose others
- LUK cached in Keychain after Touch ID unlock (single cache, per-bundle DEK derived on demand)

**Bundle Encryption**

```
Project Bundle = JSON { "_meta": {"created_at": "...", "updated_at": "..."}, "API_KEY": "sk-...", "DB_PASS": "..." }
Encrypted = AES-256-GCM(DEK, nonce, plaintext)
File format = [1B version=0x01][12B nonce][ciphertext][32B auth tag]
```

- Version byte enables future format migration
- Auth tag after ciphertext (RFC 8452 convention)
- Metadata in plaintext JSON for tracking

**File Permissions**

All store files created with restrictive permissions:
- Directories: `0700`
- Files (bundles, index, cache): `0600`
- Validated on startup — refuse to operate if permissions are too open

### Authentication

**Primary: Touch ID via macOS Keychain (Two-Item Pattern)**

Two Keychain items per session:

1. **Protected item** — LUK stored with `SecAccessControl` requiring biometry (`kSecAccessControlUserPresence`). Triggering access shows the native Touch ID dialog. Falls back to Mac login password (macOS behavior, documented limitation).
2. **Cache item** — LUK stored without access control, with a timestamp. Checked first on each access. If TTL expired, prompt via protected item.

On TTL expiry or `secrets-manager logout`, the cache item is explicitly deleted.

**Session Caching**

After Touch ID unlock, the session stays unlocked for a configurable TTL (default 15 minutes). The Keychain cache item is shared across all terminals — unlocking in one terminal unlocks for all. TTL uses a **sliding window** (resets on each successful access). On TTL expiry or `secrets-manager logout`, the cache item is explicitly deleted.

**Implementation: CGO + macOS Security Framework**

- `SecItemAdd` / `SecItemCopyMatching` with `kSecAttrAccessControl`
- `kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly`
- Go CGO bindings for the Security framework calls
- Consider [`99designs/keychain`](https://github.com/99designs/keychain) for basic operations, custom CGO only for `SecAccessControl`

### Shell Integration

A shell hook detects directory changes and triggers injection:

```bash
# zsh (via ~/.zshrc) — use chpwd hook, fires only on directory change
__secrets_manager_last_bundle=""
__secrets_manager_hook() {
  local output
  output=$(secrets-manager load)
  if [ -n "$output" ]; then
    local bundle_hash
    bundle_hash=$(echo "$output" | head -1 | cut -d: -f2)
    if [ "$bundle_hash" != "$__secrets_manager_last_bundle" ]; then
      eval "$output"
      __secrets_manager_last_bundle="$bundle_hash"
    fi
  fi
}
add-zsh-hook chpwd __secrets_manager_hook

# bash (via ~/.bashrc) — PROMPT_COMMAND fires after cd, before prompt
__secrets_manager_hook() {
  local output
  output=$(secrets-manager load)
  if [ -n "$output" ]; then
    local bundle_hash
    bundle_hash=$(echo "$output" | head -1 | cut -d: -f2)
    if [ "$bundle_hash" != "$__secrets_manager_last_bundle" ]; then
      eval "$output"
      __secrets_manager_last_bundle="$bundle_hash"
    fi
  fi
}
# Use array form for bash 5.1+ compatibility
if [[ "$BASH_VERSINFO" -ge 5 && "${BASH_VERSINFO[1]:-0}" -ge 1 ]]; then
  PROMPT_COMMAND=("$__secrets_manager_hook" "${PROMPT_COMMAND[@]}")
else
  PROMPT_COMMAND="__secrets_manager_hook${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
```

**Shell-Safe Output Format**

`secrets-manager load` outputs shell-safe assignments using single-quote escaping (equivalent to bash `printf '%q'`), ensuring all values are safe for `eval`:

```
#bundle_hash:sha256:abc123...
export API_KEY='sk-1234567890abcdef'
export DB_PASS='p@$$w"rd; with $pecial chars'
```

- First line is a comment with the bundle hash (for change detection)
- All values are single-quoted. Internal single quotes escaped as `'\''`  (close-quote, escaped-quote, open-quote)
- Secrets containing `$()`, backticks, newlines, and all shell metacharacters are inert inside single quotes
- Errors go to stderr; `load` exits 0 with no output when no bundle is found
- Hook only installs in interactive shells (`[[ $- == *i* ]]`)

### Directory Hierarchy & Inheritance

A gitignored marker file `.secrets-manager` in a directory tells the tool this is a tracked project root.

```
secrets-manager load
  → resolve PWD to canonical path (realpath, follows symlinks)
  → walk up from PWD looking for .secrets-manager marker (stops at /, boundary marker checked)
  → found at /Users/cameron/projects/myapp/.secrets-manager
  → parse marker: empty = default, "disabled" = skip, "profile: <name>" = override
  → load corresponding bundle from ~/.secrets-manager/bundles/
  → decrypt and emit shell-safe export statements
```

**Subdirectory inheritance**: walking up means all subdirectories find the same parent marker — automatic inheritance.

**Disable for subfolder**: place `.secrets-manager` with content `disabled` — loader stops and emits nothing.

**Override for subfolder**: place `.secrets-manager` with content `profile: <name>` — loads a different bundle for that subtree.

**Concurrency**: bundle writes use atomic temp-file + rename pattern. File locking via `flock` (BSD-compatible on macOS, `github.com/gofrs/flock`) prevents concurrent `set`/`remove` from corrupting bundles. `index.json` writes are also locked.

**Walk boundary**: stops at `/` (root). Configurable via `SECRETS_MANAGER_WALK_ROOT` env var.

**Marker parsing**: content trimmed of whitespace and trailing newline. Case-sensitive. Accepts exactly: empty file (default), `disabled`, or `profile: <name>`. All other content produces an error with the marker path.

**Debug mode**: `SECRETS_MANAGER_DEBUG=1` enables verbose stderr logging (directory walk, bundle resolution, Keychain access). Never logs secret values.

**Context resolution**: `set`, `get`, `list`, `remove`, `edit` resolve the active bundle the same way `load` does — walk up from PWD for marker, look up in `index.json`. If no marker found, error: "Not in a tracked directory. Run `secrets-manager init` or `cd` into a tracked project."

**Load error paths**:
- No marker found: exit 0, no output (common case, not an error)
- Marker found but bundle missing: exit 0, stderr warning: "Bundle not found. Run `secrets-manager init <path>`."
- Marker found but decrypt fails: exit 0, stderr warning: "Bundle corrupted. Run `secrets-manager verify`."
- Session expired: exit 0, no output. Next interactive command prompts for Touch ID.

**Rotate Keychain handling**: after re-encrypting with new LUK, old Keychain items deleted, new created. User re-auths on next access. On success: "Rotated N bundles. Run any command to authenticate with the new password."

**Keychain item naming**: fixed labels — `secrets-manager-luk-protected` and `secrets-manager-luk-cache` under service `secrets-manager`. Enables reliable find/delete by `rotate`, `logout`, `doctor`.

**FileVault check**: `doctor` checks FileVault status and warns if disabled. `setup` also warns.

**Marker .gitignore**: `init` auto-appends `.secrets-manager` to project `.gitignore`.

**Secret size limit**: values over 1MB print a warning (shell export practical limit ~2MB on macOS). No hard limit enforced.

### CLI Surface

```
# Setup & Help
secrets-manager                               # no args: welcome + usage summary; prompts to run setup if not configured
secrets-manager setup                         # create master password + install shell hook (idempotent: checks for existing hook via grep; use --reset to change password)
secrets-manager --version                     # print version
secrets-manager --help                        # print usage (available on every subcommand)

# Project Management
secrets-manager init [project-root]           # quick start: create marker + auto-named bundle (auto-<dirname>.enc). Defaults to current directory. Auto-adds .secrets-manager to .gitignore.
secrets-manager deinit <project-root>         # remove marker and bundle (prompts confirmation, -y to skip)
secrets-manager status                        # active profile, session TTL remaining, Keychain state, debug mode hint
secrets-manager doctor                        # diagnostics: Keychain state, hook installed?, store permissions, config validity
secrets-manager uninstall                     # remove shell hook from rc file + prompt about deleting store (offers secure erase)

# Secret Operations
secrets-manager set <KEY> [VALUE]             # add/update (omitting VALUE prompts interactively; re-auths if session expired)
secrets-manager set --stdin <KEY>             # read value from stdin
secrets-manager set --file <.env>             # bulk import from file
secrets-manager get <KEY>                     # print decrypted value to stdout
secrets-manager get <KEY> --clip              # copy to clipboard via pbcopy; prints "Copied to clipboard" on stderr
secrets-manager edit <KEY>                    # open $EDITOR with current value; saves on exit (re-auths if expired). Temp file in ~/.secrets-manager/ (not /tmp), mode 0600, deleted on exit.
secrets-manager list                          # list keys (values hidden, shows active profile name)
secrets-manager remove <KEY>                  # delete a secret (prompts confirmation, -y to skip)
secrets-manager logout                        # delete Keychain cache item, require re-auth on next access

# Profiles — named secret bundles, linkable to directories
secrets-manager profiles list                 # list all named profiles
secrets-manager profiles create <name>        # create a new named profile
secrets-manager profiles delete <name>        # delete profile + encrypted bundle (prompts confirmation, -y to skip)
secrets-manager profiles link <path> <name>   # create marker with "profile: <name>" in <path> + create bundle if needed. Auto-adds .secrets-manager to .gitignore. Errors if marker already exists (use --force to overwrite).
secrets-manager profiles unlink <path>        # remove a profile link

# Shell Integration
secrets-manager load                          # shell hook: emit shell-safe "export K=V" lines

# Password & Migration
secrets-manager rotate                        # re-encrypt all bundles with new password (two-phase atomic, progress messages, .bak rollback)
secrets-manager verify                        # integrity check: all bundles decrypt, index consistency, orphan cleanup
secrets-manager backup --to-file <path>       # export all bundles to encrypted file (encrypted with current LUK)
secrets-manager export --to-file <path>       # alias for backup
secrets-manager restore --from-file <path>    # import bundles (merges with existing, prompts on key conflict)
secrets-manager import --from-file <path>     # alias for restore

# Docker
secrets-manager docker-env [output-path]      # generate .env file (0600, auto-adds to .gitignore, named .env.secrets)
secrets-manager docker-compose [output-path]  # alias for docker-env
```

### Docker Compose Integration

```
secrets-manager docker-env [output-path]
```

Generates a `.env` file from the current project's decrypted secrets. Default output is `./.env.secrets`.
- File created with 0600 permissions
- Auto-appends `.env.secrets` to project `.gitignore` if not already present
- Prints usage hint: `docker compose --env-file=.env.secrets up`
- Could also support `--watch` mode to regenerate on secret changes

### Distribution

- Installed via **asdf** or **mise**
- Prebuilt macOS binary (Intel + Apple Silicon) or build-from-source plugin
- CGO allowed — links against macOS Security framework
- Installation command sets up the shell hook automatically
- **Code signing**: developer certificate + notarization for Gatekeeper compliance
- **Supply chain**: `go mod verify` in CI, reproducible builds, dependency audit

## Phased Implementation

| Phase | Scope |
|-------|-------|
| **1. MVP** | `setup`, `init`, `set`, `get`, `list`, `load`, `remove`, `deinit`, `verify` — master password creation, shell hook install, core encrypt/decrypt, shell-safe output, directory walk-up + marker resolution, atomic writes, file permissions, error messages, confirmation prompts, debug mode, key validation |
| **2. Touch ID** | Keychain two-item pattern, biometric unlock, session caching (sliding TTL), `status` command |
| **3. Profiles** | `profiles` subcommand (list, create, delete, link, unlink), directory-to-profile association |
| **4. Inheritance** | Marker `disabled`/override support, symlink handling, configurable walk root |
| **5. Docker** | `docker-env` subcommand, `.env` generation (0600, auto .gitignore) |
| **6. Backup/Rotate** | `backup`, `restore`, `rotate` (two-phase atomic), `edit` command, `get --clip` |
| **7. Polish** | asdf/mise plugin, shell completions, zsh + bash support |

## Open Questions

1. **Clipboard backend**: Use `pbcopy` (macOS native) or a cross-platform library for `--clip`?

## Resolved Design Decisions

### Setup & Onboarding
- **Setup flow**: `setup` creates master password + installs shell hook into detected rc file. Idempotent — greps for `__secrets_manager_hook` before appending. `--reset` flag changes password.
- **No-args behavior**: running `secrets-manager` with no subcommand shows welcome + usage, prompts to run `setup` if not configured.
- **--help/--version**: available on every subcommand.

### Shell Integration
- **Bash hook**: `PROMPT_COMMAND` (fires after cd, before prompt). Array form for bash 5.1+, string fallback for older. NOT a DEBUG trap.
- **Zsh hook**: `chpwd` (fires only on directory change).
- **Interactive guard**: hook installation wrapped in `[[ $- == *i* ]]` — never installs in non-interactive shells.
- **Shell-safe output**: all values single-quoted with `'\''`  escape — no eval injection possible.
- **Session expiry in hook**: `load` prints stderr hint on expiry. Hook sees empty stdout, no injection, user sees prompt to unlock.

### Encryption & Storage
- **scrypt parameters**: N=262144, r=8, p=1 (OWASP 2026 recommended minimum).
- **Key derivation**: LUK + HKDF per-bundle DEK — compromise containment across bundles.
- **File format**: version byte (0x01), auth tag after ciphertext, metadata in plaintext JSON.
- **File permissions**: 0700 directories, 0600 files, validated on startup.
- **Atomic writes**: all bundle mutations use temp file + rename.
- **Concurrency**: `flock` on bundle and index.json writes.
- **index.json schema**: versioned, with `profiles` and `auto` sections.

### Authentication
- **Session tokens**: Keychain two-item pattern (protected + cache), shared across terminals.
- **Session TTL**: sliding window (resets on each access).
- **Keychain item naming**: fixed labels `secrets-manager-luk-protected` / `secrets-manager-luk-cache` under service `secrets-manager`.
- **FileVault check**: `doctor` and `setup` warn if FileVault disabled.

### CLI & UX
- **CLI secret input**: interactive prompt or stdin — never pass secrets as CLI arguments.
- **Clipboard output**: `get --clip` copies via `pbcopy`, prints "Copied to clipboard" on stderr.
- **Destructive ops**: `remove`, `deinit`, `rotate`, `profiles delete` prompt for confirmation; `-y` / `--yes` to bypass.
- **Error messages**: all errors include actionable recovery hint.
- **Key naming**: alphanumeric + underscore + hyphen only; validated on `set`.
- **Secret size limit**: warning at 1MB (shell export practical limit ~2MB).
- **list output**: shows active profile name for context.
- **edit**: temp file in ~/.secrets-manager/ (not /tmp), mode 0600, deleted on exit.

### Profiles & Inheritance
- **Profiles**: named bundles via `profiles` subcommand. `init` = auto-named bundle; `profiles link` = shared named profile.
- **Auto-bundle naming**: `auto-<slugified-last-dirname>.enc`. Collision: append `-<short-hash>`.
- **Load resolution**: empty marker → check `auto` in index.json. `profile: name` → look up name in `profiles` in index.json.
- **Marker parsing**: whitespace-trimmed, case-sensitive, strict validation.
- **Marker .gitignore**: `init` and `profiles link` auto-append `.secrets-manager` to project `.gitignore`.
- **Symlink handling**: canonical paths (realpath) for all path hashing.
- **Walk boundary**: `/` by default. Configurable via `SECRETS_MANAGER_WALK_ROOT`.
- **profiles link**: errors if marker exists (use --force).

### Commands
- **docker-env**: primary name. `docker-compose` kept as alias.
- **backup/restore**: primary names. `export`/`import` kept as aliases.
- **rotate atomicity**: two-phase — decrypt all → encrypt all → atomic rename with `.bak` rollback. Progress messages.
- **Rotate Keychain**: old items deleted after re-encryption, user re-auths on next access.
- **verify orphan cleanup**: lists orphans, prompts for deletion (not automatic).
- **doctor**: diagnostics for Keychain state, hook installation, permissions, config validity.
- **uninstall**: removes shell hook from rc file, prompts about deleting store. Offers secure erase.
- **logout**: deletes Keychain cache item, requires re-auth on next access.
- **Context resolution**: `set`/`get`/`list`/`remove`/`edit` use same walk-up + index.json as `load`.

### Distribution
- **Shell support**: zsh (`chpwd`) and bash (`PROMPT_COMMAND`). Fish not supported.
- **Code signing**: developer certificate + notarization for Gatekeeper.
- **Supply chain**: `go mod verify`, reproducible builds, dependency audit.
- **Debug mode**: `SECRETS_MANAGER_DEBUG=1` for verbose stderr (never logs values).
