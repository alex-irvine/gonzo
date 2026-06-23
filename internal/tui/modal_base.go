package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderModalOverlay renders modal using full screen with padding
func (m *DashboardModel) renderModalOverlay() string {
	// Check for patterns modal first
	if m.showPatternsModal {
		return m.renderPatternsModal()
	}

	// Check for counts modal
	if m.showCountsModal {
		return m.renderCountsModal()
	}

	// Check for model selection modal
	if m.showModelSelectionModal {
		return m.renderModelSelectionModal()
	}

	// Check for severity filter modal
	if m.showSeverityFilterModal {
		return m.renderSeverityFilterModal()
	}

	// Check if this is a log details modal
	isLogDetailsModal := m.currentLogEntry != nil

	if isLogDetailsModal {
		return m.renderDetailModal()
	} else {
		return m.renderSingleModal()
	}
}

// renderSingleModal renders non-log-details modal with simple single layout
func (m *DashboardModel) renderSingleModal() string {
	// Calculate dimensions
	modalWidth := m.width - 8   // 4 chars margin on each side
	modalHeight := m.height - 6 // 3 lines margin top and bottom

	// Account for borders and headers
	contentWidth := modalWidth - 4   // Modal borders
	contentHeight := modalHeight - 4 // Header + status

	// Update viewport
	m.infoViewport.Width = contentWidth
	m.infoViewport.Height = contentHeight
	m.infoViewport.SetContent(m.modalContent)

	// Create content pane
	contentPane := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorGray).
		Render(m.infoViewport.View())

	// Header
	header := lipgloss.NewStyle().
		Width(contentWidth).
		Foreground(ColorBlue).
		Bold(true).
		Render("Top Values")

	// Status bar
	statusBar := m.renderModalStatusBar()

	// Combine all parts
	modal := lipgloss.JoinVertical(lipgloss.Left, header, contentPane, statusBar)

	// Add outer border and center
	finalModal := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBlue).
		Render(modal)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, finalModal)
}

// renderModalStatusBar renders the status bar for modals
func (m *DashboardModel) renderModalStatusBar() string {
	var statusItems []string

	if m.currentLogEntry != nil {
		// Log details (JSON tree) help text
		statusItems = append(statusItems, "↑↓: Move", "←/→/Enter: Collapse/Expand")
		if m.aiClient != nil {
			statusItems = append(statusItems, "i: AI Analysis")
		}
		statusItems = append(statusItems, "y: Yank entry", "Y: Yank node")
	} else {
		// Single modal help text (like Top Values modal)
		statusItems = append(statusItems, "↑↓/Wheel: Scroll", "PgUp/PgDn: Page")
	}

	// Always show close option
	statusItems = append(statusItems, "ESC: Close")

	// Format status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(ColorGray)

	bar := statusStyle.Render(strings.Join(statusItems, " • "))

	// Surface transient yank feedback alongside the help text.
	if m.yankFeedback != "" {
		feedback := lipgloss.NewStyle().Foreground(ColorGreen).Bold(true).Render(m.yankFeedback)
		bar = feedback + statusStyle.Render(" • ") + bar
	}
	return bar
}

// getSeverityColor returns the appropriate color for a severity level
func getSeverityColor(severity string) lipgloss.Color {
	switch severity {
	case "FATAL", "CRITICAL":
		return ColorRed
	case "ERROR":
		return ColorRed
	case "WARN":
		return ColorOrange
	case "INFO":
		return ColorBlue
	case "DEBUG", "TRACE":
		return ColorGray
	default:
		return ColorWhite
	}
}
