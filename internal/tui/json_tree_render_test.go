package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// stripANSI removes lipgloss styling so assertions can match plain text.
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestRenderJSONTree_ContainsKeysAndMarkers(t *testing.T) {
	lipgloss.SetColorProfile(0) // force no-color so output is plain-ish

	m := &DashboardModel{}
	m.buildJSONTree(LogEntry{RawLine: `{"level":"info","labels":{"ns":"default"},"value":65}`})

	out, cursor := m.renderJSONTree(80)
	if cursor != 0 {
		t.Errorf("cursor = %d, want 0", cursor)
	}
	plain := stripANSI(out)

	for _, want := range []string{"level", `"info"`, "labels", "ns", `"default"`, "value", "65"} {
		if !strings.Contains(plain, want) {
			t.Errorf("rendered tree missing %q\n---\n%s", want, plain)
		}
	}
	// labels is a branch and should carry an expand marker.
	if !strings.Contains(plain, "▼") {
		t.Errorf("expected an expand marker in:\n%s", plain)
	}
}

func TestRenderJSONTree_CollapsedHidesChildren(t *testing.T) {
	lipgloss.SetColorProfile(0)

	m := &DashboardModel{}
	m.buildJSONTree(LogEntry{RawLine: `{"labels":{"ns":"default"},"z":1}`})
	// Cursor on "labels" (index 0); collapse it.
	m.jsonToggle()

	out, _ := m.renderJSONTree(80)
	plain := stripANSI(out)
	if strings.Contains(plain, "default") {
		t.Errorf("collapsed branch should hide child value:\n%s", plain)
	}
	if !strings.Contains(plain, "▶") {
		t.Errorf("collapsed branch should show ▶ marker:\n%s", plain)
	}
	if !strings.Contains(plain, "z") {
		t.Errorf("sibling key should remain visible:\n%s", plain)
	}
}

func TestRenderJSONTree_Empty(t *testing.T) {
	m := &DashboardModel{}
	m.buildJSONTree(LogEntry{})
	out, _ := m.renderJSONTree(80)
	if !strings.Contains(stripANSI(out), "No structured data") {
		t.Errorf("expected empty placeholder, got %q", out)
	}
}

func TestRenderDetailHeader(t *testing.T) {
	lipgloss.SetColorProfile(0)

	m := &DashboardModel{}
	out := stripANSI(m.renderDetailHeader(LogEntry{Severity: "ERROR", Message: "boom"}))
	if !strings.Contains(out, "Received:") || !strings.Contains(out, "Severity:") {
		t.Errorf("header missing labels:\n%s", out)
	}
	if !strings.Contains(out, "ERROR") {
		t.Errorf("header missing severity value:\n%s", out)
	}
}

func TestEnsureCursorVisible(t *testing.T) {
	m := &DashboardModel{}
	m.infoViewport = viewport.New(80, 5)
	m.infoViewport.SetContent(strings.Repeat("line\n", 50))

	// Cursor below the window scrolls down.
	m.ensureCursorVisible(20, 5)
	if got := m.infoViewport.YOffset; got != 16 {
		t.Errorf("YOffset = %d, want 16", got)
	}
	// Cursor above the window scrolls up.
	m.ensureCursorVisible(3, 5)
	if got := m.infoViewport.YOffset; got != 3 {
		t.Errorf("YOffset = %d, want 3", got)
	}
}
