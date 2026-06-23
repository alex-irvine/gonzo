# Details Pane Rework — JSON Tree

Date: 2026-06-11

## Goal

Replace the split log-details modal (70% info / 30% AI chat) with a single
full-width pane that renders the log entry as a collapsible, colorized JSON
tree — a proper structured-logger view. Remove the chat pane entirely. Allow
yank inside the details pane, matching the main log list.

## Current State

- `renderSplitModal` (`internal/tui/modal_log_details.go`) draws a 70/30 split:
  left = `formatLogDetails` (timestamps, severity, message, attributes table,
  AI analysis), right = AI chat history + input.
- Chat is wired through many fields (`chatViewport`, `chatInput`, `chatHistory`,
  `chatActive`, `chatAiAnalyzing`, `chatAutoScroll`, `chatSpinnerFrame`,
  `modalActiveSection`), the `AIAnalysisMsg.IsChat` branch, tab toggling, and
  mouse click regions.
- `LogEntry.RawLine` holds the raw JSON log line. `Message` may contain embedded
  JSON (e.g. `[noise] {...}`). `Attributes` is the extracted key/value map.
- Yank (`y`/`Y`) works only in the main log list and is blocked while
  `m.showModal` is true.

## Design

### 1. JSON tree component — `internal/tui/json_tree.go`

- `parseToNode(raw string) (*jsonNode, bool)` — `encoding/json` decode into
  `interface{}`, wrap as a tree. Object/array → branch; scalar → leaf. If a
  string value itself parses as JSON, it becomes a sub-branch (handles the
  `message: "[noise] {...}"` case). Returns ok=false when `raw` is not JSON.
- `jsonNode`: `key`, `value`, `kind` (object/array/string/number/bool/null),
  `children []*jsonNode`, `path` (stable id for expand state).
- Model state on `DashboardModel`:
  - `jsonRoot *jsonNode`
  - `jsonExpanded map[string]bool` (path → expanded)
  - `jsonCursor int` (index into the flattened visible list)
  - `jsonVisible []*jsonNode` (recomputed flatten of visible nodes)
- `flattenVisible()` — walk root honoring `jsonExpanded`, produce `jsonVisible`.
- `renderTree(width int) string` — colored lines, cursor row highlighted,
  `▼`/`▶` markers on branches. Colors: keys gray, strings green, numbers cyan,
  bool/null orange, brackets/braces gray.
- Degrade path: `RawLine` empty or not JSON → build tree from `Attributes` map
  (sorted keys, all string leaves). Both empty → render "No data".

### 2. Modal render — `renderSplitModal` → `renderDetailModal`

- Drop the split, chat pane, tab indicators, and chat bits in the header.
- One full-width bordered pane.
- Small header block: Received / Log Time (if present) / Severity.
- Tree fills the body via `infoViewport` (reused for scrolling).
- AI Analysis section (`i` key) stays — it is info-pane content, not chat.
- Viewport auto-scrolls to keep `jsonCursor` row visible.

### 3. Key handling — `internal/tui/navigation.go`

Remove:
- `chatActive` input block.
- tab → chat toggle.
- mouse chat-region click handling.

Modal keys become:
- `up`/`k`, `down`/`j` — move tree cursor.
- `enter`, `space`, `l` — expand focused branch; `h` — collapse focused branch
  (or jump to parent if already collapsed/leaf).
- `pgup`/`pgdown` — viewport scroll.
- `i` — AI analysis (kept).
- `y` — yank whole entry (RawLine, fallback Message).
- `Y` — yank focused node value as `json.MarshalIndent`.
- `esc` — close modal.

### 4. Yank — `internal/tui/yank.go`

- `yankCurrentLogFromModal()` — whole entry via existing `yankableText`.
- `yankFocusedNode()` — `json.MarshalIndent` of the focused node's value.
- Reuse `setYankFeedback`; ensure feedback is visible while the modal is open
  (render in modal status bar).

### 5. Chat cleanup

- Remove dead chat fields, the `AIAnalysisMsg.IsChat` branch, chat spinner, and
  the chat-only `AnalyzeLogWithContext` caller. Leave the ai-client method
  itself in place (out of scope to gut).

## Testing

- Unit tests for `parseToNode`: flat object, nested object, array,
  nested-JSON-in-string, non-JSON input, attributes fallback.
- Unit tests for `flattenVisible` / expand-collapse cursor logic.
- Tree rendering, scrolling, and yank exercised manually in the TUI.

## Files Touched

- New: `internal/tui/json_tree.go`, `internal/tui/json_tree_test.go`
- Edited: `internal/tui/modal_log_details.go`, `internal/tui/navigation.go`,
  `internal/tui/update.go`, `internal/tui/yank.go`, `internal/tui/model.go`
