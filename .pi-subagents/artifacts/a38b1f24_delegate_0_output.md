## Security Review Summary

**11 findings: 0 critical, 2 medium, 9 low. No blockers.**

### Medium severity

1. **Keychain session independence contradiction** — Plan claims "independent per-terminal" sessions but the Keychain two-item pattern stores a single cache item shared across all terminals. Either key by terminal ID or change the claim to "shared session."

2. **Plaintext global salt in config.yaml** — Read access to the store enables offline password cracking. Acceptable for single-user local dev but should be documented.

### Low severity

3. **`load` stderr suppression** hides real failures (permissions, Keychain errors, missing binary)
4. **Clipboard persistence** — `get --clip` leaves secrets in clipboard indefinitely
5. **Symlink walk-up manipulation** — resolved path could point outside expected hierarchy
6. **Marker file tampering** — no integrity protection on `.secrets-manager`
7. **Cache TTL enforcement unspecified** — fixed vs. sliding window not defined
8. **No key name validation** — invalid names could break exports
9. **`profiles link` under-specified** — operation semantics undefined
10. **`rotate` multi-step atomicity** — needs explicit two-phase rollback procedure
11. **`flock` macOS compatibility** — BSD semantics differ from Linux

### Confirmed clean

Encryption scheme (AES-256-GCM + scrypt + HKDF), shell eval safety (`QuoteToASCII`), file permissions (0600/0700), CLI secret exposure prevention, error message hygiene, atomic writes, symlink canonicalization, Docker `.env` permissions, and destructive operation confirmations are all solid.