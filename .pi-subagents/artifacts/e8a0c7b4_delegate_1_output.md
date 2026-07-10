Review written to `ARCHITECTURE-REVIEW.md`. Key findings:

**Critical**: `eval "$(secrets-manager load)"` with unescaped secret values is arbitrary code execution — any secret containing `$()`, backticks, or quotes executes as shell code. Must use `%q`-style quoting.

**High**: The encryption model has an internal contradiction — per-file salts mean per-file derived keys, but the plan caches a single "decryption key" in the Keychain. Either use one global salt or implement HKDF per-file key derivation.

**High**: No version byte in the encrypted file format makes future migration impossible.

**Medium**: `PROMPT_COMMAND` string concatenation breaks bash 5.1+ arrays; no file locking on bundle writes; symlink behavior undefined for path hashing.

**Missing**: `edit`, `status`, `deinit`, `export`/`import`, `verify` commands. No atomic write strategy. No concurrent access protection.