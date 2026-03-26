# skills-tui

A terminal UI for browsing and launching [Claude Code](https://claude.ai/code) skills from Markdown files.

## What it does

`skill` presents an interactive two-level chooser:

1. Pick a **category** (subdirectory of your skills directory)
2. Pick a **skill** (Markdown file in that category)
3. Confirm, and Claude Code is launched with the skill file as the prompt

## Prerequisites

- [Claude Code](https://claude.ai/code) CLI (`claude`) installed and on your `PATH`
- A skills directory (default: `~/skills/skills`) containing category subdirectories with `.md` skill files

See [kevinpinscoe/skills](https://github.com/kevinpinscoe/skills) for an example skills repository.

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
  -o ~/skills/commands/skill
chmod +x ~/skills/commands/skill
```

### Build from source

```bash
git clone https://github.com/kevinpinscoe/skills-tui.git
cd skills-tui
make build   # installs to ~/skills/skill
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
~/skills/
├── commands/
│   └── skill          # this binary
└── skills/
    ├── aws/
    │   └── deploy.md
    ├── backup/
    │   └── snapshot.md
    └── monday.com/
        └── create-item.md
```

Skills are two levels deep: **category directory** → **skill `.md` file**.

## Keyboard shortcuts

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate |
| `/` | Filter list |
| `Enter` | Select |
| `q` / `Esc` / `Ctrl+C` | Quit |
