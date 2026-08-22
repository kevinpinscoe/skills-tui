# skills-tui

A terminal UI for browsing and launching [Claude Code](https://claude.ai/code) skills from Markdown files.

## What it does

`skills` presents an interactive two-level chooser:

1. Pick a **category** — synthesized from each skill's `SKILL.md` `category:` frontmatter field, not a directory (a skill with no `category:` groups under `uncategorized`)
2. Pick a **skill** (a subdirectory of your skills directory containing either a `run.sh` script or a `SKILL.md` file)
3. Confirm, and the skill runs:
   - If the skill directory contains a `run.sh`, `skills` changes into that directory and executes `run.sh` as a shell script (stdin is wired through, so the script can prompt for input).
   - Otherwise, Claude Code is launched with the contents of `SKILL.md` as the prompt.

## Prerequisites

- [Claude Code](https://claude.ai/code) CLI (`claude`) installed and on your `PATH`
- A skills directory (default: `~/.claude/skills` — Claude Code's own skill path) containing skill subdirectories, flat, one level deep. Each skill subdirectory must contain either a `run.sh` script or a `SKILL.md` file (or both — `run.sh` takes precedence). An entry matched by the skills directory's own `.gitignore` is excluded from the chooser — this is how third-party content installed into the same directory (e.g. a Claude Code plugin's own skills) stays out of the list without `skills` needing to know anything about where it came from.

See [https://github.com/kevinpinscoe/skills](https://github.com/kevinpinscoe/skills) for an example skills repository.

## Installation

### Package managers

#### Homebrew (macOS/Linux)

```bash
brew tap kevinpinscoe/homebrew-tap
brew install --cask skills-tui
```

Upgrading from a release before v3.0.1? This tool shipped as a formula until
then. Remove the old one first — a formula and a cask of the same name cannot
coexist:

```bash
brew uninstall skills-tui
```

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
| Linux x86-64 | `skills-linux-amd64` |
| macOS Apple Silicon | `skills-darwin-arm64` |
| Raspberry Pi (64-bit) | `skills-linux-arm64` |

Each release includes a `checksums.txt` (SHA-256) and a `commit.txt` recording the exact git commit the binaries were built from.

```bash
# Linux x86-64
BINARY=skills-linux-amd64

# macOS Apple Silicon
BINARY=skills-darwin-arm64

# Raspberry Pi 64-bit
BINARY=skills-linux-arm64
```

```bash
BASE=https://github.com/kevinpinscoe/skills-tui/releases/latest/download

# Download the binary and verification files
curl -fsSL "$BASE/$BINARY"       -o ~/.local/bin/skills
curl -fsSL "$BASE/checksums.txt" -o /tmp/skills-checksums.txt
curl -fsSL "$BASE/commit.txt"    -o /tmp/skills-commit.txt

# Verify the checksum
echo "$(grep "$BINARY" /tmp/skills-checksums.txt | awk '{print $1}')  $HOME/.local/bin/skills" \
  | shasum -a 256 --check

# Confirm the source commit (optional — cross-reference with GitHub)
cat /tmp/skills-commit.txt

chmod +x ~/.local/bin/skills
```

### Build from source

```bash
git clone https://github.com/kevinpinscoe/skills-tui.git
cd skills-tui
make install   # installs to ~/.local/bin/skills
```

## Usage

```
skills [--help] [--version] [--list] [--sort=<order>]
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
| `SKILLS_DIR` | `~/.claude/skills` | Path to the root skills directory |
| `SKILL_SORT` | `alpha` | Default sort order; overridden by `--sort` |

## Skills directory layout

```
~/.local/bin/
└── skills             # this binary

~/.claude/skills/
├── .gitignore          # excludes anything not owned by your skills repo
├── aws-deploy/
│   └── SKILL.md         # category: aws
├── backup-snapshot/
│   └── SKILL.md         # category: backup
└── youtrack-create-ticket/
    ├── run.sh            # category: youtrack
    └── create-ticket.py
```

Skills are flat, one level deep: **skill directory** → **`run.sh` or `SKILL.md`**. There is no
category directory — a skill's category comes from a `category:` field in its `SKILL.md`
frontmatter, which `skills` reads to build the first-level chooser.

When a skill uses `run.sh`, the script is executed with its directory as the working directory, so it can reference co-located files (e.g. `./create-ticket.py`) by relative path.

## Examples

Example skill files and a ready-to-use skills repository can be found at [https://github.com/kevinpinscoe/skills](https://github.com/kevinpinscoe/skills).

## macOS Gatekeeper

Binaries downloaded from the internet are subject to macOS Gatekeeper. Starting with the release that includes ad-hoc signing, the `skills-darwin-arm64` binary is signed with `codesign --sign -` during the release build, which satisfies Gatekeeper for locally-run binaries without requiring an Apple Developer account.

If the binary is killed immediately on launch (exit code 137 / SIGKILL), Gatekeeper has rejected the signature on a downloaded file. Two attributes can cause this:

- **`com.apple.quarantine`** — set when a browser downloads the file. Strip it with:

  ```bash
  xattr -d com.apple.quarantine ~/.local/bin/skills
  ```

- **`com.apple.provenance`** — set on macOS 14+ for files downloaded by `curl`. It is a protected attribute and **cannot be removed with `xattr`**. Re-sign the binary locally instead, which makes Gatekeeper trust your local ad-hoc signature:

  ```bash
  codesign --force --sign - ~/.local/bin/skills
  ```

Or, to manually ad-hoc sign a binary you built from source:

```bash
codesign --sign - ~/.local/bin/skills
```

Note: `spctl --assess --type execute ~/.local/bin/skills` may still print `rejected` for ad-hoc-signed binaries even after the workarounds above. That is `spctl`'s static assessment, not the actual execution gate — if `skills --version` runs successfully, the binary is fine.

## Keyboard shortcuts

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate |
| `/` | Filter list |
| `Enter` | Select |
| `←` / `Esc` / `q` | Go back to category list (from skill list) |
| `Esc` / `q` / `Ctrl+C` | Quit (from category list) |

## Security

Release artifacts are cosign-signed and each release publishes an SPDX SBOM.
See [SECURITY.md](SECURITY.md) for verification steps and how to report a
vulnerability.

## License

MIT — see [LICENSE](LICENSE).
