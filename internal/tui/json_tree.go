package tui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// jsonKind classifies a node in the parsed JSON tree.
type jsonKind int

const (
	kindObject jsonKind = iota
	kindArray
	kindString
	kindNumber
	kindBool
	kindNull
)

type jsonLineMode int

const (
	jsonLineModeWrap jsonLineMode = iota
	jsonLineModeScroll
)

// jsonNode is one node in the structured log tree. Branches (object/array)
// carry children; leaves carry a scalar value. path is a stable identifier used
// to remember expand/collapse state across re-renders.
type jsonNode struct {
	key      string // display key: object field name, "[i]" for array elements, "" for root
	kind     jsonKind
	value    interface{} // decoded subtree (for yank via MarshalIndent)
	children []*jsonNode
	path     string
}

func (n *jsonNode) isBranch() bool {
	return n.kind == kindObject || n.kind == kindArray
}

// flatNode is a node paired with its render depth in the flattened view.
type flatNode struct {
	node  *jsonNode
	depth int
}

// parseToNode decodes raw as JSON and builds a tree. ok is false when raw is
// not valid JSON.
func parseToNode(raw string) (*jsonNode, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, false
	}
	var v interface{}
	if err := json.Unmarshal([]byte(trimmed), &v); err != nil {
		return nil, false
	}
	return buildNode("", "root", v), true
}

// nodeFromAttributes builds a tree from the extracted attribute map. Used as a
// fallback when the raw line is absent or not JSON.
func nodeFromAttributes(attrs map[string]string) *jsonNode {
	m := make(map[string]interface{}, len(attrs))
	for k, v := range attrs {
		m[k] = v
	}
	return buildNode("", "root", m)
}

// buildNode recursively wraps a decoded JSON value into a jsonNode. Strings that
// are themselves valid JSON objects/arrays are expanded into sub-branches.
func buildNode(key, path string, v interface{}) *jsonNode {
	switch val := v.(type) {
	case map[string]interface{}:
		n := &jsonNode{key: key, kind: kindObject, value: v, path: path}
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			n.children = append(n.children, buildNode(k, path+"."+k, val[k]))
		}
		return n
	case []interface{}:
		n := &jsonNode{key: key, kind: kindArray, value: v, path: path}
		for i, item := range val {
			ck := fmt.Sprintf("[%d]", i)
			n.children = append(n.children, buildNode(ck, fmt.Sprintf("%s[%d]", path, i), item))
		}
		return n
	case string:
		// Expand strings that are wholly valid JSON objects/arrays.
		if nested, ok := parseNestedJSON(val); ok {
			n := buildNode(key, path, nested)
			return n
		}
		return &jsonNode{key: key, kind: kindString, value: val, path: path}
	case float64:
		return &jsonNode{key: key, kind: kindNumber, value: val, path: path}
	case bool:
		return &jsonNode{key: key, kind: kindBool, value: val, path: path}
	default: // nil
		return &jsonNode{key: key, kind: kindNull, value: nil, path: path}
	}
}

// parseNestedJSON returns the decoded value if s is wholly a JSON object or
// array. Plain strings (including JSON scalars and prefixed text) return false.
func parseNestedJSON(s string) (interface{}, bool) {
	t := strings.TrimSpace(s)
	if len(t) < 2 {
		return nil, false
	}
	if (t[0] != '{' || t[len(t)-1] != '}') && (t[0] != '[' || t[len(t)-1] != ']') {
		return nil, false
	}
	var v interface{}
	if err := json.Unmarshal([]byte(t), &v); err != nil {
		return nil, false
	}
	return v, true
}

// isExpanded reports whether a branch at path is expanded. Branches default to
// expanded; the map only records explicit collapses/toggles.
func isExpanded(expanded map[string]bool, path string) bool {
	v, ok := expanded[path]
	return !ok || v
}

// flattenVisible walks root honoring the expanded set and returns the ordered
// list of visible nodes with depths. The root node itself is not rendered when
// it is a branch; its children start at depth 0.
func flattenVisible(root *jsonNode, expanded map[string]bool) []flatNode {
	if root == nil {
		return nil
	}
	var out []flatNode
	var walk func(n *jsonNode, depth int)
	walk = func(n *jsonNode, depth int) {
		out = append(out, flatNode{node: n, depth: depth})
		if n.isBranch() && isExpanded(expanded, n.path) {
			for _, c := range n.children {
				walk(c, depth+1)
			}
		}
	}
	if root.isBranch() {
		for _, c := range root.children {
			walk(c, 0)
		}
	} else {
		walk(root, 0)
	}
	return out
}

// --- model integration ---

// buildJSONTree (re)builds the tree for the current log entry: raw line first,
// falling back to attributes. Resets cursor and expand state.
func (m *DashboardModel) buildJSONTree(entry LogEntry) {
	if root, ok := parseToNode(entry.RawLine); ok {
		m.jsonRoot = root
	} else if len(entry.Attributes) > 0 {
		m.jsonRoot = nodeFromAttributes(entry.Attributes)
	} else if entry.Message != "" {
		// Last resort: a single string leaf so the pane is never empty.
		m.jsonRoot = &jsonNode{key: "message", kind: kindString, value: entry.Message, path: "root.message"}
	} else {
		m.jsonRoot = nil
	}
	m.jsonExpanded = make(map[string]bool)
	m.jsonCursor = 0
	m.jsonHorizontalOffset = 0
	m.jsonVisible = flattenVisible(m.jsonRoot, m.jsonExpanded)
}

func (m *DashboardModel) jsonToggleLineMode() {
	if m.jsonLineMode == jsonLineModeWrap {
		m.jsonLineMode = jsonLineModeScroll
		return
	}
	m.jsonLineMode = jsonLineModeWrap
	m.jsonHorizontalOffset = 0
}

func (m *DashboardModel) jsonScrollLeft(cols int) {
	if cols < 1 {
		cols = 1
	}
	m.jsonHorizontalOffset -= cols
	if m.jsonHorizontalOffset < 0 {
		m.jsonHorizontalOffset = 0
	}
}

func (m *DashboardModel) jsonScrollRight(cols int) {
	if cols < 1 {
		cols = 1
	}
	m.jsonHorizontalOffset += cols
}

func (m *DashboardModel) jsonLineModeLabel() string {
	if m.jsonLineMode == jsonLineModeScroll {
		return "Scroll"
	}
	return "Wrap"
}

func (m *DashboardModel) jsonReflatten() {
	m.jsonVisible = flattenVisible(m.jsonRoot, m.jsonExpanded)
	if m.jsonCursor >= len(m.jsonVisible) {
		m.jsonCursor = len(m.jsonVisible) - 1
	}
	if m.jsonCursor < 0 {
		m.jsonCursor = 0
	}
}

func (m *DashboardModel) jsonFocusedNode() *jsonNode {
	if m.jsonCursor < 0 || m.jsonCursor >= len(m.jsonVisible) {
		return nil
	}
	return m.jsonVisible[m.jsonCursor].node
}

func (m *DashboardModel) jsonCursorMove(delta int) {
	if len(m.jsonVisible) == 0 {
		return
	}
	m.jsonCursor += delta
	if m.jsonCursor < 0 {
		m.jsonCursor = 0
	}
	if m.jsonCursor >= len(m.jsonVisible) {
		m.jsonCursor = len(m.jsonVisible) - 1
	}
}

// jsonToggle expands a collapsed branch or collapses an expanded one.
func (m *DashboardModel) jsonToggle() {
	n := m.jsonFocusedNode()
	if n == nil || !n.isBranch() {
		return
	}
	m.jsonExpanded[n.path] = !isExpanded(m.jsonExpanded, n.path)
	m.jsonReflatten()
}

// jsonExpand expands the focused branch (no-op if already expanded or a leaf).
func (m *DashboardModel) jsonExpand() {
	n := m.jsonFocusedNode()
	if n == nil || !n.isBranch() || isExpanded(m.jsonExpanded, n.path) {
		return
	}
	m.jsonExpanded[n.path] = true
	m.jsonReflatten()
}

// jsonCollapse collapses the focused branch; if it is a leaf or already
// collapsed, moves the cursor to the parent branch instead.
func (m *DashboardModel) jsonCollapse() {
	n := m.jsonFocusedNode()
	if n == nil {
		return
	}
	if n.isBranch() && isExpanded(m.jsonExpanded, n.path) {
		m.jsonExpanded[n.path] = false
		m.jsonReflatten()
		return
	}
	// Jump to parent: the nearest earlier node with a shallower depth.
	curDepth := m.jsonVisible[m.jsonCursor].depth
	for i := m.jsonCursor - 1; i >= 0; i-- {
		if m.jsonVisible[i].depth < curDepth {
			m.jsonCursor = i
			return
		}
	}
}

// renderJSONTree renders the visible tree to a string and returns it along with
// the cursor's line index so the caller can keep it in view.
func (m *DashboardModel) renderJSONTree(width int) (string, int) {
	if m.jsonRoot == nil || len(m.jsonVisible) == 0 {
		return lipgloss.NewStyle().Foreground(ColorGray).Italic(true).Render("No structured data"), 0
	}
	if width < 10 {
		width = 10
	}

	keyStyle := lipgloss.NewStyle().Foreground(ColorBlue)
	strStyle := lipgloss.NewStyle().Foreground(ColorGreen)
	numStyle := lipgloss.NewStyle().Foreground(ColorYellow)
	boolStyle := lipgloss.NewStyle().Foreground(ColorOrange)
	nullStyle := lipgloss.NewStyle().Foreground(ColorGray)
	punctStyle := lipgloss.NewStyle().Foreground(ColorGray)
	cursorStyle := lipgloss.NewStyle().Background(ColorDarkGray).Width(width)

	lines := make([]string, 0, len(m.jsonVisible))
	lineWidths := make([]int, 0, len(m.jsonVisible))
	for _, fn := range m.jsonVisible {
		n := fn.node
		indent := strings.Repeat("  ", fn.depth)

		var marker string
		if n.isBranch() {
			if isExpanded(m.jsonExpanded, n.path) {
				marker = "▼ "
			} else {
				marker = "▶ "
			}
		} else {
			marker = "  "
		}

		var label string
		if n.key != "" {
			label = keyStyle.Render(n.key) + punctStyle.Render(": ")
		}

		var val string
		switch n.kind {
		case kindObject:
			val = punctStyle.Render(fmt.Sprintf("{%d}", len(n.children)))
		case kindArray:
			val = punctStyle.Render(fmt.Sprintf("[%d]", len(n.children)))
		case kindString:
			val = strStyle.Render(strconv.Quote(n.value.(string)))
		case kindNumber:
			val = numStyle.Render(formatNumber(n.value.(float64)))
		case kindBool:
			val = boolStyle.Render(strconv.FormatBool(n.value.(bool)))
		case kindNull:
			val = nullStyle.Render("null")
		}

		line := indent + punctStyle.Render(marker) + label + val
		lines = append(lines, line)
		lineWidths = append(lineWidths, lipgloss.Width(line))
	}

	maxLineWidth := 0
	for _, w := range lineWidths {
		if w > maxLineWidth {
			maxLineWidth = w
		}
	}

	if m.jsonLineMode == jsonLineModeWrap {
		m.jsonHorizontalOffset = 0
	} else {
		maxOffset := max(0, maxLineWidth-width)
		if m.jsonHorizontalOffset > maxOffset {
			m.jsonHorizontalOffset = maxOffset
		}
		if m.jsonHorizontalOffset < 0 {
			m.jsonHorizontalOffset = 0
		}
	}

	var out []string
	cursorLine := 0
	for i, line := range lines {
		if m.jsonLineMode == jsonLineModeWrap {
			wrapped := wrapANSILine(line, width)
			if len(wrapped) == 0 {
				wrapped = []string{""}
			}
			if i == m.jsonCursor {
				cursorLine = len(out)
			}
			for _, seg := range wrapped {
				if i == m.jsonCursor {
					seg = cursorStyle.Render(padANSI(seg, width))
				}
				out = append(out, seg)
			}
			continue
		}

		seg := ansi.Cut(line, m.jsonHorizontalOffset, m.jsonHorizontalOffset+width)
		if i == m.jsonCursor {
			cursorLine = len(out)
			seg = cursorStyle.Render(padANSI(seg, width))
		}
		out = append(out, seg)
	}

	return strings.Join(out, "\n"), cursorLine
}

// formatNumber renders a JSON number without a trailing ".0" for integers.
func formatNumber(f float64) string {
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func wrapANSILine(s string, width int) []string {
	if width < 1 {
		return []string{s}
	}
	lineWidth := lipgloss.Width(s)
	if lineWidth <= width {
		return []string{s}
	}

	out := make([]string, 0, (lineWidth/width)+1)
	for start := 0; start < lineWidth; start += width {
		end := min(start+width, lineWidth)
		out = append(out, ansi.Cut(s, start, end))
	}
	return out
}

func padANSI(s string, width int) string {
	gap := width - lipgloss.Width(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}

// jsonFocusedYank returns the focused node's value as indented JSON for yank.
func (m *DashboardModel) jsonFocusedYank() (string, bool) {
	n := m.jsonFocusedNode()
	if n == nil {
		return "", false
	}
	out, err := json.MarshalIndent(n.value, "", "  ")
	if err != nil {
		return "", false
	}
	return string(out), true
}
