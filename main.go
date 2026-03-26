package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var titleStyle = lipgloss.NewStyle().MarginLeft(2)

type item struct {
	title string
	path  string
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

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			fmt.Println("skill — browse and launch skills via Claude Code")
			fmt.Println()
			fmt.Println("Usage: skill [--help]")
			fmt.Println()
			fmt.Println("  Presents an interactive chooser to select a skill category,")
			fmt.Println("  then a skill, then launches Claude Code with that skill as")
			fmt.Println("  the initial prompt.")
			fmt.Println()
			fmt.Println("Environment:")
			fmt.Println("  SKILLS_DIR   Root skills directory (default: ~/skills/skills)")
			os.Exit(0)
		}
	}

	skillsDir := os.Getenv("SKILLS_DIR")
	if skillsDir == "" {
		skillsDir = "~/skills/skills"
	}
	skillsDir = expandHome(skillsDir)

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading skills directory %s: %v\n", skillsDir, err)
		os.Exit(1)
	}

	var categories []item
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(skillsDir, entry.Name())
		subEntries, _ := os.ReadDir(subDir)
		for _, sub := range subEntries {
			if !sub.IsDir() && strings.HasSuffix(sub.Name(), ".md") {
				categories = append(categories, item{
					title: entry.Name(),
					path:  subDir,
				})
				break
			}
		}
	}

	if len(categories) == 0 {
		fmt.Fprintln(os.Stderr, "no skill categories found in", skillsDir)
		os.Exit(1)
	}

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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		displayName := strings.TrimSuffix(entry.Name(), ".md")
		displayName = strings.ReplaceAll(displayName, "-", " ")
		skills = append(skills, item{
			title: displayName,
			path:  filepath.Join(chosenCategory.path, entry.Name()),
		})
	}

	if len(skills) == 0 {
		fmt.Fprintln(os.Stderr, "no skills found in category", chosenCategory.title)
		os.Exit(1)
	}

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

	content, err := os.ReadFile(chosenSkill.path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading skill file: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("claude", string(content))
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
