package vault

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	wikilinkPattern = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
)

type Vault struct {
	Root        string
	Notes       []*Note
	ParseErrors []ParseError
}

type Note struct {
	Path       string
	RelPath    string
	Stem       string
	Title      string
	HasTitle   bool
	Aliases    []string
	Tags       []string
	Status     string
	Body       string
	Metadata   map[string]any
	ModifiedAt time.Time
}

type ResolveResult struct {
	Note     *Note
	Warning  string
	Matched  string
	Resolved bool
}

type Conflict struct {
	Name       string
	Matched    string
	Winner     *Note
	Candidates []*Note
}

type UnresolvedLink struct {
	Source *Note
	Target string
	Count  int
}

// ParseError records a frontmatter YAML parse failure for a specific note.
// The note is still indexed with empty metadata; only its frontmatter fields
// (title, aliases, tags, status) are unavailable.
type ParseError struct {
	RelPath string
	Err     error
}

type PublishOptions struct {
	OutputDir           string
	SkipDrafts          bool
	MarkUnresolved      bool
	PreserveFrontmatter bool
}

type PeriodicKind string

const (
	PeriodicDaily   PeriodicKind = "daily"
	PeriodicWeekly  PeriodicKind = "weekly"
	PeriodicMonthly PeriodicKind = "monthly"
)

func Load(root string) (*Vault, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	v := &Vault{Root: root}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			note, parseErr, err := parseNote(root, path)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			v.Notes = append(v.Notes, note)
			if parseErr != nil {
				v.ParseErrors = append(v.ParseErrors, *parseErr)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(v.Notes, func(i, j int) bool {
		if v.Notes[i].ModifiedAt.Equal(v.Notes[j].ModifiedAt) {
			return v.Notes[i].RelPath < v.Notes[j].RelPath
		}
		return v.Notes[i].ModifiedAt.After(v.Notes[j].ModifiedAt)
	})

	return v, nil
}

func parseNote(root, path string) (*Note, *ParseError, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	metadata := map[string]any{}
	body := string(raw)
	var parseErr *ParseError
	if fm, content, ok := splitFrontmatter(raw); ok {
		if err := yaml.Unmarshal(fm, &metadata); err != nil {
			// Soft error: index the note with empty metadata so it remains
			// visible; surface the parse failure through ParseError.
			metadata = map[string]any{}
			body = string(content)
			parseErr = &ParseError{Err: err}
		} else {
			body = string(content)
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}

	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return nil, nil, err
	}
	relPath = filepath.ToSlash(relPath)
	stem := strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(relPath))
	title := stem
	hasTitle := false
	if value, ok := metadata["title"].(string); ok && strings.TrimSpace(value) != "" {
		title = strings.TrimSpace(value)
		hasTitle = true
	}

	var aliases []string
	if rawAliases, ok := metadata["aliases"]; ok {
		aliases = stringSlice(rawAliases)
	}

	tags := mergeTags(stringSlice(metadata["tags"]), ExtractInlineTags(body))
	status, _ := metadata["status"].(string)

	note := &Note{
		Path:       path,
		RelPath:    relPath,
		Stem:       stem,
		Title:      title,
		HasTitle:   hasTitle,
		Aliases:    aliases,
		Tags:       tags,
		Status:     strings.TrimSpace(status),
		Body:       body,
		Metadata:   metadata,
		ModifiedAt: info.ModTime(),
	}
	if parseErr != nil {
		parseErr.RelPath = relPath
	}
	return note, parseErr, nil
}

func splitFrontmatter(raw []byte) ([]byte, []byte, bool) {
	if !bytes.HasPrefix(raw, []byte("---\n")) && !bytes.HasPrefix(raw, []byte("---\r\n")) {
		return nil, raw, false
	}
	content := strings.ReplaceAll(string(raw), "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return nil, raw, false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			fm := strings.Join(lines[1:i], "\n")
			body := strings.Join(lines[i+1:], "\n")
			return []byte(fm), []byte(body), true
		}
	}
	return nil, raw, false
}

func ExtractInlineTags(body string) []string {
	lines := strings.Split(body, "\n")
	inFence := false
	inHTML := false
	seen := map[string]struct{}{}
	var tags []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") && !strings.HasPrefix(trimmed, "</") {
			inHTML = true
		}
		if inHTML {
			if strings.HasPrefix(trimmed, "</") {
				inHTML = false
			}
			continue
		}

		scan := []rune(stripInlineCode(line))
		for i := 0; i < len(scan); i++ {
			if scan[i] != '#' {
				continue
			}
			if i > 0 && !isWhitespace(scan[i-1]) {
				continue
			}
			j := i + 1
			for j < len(scan) && isTagRune(scan[j]) {
				j++
			}
			if j == i+1 {
				continue
			}
			tag := strings.TrimRight(string(scan[i+1:j]), ".,;:!?")
			if tag == "" || strings.HasPrefix(tag, "/") || strings.HasSuffix(tag, "/") {
				continue
			}
			key := strings.ToLower(tag)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			tags = append(tags, tag)
		}
	}

	sort.Strings(tags)
	return tags
}

func stripInlineCode(line string) string {
	var out strings.Builder
	for i := 0; i < len(line); {
		if line[i] != '`' || isEscapedBacktick(line, i) {
			out.WriteByte(line[i])
			i++
			continue
		}

		runLen := 1
		for i+runLen < len(line) && line[i+runLen] == '`' {
			runLen++
		}
		closeIdx := findClosingBacktickRun(line, i+runLen, runLen)
		if closeIdx == -1 {
			out.WriteString(line[i : i+runLen])
			i += runLen
			continue
		}
		i = closeIdx + runLen
	}
	return out.String()
}

func isEscapedBacktick(line string, idx int) bool {
	backslashes := 0
	for i := idx - 1; i >= 0 && line[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func findClosingBacktickRun(line string, start, runLen int) int {
	for i := start; i < len(line); i++ {
		if line[i] != '`' || isEscapedBacktick(line, i) {
			continue
		}
		match := true
		for j := 1; j < runLen; j++ {
			if i+j >= len(line) || line[i+j] != '`' {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func (v *Vault) Resolve(target string) ResolveResult {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return ResolveResult{}
	}

	if note, warning := chooseMatch("title", target, v.Notes, func(n *Note) []string {
		if !n.HasTitle {
			return nil
		}
		return []string{n.Title}
	}); note != nil {
		return ResolveResult{Note: note, Warning: warning, Matched: "title", Resolved: true}
	}
	if note, warning := chooseMatch("alias", target, v.Notes, func(n *Note) []string { return n.Aliases }); note != nil {
		return ResolveResult{Note: note, Warning: warning, Matched: "alias", Resolved: true}
	}
	if note, warning := chooseMatch("filename", target, v.Notes, func(n *Note) []string { return []string{n.Stem} }); note != nil {
		return ResolveResult{Note: note, Warning: warning, Matched: "filename", Resolved: true}
	}
	return ResolveResult{}
}

func chooseMatch(kind, target string, notes []*Note, names func(*Note) []string) (*Note, string) {
	var matches []*Note
	for _, note := range notes {
		for _, candidate := range names(note) {
			if strings.EqualFold(strings.TrimSpace(candidate), target) {
				matches = append(matches, note)
				break
			}
		}
	}
	if len(matches) == 0 {
		return nil, ""
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].ModifiedAt.Equal(matches[j].ModifiedAt) {
			return matches[i].RelPath < matches[j].RelPath
		}
		return matches[i].ModifiedAt.After(matches[j].ModifiedAt)
	})
	if len(matches) == 1 {
		return matches[0], ""
	}
	return matches[0], fmt.Sprintf("%s conflict for %q, resolved to most recently modified note: %s", kind, target, matches[0].RelPath)
}

func (v *Vault) Conflicts() []Conflict {
	var conflicts []Conflict
	conflicts = append(conflicts, collectConflicts("title", v.Notes, func(n *Note) []string {
		if !n.HasTitle {
			return nil
		}
		return []string{n.Title}
	})...)
	conflicts = append(conflicts, collectConflicts("alias", v.Notes, func(n *Note) []string { return n.Aliases })...)
	conflicts = append(conflicts, collectConflicts("filename", v.Notes, func(n *Note) []string { return []string{n.Stem} })...)
	sort.Slice(conflicts, func(i, j int) bool {
		if conflicts[i].Matched == conflicts[j].Matched {
			return conflicts[i].Name < conflicts[j].Name
		}
		return conflicts[i].Matched < conflicts[j].Matched
	})
	return conflicts
}

func (v *Vault) UnresolvedLinks() []UnresolvedLink {
	type unresolvedKey struct {
		source string
		target string
	}

	notesByPath := map[string]*Note{}
	counts := map[unresolvedKey]int{}
	display := map[unresolvedKey]string{}

	for _, note := range v.Notes {
		notesByPath[note.RelPath] = note
		eachWikilinkOutsideCode(note.Body, func(target, _ string) {
			target = strings.TrimSpace(target)
			if target == "" {
				return
			}
			if v.Resolve(target).Resolved {
				return
			}
			key := unresolvedKey{
				source: note.RelPath,
				target: strings.ToLower(target),
			}
			counts[key]++
			if _, ok := display[key]; !ok {
				display[key] = target
			}
		})
	}

	var unresolved []UnresolvedLink
	for key, count := range counts {
		unresolved = append(unresolved, UnresolvedLink{
			Source: notesByPath[key.source],
			Target: display[key],
			Count:  count,
		})
	}
	sort.Slice(unresolved, func(i, j int) bool {
		if unresolved[i].Source.RelPath == unresolved[j].Source.RelPath {
			return strings.ToLower(unresolved[i].Target) < strings.ToLower(unresolved[j].Target)
		}
		return unresolved[i].Source.RelPath < unresolved[j].Source.RelPath
	})
	return unresolved
}

func collectConflicts(kind string, notes []*Note, names func(*Note) []string) []Conflict {
	buckets := map[string][]*Note{}
	display := map[string]string{}
	for _, note := range notes {
		for _, candidate := range names(note) {
			candidate = strings.TrimSpace(candidate)
			if candidate == "" {
				continue
			}
			key := strings.ToLower(candidate)
			buckets[key] = appendUniqueNote(buckets[key], note)
			if _, ok := display[key]; !ok {
				display[key] = candidate
			}
		}
	}

	var conflicts []Conflict
	for key, candidates := range buckets {
		if len(candidates) < 2 {
			continue
		}
		sortNotesByRecency(candidates)
		conflicts = append(conflicts, Conflict{
			Name:       display[key],
			Matched:    kind,
			Winner:     candidates[0],
			Candidates: append([]*Note(nil), candidates...),
		})
	}
	return conflicts
}

func (v *Vault) Publish(opts PublishOptions) ([]string, error) {
	if opts.OutputDir == "" {
		return nil, errors.New("publish output directory is required")
	}
	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return nil, err
	}

	var warnings []string
	for _, note := range v.Notes {
		if opts.SkipDrafts && strings.EqualFold(note.Status, "draft") {
			continue
		}
		dst := filepath.Join(opts.OutputDir, filepath.FromSlash(note.RelPath))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return warnings, err
		}
		rendered, renderedWarnings := v.renderPublishedNote(note, dst, opts)
		warnings = append(warnings, renderedWarnings...)
		if err := os.WriteFile(dst, []byte(rendered), 0o644); err != nil {
			return warnings, err
		}
	}
	return warnings, nil
}

func (v *Vault) renderPublishedNote(note *Note, dst string, opts PublishOptions) (string, []string) {
	var warnings []string
	body := replaceWikilinksOutsideCode(note.Body, func(match string) string {
		parts := wikilinkPattern.FindStringSubmatch(match)
		target := strings.TrimSpace(parts[1])
		display := target
		if strings.TrimSpace(parts[2]) != "" {
			display = strings.TrimSpace(parts[2])
		}
		resolved := v.Resolve(target)
		if !resolved.Resolved {
			warnings = append(warnings, fmt.Sprintf("unresolved link %q in %s", target, note.RelPath))
			if opts.MarkUnresolved {
				return display + " [unresolved]"
			}
			return display
		}
		if resolved.Warning != "" {
			warnings = append(warnings, resolved.Warning)
		}
		targetPath, err := filepath.Rel(filepath.Dir(dst), filepath.Join(opts.OutputDir, filepath.FromSlash(resolved.Note.RelPath)))
		if err != nil {
			warnings = append(warnings, err.Error())
			return display
		}
		return fmt.Sprintf("[%s](%s)", display, filepath.ToSlash(targetPath))
	})

	if !opts.PreserveFrontmatter || len(note.Metadata) == 0 {
		return body, warnings
	}
	return encodeFrontmatter(note.Metadata) + body, warnings
}

func eachWikilinkOutsideCode(body string, visit func(target, display string)) {
	lines := strings.Split(body, "\n")
	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		eachInlineWikilink(line, visit)
	}
}

func replaceWikilinksOutsideCode(body string, replacer func(string) string) string {
	lines := strings.Split(body, "\n")
	inFence := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		lines[i] = replaceInlineWikilinks(line, replacer)
	}
	return strings.Join(lines, "\n")
}

func eachInlineWikilink(line string, visit func(target, display string)) {
	inCode := false
	start := 0
	for i, r := range line {
		if r != '`' {
			continue
		}
		if !inCode {
			scanInlineWikilinks(line[start:i], visit)
		}
		inCode = !inCode
		start = i + 1
	}
	if !inCode {
		scanInlineWikilinks(line[start:], visit)
	}
}

func scanInlineWikilinks(segment string, visit func(target, display string)) {
	matches := wikilinkPattern.FindAllStringSubmatch(segment, -1)
	for _, match := range matches {
		target := ""
		display := ""
		if len(match) > 1 {
			target = match[1]
		}
		if len(match) > 2 {
			display = match[2]
		}
		visit(target, display)
	}
}

func replaceInlineWikilinks(line string, replacer func(string) string) string {
	var out strings.Builder
	inCode := false
	start := 0
	for i, r := range line {
		if r != '`' {
			continue
		}
		segment := line[start:i]
		if inCode {
			out.WriteString(segment)
			out.WriteRune('`')
		} else {
			out.WriteString(wikilinkPattern.ReplaceAllStringFunc(segment, replacer))
			out.WriteRune('`')
		}
		inCode = !inCode
		start = i + 1
	}
	remainder := line[start:]
	if inCode {
		out.WriteString(remainder)
	} else {
		out.WriteString(wikilinkPattern.ReplaceAllStringFunc(remainder, replacer))
	}
	return out.String()
}

func (v *Vault) CreateNote(relDir, title string, metadata map[string]any, body string) (*Note, error) {
	slug := Slugify(title)
	if slug == "" {
		return nil, errors.New("title is required")
	}
	relPath := filepath.ToSlash(filepath.Join(relDir, slug+".md"))
	fullPath := filepath.Join(v.Root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, err
	}
	if _, err := os.Stat(fullPath); err == nil {
		return nil, fmt.Errorf("%s already exists", relPath)
	}

	if metadata == nil {
		metadata = map[string]any{}
	}
	if _, ok := metadata["title"]; !ok {
		metadata["title"] = title
	}
	rendered := encodeFrontmatter(metadata) + strings.TrimLeft(body, "\n")
	if err := os.WriteFile(fullPath, []byte(rendered), 0o644); err != nil {
		return nil, err
	}
	note, _, err := parseNote(v.Root, fullPath)
	return note, err
}


func (v *Vault) EnsurePeriodicNote(kind PeriodicKind, when time.Time) (*Note, bool, error) {
	relPath, metadata, body, err := periodicSpec(kind, when)
	if err != nil {
		return nil, false, err
	}
	if note := v.NoteByRelPath(relPath); note != nil {
		return note, false, nil
	}

	fullPath := filepath.Join(v.Root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, false, err
	}
	rendered := encodeFrontmatter(metadata) + body
	if err := os.WriteFile(fullPath, []byte(rendered), 0o644); err != nil {
		return nil, false, err
	}
	note, _, err := parseNote(v.Root, fullPath)
	if err != nil {
		return nil, false, err
	}
	v.Notes = append(v.Notes, note)
	sortNotesByRecency(v.Notes)
	return note, true, nil
}

// PeriodicRelPath returns the vault-relative path for a periodic note of the
// given kind at the given time. It does not create or read any file.
func PeriodicRelPath(kind PeriodicKind, when time.Time) (string, error) {
	relPath, _, _, err := periodicSpec(kind, when)
	return relPath, err
}

// DraftNotes returns notes whose status field is "draft" (case-insensitive).
func (v *Vault) DraftNotes() []*Note {
	var drafts []*Note
	for _, note := range v.Notes {
		if strings.EqualFold(note.Status, "draft") {
			drafts = append(drafts, note)
		}
	}
	return drafts
}

// StaleNotes returns notes whose modification time is older than days days
// before asOf. Returns nil if days is zero or negative.
func (v *Vault) StaleNotes(days int, asOf time.Time) []*Note {
	if days <= 0 {
		return nil
	}
	cutoff := asOf.AddDate(0, 0, -days)
	var stale []*Note
	for _, note := range v.Notes {
		if note.ModifiedAt.Before(cutoff) {
			stale = append(stale, note)
		}
	}
	return stale
}

func (v *Vault) Capture(text string) error {
	inboxPath := filepath.Join(v.Root, "inbox.md")
	existing, err := os.ReadFile(inboxPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	timestamp := time.Now().Format("2006-01-02 15:04")
	entry := fmt.Sprintf("- [%s] %s\n", timestamp, strings.TrimSpace(text))
	updated := string(existing)
	if updated != "" && !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	updated += entry
	return os.WriteFile(inboxPath, []byte(updated), 0o644)
}

func (v *Vault) RenameNote(relPath, newTitle string) (string, []string, error) {
	note := v.NoteByRelPath(relPath)
	if note == nil {
		return "", nil, fmt.Errorf("note not found: %s", relPath)
	}

	newStem := Slugify(newTitle)
	if newStem == "" {
		return "", nil, errors.New("new title is required")
	}
	newRelPath := filepath.ToSlash(filepath.Join(filepath.Dir(note.RelPath), newStem+".md"))
	newAbsPath := filepath.Join(v.Root, filepath.FromSlash(newRelPath))
	if newRelPath == note.RelPath {
		return "", nil, fmt.Errorf("new title produces the same filename: %s", newRelPath)
	}
	if _, err := os.Stat(newAbsPath); err == nil {
		return "", nil, fmt.Errorf("a note already exists at %s", newRelPath)
	}
	if err := os.MkdirAll(filepath.Dir(newAbsPath), 0o755); err != nil {
		return "", nil, err
	}

	oldTargets := []string{note.Title, note.Stem}
	oldMetadata := cloneMap(note.Metadata)
	if _, ok := oldMetadata["title"]; ok {
		oldMetadata["title"] = newTitle
	}
	content := encodeFrontmatter(oldMetadata) + strings.TrimLeft(note.Body, "\n")
	if len(oldMetadata) == 0 {
		content = note.Body
	}
	if err := os.WriteFile(note.Path, []byte(content), 0o644); err != nil {
		return "", nil, err
	}
	if err := os.Rename(note.Path, newAbsPath); err != nil {
		return "", nil, err
	}

	reloaded, err := Load(v.Root)
	if err != nil {
		return "", nil, err
	}
	warnings, err := reloaded.rewriteInboundLinks(oldTargets, newTitle)
	if err != nil {
		return "", warnings, err
	}

	*v = *reloaded
	return newRelPath, warnings, nil
}

func (v *Vault) rewriteInboundLinks(oldTargets []string, newTitle string) ([]string, error) {
	var warnings []string
	for _, note := range v.Notes {
		updated := wikilinkPattern.ReplaceAllStringFunc(note.Body, func(match string) string {
			parts := wikilinkPattern.FindStringSubmatch(match)
			target := strings.TrimSpace(parts[1])
			display := strings.TrimSpace(parts[2])
			for _, oldTarget := range oldTargets {
				if strings.EqualFold(target, oldTarget) {
					if display != "" {
						return fmt.Sprintf("[[%s|%s]]", newTitle, display)
					}
					return fmt.Sprintf("[[%s]]", newTitle)
				}
			}
			return match
		})
		if updated == note.Body {
			continue
		}
		rendered := encodeFrontmatter(note.Metadata) + strings.TrimLeft(updated, "\n")
		if len(note.Metadata) == 0 {
			rendered = updated
		}
		if err := os.WriteFile(note.Path, []byte(rendered), 0o644); err != nil {
			warnings = append(warnings, err.Error())
		}
	}
	return warnings, nil
}

func (v *Vault) NoteByRelPath(relPath string) *Note {
	normalized := filepath.ToSlash(relPath)
	for _, note := range v.Notes {
		if note.RelPath == normalized {
			return note
		}
	}
	return nil
}

func encodeFrontmatter(metadata map[string]any) string {
	if len(metadata) == 0 {
		return ""
	}
	raw, err := yaml.Marshal(metadata)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("---\n%s---\n\n", string(raw))
}

func Slugify(input string) string {
	input = strings.ReplaceAll(input, "/", "-")
	input = strings.ReplaceAll(input, "\\", "-")
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return ""
	}
	var out strings.Builder
	lastDash := false
	for _, r := range input {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '/':
			if !lastDash && out.Len() > 0 {
				out.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(out.String(), "-")
}

func mergeTags(sources ...[]string) []string {
	seen := map[string]string{}
	for _, source := range sources {
		for _, tag := range source {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			key := strings.ToLower(tag)
			if _, ok := seen[key]; !ok {
				seen[key] = tag
			}
		}
	}
	var tags []string
	for _, tag := range seen {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		var items []string
		for _, item := range typed {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				items = append(items, strings.TrimSpace(s))
			}
		}
		return items
	default:
		return nil
	}
}

func appendUniqueNote(notes []*Note, candidate *Note) []*Note {
	for _, note := range notes {
		if note.RelPath == candidate.RelPath {
			return notes
		}
	}
	return append(notes, candidate)
}

func sortNotesByRecency(notes []*Note) {
	sort.Slice(notes, func(i, j int) bool {
		if notes[i].ModifiedAt.Equal(notes[j].ModifiedAt) {
			return notes[i].RelPath < notes[j].RelPath
		}
		return notes[i].ModifiedAt.After(notes[j].ModifiedAt)
	})
}

func cloneMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func isTagRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r == '_' || r == '-' || r == '/':
		return true
	default:
		return false
	}
}

func periodicSpec(kind PeriodicKind, when time.Time) (string, map[string]any, string, error) {
	when = when.In(time.Local)
	switch kind {
	case PeriodicDaily:
		stamp := when.Format("2006-01-02")
		return filepath.ToSlash(filepath.Join("daily", stamp+".md")), map[string]any{
			"title":  stamp,
			"date":   stamp,
			"status": "active",
			"source": "human",
			"tags":   []string{"daily"},
		}, "# " + stamp + "\n\n", nil
	case PeriodicWeekly:
		year, week := when.ISOWeek()
		stamp := fmt.Sprintf("%04d-W%02d", year, week)
		return filepath.ToSlash(filepath.Join("weekly", stamp+".md")), map[string]any{
			"title":  stamp,
			"date":   when.Format("2006-01-02"),
			"status": "active",
			"source": "human",
			"tags":   []string{"weekly"},
		}, "# " + stamp + "\n\n", nil
	case PeriodicMonthly:
		stamp := when.Format("2006-01")
		return filepath.ToSlash(filepath.Join("monthly", stamp+".md")), map[string]any{
			"title":  stamp,
			"date":   when.Format("2006-01-02"),
			"status": "active",
			"source": "human",
			"tags":   []string{"monthly"},
		}, "# " + stamp + "\n\n", nil
	default:
		return "", nil, "", fmt.Errorf("unsupported periodic note kind: %s", kind)
	}
}
