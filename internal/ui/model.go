package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spinchange/yanp-tui/internal/config"
	"github.com/spinchange/yanp-tui/internal/vault"
)

type mode int

const (
	modeBrowse mode = iota
	modeDashboard
	modeHealth
	modeHelp
	modeFilter
	modeFirstRun
	modeSwitchVault
	modeCreateVault
	modeNew
	modeCapture
	modeRename
	modePublish
)

type reloadMsg struct {
	v   *vault.Vault
	err error
}

type actionMsg struct {
	status    string
	vaultPath string
	notePath  string
	cfg       *config.Config
	err       error
}

type Model struct {
	cfg       config.Config
	vaultPath string
	vault     *vault.Vault
	allNotes  []*vault.Note
	notes     []*vault.Note
	filter    string
	dashIndex int
	selected  int
	width     int
	height    int
	viewport  viewport.Model
	input     textinput.Model
	mode      mode
	status    string
	setupNote string
	err       error
	style     styles
}

type dashboardItem struct {
	label   string
	detail  string
	relPath string
	action  string
}

type styles struct {
	title      lipgloss.Style
	subtle     lipgloss.Style
	selected   lipgloss.Style
	panel      lipgloss.Style
	statusGood lipgloss.Style
	statusBad  lipgloss.Style
}

func New(vaultPath string, cfg config.Config) (Model, error) {
	input := textinput.New()
	input.CharLimit = 260

	viewport := viewport.New(0, 0)
	style := styles{
		title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("221")),
		subtle:     lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		selected:   lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Padding(0, 1),
		panel:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1),
		statusGood: lipgloss.NewStyle().Foreground(lipgloss.Color("120")),
		statusBad:  lipgloss.NewStyle().Foreground(lipgloss.Color("204")),
	}

	m := Model{
		cfg:       cfg,
		vaultPath: vaultPath,
		viewport:  viewport,
		input:     input,
		mode:      modeDashboard,
		status:    "YANP TUI",
		style:     style,
	}
	if strings.TrimSpace(vaultPath) == "" {
		m.mode = modeFirstRun
		m.setupNote = "Pick an existing vault folder or create a new one."
		m.status = "No vault configured yet"
		m.input.Placeholder = "Path to your vault folder"
		m.input.Focus()
		m.refreshPreview()
		return m, nil
	}
	if err := m.loadVault(vaultPath); err != nil {
		if os.IsNotExist(err) {
			m.mode = modeFirstRun
			m.setupNote = "That configured vault path does not exist yet. Choose a vault folder or create one."
			m.status = "Configured vault not found"
			m.input.Placeholder = "Path to your vault folder"
			m.input.SetValue(vaultPath)
			m.input.Focus()
			m.err = nil
			m.refreshPreview()
			return m, nil
		}
		return Model{}, err
	}
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = max(20, msg.Width-42)
		m.viewport.Height = max(8, msg.Height-8)
		m.refreshPreview()
	case tea.KeyMsg:
		if m.mode == modeFirstRun {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.input.SetValue("")
				m.status = "First-run setup"
				return m, nil
			case "V":
				m.startPrompt(modeCreateVault, "Path for a new vault folder")
				return m, nil
			}
			return m.updatePrompt(msg)
		}
		if m.mode == modeHelp {
			switch msg.String() {
			case "esc", "?", "h":
				m.mode = modeDashboard
				m.status = "Back to dashboard"
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		if m.mode == modeHealth {
			switch msg.String() {
			case "esc", "g":
				m.mode = modeDashboard
				m.status = "Back to dashboard"
				return m, nil
			case "?", "h":
				m.mode = modeHelp
				m.status = "Help"
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		if m.mode == modeDashboard {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "?", "h":
				m.mode = modeHelp
				m.status = "Help"
				return m, nil
			case "enter", "l":
				return m.activateDashboardItem()
			case "/":
				m.startPrompt(modeFilter, "Filter notes by title, tag, alias, path, or text")
				m.input.SetValue(m.currentFilter())
				return m, nil
			case "j", "down":
				items := m.dashboardItems()
				if len(items) > 0 && m.dashIndex < len(items)-1 {
					m.dashIndex++
				}
				return m, nil
			case "k", "up":
				if m.dashIndex > 0 {
					m.dashIndex--
				}
				return m, nil
			case "r":
				m.status = "Refreshing vault index..."
				return m, loadVaultCmd(m.vaultPath)
			case "d":
				return m, ensurePeriodicCmd(m.vaultPath, vault.PeriodicDaily)
			case "w":
				return m, ensurePeriodicCmd(m.vaultPath, vault.PeriodicWeekly)
			case "m":
				return m, ensurePeriodicCmd(m.vaultPath, vault.PeriodicMonthly)
			case "v":
				m.startPrompt(modeSwitchVault, "Path to an existing vault folder")
				m.input.SetValue(m.vaultPath)
				return m, nil
			case "V":
				m.startPrompt(modeCreateVault, "Path for a new vault folder")
				if strings.TrimSpace(m.vaultPath) != "" {
					m.input.SetValue(filepath.Dir(m.vaultPath))
				}
				return m, nil
			case "n":
				m.startPrompt(modeNew, "Title for new note")
				return m, nil
			case "c":
				m.startPrompt(modeCapture, "Capture text for inbox.md")
				return m, nil
			}
			return m, nil
		}
		if m.mode != modeBrowse {
			return m.updatePrompt(msg)
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?", "h":
			m.mode = modeHelp
			m.status = "Help"
			return m, nil
		case "g":
			m.mode = modeDashboard
			m.status = "Dashboard"
			return m, nil
		case "/":
			m.startPrompt(modeFilter, "Filter notes by title, tag, alias, path, or text")
			m.input.SetValue(m.currentFilter())
			return m, nil
		case "esc":
			if m.currentFilter() != "" {
				m.applyFilter("")
				m.status = "Cleared filter"
			}
		case "j", "down":
			if m.selected < len(m.notes)-1 {
				m.selected++
				m.refreshPreview()
			}
		case "k", "up":
			if m.selected > 0 {
				m.selected--
				m.refreshPreview()
			}
		case "r":
			m.status = "Refreshing vault index..."
			return m, loadVaultCmd(m.vaultPath)
		case "d":
			return m, ensurePeriodicCmd(m.vaultPath, vault.PeriodicDaily)
		case "w":
			return m, ensurePeriodicCmd(m.vaultPath, vault.PeriodicWeekly)
		case "m":
			return m, ensurePeriodicCmd(m.vaultPath, vault.PeriodicMonthly)
		case "v":
			m.startPrompt(modeSwitchVault, "Path to an existing vault folder")
			m.input.SetValue(m.vaultPath)
		case "V":
			m.startPrompt(modeCreateVault, "Path for a new vault folder")
			if strings.TrimSpace(m.vaultPath) != "" {
				m.input.SetValue(filepath.Dir(m.vaultPath))
			}
		case "n":
			m.startPrompt(modeNew, "Title for new note")
		case "c":
			m.startPrompt(modeCapture, "Capture text for inbox.md")
		case "R":
			if m.currentNote() != nil {
				m.startPrompt(modeRename, "New title for selected note")
				m.input.SetValue(m.currentNote().Title)
			}
		case "p":
			m.startPrompt(modePublish, "Publish output directory")
			m.input.SetValue(filepath.Join(m.vaultPath, "_published"))
		}
	case reloadMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = "Vault refresh failed"
			return m, nil
		}
		m.vault = msg.v
		m.allNotes = msg.v.Notes
		m.notes = msg.v.Notes
		if m.selected >= len(m.notes) {
			m.selected = max(0, len(m.notes)-1)
		}
		items := m.dashboardItems()
		if m.dashIndex >= len(items) {
			m.dashIndex = max(0, len(items)-1)
		}
		if filter := m.currentFilter(); filter != "" {
			m.applyFilter(filter)
		}
		m.status = fmt.Sprintf("Refreshed %d notes", len(m.notes))
		m.err = nil
		m.refreshPreview()
	case actionMsg:
		m.input.Blur()
		if msg.err != nil {
			m.err = msg.err
			m.status = msg.err.Error()
			return m, nil
		}
		if msg.cfg != nil {
			m.cfg = *msg.cfg
		}
		if strings.TrimSpace(msg.vaultPath) != "" {
			if err := m.loadVault(msg.vaultPath); err != nil {
				m.err = err
				m.status = err.Error()
				return m, nil
			}
			m.mode = modeDashboard
			if strings.TrimSpace(msg.notePath) != "" {
				if idx := m.indexOfNote(msg.notePath); idx >= 0 {
					m.selected = idx
					m.mode = modeBrowse
					m.refreshPreview()
				}
			}
		} else {
			m.mode = modeBrowse
		}
		m.err = nil
		m.status = msg.status
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updatePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeBrowse
		m.input.Blur()
		m.status = "Cancelled"
		return m, nil
	case "enter":
		value := strings.TrimSpace(m.input.Value())
		switch m.mode {
		case modeFirstRun, modeSwitchVault:
			return m, switchVaultCmd(m.cfg, value)
		case modeCreateVault:
			return m, createVaultCmd(m.cfg, value)
		case modeFilter:
			m.mode = modeBrowse
			m.input.Blur()
			m.applyFilter(value)
			if value == "" {
				m.status = fmt.Sprintf("Showing all %d notes", len(m.notes))
			} else {
				m.status = fmt.Sprintf("Filter %q matched %d notes", value, len(m.notes))
			}
			return m, nil
		case modeNew:
			return m, createNoteCmd(m.vaultPath, value)
		case modeCapture:
			return m, captureCmd(m.vaultPath, value)
		case modeRename:
			if note := m.currentNote(); note != nil {
				return m, renameCmd(m.vaultPath, note.RelPath, value)
			}
		case modePublish:
			return m, publishCmd(m.vaultPath, value)
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.mode == modeFirstRun {
		return m.renderFirstRun()
	}
	if m.mode == modeHelp {
		return m.renderHelp()
	}
	if m.mode == modeHealth {
		return m.renderHealth()
	}
	if m.mode == modeDashboard {
		return m.renderDashboard()
	}

	leftWidth := min(38, max(28, m.width/3))
	rightWidth := max(20, m.width-leftWidth-5)

	header := m.style.title.Render("YANP TUI") + "\n" +
		m.style.subtle.Render("j/k move  / filter  g dashboard  v switch vault  V new vault  n new  c capture  R rename  p publish  r refresh  ? help  q quit")

	list := m.renderList(leftWidth)
	preview := m.style.panel.Width(rightWidth).Height(max(10, m.height-8)).Render(m.viewport.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, list, preview)
	statusStyle := m.style.statusGood
	statusText := m.status
	if m.err != nil {
		statusStyle = m.style.statusBad
		statusText = m.err.Error()
	}
	footer := statusStyle.Render(statusText)
	if m.mode != modeBrowse {
		footer += "\n" + m.style.subtle.Render(promptLabel(m.mode)) + "\n" + m.input.View()
	} else if m.currentFilter() != "" {
		footer += "\n" + m.style.subtle.Render("Active filter: "+m.currentFilter()+"  (esc clears)")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(header + "\n\n" + body + "\n\n" + footer)
}

func (m Model) renderList(width int) string {
	var rows []string
	for i, note := range m.notes {
		row := fmt.Sprintf("%-20s  %s", truncate(note.Title, 20), note.RelPath)
		if i == m.selected {
			rows = append(rows, m.style.selected.Width(width-4).Render(row))
			continue
		}
		rows = append(rows, lipgloss.NewStyle().Width(width-2).Render(row))
	}
	if len(rows) == 0 {
		rows = append(rows, "No notes found.")
	}
	return m.style.panel.Width(width).Height(max(10, m.height-8)).Render(strings.Join(rows, "\n"))
}

func (m *Model) refreshPreview() {
	note := m.currentNote()
	if note == nil {
		m.viewport.SetContent("No note selected.")
		return
	}

	var meta []string
	meta = append(meta, fmt.Sprintf("Path: %s", note.RelPath))
	meta = append(meta, fmt.Sprintf("Modified: %s", note.ModifiedAt.Format(time.RFC822)))
	if len(note.Tags) > 0 {
		sortedTags := append([]string(nil), note.Tags...)
		sort.Strings(sortedTags)
		meta = append(meta, "Tags: "+strings.Join(sortedTags, ", "))
	}
	if len(note.Aliases) > 0 {
		meta = append(meta, "Aliases: "+strings.Join(note.Aliases, ", "))
	}
	if note.Status != "" {
		meta = append(meta, "Status: "+note.Status)
	}
	content := m.style.title.Render(note.Title) + "\n" +
		m.style.subtle.Render(strings.Join(meta, "\n")) + "\n\n" +
		note.Body
	m.viewport.SetContent(content)
}

func (m *Model) startPrompt(nextMode mode, placeholder string) {
	m.mode = nextMode
	m.input.SetValue("")
	m.input.Placeholder = placeholder
	m.input.Focus()
}

func (m Model) currentNote() *vault.Note {
	if len(m.notes) == 0 || m.selected < 0 || m.selected >= len(m.notes) {
		return nil
	}
	return m.notes[m.selected]
}

func promptLabel(mode mode) string {
	switch mode {
	case modeDashboard:
		return "Dashboard"
	case modeHealth:
		return "Vault health"
	case modeHelp:
		return "Help"
	case modeFirstRun:
		return "First-run setup"
	case modeFilter:
		return "Filter note list"
	case modeSwitchVault:
		return "Switch to an existing vault"
	case modeCreateVault:
		return "Create a new vault"
	case modeNew:
		return "Create note in the vault root"
	case modeCapture:
		return "Append a capture entry to inbox.md"
	case modeRename:
		return "Rename selected note and rewrite inbound wikilinks"
	case modePublish:
		return "Publish notes to a separate CommonMark directory"
	default:
		return ""
	}
}

func (m Model) renderHelp() string {
	header := m.style.title.Render("YANP TUI Help")
	body := strings.Join([]string{
		"YANP TUI is a terminal interface for a YANP vault.",
		"",
		"Keys",
		"  enter / l   Open the selected dashboard item",
		"  j / down    Move selection down",
		"  k / up      Move selection up",
		"  /           Filter notes by title, alias, tag, path, or text",
		"  d           Open or create today's daily note",
		"  w           Open or create this week's note",
		"  m           Open or create this month's note",
		"  g           Return to the dashboard",
		"  v           Switch to a different vault location",
		"  V           Create and switch to a new vault",
		"  esc         Leave health/help or cancel a prompt",
		"  n           Create a new note in the vault root",
		"  c           Capture a quick entry into inbox.md",
		"  R           Rename the selected note and rewrite inbound wikilinks",
		"  p           Publish the vault to a separate CommonMark directory",
		"  r           Refresh the vault index",
		"  ? or h      Open this help screen",
		"  q           Quit",
		"",
		"YANP Basics",
		"  - Notes are plain Markdown files with optional YAML frontmatter.",
		"  - Internal links use wikilinks like [[YANP TUI]].",
		"  - Resolution order is title, aliases, then filename stem.",
		"  - Publish rewrites wikilinks to standard relative Markdown links.",
		"",
		"Current Screen",
		"  - Dashboard: quick summary and selectable shortcuts",
		"  - Health: duplicate-target and unresolved-link report",
		"  - Left panel: note list",
		"  - Right panel: preview of the selected note",
		"  - Footer: status and prompts",
		"",
		"Press esc, h, or ? to return.",
	}, "\n")

	content := m.style.panel.Width(max(50, m.width-6)).Render(header + "\n\n" + body)
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (m Model) renderDashboard() string {
	header := m.style.title.Render("YANP TUI") + "\n" +
		m.style.subtle.Render("j/k choose  enter open  d daily  w weekly  m monthly  / filter  v switch vault  V new vault  n new  c capture  r refresh  ? help  q quit")

	recent := m.renderDashboardItems()
	overview := m.renderOverview()
	statusStyle := m.style.statusGood
	statusText := m.status
	if m.err != nil {
		statusStyle = m.style.statusBad
		statusText = m.err.Error()
	}
	footer := statusStyle.Render(statusText)
	if m.currentFilter() != "" {
		footer += "\n" + m.style.subtle.Render("Active filter: "+m.currentFilter())
	}
	if m.mode == modeFilter {
		footer += "\n" + m.style.subtle.Render(promptLabel(m.mode)) + "\n" + m.input.View()
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, overview, recent)
	return lipgloss.NewStyle().Padding(1, 2).Render(header + "\n\n" + body + "\n\n" + footer)
}

func (m Model) renderOverview() string {
	total := len(m.allNotes)
	filtered := len(m.notes)
	conflicts := m.vault.Conflicts()
	unresolved := m.vault.UnresolvedLinks()
	inbox := "missing"
	inboxEntries := 0
	if m.vault.NoteByRelPath("inbox.md") != nil {
		inbox = "present"
		inboxEntries = inboxEntryCount(m.vault.NoteByRelPath("inbox.md"))
	}
	dailySummary := periodicSummary(m.vault, vault.PeriodicDaily, time.Now())
	weeklySummary := periodicSummary(m.vault, vault.PeriodicWeekly, time.Now())
	monthlySummary := periodicSummary(m.vault, vault.PeriodicMonthly, time.Now())

	sections := []string{
		m.style.title.Render("Overview"),
		"",
		fmt.Sprintf("Vault: %s", m.vaultPath),
		fmt.Sprintf("Total notes: %d", total),
		fmt.Sprintf("Visible notes: %d", filtered),
		fmt.Sprintf("Inbox: %s", inbox),
		fmt.Sprintf("Inbox entries: %d", inboxEntries),
		fmt.Sprintf("Conflicts: %d", len(conflicts)),
		fmt.Sprintf("Unresolved links: %d", unresolvedLinkCount(unresolved)),
	}

	filter := m.currentFilter()
	if filter != "" {
		sections = append(sections, fmt.Sprintf("Filter: %s", filter))
	}

	sections = append(sections,
		"",
		m.style.title.Render("Current Period"),
		"",
		fmt.Sprintf("Today: %s", dailySummary),
		fmt.Sprintf("This week: %s", weeklySummary),
		fmt.Sprintf("This month: %s", monthlySummary),
		"",
		m.style.title.Render("Vault health"),
		"",
	)
	if len(conflicts) == 0 && len(unresolved) == 0 {
		sections = append(sections, "No duplicate targets or unresolved wikilinks detected.")
	} else {
		if len(conflicts) == 0 {
			sections = append(sections, "No duplicate title, alias, or filename targets detected.")
		} else {
			limit := min(3, len(conflicts))
			for i := 0; i < limit; i++ {
				conflict := conflicts[i]
				sections = append(sections, fmt.Sprintf("- %s conflict: %s", conflict.Matched, conflict.Name))
			}
		}
		if len(unresolved) == 0 {
			sections = append(sections, "No unresolved wikilinks detected.")
		} else {
			limit := min(3, len(unresolved))
			for i := 0; i < limit; i++ {
				link := unresolved[i]
				sections = append(sections, fmt.Sprintf("- unresolved link: %s (%s)", link.Target, link.Source.RelPath))
			}
		}
		sections = append(sections, "Open the health report from the dashboard for details.")
	}

	sections = append(sections,
		"",
		m.style.title.Render("Next actions"),
		"",
		"Use j/k to choose a dashboard item.",
		"Press d, w, or m to jump into the current period note.",
		"Use inbox and current-period notes as your main landing points.",
		"Press enter to open the selected target.",
		"Press / to search and narrow the list.",
		"Press n to create a new note.",
		"Press c to capture directly into inbox.md.",
	)

	return m.style.panel.Width(max(36, m.width/2-4)).Height(max(12, m.height-8)).Render(strings.Join(sections, "\n"))
}

func (m Model) renderFirstRun() string {
	header := m.style.title.Render("YANP First-Run Setup")
	lines := []string{
		"YANP-TUI needs a real vault location before it can open notes.",
		"",
		"Enter a folder path and press Enter to use an existing vault.",
		"Press Shift+V from the app later if you want to create a new vault folder.",
		"",
		"Examples:",
		"  G:\\Notes",
		"  D:\\PKM\\vault",
	}
	if strings.TrimSpace(m.setupNote) != "" {
		lines = append(lines, "", m.setupNote)
	}
	body := m.style.panel.Width(max(50, m.width-6)).Render(header + "\n\n" + strings.Join(lines, "\n"))
	footer := m.style.subtle.Render(promptLabel(m.mode)) + "\n" + m.input.View()
	return lipgloss.NewStyle().Padding(1, 2).Render(body + "\n\n" + footer)
}

func (m Model) renderDashboardItems() string {
	lines := []string{m.style.title.Render("Shortcuts"), ""}
	items := m.dashboardItems()
	if len(items) == 0 {
		lines = append(lines, "No shortcuts available.")
	} else {
		for i, item := range items {
			row := fmt.Sprintf("%s", item.label)
			detail := item.detail
			if i == m.dashIndex {
				lines = append(lines, m.style.selected.Width(max(28, m.width/2-10)).Render(row))
			} else {
				lines = append(lines, row)
			}
			if detail != "" {
				lines = append(lines, "  "+detail)
			}
		}
	}
	return m.style.panel.Width(max(36, m.width/2-4)).Height(max(12, m.height-8)).Render(strings.Join(lines, "\n"))
}

func (m Model) dashboardItems() []dashboardItem {
	var items []dashboardItem
	items = append(items, dashboardItem{
		label:  "Browse visible notes",
		detail: fmt.Sprintf("%d notes in the current list", len(m.notes)),
		action: "browse",
	})
	conflicts := m.vault.Conflicts()
	unresolved := m.vault.UnresolvedLinks()
	if len(conflicts) > 0 || len(unresolved) > 0 {
		items = append(items, dashboardItem{
			label:  "Open vault health report",
			detail: fmt.Sprintf("%d conflict(s), %d unresolved link(s)", len(conflicts), unresolvedLinkCount(unresolved)),
			action: "health",
		})
	}
	if note := m.vault.NoteByRelPath("inbox.md"); note != nil {
		items = append(items, dashboardItem{
			label:   "Open inbox",
			detail:  note.RelPath,
			relPath: note.RelPath,
			action:  "note",
		})
	}
	items = append(items, dashboardItem{
		label:  "Open today's daily note",
		detail: "Create it if it does not exist yet",
		action: "daily",
	})
	items = append(items, dashboardItem{
		label:  "Open this week's note",
		detail: "Create it if it does not exist yet",
		action: "weekly",
	})
	items = append(items, dashboardItem{
		label:  "Open this month's note",
		detail: "Create it if it does not exist yet",
		action: "monthly",
	})
	todayRel := filepath.ToSlash(filepath.Join("daily", time.Now().Format("2006-01-02")+".md"))
	if note := m.vault.NoteByRelPath(todayRel); note != nil {
		items = append(items, dashboardItem{
			label:   "Open today's daily note",
			detail:  note.RelPath,
			relPath: note.RelPath,
			action:  "note",
		})
	}
	limit := min(5, len(m.notes))
	for i := 0; i < limit; i++ {
		note := m.notes[i]
		items = append(items, dashboardItem{
			label:   fmt.Sprintf("Recent: %s", note.Title),
			detail:  note.RelPath,
			relPath: note.RelPath,
			action:  "note",
		})
	}
	return items
}

func (m Model) activateDashboardItem() (tea.Model, tea.Cmd) {
	items := m.dashboardItems()
	if len(items) == 0 {
		m.status = "No dashboard items available"
		return m, nil
	}
	if m.dashIndex < 0 || m.dashIndex >= len(items) {
		m.dashIndex = 0
	}
	item := items[m.dashIndex]
	switch item.action {
	case "browse":
		m.mode = modeBrowse
		m.status = "Browsing notes"
		m.refreshPreview()
		return m, nil
	case "note":
		if idx := m.indexOfNote(item.relPath); idx >= 0 {
			m.selected = idx
			m.mode = modeBrowse
			m.status = "Opened " + item.relPath
			m.refreshPreview()
			return m, nil
		}
		m.status = "Shortcut target is not in the current note list"
		return m, nil
	case "health":
		m.mode = modeHealth
		m.status = fmt.Sprintf("Vault health: %d conflict(s), %d unresolved link(s)", len(m.vault.Conflicts()), unresolvedLinkCount(m.vault.UnresolvedLinks()))
		return m, nil
	case "daily":
		return m, ensurePeriodicCmd(m.vaultPath, vault.PeriodicDaily)
	case "weekly":
		return m, ensurePeriodicCmd(m.vaultPath, vault.PeriodicWeekly)
	case "monthly":
		return m, ensurePeriodicCmd(m.vaultPath, vault.PeriodicMonthly)
	default:
		return m, nil
	}
}

func (m Model) renderHealth() string {
	header := m.style.title.Render("YANP Vault Health") + "\n" +
		m.style.subtle.Render("g or esc return to dashboard  ? help  q quit")
	conflicts := m.vault.Conflicts()
	unresolved := m.vault.UnresolvedLinks()

	lines := []string{}
	if len(conflicts) == 0 && len(unresolved) == 0 {
		lines = append(lines, "No duplicate targets or unresolved wikilinks detected.")
	} else {
		lines = append(lines, "Conflicts", "")
		if len(conflicts) == 0 {
			lines = append(lines, "No duplicate title, alias, or filename targets detected.", "")
		} else {
			lines = append(lines, fmt.Sprintf("Detected %d conflict(s).", len(conflicts)), "")
			for _, conflict := range conflicts {
				lines = append(lines, fmt.Sprintf("%s conflict: %s", titleLabel(conflict.Matched), conflict.Name))
				if conflict.Winner != nil {
					lines = append(lines, fmt.Sprintf("  winner: %s", conflict.Winner.RelPath))
				}
				for _, candidate := range conflict.Candidates {
					lines = append(lines, fmt.Sprintf("  candidate: %s", candidate.RelPath))
				}
				lines = append(lines, "")
			}
		}

		lines = append(lines, "Unresolved wikilinks", "")
		if len(unresolved) == 0 {
			lines = append(lines, "No unresolved wikilinks detected.")
		} else {
			lines = append(lines, fmt.Sprintf("Detected %d unresolved wikilink(s).", unresolvedLinkCount(unresolved)), "")
			for _, link := range unresolved {
				lines = append(lines, fmt.Sprintf("%s -> %s", link.Source.RelPath, link.Target))
				if link.Count > 1 {
					lines = append(lines, fmt.Sprintf("  occurrences: %d", link.Count))
				}
				lines = append(lines, "")
			}
		}
	}

	content := m.style.panel.Width(max(50, m.width-6)).Render(header + "\n\n" + strings.Join(lines, "\n"))
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (m *Model) applyFilter(query string) {
	query = strings.TrimSpace(query)
	m.filter = query
	if query == "" {
		m.notes = m.allNotes
		m.selected = 0
		m.refreshPreview()
		return
	}
	needle := strings.ToLower(query)
	var filtered []*vault.Note
	for _, note := range m.allNotes {
		if noteMatches(note, needle) {
			filtered = append(filtered, note)
		}
	}
	m.notes = filtered
	m.selected = 0
	m.refreshPreview()
}

func (m Model) currentFilter() string {
	return m.filter
}

func (m Model) indexOfNote(relPath string) int {
	for i, note := range m.notes {
		if note.RelPath == relPath {
			return i
		}
	}
	return -1
}

func noteMatches(note *vault.Note, needle string) bool {
	if strings.Contains(strings.ToLower(note.Title), needle) || strings.Contains(strings.ToLower(note.RelPath), needle) || strings.Contains(strings.ToLower(note.Body), needle) {
		return true
	}
	for _, alias := range note.Aliases {
		if strings.Contains(strings.ToLower(alias), needle) {
			return true
		}
	}
	for _, tag := range note.Tags {
		if strings.Contains(strings.ToLower(tag), needle) {
			return true
		}
	}
	return false
}

func (m *Model) loadVault(root string) error {
	v, err := vault.Load(root)
	if err != nil {
		return err
	}
	m.vaultPath = root
	m.vault = v
	m.allNotes = v.Notes
	m.notes = v.Notes
	m.selected = 0
	m.dashIndex = 0
	m.refreshPreview()
	return nil
}

func loadVaultCmd(root string) tea.Cmd {
	return func() tea.Msg {
		v, err := vault.Load(root)
		return reloadMsg{v: v, err: err}
	}
}

func switchVaultCmd(cfg config.Config, root string) tea.Cmd {
	return func() tea.Msg {
		root = filepath.Clean(strings.TrimSpace(root))
		if root == "" {
			return actionMsg{err: fmt.Errorf("vault path is required")}
		}
		info, err := os.Stat(root)
		if err != nil {
			return actionMsg{err: err}
		}
		if !info.IsDir() {
			return actionMsg{err: fmt.Errorf("%s is not a directory", root)}
		}
		cfg.Vault = root
		if cfg.Defaults.StaleDays == 0 {
			cfg.Defaults.StaleDays = 30
		}
		if cfg.Defaults.DashboardLimit == 0 {
			cfg.Defaults.DashboardLimit = 5
		}
		if err := config.Save(cfg); err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{
			status:    "Using vault " + root,
			vaultPath: root,
			cfg:       &cfg,
		}
	}
}

func createVaultCmd(cfg config.Config, root string) tea.Cmd {
	return func() tea.Msg {
		root = filepath.Clean(strings.TrimSpace(root))
		if root == "" {
			return actionMsg{err: fmt.Errorf("new vault path is required")}
		}
		if err := os.MkdirAll(root, 0o755); err != nil {
			return actionMsg{err: err}
		}
		for _, rel := range []string{"daily", "weekly", "monthly", "assets", "templates"} {
			if err := os.MkdirAll(filepath.Join(root, rel), 0o755); err != nil {
				return actionMsg{err: err}
			}
		}
		inboxPath := filepath.Join(root, "inbox.md")
		if _, err := os.Stat(inboxPath); os.IsNotExist(err) {
			if err := os.WriteFile(inboxPath, []byte("# Inbox\n\n"), 0o644); err != nil {
				return actionMsg{err: err}
			}
		}
		cfg.Vault = root
		if cfg.Templates == "" {
			cfg.Templates = filepath.Join(root, "templates")
		}
		if cfg.Defaults.StaleDays == 0 {
			cfg.Defaults.StaleDays = 30
		}
		if cfg.Defaults.DashboardLimit == 0 {
			cfg.Defaults.DashboardLimit = 5
		}
		if err := config.Save(cfg); err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{
			status:    "Created and selected vault " + root,
			vaultPath: root,
			cfg:       &cfg,
		}
	}
}

func ensurePeriodicCmd(root string, kind vault.PeriodicKind) tea.Cmd {
	return func() tea.Msg {
		v, err := vault.Load(root)
		if err != nil {
			return actionMsg{err: err}
		}
		note, created, err := v.EnsurePeriodicNote(kind, time.Now())
		if err != nil {
			return actionMsg{err: err}
		}
		verb := "Opened"
		if created {
			verb = "Created"
		}
		return actionMsg{
			status:    fmt.Sprintf("%s %s", verb, note.RelPath),
			vaultPath: root,
			notePath:  note.RelPath,
		}
	}
}

func createNoteCmd(root, title string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(title) == "" {
			return actionMsg{err: fmt.Errorf("title is required")}
		}
		v, err := vault.Load(root)
		if err != nil {
			return actionMsg{err: err}
		}
		_, err = v.CreateNote("", title, map[string]any{
			"title":  title,
			"status": "active",
			"source": "human",
			"date":   time.Now().Format("2006-01-02"),
		}, "# "+title+"\n\n")
		if err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{status: "Note created"}
	}
}

func captureCmd(root, text string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(text) == "" {
			return actionMsg{err: fmt.Errorf("capture text is required")}
		}
		v, err := vault.Load(root)
		if err != nil {
			return actionMsg{err: err}
		}
		if err := v.Capture(text); err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{status: "Captured to inbox.md"}
	}
}

func renameCmd(root, relPath, title string) tea.Cmd {
	return func() tea.Msg {
		v, err := vault.Load(root)
		if err != nil {
			return actionMsg{err: err}
		}
		newPath, warnings, err := v.RenameNote(relPath, title)
		status := "Renamed to " + newPath
		if len(warnings) > 0 {
			status += fmt.Sprintf(" (%d warnings)", len(warnings))
		}
		return actionMsg{status: status, err: err}
	}
}

func publishCmd(root, outputDir string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(outputDir) == "" {
			return actionMsg{err: fmt.Errorf("publish output directory is required")}
		}
		v, err := vault.Load(root)
		if err != nil {
			return actionMsg{err: err}
		}
		warnings, err := v.Publish(vault.PublishOptions{
			OutputDir:           outputDir,
			MarkUnresolved:      true,
			PreserveFrontmatter: true,
		})
		status := fmt.Sprintf("Published %d notes to %s", len(v.Notes), outputDir)
		if len(warnings) > 0 {
			status += fmt.Sprintf(" with %d warnings", len(warnings))
		}
		return actionMsg{status: status, err: err}
	}
}

func truncate(input string, maxLen int) string {
	runes := []rune(input)
	if len(runes) <= maxLen {
		return input
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func titleLabel(input string) string {
	if input == "" {
		return ""
	}
	return strings.ToUpper(input[:1]) + input[1:]
}

func unresolvedLinkCount(links []vault.UnresolvedLink) int {
	total := 0
	for _, link := range links {
		total += link.Count
	}
	return total
}

func inboxEntryCount(note *vault.Note) int {
	if note == nil {
		return 0
	}
	count := 0
	for _, line := range strings.Split(note.Body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			count++
		}
	}
	return count
}

func periodicSummary(v *vault.Vault, kind vault.PeriodicKind, when time.Time) string {
	relPath, _, _, _, err := vaultPeriodicSpec(kind, when)
	if err != nil {
		return "unavailable"
	}
	note := v.NoteByRelPath(relPath)
	if note == nil {
		return "not created yet"
	}
	return note.RelPath
}

func vaultPeriodicSpec(kind vault.PeriodicKind, when time.Time) (string, string, map[string]any, string, error) {
	switch kind {
	case vault.PeriodicDaily:
		stamp := when.In(time.Local).Format("2006-01-02")
		return filepath.ToSlash(filepath.Join("daily", stamp+".md")), "", nil, "", nil
	case vault.PeriodicWeekly:
		year, week := when.In(time.Local).ISOWeek()
		return filepath.ToSlash(filepath.Join("weekly", fmt.Sprintf("%04d-W%02d.md", year, week))), "", nil, "", nil
	case vault.PeriodicMonthly:
		stamp := when.In(time.Local).Format("2006-01")
		return filepath.ToSlash(filepath.Join("monthly", stamp+".md")), "", nil, "", nil
	default:
		return "", "", nil, "", fmt.Errorf("unsupported periodic note kind: %s", kind)
	}
}
