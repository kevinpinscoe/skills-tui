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

### Package managers

#### APT (Debian/Ubuntu)

```bash
curl -sL https://kevinpinscoe.github.io/apt/gpg.key \
  | sudo gpg --dearmor -o /etc/apt/keyrings/kevinpinscoe.gpg

echo "deb [signed-by=/etc/apt/keyrings/kevinpinscoe.gpg] \
  https://kevinpinscoe.github.io/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/kevinpinscoe.list

sudo apt update
sudo apt install skills-tui
```

#### DNF (Fedora/RHEL)

```bash
sudo curl -fsSL https://kevinpinscoe.github.io/rpm/kevinpinscoe.repo \
  -o /etc/yum.repos.d/kevinpinscoe.repo
sudo dnf install skills-tui
```

### Download a pre-built binary

Grab the latest release for your platform from the [Releases](https://github.com/kevinpinscoe/skills-tui/releases) page:

| Platform | Binary |
|---|---|
| Linux x86-64 | `skill-linux-amd64` |
| macOS Apple Silicon | `skill-darwin-arm64` |
| Raspberry Pi (64-bit) | `skill-linux-arm64` |

Each release includes a `checksums.txt` (SHA-256) and a `commit.txt` recording the exact git commit the binaries were built from.

```bash
# Linux x86-64
BINARY=skill-linux-amd64

# macOS Apple Silicon
BINARY=skill-darwin-arm64

# Raspberry Pi 64-bit
BINARY=skill-linux-arm64
```

```bash
BASE=https://github.com/kevinpinscoe/skills-tui/releases/latest/download

# Download the binary and verification files
curl -fsSL "$BASE/$BINARY"       -o ~/.local/bin/skill
curl -fsSL "$BASE/checksums.txt" -o /tmp/skill-checksums.txt
curl -fsSL "$BASE/commit.txt"    -o /tmp/skill-commit.txt

# Verify the checksum
echo "$(grep "$BINARY" /tmp/skill-checksums.txt | awk '{print $1}')  $HOME/.local/bin/skill" \
  | shasum -a 256 --check

# Confirm the source commit (optional — cross-reference with GitHub)
cat /tmp/skill-commit.txt

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
skill [--help] [--version] [--list] [--sort=<order>]
```

### Flags

| Flag | Description |
|---|---|
| `--help`, `-h` | Show usage and exit |
| `--version`, `-v` | Print version and the resolved skills directory |
| `--list` | Print categories and their skill directories with each directory's mtime, then exit (no chooser, plain text) |
| `--sort=<order>` | Order categories and skills; see *Sort orders* below |

### Sort orders

| Order | Behavior |
|---|---|
| `alpha` | Case-insensitive name, A→Z (default) |
| `mtime` | Directory mod time, newest first |
| `recent` | Newest `run.sh` / `SKILL.md` inside, newest first |

`--sort` and `SKILL_SORT` apply to both the interactive chooser and `--list` output.

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `SKILLS_DIR` | `~/skills/skills` | Path to the root skills directory |
| `SKILL_SORT` | `alpha` | Default sort order; overridden by `--sort` |

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

If the binary is killed immediately on launch (exit code 137 / SIGKILL), Gatekeeper has rejected the signature on a downloaded file. Two attributes can cause this:

- **`com.apple.quarantine`** — set when a browser downloads the file. Strip it with:

  ```bash
  xattr -d com.apple.quarantine ~/.local/bin/skill
  ```

- **`com.apple.provenance`** — set on macOS 14+ for files downloaded by `curl`. It is a protected attribute and **cannot be removed with `xattr`**. Re-sign the binary locally instead, which makes Gatekeeper trust your local ad-hoc signature:

  ```bash
  codesign --force --sign - ~/.local/bin/skill
  ```

Or, to manually ad-hoc sign a binary you built from source:

```bash
codesign --sign - ~/.local/bin/skill
```

Note: `spctl --assess --type execute ~/.local/bin/skill` may still print `rejected` for ad-hoc-signed binaries even after the workarounds above. That is `spctl`'s static assessment, not the actual execution gate — if `skill --version` runs successfully, the binary is fine.

## Keyboard shortcuts

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate |
| `/` | Filter list |
| `Enter` | Select |
| `←` / `Esc` / `q` | Go back to category list (from skill list) |
| `Esc` / `q` / `Ctrl+C` | Quit (from category list) |
