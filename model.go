package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	groups       [][]string
	currentGroup int
	currentFile  int
	selected     map[int]map[int]struct{}
	quitting     bool
	err          error
	scanning     bool
	resultsChan  <-chan []string
	showConfirm  bool
	toDelete     []string
}

type scanCompleteMsg struct{}

var (
	fileStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	selectedFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("227")).Bold(true)
	deleteStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	confirmStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("196")).Padding(1)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
)

func initialModel(resultsChan <-chan []string) model {
	return model{
		resultsChan: resultsChan,
		selected:    make(map[int]map[int]struct{}),
		scanning:    true,
		groups:      make([][]string, 0),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		waitForResults(m.resultsChan),
		tea.EnterAltScreen,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showConfirm {
			switch msg.String() {
			case "y", "Y":
				return m.deleteFiles()
			case "n", "N", "esc":
				m.showConfirm = false
				m.toDelete = nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if len(m.groups) == 0 {
				return m, nil
			}
			if m.currentFile > 0 {
				m.currentFile--
			}

		case "down", "j":
			if len(m.groups) == 0 {
				return m, nil
			}
			if m.currentFile < len(m.groups[m.currentGroup])-1 {
				m.currentFile++
			}

		case "left", "h":
			if len(m.groups) == 0 {
				return m, nil
			}
			if m.currentGroup > 0 {
				m.currentGroup--
				m.currentFile = 0
			}

		case "right", "l":
			if len(m.groups) == 0 {
				return m, nil
			}
			if m.currentGroup < len(m.groups)-1 {
				m.currentGroup++
				m.currentFile = 0
			}

		case " ":
			if len(m.groups) == 0 {
				return m, nil
			}
			group := m.currentGroup
			file := m.currentFile

			if _, exists := m.selected[group]; !exists {
				m.selected[group] = make(map[int]struct{})
			}

			if _, selected := m.selected[group][file]; selected {
				delete(m.selected[group], file)
			} else {
				m.selected[group][file] = struct{}{}
			}

		case "d":
			if len(m.groups) == 0 {
				return m, nil
			}
			m.toDelete = m.getSelectedFiles()
			if len(m.toDelete) > 0 {
				m.showConfirm = true
			}

		}

	case scanCompleteMsg:
		m.scanning = false
		return m, nil

	case []string:
		m.scanning = false
		m.groups = append(m.groups, msg)
		return m, waitForResults(m.resultsChan)

	case error:
		m.err = msg
		return m, tea.Quit

	case tea.WindowSizeMsg:
		return m, nil
	}

	return m, nil
}

func (m model) getSelectedFiles() []string {
	var toDelete []string
	for groupIdx, files := range m.selected {
		if groupIdx >= len(m.groups) {
			continue
		}
		for fileIdx := range files {
			if fileIdx < len(m.groups[groupIdx]) {
				toDelete = append(toDelete, m.groups[groupIdx][fileIdx])
			}
		}
	}
	return toDelete
}

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	if m.quitting {
		return "\n  Scan completed. Exiting...\n\n"
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Duplicate File Finder") + "\n\n")

	if m.showConfirm {
		return confirmStyle.Render(
			fmt.Sprintf("Delete %d selected files?\n\n", len(m.toDelete))+
				"[y] Yes  [n] No\n") +
			helpStyle.Render("(This cannot be undone)")
	}

	if m.scanning {
		b.WriteString("🔍 Scanning for duplicates...\n")
	} else if len(m.groups) == 0 {
		b.WriteString("🎉 No duplicates found!\n")
	} else {
		current := m.groups[m.currentGroup]
		b.WriteString(fmt.Sprintf(" Group %d/%d (%d files)\n",
			m.currentGroup+1, len(m.groups), len(current)))

		for i, file := range current {
			var line strings.Builder
			if _, selected := m.selected[m.currentGroup][i]; selected {
				line.WriteString(selectedFileStyle.Render("◉ "))
			} else {
				line.WriteString("◌ ")
			}

			if i == m.currentFile {
				line.WriteString("➔ ")
			} else {
				line.WriteString("  ")
			}

			line.WriteString(fileStyle.Render(file))

			if _, selected := m.selected[m.currentGroup][i]; selected {
				line.WriteString(deleteStyle.Render(" (marked for deletion)"))
			}

			b.WriteString(line.String() + "\n")
		}
	}

	helpText := ""
	if !m.showConfirm {
		if m.scanning {
			helpText = helpStyle.Render("(Press q to quit)")
		} else if len(m.groups) > 0 {
			helpText = helpStyle.Render(
				"↑/↓: Navigate files • ←/→: Switch groups • Space: Select • d: Delete selected • q: Quit",
			)
		} else {
			helpText = helpStyle.Render("(Press q to quit)")
		}
	}

	if helpText != "" {
		b.WriteString("\n\n" + helpText)
	}

	return b.String()
}

func (m model) deleteFiles() (tea.Model, tea.Cmd) {
	deleted := 0
	for _, path := range m.toDelete {
		if err := os.Remove(path); err == nil {
			deleted++
			m.removeDeletedFile(path)
		}
	}

	m.selected = make(map[int]map[int]struct{})
	m.toDelete = nil
	m.showConfirm = false

	return m, tea.Batch(
		tea.Printf("%s %d files deleted", deleteStyle.Render("✔"), deleted),
	)
}

func (m model) removeDeletedFile(path string) {
	for groupIdx, group := range m.groups {
		for fileIdx, filePath := range group {
			if filePath == path {
				m.groups[groupIdx] = append(group[:fileIdx], group[fileIdx+1:]...)

				if groupIdx == m.currentGroup && fileIdx <= m.currentFile {
					m.currentFile = max(0, m.currentFile-1)
				}
				break
			}
		}

		if len(m.groups[groupIdx]) == 0 {
			m.groups = append(m.groups[:groupIdx], m.groups[groupIdx+1:]...)
			m.currentGroup = max(0, min(m.currentGroup, len(m.groups)-1))
		}
	}
}

func waitForResults(results <-chan []string) tea.Cmd {
	return func() tea.Msg {
		group, ok := <-results
		if !ok {
			return scanCompleteMsg{}
		}
		return group
	}
}

type updateMsg struct {
	deleted []string
}

func renderItems(items []string, current int, selected map[int]struct{}) string {
	var b strings.Builder
	for _, item := range items {
		prefix := "  "
		if _, ok := selected[current]; ok {
			prefix = selectedStyle.Render("✔ ")
		}
		b.WriteString(fmt.Sprintf("%s %s\n", prefix, itemStyle.Render(item)))
	}
	return b.String()
}
