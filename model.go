package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	interactiveMode bool
	ignoreNamesFlag string
	ignoreRegexFlag string
	workers         uint
)

type model struct {
	groups       [][]string
	currentGroup int
	selected     map[int]struct{}
	quitting     bool
	err          error
	scanning     bool
	resultsChan  <-chan []string
}

var (
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("227")).Bold(true)
	groupStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2)
	itemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
)

func initialModel(resultsChan <-chan []string) model {
	return model{
		resultsChan: resultsChan,
		selected:    make(map[int]struct{}),
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

func waitForResults(results <-chan []string) tea.Cmd {
	return func() tea.Msg {
		if group, ok := <-results; ok {
			return group
		}
		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.currentGroup > 0 {
				m.currentGroup--
			}

		case "down", "j":
			if m.currentGroup < len(m.groups)-1 {
				m.currentGroup++
			}

		case " ":
			if _, ok := m.selected[m.currentGroup]; ok {
				delete(m.selected, m.currentGroup)
			} else {
				m.selected[m.currentGroup] = struct{}{}
			}

		case "d":
			if len(m.selected) > 0 {
				return m, m.deleteSelected()
			}
		}

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

func (m model) deleteSelected() tea.Cmd {
	return func() tea.Msg {
		var deleted []string
		for idx := range m.selected {
			if idx >= len(m.groups) {
				continue
			}
			group := m.groups[idx]
			if len(group) == 0 {
				continue
			}
			for i := 1; i < len(group); i++ {
				if err := os.Remove(group[i]); err == nil {
					deleted = append(deleted, group[i])
				}
			}
		}
		m.selected = make(map[int]struct{})
		return updateMsg{deleted: deleted}
	}
}

type updateMsg struct {
	deleted []string
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

	if m.scanning {
		b.WriteString("🔍 Scanning for duplicates...\n")
		b.WriteString(helpStyle.Render("(Press q to quit)\n"))
		return b.String()
	}

	if len(m.groups) == 0 {
		b.WriteString("🎉 No duplicates found!\n")
		return b.String()
	}

	if len(m.groups) > 0 && m.currentGroup < len(m.groups) {
		group := m.groups[m.currentGroup]
		groupBox := groupStyle.Render(
			fmt.Sprintf("Duplicate Group %d/%d\n\n", m.currentGroup+1, len(m.groups)) +
				renderItems(group, m.currentGroup, m.selected),
		)
		b.WriteString(groupBox + "\n\n")
	}

	help := helpStyle.Render(
		"↑/↓: Navigate • Space: Select • d: Delete selected • q: Quit",
	)
	b.WriteString(help)

	return b.String()
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
