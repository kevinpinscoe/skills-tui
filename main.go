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

const uncategorized = "uncategorized"

type item struct {
	title    string
	path     string
	category string // for a skill item: its category. For a synthesized category item: the category key itself.
	mtime    time.Time
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

type appState int

const (
	stateCategory appState = iota
	stateSkill
)

type model struct {
	appState       appState
	categoryList   list.Model
	skillList      list.Model
	allSkills      []item
	chosenCategory item
	chosenSkill    item
	sortMode       sortMode
	quitting       bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q", "esc":
			if m.appState == stateSkill && m.skillList.FilterState() == list.Unfiltered {
				m.appState = stateCategory
				return m, nil
			}
			if m.appState == stateCategory {
				m.quitting = true
				return m, tea.Quit
			}
		case "left":
			if m.appState == stateSkill && m.skillList.FilterState() == list.Unfiltered {
				m.appState = stateCategory
				return m, nil
			}
		case "enter":
			switch m.appState {
			case stateCategory:
				if i, ok := m.categoryList.SelectedItem().(item); ok {
					skills := skillsInCategory(m.allSkills, i.category, m.sortMode)
					if len(skills) == 0 {
						return m, nil
					}
					m.chosenCategory = i
					m.skillList = newList("Select Skill — "+i.title, skills)
					m.appState = stateSkill
				}
				return m, nil
			case stateSkill:
				if i, ok := m.skillList.SelectedItem().(item); ok {
					m.chosenSkill = i
				}
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.categoryList.SetWidth(msg.Width)
		m.skillList.SetWidth(msg.Width)
		return m, nil
	}

	var cmd tea.Cmd
	switch m.appState {
	case stateCategory:
		m.categoryList, cmd = m.categoryList.Update(msg)
	case stateSkill:
		m.skillList, cmd = m.skillList.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	if m.quitting || m.chosenSkill.path != "" {
		return ""
	}
	switch m.appState {
	case stateCategory:
		return "\n" + m.categoryList.View()
	case stateSkill:
		return "\n" + m.skillList.View()
	}
	return ""
}

func newList(title string, items []item) list.Model {
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
	return l
}

// skillsInCategory filters the already-loaded flat skill list down to one
// category, sorted per mode. No disk access — the category level is
// synthesized from SKILL.md frontmatter, not a real directory.
func skillsInCategory(all []item, category string, mode sortMode) []item {
	var skills []item
	for _, sk := range all {
		if sk.category == category {
			skills = append(skills, sk)
		}
	}
	sortItems(skills, mode)
	return skills
}

// loadSkills walks SKILLS_DIR one level deep (Claude Code's own flat skill
// layout: skill-name/SKILL.md) and returns every entry that looks like a
// skill and isn't excluded by <skillsDir>/.gitignore. Replaces the old
// two-level category/skill walk and its several hand-duplicated copies.
func loadSkills(skillsDir string, mode sortMode) ([]item, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}
	patterns := loadIgnorePatterns(skillsDir)

	var skills []item
	for _, entry := range entries {
		name := entry.Name()
		if !isDir(skillsDir, entry) || name == "archived" {
			continue
		}
		if isIgnored(name, patterns) {
			continue
		}
		skillDir := filepath.Join(skillsDir, name)
		if !hasRunnable(skillDir) {
			continue
		}
		sk := item{
			title:    strings.ReplaceAll(name, "-", " "),
			path:     skillDir,
			category: readCategory(skillDir),
		}
		switch mode {
		case sortMtime:
			sk.mtime = dirMtime(skillDir)
		case sortRecent:
			sk.mtime = skillRecentMtime(skillDir)
		}
		skills = append(skills, sk)
	}
	return skills, nil
}

// groupByCategory synthesizes the category-chooser's first-level list from
// each skill's already-loaded category field — never from a directory, since
// there isn't one. A skill with no category: frontmatter groups under the
// literal "uncategorized" bucket rather than being dropped or erroring.
func groupByCategory(skills []item, mode sortMode) []item {
	order := []string{}
	seen := map[string]bool{}
	newest := map[string]time.Time{}
	for _, sk := range skills {
		if !seen[sk.category] {
			seen[sk.category] = true
			order = append(order, sk.category)
		}
		if sk.mtime.After(newest[sk.category]) {
			newest[sk.category] = sk.mtime
		}
	}
	categories := make([]item, 0, len(order))
	for _, cat := range order {
		categories = append(categories, item{title: cat, path: cat, category: cat, mtime: newest[cat]})
	}
	sortItems(categories, mode)
	return categories
}

func runChooser(categories []item, allSkills []item, mode sortMode) (item, bool) {
	m := model{
		appState:     stateCategory,
		categoryList: newList("Skill Category", categories),
		skillList:    newList("", nil),
		allSkills:    allSkills,
		sortMode:     mode,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "chooser error:", err)
		os.Exit(1)
	}
	final := result.(model)
	if final.quitting || final.chosenSkill.path == "" {
		return item{}, false
	}
	return final.chosenSkill, true
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

// frontmatterField returns the value of a top-level "field: value" line
// inside a SKILL.md's YAML frontmatter block, or "" if the file, the
// frontmatter block, or the field is absent. A hand-written line scan is
// enough for the one field this tool reads — not a general YAML parser.
func frontmatterField(skillMD string, field string) string {
	content, err := os.ReadFile(skillMD)
	if err != nil {
		return ""
	}
	s := string(content)
	if !strings.HasPrefix(s, "---") {
		return ""
	}
	end := strings.Index(s[3:], "\n---")
	if end == -1 {
		return ""
	}
	block := s[3 : 3+end]
	prefix := field + ":"
	for _, line := range strings.Split(block, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

// readCategory returns a skill directory's category: frontmatter value,
// falling back to the literal "uncategorized" bucket when the SKILL.md is
// missing the field (or there is no SKILL.md, only a run.sh).
func readCategory(skillDir string) string {
	if v := frontmatterField(filepath.Join(skillDir, "SKILL.md"), "category"); v != "" {
		return v
	}
	return uncategorized
}

// loadIgnorePatterns reads <skillsDir>/.gitignore, if present, into a small
// set of glob patterns matched against top-level entry names. This is
// deliberately not a full gitignore engine — SKILLS_DIR is flat (one level),
// so there is nothing nested to match against, and negation (!pattern) is
// out of scope until a real need for it shows up.
func loadIgnorePatterns(skillsDir string) []string {
	content, err := os.ReadFile(filepath.Join(skillsDir, ".gitignore"))
	if err != nil {
		return nil
	}
	var patterns []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "**/")
		line = strings.TrimSuffix(line, "/")
		if line == "" {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

func isIgnored(name string, patterns []string) bool {
	for _, p := range patterns {
		if ok, err := filepath.Match(p, name); err == nil && ok {
			return true
		}
	}
	return false
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

func printInventory(categories []item, allSkills []item, mode sortMode) {
	for i, cat := range categories {
		if i > 0 {
			fmt.Println()
		}
		fmt.Println(cat.title)
		skills := skillsInCategory(allSkills, cat.category, mode)
		for _, sk := range skills {
			fmt.Printf("  %-32s  %s\n", sk.title, dirMtime(sk.path).Format("2006-01-02 15:04"))
		}
	}
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
	return expandHome("~/.claude/skills"), false
}

func main() {
	mode := sortAlpha
	listMode := false
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
			fmt.Println("skills — browse and launch skills via Claude Code")
			fmt.Println()
			fmt.Println("Usage: skills [--help] [--version] [--list] [--sort=<order>]")
			fmt.Println()
			fmt.Println("  Presents an interactive chooser to select a skill category,")
			fmt.Println("  then a skill, then launches Claude Code with that skill as")
			fmt.Println("  the initial prompt. Categories are read from each skill's")
			fmt.Println("  SKILL.md category: frontmatter, not a directory.")
			fmt.Println()
			fmt.Println("Flags:")
			fmt.Println("  --list           Print skill directories and their mtimes, then exit")
			fmt.Println("  --sort=<order>   Order categories and skills (see below)")
			fmt.Println()
			fmt.Println("Sort orders:")
			fmt.Println("  alpha    name, A→Z (default)")
			fmt.Println("  mtime    directory mod time, newest first")
			fmt.Println("  recent   newest run.sh / SKILL.md inside, newest first")
			fmt.Println()
			fmt.Println("Environment:")
			fmt.Println("  SKILLS_DIR   Root skills directory (default: ~/.claude/skills)")
			fmt.Println("  SKILL_SORT   Default sort order (overridden by --sort)")
			os.Exit(0)
		case arg == "--version" || arg == "-v":
			dir, fromEnv := resolveSkillsDir()
			source := "default"
			if fromEnv {
				source = "SKILLS_DIR"
			}
			fmt.Printf("skills %s\n", version)
			fmt.Println()
			fmt.Println("Usage: skills [--help] [--version] [--list] [--sort=<order>]")
			fmt.Println("  Browse skill categories and launch Claude Code with the selected skill.")
			fmt.Println()
			fmt.Printf("Skills directory: %s (%s)\n", dir, source)
			os.Exit(0)
		case arg == "--list":
			listMode = true
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

	allSkills, err := loadSkills(skillsDir, mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading skills directory %s: %v\n", skillsDir, err)
		os.Exit(1)
	}
	if len(allSkills) == 0 {
		fmt.Fprintln(os.Stderr, "no skills found in", skillsDir)
		os.Exit(1)
	}

	categories := groupByCategory(allSkills, mode)

	if listMode {
		printInventory(categories, allSkills, mode)
		return
	}

	chosenSkill, ok := runChooser(categories, allSkills, mode)
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
