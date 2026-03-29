package ui

import (
	"fmt"
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
	modeHelp
	modeFilter
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
	status string
	err    error
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
	v, err := vault.Load(vaultPath)
	if err != nil {
		return Model{}, err
	}

	input := textinput.New()
	input.CharLimit = 200

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
		vault:     v,
		allNotes:  v.Notes,
		notes:     v.Notes,
		viewport:  viewport,
		input:     input,
		mode:      modeDashboard,
		status:    fmt.Sprintf("Loaded %d notes from %s", len(v.Notes), vaultPath),
		style:     style,
	}
	m.refreshPreview()
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
		m.mode = modeBrowse
		m.input.Blur()
		if msg.err != nil {
			m.err = msg.err
			m.status = msg.err.Error()
			return m, nil
		}
		m.err = nil
		m.status = msg.status
		return m, loadVaultCmd(m.vaultPath)
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
	if m.mode == modeHelp {
		return m.renderHelp()
	}
	if m.mode == modeDashboard {
		return m.renderDashboard()
	}

	leftWidth := min(38, max(28, m.width/3))
	rightWidth := max(20, m.width-leftWidth-5)

	header := m.style.title.Render("YANP TUI") + "\n" +
		m.style.subtle.Render("j/k move  / filter  g dashboard  n new  c capture  R rename  p publish  r refresh  ? help  q quit")

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
	case modeHelp:
		return "Help"
	case modeFilter:
		return "Filter note list"
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
		"  g           Return to the dashboard",
		"  n           Create a new note in the vault root",
		"  c           Capture a quick entry into inbox.md",
		"  R           Rename the selected note and rewrite inbound wikilinks",
		"  p           Publish the vault to a separate CommonMark directory",
		"  r           Refresh the vault index",
		"  ? or h      Open this help screen",
		"  esc         Close help or cancel a prompt",
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
		m.style.subtle.Render("j/k choose  enter open  / filter  n new  c capture  r refresh  ? help  q quit")

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
	inbox := "missing"
	if m.vault.NoteByRelPath("inbox.md") != nil {
		inbox = "present"
	}

	sections := []string{
		m.style.title.Render("Overview"),
		"",
		fmt.Sprintf("Vault: %s", m.vaultPath),
		fmt.Sprintf("Total notes: %d", total),
		fmt.Sprintf("Visible notes: %d", filtered),
		fmt.Sprintf("Inbox: %s", inbox),
	}

	filter := m.currentFilter()
	if filter != "" {
		sections = append(sections, fmt.Sprintf("Filter: %s", filter))
	}

	sections = append(sections,
		"",
		m.style.title.Render("Next actions"),
		"",
		"Use j/k to choose a dashboard item.",
		"Press enter to open the selected target.",
		"Press / to search and narrow the list.",
		"Press n to create a new note.",
		"Press c to capture directly into inbox.md.",
	)

	return m.style.panel.Width(max(36, m.width/2-4)).Height(max(12, m.height-8)).Render(strings.Join(sections, "\n"))
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
	if note := m.vault.NoteByRelPath("inbox.md"); note != nil {
		items = append(items, dashboardItem{
			label:   "Open inbox",
			detail:  note.RelPath,
			relPath: note.RelPath,
			action:  "note",
		})
	}
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
	default:
		return m, nil
	}
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

func loadVaultCmd(root string) tea.Cmd {
	return func() tea.Msg {
		v, err := vault.Load(root)
		return reloadMsg{v: v, err: err}
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
