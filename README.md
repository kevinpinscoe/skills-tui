# skills-tui

A terminal UI for browsing and launching [Claude Code](https://claude.ai/code) skills from Markdown files.

## What it does

`skill` presents an interactive two-level chooser:

1. Pick a **category** (subdirectory of your skills directory)
2. Pick a **skill** (subdirectory containing a `SKILL.md` file)
3. Confirm, and Claude Code is launched with the skill file as the prompt

## Prerequisites

- [Claude Code](https://claude.ai/code) CLI (`claude`) installed and on your `PATH`
- A skills directory (default: `~/skills/skills`) containing category subdirectories with skill subdirectories each containing a `SKILL.md` file

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
    └── monday.com/
        └── create-item/
            └── SKILL.md
```

Skills are three levels deep: **category directory** → **skill directory** → **`SKILL.md`**.

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
