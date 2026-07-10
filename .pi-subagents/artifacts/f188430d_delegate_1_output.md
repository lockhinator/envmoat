**Review findings:**

**Internal contradiction found:**

1. **`index.json` `links` field is defined but never used.** The schema documents `links` as `"path → bundle mapping"`, yet:
   - No CLI command writes to `links` (`init` writes to `auto`, `profiles link` writes a marker file on disk)
   - The load resolution flow ("empty marker → check `auto` in index.json; `profile: name` → look up in `profiles`") never reads from `links`
   - The marker-file approach (`.secrets-manager` with `profile: <name>`) fully replaces whatever `links` was intended to do

   Either `links` should be removed from the schema, or a command/resolution path should be added that uses it. As written, it's dead schema.

No missing core features — all seven stated core requirements (central store, encryption, biometric unlock, terminal injection, hierarchical inheritance, Docker Compose integration, asdf/mise distribution) are covered across the phased implementation.

---