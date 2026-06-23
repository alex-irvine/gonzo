package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderDetailModal renders the log details modal as a single full-width pane
// containing a small header and a collapsible JSON tree of the entry.
func (m *DashboardModel) renderDetailModal() string {
	// Calculate dimensions
	modalWidth := m.width - 8   // 4 chars margin on each side
	modalHeight := m.height - 6 // 3 lines margin top and bottom

	// Account for borders and headers
	contentWidth := modalWidth - 4   // Modal borders
	contentHeight := modalHeight - 6 // Header + status

	// Content area inside the pane border.
	contentAreaWidth := contentWidth - 2
	if contentAreaWidth < 10 {
		contentAreaWidth = 10
	}

	// Header: timestamps + severity.
	headerBlock := m.renderDetailHeader(*m.currentLogEntry)
	headerLines := strings.Count(headerBlock, "\n") + 1

	treeHeight := contentHeight - headerLines - 1 // -1 for the blank separator line
	if treeHeight < 1 {
		treeHeight = 1
	}

	// Render the JSON tree and keep the cursor visible in the viewport.
	treeContent, cursorLine := m.renderJSONTree(contentAreaWidth)

	m.infoViewport.Width = contentAreaWidth
	m.infoViewport.Height = treeHeight
	m.infoViewport.SetContent(treeContent)
	m.ensureCursorVisible(cursorLine, treeHeight)

	body := lipgloss.JoinVertical(lipgloss.Left, headerBlock, "", m.infoViewport.View())

	// Pane with border.
	pane := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorBlue).
		Render(body)

	// Header bar with AI status on the right.
	headerTitle := "Log Details"
	var aiStatus string
	if m.aiConfigured {
		aiStatus = fmt.Sprintf("%s: %s", m.aiServiceName, m.aiModelName)
	} else {
		aiStatus = "AI Not Available"
	}

	headerLeft := lipgloss.NewStyle().Foreground(ColorBlue).Bold(true).Render(headerTitle)
	headerRight := lipgloss.NewStyle().
		Foreground(func() lipgloss.Color {
			if m.aiConfigured {
				return ColorGreen
			}
			return ColorOrange
		}()).
		Render(aiStatus)

	headerSpacing := contentWidth - len(headerTitle) - len(aiStatus)
	if headerSpacing < 1 {
		headerSpacing = 1
	}
	headerBar := lipgloss.JoinHorizontal(lipgloss.Top,
		headerLeft,
		strings.Repeat(" ", headerSpacing),
		headerRight,
	)

	statusBar := m.renderModalStatusBar()

	modal := lipgloss.JoinVertical(lipgloss.Left, headerBar, pane, statusBar)

	finalModal := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBlue).
		Render(modal)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, finalModal)
}

// renderDetailHeader renders the fixed header block (timestamps, severity) shown
// above the JSON tree.
func (m *DashboardModel) renderDetailHeader(entry LogEntry) string {
	labelStyle := lipgloss.NewStyle().Foreground(ColorGray).Width(10)
	valueStyle := lipgloss.NewStyle().Foreground(ColorWhite)

	severityStyle := lipgloss.NewStyle().Bold(true).Foreground(getSeverityColor(normalizeSeverityLevel(entry.Severity)))

	var b strings.Builder
	b.WriteString(labelStyle.Render("Received:") + " " +
		valueStyle.Render(entry.Timestamp.Format("2006-01-02 15:04:05.000")))
	if !entry.OrigTimestamp.IsZero() {
		b.WriteString("\n" + labelStyle.Render("Log Time:") + " " +
			valueStyle.Render(entry.OrigTimestamp.Format("2006-01-02 15:04:05.000")))
	}
	b.WriteString("\n" + labelStyle.Render("Severity:") + " " + severityStyle.Render(entry.Severity))

	// Surface AI analysis inline when present.
	if m.aiAnalyzing {
		spinner := fmt.Sprintf("%s Analyzing...", m.getSpinner())
		b.WriteString("\n" + labelStyle.Render("AI:") + " " +
			lipgloss.NewStyle().Foreground(ColorYellow).Render(spinner))
	} else if m.aiAnalysisResult != "" && m.aiAnalysisResult != "Analyzing..." {
		b.WriteString("\n" + labelStyle.Render("AI:") + " " +
			valueStyle.Render(m.aiAnalysisResult))
	}

	return b.String()
}

// ensureCursorVisible scrolls infoViewport so the given cursor line stays within
// the visible window of the given height.
func (m *DashboardModel) ensureCursorVisible(cursorLine, height int) {
	offset := m.infoViewport.YOffset
	if cursorLine < offset {
		m.infoViewport.SetYOffset(cursorLine)
	} else if cursorLine >= offset+height {
		m.infoViewport.SetYOffset(cursorLine - height + 1)
	}
}
