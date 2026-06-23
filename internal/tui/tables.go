package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alex-irvine/gonzo/internal/memory"

	"github.com/charmbracelet/lipgloss"
)

// formatAttributeValuesModal formats the attribute values modal showing individual values and their counts with full width layout
func (m *DashboardModel) formatAttributeValuesModal(entry *memory.AttributeStatsEntry, maxWidth int) string {
	var modal strings.Builder

	// Title section
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	modal.WriteString(titleStyle.Render(fmt.Sprintf("Attribute Values for \"%s\"", entry.Key)) + "\n\n")

	if len(entry.Values) == 0 {
		helpStyle := lipgloss.NewStyle().Foreground(ColorGray).Italic(true)
		modal.WriteString(helpStyle.Render("No values recorded for this attribute.") + "\n")
		return modal.String()
	}

	// Add summary section at the top
	summaryStyle := lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	modal.WriteString(summaryStyle.Render("Summary:") + "\n")

	summaryDetailStyle := lipgloss.NewStyle().Foreground(ColorWhite)
	modal.WriteString(summaryDetailStyle.Render(fmt.Sprintf("Total occurrences: %d", entry.TotalCount)) + "\n")
	modal.WriteString(summaryDetailStyle.Render(fmt.Sprintf("Unique values: %d", entry.UniqueValueCount)) + "\n\n")

	// Convert map to sorted slice for consistent display
	type ValueCount struct {
		Value string
		Count int64
	}

	values := make([]ValueCount, 0, len(entry.Values))
	for value, count := range entry.Values {
		values = append(values, ValueCount{Value: value, Count: count})
	}

	// Sort by count (descending), then by value name for ties
	sort.Slice(values, func(i, j int) bool {
		if values[i].Count == values[j].Count {
			return values[i].Value < values[j].Value
		}
		return values[i].Count > values[j].Count
	})

	// Calculate optimal column widths using full available width
	availableWidth := maxWidth

	// Reserve space for count column and separators
	countColumnWidth := 18 // Fixed width for "Count" column including percentage
	separatorWidth := 3    // " │ " separator

	// Use remaining space for value column
	valueColumnWidth := availableWidth - countColumnWidth - separatorWidth

	// Find actual max value length in data
	actualMaxValueLen := 0
	for _, vc := range values {
		if len(vc.Value) > actualMaxValueLen {
			actualMaxValueLen = len(vc.Value)
		}
	}

	// Use full available space for values, but ensure minimum readable width
	maxValueLength := max(valueColumnWidth, 20) // Use full space available, minimum 20

	// Table header
	headerStyle := lipgloss.NewStyle().Foreground(ColorWhite).Bold(true)
	header := fmt.Sprintf("%-*s │ %s", maxValueLength, "Value", "Count")
	modal.WriteString(headerStyle.Render(header) + "\n")

	// Divider line
	dividerStyle := lipgloss.NewStyle().Foreground(ColorGray)
	modal.WriteString(dividerStyle.Render(strings.Repeat("─", len(header))) + "\n")

	// Display ALL values with counts in table format (no artificial limit - let scrolling handle it)
	for _, vc := range values {
		displayValue := vc.Value
		if len(displayValue) > maxValueLength {
			displayValue = displayValue[:maxValueLength-3] + "..."
		}

		// Calculate percentage
		percentage := float64(vc.Count) * 100.0 / float64(entry.TotalCount)

		// Style the value and count
		valueStyle := lipgloss.NewStyle().Foreground(ColorBlue)
		countStyle := lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)

		// Format with proper table alignment
		line := fmt.Sprintf("%s │ %s",
			valueStyle.Render(fmt.Sprintf("%-*s", maxValueLength, displayValue)),
			countStyle.Render(fmt.Sprintf("%d (%.1f%%)", vc.Count, percentage)))

		modal.WriteString(line + "\n")
	}

	return modal.String()
}

// formatDuration formats a duration for user display
func (m *DashboardModel) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Nanoseconds()/1000000)
	}
	if d < time.Minute {
		if d%time.Second == 0 {
			return fmt.Sprintf("%ds", int(d.Seconds()))
		}
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		if d%time.Minute == 0 {
			return fmt.Sprintf("%dm", int(d.Minutes()))
		}
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return d.String()
}
