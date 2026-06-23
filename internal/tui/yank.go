package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
)

// visualRange returns the inclusive [low, high] log indices covered by the current
// visual selection. Caller must check visualMode first.
func (m *DashboardModel) visualRange() (int, int) {
	a, b := m.visualAnchorIndex, m.selectedLogIndex
	if a > b {
		a, b = b, a
	}
	if a < 0 {
		a = 0
	}
	if b >= len(m.logEntries) {
		b = len(m.logEntries) - 1
	}
	return a, b
}

// yankableText returns the canonical copyable form of a log entry: RawLine
// when present, otherwise Message.
func yankableText(e LogEntry) string {
	if e.RawLine != "" {
		return e.RawLine
	}
	return e.Message
}

// yankCurrentLog copies the entry at selectedLogIndex to the clipboard.
func (m *DashboardModel) yankCurrentLog() {
	if m.selectedLogIndex < 0 || m.selectedLogIndex >= len(m.logEntries) {
		m.setYankFeedback("Nothing to yank")
		return
	}
	text := yankableText(m.logEntries[m.selectedLogIndex])
	if err := clipboard.WriteAll(text); err != nil {
		m.setYankFeedback(fmt.Sprintf("Yank failed: %v", err))
		return
	}
	m.setYankFeedback("Yanked 1 log entry")
}

// yankVisualSelection copies all entries currently covered by the visual range.
// Exits visual mode afterwards.
func (m *DashboardModel) yankVisualSelection() {
	if !m.visualMode {
		m.yankCurrentLog()
		return
	}
	lo, hi := m.visualRange()
	if lo < 0 || hi < 0 || lo >= len(m.logEntries) {
		m.visualMode = false
		m.setYankFeedback("Nothing to yank")
		return
	}
	var b strings.Builder
	count := 0
	for i := lo; i <= hi && i < len(m.logEntries); i++ {
		if count > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(yankableText(m.logEntries[i]))
		count++
	}
	if err := clipboard.WriteAll(b.String()); err != nil {
		m.setYankFeedback(fmt.Sprintf("Yank failed: %v", err))
	} else {
		m.setYankFeedback(fmt.Sprintf("Yanked %d log entries", count))
	}
	m.visualMode = false
}

// yankCurrentLogFromModal copies the entry currently shown in the details modal
// to the clipboard. Mirrors yankCurrentLog but sourced from currentLogEntry so
// it works while the modal is open.
func (m *DashboardModel) yankCurrentLogFromModal() {
	if m.currentLogEntry == nil {
		m.setYankFeedback("Nothing to yank")
		return
	}
	text := yankableText(*m.currentLogEntry)
	if err := clipboard.WriteAll(text); err != nil {
		m.setYankFeedback(fmt.Sprintf("Yank failed: %v", err))
		return
	}
	m.setYankFeedback("Yanked log entry")
}

// yankFocusedNode copies the focused JSON tree node (as indented JSON) to the
// clipboard.
func (m *DashboardModel) yankFocusedNode() {
	text, ok := m.jsonFocusedYank()
	if !ok {
		m.setYankFeedback("Nothing to yank")
		return
	}
	if err := clipboard.WriteAll(text); err != nil {
		m.setYankFeedback(fmt.Sprintf("Yank failed: %v", err))
		return
	}
	m.setYankFeedback("Yanked node")
}

// setYankFeedback stores a transient message and resets the tick counter that
// will clear it on the next handful of dashboard updates.
func (m *DashboardModel) setYankFeedback(msg string) {
	m.yankFeedback = msg
	m.yankFeedbackTick = 0
}
