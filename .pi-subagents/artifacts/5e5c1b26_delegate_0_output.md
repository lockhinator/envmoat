The plan is **not clean** — there are 4 medium-severity issues that need addressing:

---

### Medium-Severity Findings

1. **Export encryption mechanism unspecified** — `export --to-file` is listed in the CLI but the encryption scheme for the exported file is not defined. Does it use the LUK (meaning anyone with the master password can decrypt the export), or a separate password/key? If separate, how is that key protected? This must be specified before implementation.

2. **No secure deletion on uninstall** — `secrets-manager uninstall` should offer secure erasure of `~/.secrets-manager/` (e.g., `shred` or multi-pass overwrite) rather than simple `rm`. If uninstall is triggered by suspected compromise, standard deletion leaves encrypted bundles recoverable from disk.

3. **No code signing / notarization** — The plan targets macOS distribution via asdf/mise but doesn't mention code signing the binary. Unsigned binaries trigger Gatekeeper warnings. For a security tool handling credentials, code signing is essential for supply chain integrity.

4. **No supply chain security** — No mention of reproducible builds, SBOM generation, or dependency auditing (`go mod verify`, Dependabot, etc.). A secrets manager should have stronger supply chain guarantees than a typical CLI tool.

---

### Low-Severity Findings (15 items, mostly inherent tradeoffs)

- **Auth tag 32 bytes** — Non-standard (GCM typically uses 16). Verify implementation doesn't concatenate extra data.
- **Unprotected cache during TTL** — LUK sits in an unprotected Keychain item for up to 15 minutes.
- **Touch ID fallback to login password** — macOS API limitation; documented.
- **Shared sessions across terminals** — Compromise of one terminal exposes all.
- **`eval` attack surface** — Mitigated by `QuoteToASCII` but inherently risky.
- **`get <KEY>` prints to stdout** — Value in terminal history. `--clip` is the safer option.
- **`.bak` files after rotate** — Old encrypted bundles remain on disk.
- **No timeout on `load`** — Hung `load` blocks every `cd` in bash.
- **Process memory / swap** — LUK could be paged to disk.
- **No rate limiting** — Touch ID has hardware limits; scrypt slows brute-force but doesn't prevent it.
- **Marker file manipulation** — Threat model dependent (attacker with write access already has env vars).
- **`flock` on NFS** — May not work on network filesystems.
- **Debug mode metadata leakage** — Could log sensitive paths or Keychain item names.
- **Config.yaml modification** — Attacker with read/write access could weaken settings (e.g., extend TTL).
- **TOCTOU race** — Between marker detection and bundle resolution.

---