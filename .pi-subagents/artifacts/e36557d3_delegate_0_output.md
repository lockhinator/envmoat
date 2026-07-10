## Security Review — PLAN.md

Reviewed the full plan against encryption, authentication, storage, shell integration, CLI, and operational security. Prior architecture review identified a critical bash hook blocker (`$REPLY` guard) and other issues; the updated PLAN.md resolves all of them. Here are the remaining findings:

---

### No Blockers

All prior critical issues are resolved in the current plan:
- Shell `eval` injection → `%q`-style escaping (`strconv.QuoteToASCII`) specified
- Bash hook `$REPLY` guard → replaced with `$BASH_COMMAND` regex
- Key derivation contradiction → single global salt + HKDF per-bundle DEK
- Missing version byte → `0x01` version byte present
- No file locking → `flock` specified
- Symlink handling → `realpath` canonicalization specified

---

### Medium Issues

**1. Shared Keychain cache — no per-terminal session isolation**

The two-item Keychain pattern stores a single cache item shared across all terminals. Unlocking in one terminal unlocks for all. The plan acknowledges this ("shared across all terminals") but doesn't address threat scenarios where one terminal session is compromised (e.g., a malicious script runs in a tmux pane). If isolation matters, the cache item should be keyed by TTY or terminal session ID.

**2. `.secrets-manager` marker not auto-added to `.gitignore`**

The plan says the marker is "gitignored" but `init` doesn't explicitly add `.secrets-manager` to the project's `.gitignore`. The Docker command does auto-add `.env.secrets`, creating an inconsistency. If the marker is committed, it leaks project structure and bundle association. `init` should auto-append to `.gitignore` like the Docker command does.

---

### Low Issues

**3. Keychain item naming not specified**

The plan describes "two Keychain items per session" but doesn't specify the `SecAttrAccount` / `SecAttrService` labels used. Without a fixed naming convention (e.g., `secrets-manager-luk-protected`, `secrets-manager-luk-cache`), the `rotate` and `logout` commands can't reliably find and delete old items, and `doctor` can't audit Keychain state.

**4. No FileVault / disk encryption prerequisite check**

The plan stores encrypted bundles on disk and the LUK in the Keychain. If FileVault is disabled, the bundles are vulnerable to cold-boot or physical disk extraction attacks. The `setup` or `doctor` command should check for FileVault and warn if disabled.

**5. `index.json` stores plaintext path-to-bundle mapping**

`index.json` is stored unencrypted and maps absolute filesystem paths to bundle filenames. An attacker with read access to `~/.secrets-manager/` learns which projects use the tool and their directory structure. Mitigation: encrypt `index.json` or store only a hash of the path instead of the full path.

**6. Per-bundle DEK isolation is theoretical, not practical**

The HKDF uses `bundle_filename` as the context string. Bundle filenames are stored in plaintext in `index.json`, so anyone who can read the index can derive all DEKs from the LUK. The per-bundle DEK scheme only provides isolation if the LUK is compromised but the index is not — an unlikely split. If true per-bundle isolation is desired, each bundle needs its own random salt stored alongside it.

**7. No backup exclusion for `~/.secrets-manager/`**

The plan doesn't address Time Machine or other backup tools. Encrypted bundles in backups are acceptable, but `index.json` and `config.yaml` (with the global salt) in backups could aid an attacker who also obtains the Keychain database from a backup. Consider adding `com.apple.metadata:com.apple.backupd.excluded` or documenting manual exclusion.

**8. Login password change → orphaned Keychain items**

If the user changes their macOS login password, Keychain items may become inaccessible. The plan mentions `doctor` for diagnostics but doesn't specify a repair flow for this scenario.

---

### Positive Security Properties (confirmed)

| Area | Status |
|---|---|
| scrypt parameters (N=262144, r=8, p=1) | ✅ OWASP 2026 recommended |
| AES-256-GCM with 12-byte random nonce | ✅ Standard, authenticated encryption |
| File format versioning + auth tag | ✅ RFC 8452 compliant |
| File permissions (0600/0700) + startup validation | ✅ |
| Atomic writes (temp + rename) | ✅ |
| Shell-safe output (`strconv.QuoteToASCII`) | ✅ No eval injection |
| Secrets never passed as CLI args | ✅ |
| Interactive prompt / stdin for secret input | ✅ |
| Destructive operations require confirmation | ✅ |
| `kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly` | ✅ No iCloud sync, device-bound |
| Code signing + notarization | ✅ |
| `go mod verify`, reproducible builds | ✅ |
| Secure deletion on uninstall (multi-pass) | ✅ |
| Debug mode never logs secret values | ✅ |
| Bundle-hash dedup prevents redundant injection | ✅ |
| Interactive shell guard | ✅ |
| Sliding TTL with explicit cache deletion | ✅ |

---

### Verdict

**No blockers.** The plan is architecturally sound. The 2 medium and 6 low issues are refinements that can be addressed during implementation without changing the overall design. The core threat model (Keychain-protected LUK → encrypted bundles, shell-safe injection, no CLI arg leakage) is well-executed.