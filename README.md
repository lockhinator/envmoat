# envmoat

Keep your secrets out of AI agent context. Encrypted environment variables, auto-injected into your terminal on `cd`, invisible to LLM agents scanning your codebase.

macOS + Linux · Go · [Touch ID](#authentication) · [asdf](#installation) / [mise](#installation)

## The Problem

AI coding agents (Claude, Cursor, Copilot) read your entire project tree. `.env` files, `.env.local`, `.env.development` — they're all in your repo directory, all visible to the agent, all ending up in the model's context window and potentially transmitted to remote servers.

envmoat stores encrypted secrets **outside** your project, in a central store at `~/.envmoat/`. No `.env` files. No secrets on disk where agents can read them. Secrets exist only in your shell session's memory, injected automatically when you `cd` into a tracked project.

## How It Works

```
~/.envmoat/
├── config.yaml          # global settings (TTL, salt)
├── bundles/
│   ├── auto-myapp.enc   # encrypted secret bundle
│   └── staging.enc      # named profile bundle
└── index.json           # path → bundle mapping (no secret values)
```

1. Run `envmoat init` in your project — creates a `.envmoat` marker file (gitignored)
2. Run `envmoat set API_KEY` — stores the value encrypted in `~/.envmoat/bundles/`
3. `cd` into the project — shell hook detects the marker, decrypts, injects env vars into your session
4. AI agents scanning your project see only an empty `.envmoat` marker — zero secrets

## Quick Start

```bash
# Install via asdf
asdf plugin add envmoat https://github.com/lockinator/asdf-envmoat.git
asdf install envmoat latest

# Or via mise
mise use --add envmoat

# First run
envmoat setup    # create master password + install shell hook

# In your project
cd ~/projects/myapp
envmoat init     # create marker + bundle
envmoat set API_KEY    # prompts for value
envmoat set DB_PASS
envmoat list     # shows keys (values hidden)

# cd away and back — secrets auto-inject
cd ..
cd myapp         # API_KEY and DB_PASS now in your shell
```

## Commands

```
# Setup
envmoat setup                          # create password + install shell hook
envmoat setup --reset                  # change master password

# Project Management
envmoat init [project-root]            # create marker + auto-named bundle
envmoat deinit <project-root>          # remove marker and bundle
envmoat status                         # active profile, session TTL, Keychain state
envmoat doctor                         # diagnostics: Keychain, hook, permissions
envmoat uninstall                      # remove hook + optionally delete store

# Secrets
envmoat set <KEY> [VALUE]              # add/update (omit VALUE for prompt)
envmoat set --stdin <KEY>              # read from stdin
envmoat set --file <.env>              # bulk import
envmoat get <KEY>                      # print value to stdout
envmoat get <KEY> --clip               # copy to clipboard
envmoat edit <KEY>                     # open $EDITOR with current value
envmoat list                           # list keys (values hidden)
envmoat remove <KEY>                   # delete a secret
envmoat logout                         # clear session, require re-auth

# Profiles (named bundles)
envmoat profiles list                  # list named profiles
envmoat profiles create <name>         # create named profile
envmoat profiles delete <name>         # delete profile + bundle
envmoat profiles link <path> <name>    # link project to profile
envmoat profiles unlink <path>         # remove profile link

# Shell Integration
envmoat load                           # emit "export K=V" lines (for shell hook)

# Maintenance
envmoat rotate                         # re-encrypt all bundles with new password
envmoat verify                         # integrity check + orphan cleanup
envmoat backup --to-file <path>        # encrypted export
envmoat restore --from-file <path>     # import from backup
```

## Authentication

**macOS**: Touch ID via Keychain `SecAccessControl`. Falls back to login password. Session stays unlocked for 15 minutes (sliding window), shared across all terminals.

**Linux**: GNOME Keyring or KWallet via DBus. PIN prompt via system dialog. Same session caching model.

## Shell Support

- **zsh**: `chpwd` hook (fires on directory change)
- **bash**: `PROMPT_COMMAND` hook (fires before prompt)
- Hook installs only in interactive shells
- Auto-installed by `envmoat setup`

## Directory Inheritance

Subdirectories inherit parent secrets automatically — the shell hook walks up from `PWD` looking for the `.envmoat` marker.

```
/projects/myapp/.envmoat          # marker — loads auto-myapp.enc
/projects/myapp/frontend/         # inherits parent marker
/projects/myapp/frontend/.envmoat # content: "disabled" — blocks inheritance
/projects/myapp/api/.envmoat      # content: "profile: api-staging" — overrides with named profile
```

## Security

- **Encryption**: AES-256-GCM with per-bundle DEK derived via HKDF-SHA256
- **Key derivation**: scrypt (N=262144, r=8, p=1) — OWASP 2026 recommended minimum
- **Compromise containment**: each bundle has a unique DEK — compromising one doesn't expose others
- **File permissions**: 0700 directories, 0600 files, validated on startup
- **Atomic writes**: temp file + rename for all bundle mutations
- **No secrets in project directories**: only an empty marker file

## Agent Safety

envmoat is designed specifically to keep secrets invisible to AI coding agents:

- ✅ Secrets stored encrypted outside project directories
- ✅ No `.env` files written to project directories
- ✅ Shell injection happens in your session memory only
- ✅ `.envmoat` marker contains zero secret data
- ✅ `index.json` has path→bundle mapping only, no values
- ✅ Clipboard copy never prints values to stdout

## Installation

### asdf

```bash
asdf plugin add envmoat https://github.com/lockinator/asdf-envmoat.git
asdf install envmoat latest
```

### mise

```bash
mise use --add envmoat
```

### From Source

```bash
go install github.com/lockinator/envmoat@latest
```

## Debug Mode

```bash
export ENVMOAT_DEBUG=1
envmoat load    # verbose stderr logging (never logs secret values)
```

## License

TBD
