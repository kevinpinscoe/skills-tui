# skills-tui

A terminal UI for browsing and launching [Claude Code](https://claude.ai/code) skills from Markdown files.

## What it does

`skill` presents an interactive two-level chooser:

1. Pick a **category** (subdirectory of your skills directory)
2. Pick a **skill** (subdirectory containing either a `run.sh` script or a `SKILL.md` file)
3. Confirm, and the skill runs:
   - If the skill directory contains a `run.sh`, `skill` changes into that directory and executes `run.sh` as a shell script (stdin is wired through, so the script can prompt for input).
   - Otherwise, Claude Code is launched with the contents of `SKILL.md` as the prompt.

## Prerequisites

- [Claude Code](https://claude.ai/code) CLI (`claude`) installed and on your `PATH`
- A skills directory (default: `~/skills/skills`) containing category subdirectories with skill subdirectories. Each skill subdirectory must contain either a `run.sh` script or a `SKILL.md` file (or both — `run.sh` takes precedence).

See [https://github.com/kevinpinscoe/skills](https://github.com/kevinpinscoe/skills) for an example skills repository.

## Installation

### Download a pre-built binary

Grab the latest release for your platform from the [Releases](https://github.com/kevinpinscoe/skills-tui/releases) page:

| Platform | Binary |
|---|---|
| Linux x86-64 | `skill-linux-amd64` |
| macOS Apple Silicon | `skill-darwin-arm64` |
| Raspberry Pi (64-bit) | `skill-linux-arm64` |

```bash
# Example for macOS Apple Silicon
curl -L https://github.com/kevinpinscoe/skills-tui/releases/latest/download/skill-darwin-arm64 \
  -o ~/.local/bin/skill
chmod +x ~/.local/bin/skill
```

### Build from source

```bash
git clone https://github.com/kevinpinscoe/skills-tui.git
cd skills-tui
make install   # installs to ~/.local/bin/skill
```

## Usage

```
skill [--help]
```

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `SKILLS_DIR` | `~/skills/skills` | Path to the root skills directory |

## Skills directory layout

```
~/.local/bin/
└── skill              # this binary

~/skills/
└── skills/
    ├── aws/
    │   └── deploy/
    │       └── SKILL.md
    ├── backup/
    │   └── snapshot/
    │       └── SKILL.md
    └── YouTrack/
        └── create-ticket/
            ├── run.sh
            └── create-ticket.py
```

Skills are three levels deep: **category directory** → **skill directory** → **`run.sh` or `SKILL.md`**.

When a skill uses `run.sh`, the script is executed with its directory as the working directory, so it can reference co-located files (e.g. `./create-ticket.py`) by relative path.

## Examples

Example skill files and a ready-to-use skills repository can be found at [https://github.com/kevinpinscoe/skills](https://github.com/kevinpinscoe/skills).

## macOS Gatekeeper

Binaries downloaded from the internet are subject to macOS Gatekeeper. Starting with the release that includes ad-hoc signing, the `skill-darwin-arm64` binary is signed with `codesign --sign -` during the release build, which satisfies Gatekeeper for locally-run binaries without requiring an Apple Developer account.

If you still see a Gatekeeper rejection (e.g. `spctl --assess ~/.local/bin/skill` prints `rejected`), it may be because macOS has quarantined the file during download. Remove the quarantine attribute to allow it:

```bash
xattr -d com.apple.quarantine ~/.local/bin/skill
```

Or, to manually ad-hoc sign a binary you built from source:

```bash
codesign --sign - ~/.local/bin/skill
```

## Keyboard shortcuts

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate |
| `/` | Filter list |
| `Enter` | Select |
| `q` / `Esc` / `Ctrl+C` | Quit |
