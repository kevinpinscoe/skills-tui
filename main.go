package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var version = "dev"

var titleStyle = lipgloss.NewStyle().MarginLeft(2)

type item struct {
	title string
	path  string
	mtime time.Time
}

type sortMode int

const (
	sortAlpha sortMode = iota
	sortMtime
	sortRecent
)

func parseSortMode(s string) (sortMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "alpha":
		return sortAlpha, nil
	case "mtime":
		return sortMtime, nil
	case "recent":
		return sortRecent, nil
	}
	return 0, fmt.Errorf("invalid sort mode %q (want alpha, mtime, or recent)", s)
}

func sortItems(items []item, mode sortMode) {
	switch mode {
	case sortAlpha:
		sort.SliceStable(items, func(i, j int) bool {
			return strings.ToLower(items[i].title) < strings.ToLower(items[j].title)
		})
	case sortMtime, sortRecent:
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].mtime.After(items[j].mtime)
		})
	}
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.path }
func (i item) FilterValue() string { return i.title }

type model struct {
	list     list.Model
	choice   item
	quitting bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.choice = i
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice.path != "" || m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

func chooseFromList(title string, items []item) (item, bool) {
	listItems := make([]list.Item, len(items))
	for i, it := range items {
		listItems[i] = it
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	l := list.New(listItems, delegate, 60, 16)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	m := model{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "chooser error:", err)
		os.Exit(1)
	}

	final := result.(model)
	if final.quitting {
		return item{}, false
	}
	return final.choice, true
}

func stripFrontmatter(content []byte) []byte {
	s := string(content)
	if !strings.HasPrefix(s, "---") {
		return content
	}
	end := strings.Index(s[3:], "\n---")
	if end == -1 {
		return content
	}
	rest := s[3+end+4:] // skip opening ---, content, and closing ---
	return []byte(strings.TrimLeft(rest, "\n"))
}

func isDir(parent string, entry os.DirEntry) bool {
	if entry.IsDir() {
		return true
	}
	if entry.Type()&os.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(filepath.Join(parent, entry.Name()))
	if err != nil {
		return false
	}
	return info.IsDir()
}

func hasRunnable(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "run.sh")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err == nil {
		return true
	}
	return false
}

func dirMtime(dir string) time.Time {
	info, err := os.Stat(dir)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func skillRecentMtime(skillDir string) time.Time {
	var newest time.Time
	for _, name := range []string{"run.sh", "SKILL.md"} {
		if info, err := os.Stat(filepath.Join(skillDir, name)); err == nil {
			if info.ModTime().After(newest) {
				newest = info.ModTime()
			}
		}
	}
	if newest.IsZero() {
		return dirMtime(skillDir)
	}
	return newest
}

func categoryRecentMtime(categoryDir string) time.Time {
	entries, err := os.ReadDir(categoryDir)
	if err != nil {
		return dirMtime(categoryDir)
	}
	var newest time.Time
	for _, e := range entries {
		if !isDir(categoryDir, e) || e.Name() == "archived" {
			continue
		}
		skillDir := filepath.Join(categoryDir, e.Name())
		if !hasRunnable(skillDir) {
			continue
		}
		if t := skillRecentMtime(skillDir); t.After(newest) {
			newest = t
		}
	}
	if newest.IsZero() {
		return dirMtime(categoryDir)
	}
	return newest
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func resolveSkillsDir() (path string, fromEnv bool) {
	if v := os.Getenv("SKILLS_DIR"); v != "" {
		return expandHome(v), true
	}
	return expandHome("~/skills/skills"), false
}

func main() {
	mode := sortAlpha
	if v := os.Getenv("SKILL_SORT"); v != "" {
		m, err := parseSortMode(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, "SKILL_SORT:", err)
			os.Exit(2)
		}
		mode = m
	}
	for _, arg := range os.Args[1:] {
		switch {
		case arg == "--help" || arg == "-h":
			fmt.Println("skill — browse and launch skills via Claude Code")
			fmt.Println()
			fmt.Println("Usage: skill [--help] [--version] [--sort=<order>]")
			fmt.Println()
			fmt.Println("  Presents an interactive chooser to select a skill category,")
			fmt.Println("  then a skill, then launches Claude Code with that skill as")
			fmt.Println("  the initial prompt.")
			fmt.Println()
			fmt.Println("Sort orders:")
			fmt.Println("  alpha    name, A→Z (default)")
			fmt.Println("  mtime    directory mod time, newest first")
			fmt.Println("  recent   newest run.sh / SKILL.md inside, newest first")
			fmt.Println()
			fmt.Println("Environment:")
			fmt.Println("  SKILLS_DIR   Root skills directory (default: ~/skills/skills)")
			fmt.Println("  SKILL_SORT   Default sort order (overridden by --sort)")
			os.Exit(0)
		case arg == "--version" || arg == "-v":
			dir, fromEnv := resolveSkillsDir()
			source := "default"
			if fromEnv {
				source = "SKILLS_DIR"
			}
			fmt.Printf("skill %s\n", version)
			fmt.Println()
			fmt.Println("Usage: skill [--help] [--version] [--sort=<order>]")
			fmt.Println("  Browse skill categories and launch Claude Code with the selected skill.")
			fmt.Println()
			fmt.Printf("Skills directory: %s (%s)\n", dir, source)
			os.Exit(0)
		case strings.HasPrefix(arg, "--sort="):
			m, err := parseSortMode(strings.TrimPrefix(arg, "--sort="))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			mode = m
		}
	}

	skillsDir, _ := resolveSkillsDir()

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading skills directory %s: %v\n", skillsDir, err)
		os.Exit(1)
	}

	var categories []item
	for _, entry := range entries {
		if !isDir(skillsDir, entry) {
			continue
		}
		if entry.Name() == "archived" {
			continue
		}
		subDir := filepath.Join(skillsDir, entry.Name())
		subEntries, _ := os.ReadDir(subDir)
		for _, sub := range subEntries {
			if !isDir(subDir, sub) {
				continue
			}
			if sub.Name() == "archived" {
				continue
			}
			skillDir := filepath.Join(subDir, sub.Name())
			if hasRunnable(skillDir) {
				cat := item{title: entry.Name(), path: subDir}
				switch mode {
				case sortMtime:
					cat.mtime = dirMtime(subDir)
				case sortRecent:
					cat.mtime = categoryRecentMtime(subDir)
				}
				categories = append(categories, cat)
				break
			}
		}
	}

	if len(categories) == 0 {
		fmt.Fprintln(os.Stderr, "no skill categories found in", skillsDir)
		os.Exit(1)
	}

	sortItems(categories, mode)

	chosenCategory, ok := chooseFromList("Skill Category", categories)
	if !ok {
		os.Exit(0)
	}

	subEntries, err := os.ReadDir(chosenCategory.path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading category directory: %v\n", err)
		os.Exit(1)
	}

	var skills []item
	for _, entry := range subEntries {
		if !isDir(chosenCategory.path, entry) {
			continue
		}
		if entry.Name() == "archived" {
			continue
		}
		skillDir := filepath.Join(chosenCategory.path, entry.Name())
		if !hasRunnable(skillDir) {
			continue
		}
		displayName := strings.ReplaceAll(entry.Name(), "-", " ")
		sk := item{title: displayName, path: skillDir}
		switch mode {
		case sortMtime:
			sk.mtime = dirMtime(skillDir)
		case sortRecent:
			sk.mtime = skillRecentMtime(skillDir)
		}
		skills = append(skills, sk)
	}

	if len(skills) == 0 {
		fmt.Fprintln(os.Stderr, "no skills found in category", chosenCategory.title)
		os.Exit(1)
	}

	sortItems(skills, mode)

	chosenSkill, ok := chooseFromList("Select Skill — "+chosenCategory.title, skills)
	if !ok {
		os.Exit(0)
	}

	fmt.Printf("Run skill \"%s\"? [y/N] ", chosenSkill.title)
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}

	runScript := filepath.Join(chosenSkill.path, "run.sh")
	skillFile := filepath.Join(chosenSkill.path, "SKILL.md")

	var cmd *exec.Cmd
	if _, err := os.Stat(runScript); err == nil {
		cmd = exec.Command("bash", "run.sh")
		cmd.Dir = chosenSkill.path
	} else {
		content, err := os.ReadFile(skillFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading skill file: %v\n", err)
			os.Exit(1)
		}
		cmd = exec.Command("claude", string(stripFrontmatter(content)))
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "error running claude: %v\n", err)
		os.Exit(1)
	}
}
