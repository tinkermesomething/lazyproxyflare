package caddy

import (
	"fmt"
	"strings"
)

// isBannerComment reports whether a comment line is a section divider such as
// "# =====" rather than a comment that belongs to the block below it.
func isBannerComment(trimmed string) bool {
	return strings.HasPrefix(trimmed, "# ===") || strings.HasPrefix(trimmed, "#===") ||
		strings.HasPrefix(trimmed, "# ---") || strings.HasPrefix(trimmed, "#---")
}

// commentStart returns the index of the first line of the comment block that
// sits directly above lines[i], or i when there is none. Banner dividers and
// blank lines end the block.
func commentStart(lines []string, i int) int {
	start := i
	for start > 0 {
		t := strings.TrimSpace(lines[start-1])
		if !strings.HasPrefix(t, "#") || isBannerComment(t) {
			break
		}
		start--
	}
	return start
}

// leadingComment returns the comment text directly above lines[i], joined by
// spaces and without the "#" markers.
func leadingComment(lines []string, i int) string {
	var parts []string
	for _, l := range lines[commentStart(lines, i):i] {
		parts = append(parts, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "#")))
	}
	return strings.Join(parts, " ")
}

// FindSnippet parses src and returns the snippet called name.
func FindSnippet(src, name string) (*Snippet, error) {
	var found *Snippet
	for _, s := range ParseCaddyfileWithSnippets(src).Snippets {
		if s.Name == name {
			if found != nil {
				return nil, fmt.Errorf("snippet %q is defined more than once", name)
			}
			s := s
			found = &s
		}
	}
	if found == nil {
		return nil, fmt.Errorf("snippet %q not found in the Caddyfile", name)
	}
	return found, nil
}

// braceBalance returns opens minus closes, ignoring braces inside quotes.
func braceBalance(s string) int {
	n := 0
	for _, line := range strings.Split(s, "\n") {
		open, close := countBracesOutsideQuotes(line)
		n += open - close
	}
	return n
}

// ReplaceSnippetBody returns src with the body of snippet name replaced.
// expectedOld is the body the caller last showed the user: when the file has
// changed since (a hand edit, another tool), nothing is replaced, so the edit
// cannot overwrite changes the user never saw. The snippet is located in src
// itself, never by line numbers from an earlier parse.
func ReplaceSnippetBody(src, name, expectedOld, newBody string) (string, error) {
	s, err := FindSnippet(src, name)
	if err != nil {
		return "", err
	}
	if s.Content != expectedOld {
		return "", fmt.Errorf("snippet %q changed in the Caddyfile since it was opened; reopen it", name)
	}
	if braceBalance(newBody) != 0 {
		return "", fmt.Errorf("unbalanced braces in the new snippet body")
	}
	body := strings.TrimRight(newBody, "\n \t")
	lines := strings.Split(src, "\n")
	out := append([]string{}, lines[:s.LineStart]...)
	if body != "" {
		out = append(out, strings.Split(body, "\n")...)
	}
	out = append(out, lines[s.LineEnd-1:]...)
	return strings.Join(out, "\n"), nil
}

// RemoveSnippetBlock returns src without snippet name: the block, the comment
// lines directly above it, and one of the blank lines around it so no double
// gap is left behind. Section banners are kept.
func RemoveSnippetBlock(src, name string) (string, error) {
	s, err := FindSnippet(src, name)
	if err != nil {
		return "", err
	}
	lines := strings.Split(src, "\n")
	start := commentStart(lines, s.LineStart-1)
	end := s.LineEnd // exclusive index after the closing brace
	beforeBlank := start == 0 || strings.TrimSpace(lines[start-1]) == ""
	if beforeBlank && end < len(lines) && strings.TrimSpace(lines[end]) == "" {
		end++
	}
	out := append(append([]string{}, lines[:start]...), lines[end:]...)
	return strings.Join(dropEmptyCreatorBanner(out), "\n"), nil
}

const (
	creatorBannerStart = "# === Snippets created by LazyProxyFlare ==="
	creatorBannerEnd   = "# === End of snippets ==="
)

// dropEmptyCreatorBanner removes the banner the snippet wizard writes when
// nothing is left between its start and end lines, plus the blank line after it.
func dropEmptyCreatorBanner(lines []string) []string {
	for i, l := range lines {
		if strings.TrimSpace(l) != creatorBannerStart {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			t := strings.TrimSpace(lines[j])
			if t == creatorBannerEnd {
				end := j + 1
				if end < len(lines) && strings.TrimSpace(lines[end]) == "" {
					end++
				}
				return append(append([]string{}, lines[:i]...), lines[end:]...)
			}
			if t != "" && !strings.HasPrefix(t, "# Generated:") {
				return lines // something still lives in the banner
			}
		}
		return lines
	}
	return lines
}
