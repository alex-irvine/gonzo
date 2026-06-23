package tui

import "testing"

func TestParseToNode_FlatObject(t *testing.T) {
	root, ok := parseToNode(`{"level":"info","value":65}`)
	if !ok {
		t.Fatal("expected valid JSON")
	}
	if root.kind != kindObject {
		t.Fatalf("root kind = %v, want object", root.kind)
	}
	if len(root.children) != 2 {
		t.Fatalf("children = %d, want 2", len(root.children))
	}
	// Keys are sorted: level, value.
	if root.children[0].key != "level" || root.children[0].kind != kindString {
		t.Errorf("child[0] = %+v", root.children[0])
	}
	if root.children[1].key != "value" || root.children[1].kind != kindNumber {
		t.Errorf("child[1] = %+v", root.children[1])
	}
}

func TestParseToNode_Nested(t *testing.T) {
	root, ok := parseToNode(`{"labels":{"ns":"default","lvl":"err"}}`)
	if !ok {
		t.Fatal("expected valid JSON")
	}
	labels := root.children[0]
	if labels.key != "labels" || labels.kind != kindObject {
		t.Fatalf("labels = %+v", labels)
	}
	if len(labels.children) != 2 {
		t.Fatalf("labels children = %d, want 2", len(labels.children))
	}
	if labels.children[0].path != "root.labels.lvl" {
		t.Errorf("path = %q", labels.children[0].path)
	}
}

func TestParseToNode_Array(t *testing.T) {
	root, ok := parseToNode(`{"items":[1,2,3]}`)
	if !ok {
		t.Fatal("expected valid JSON")
	}
	items := root.children[0]
	if items.kind != kindArray || len(items.children) != 3 {
		t.Fatalf("items = %+v", items)
	}
	if items.children[2].key != "[2]" || items.children[2].path != "root.items[2]" {
		t.Errorf("child[2] = %+v", items.children[2])
	}
}

func TestParseToNode_NestedJSONString(t *testing.T) {
	root, ok := parseToNode(`{"msg":"{\"a\":1}"}`)
	if !ok {
		t.Fatal("expected valid JSON")
	}
	msg := root.children[0]
	if msg.kind != kindObject {
		t.Fatalf("nested JSON string should expand to object, got %v", msg.kind)
	}
	if msg.children[0].key != "a" {
		t.Errorf("nested child = %+v", msg.children[0])
	}
}

func TestParseToNode_PrefixedStringStaysLeaf(t *testing.T) {
	root, ok := parseToNode(`{"msg":"[noise] {\"a\":1}"}`)
	if !ok {
		t.Fatal("expected valid JSON")
	}
	if root.children[0].kind != kindString {
		t.Errorf("prefixed string should stay a leaf, got %v", root.children[0].kind)
	}
}

func TestParseToNode_NotJSON(t *testing.T) {
	if _, ok := parseToNode("plain text log line"); ok {
		t.Error("expected non-JSON to fail")
	}
	if _, ok := parseToNode(""); ok {
		t.Error("expected empty to fail")
	}
}

func TestNodeFromAttributes(t *testing.T) {
	root := nodeFromAttributes(map[string]string{"b": "2", "a": "1"})
	if root.kind != kindObject || len(root.children) != 2 {
		t.Fatalf("root = %+v", root)
	}
	// Sorted keys.
	if root.children[0].key != "a" || root.children[1].key != "b" {
		t.Errorf("keys not sorted: %q %q", root.children[0].key, root.children[1].key)
	}
}

func TestFlattenVisible_DefaultExpanded(t *testing.T) {
	root, _ := parseToNode(`{"a":{"b":1},"c":2}`)
	flat := flattenVisible(root, map[string]bool{})
	// a (branch), b (leaf), c (leaf) => 3 visible by default.
	if len(flat) != 3 {
		t.Fatalf("visible = %d, want 3", len(flat))
	}
	if flat[0].depth != 0 || flat[1].depth != 1 || flat[2].depth != 0 {
		t.Errorf("depths = %d,%d,%d", flat[0].depth, flat[1].depth, flat[2].depth)
	}
}

func TestFlattenVisible_Collapsed(t *testing.T) {
	root, _ := parseToNode(`{"a":{"b":1},"c":2}`)
	expanded := map[string]bool{"root.a": false}
	flat := flattenVisible(root, expanded)
	// a (collapsed branch), c => 2 visible.
	if len(flat) != 2 {
		t.Fatalf("visible = %d, want 2", len(flat))
	}
	if flat[0].node.key != "a" || flat[1].node.key != "c" {
		t.Errorf("got %q,%q", flat[0].node.key, flat[1].node.key)
	}
}

func TestToggleExpandCollapse(t *testing.T) {
	m := &DashboardModel{}
	m.buildJSONTree(LogEntry{RawLine: `{"a":{"b":1},"c":2}`})
	if len(m.jsonVisible) != 3 {
		t.Fatalf("initial visible = %d, want 3", len(m.jsonVisible))
	}
	// Cursor on "a" (index 0). Collapse it.
	m.jsonToggle()
	if len(m.jsonVisible) != 2 {
		t.Fatalf("after collapse visible = %d, want 2", len(m.jsonVisible))
	}
	// Expand again.
	m.jsonToggle()
	if len(m.jsonVisible) != 3 {
		t.Fatalf("after expand visible = %d, want 3", len(m.jsonVisible))
	}
}

func TestBuildJSONTree_AttributesFallback(t *testing.T) {
	m := &DashboardModel{}
	m.buildJSONTree(LogEntry{RawLine: "not json", Attributes: map[string]string{"k": "v"}})
	if m.jsonRoot == nil || len(m.jsonRoot.children) != 1 {
		t.Fatalf("expected attributes fallback, got %+v", m.jsonRoot)
	}
	if m.jsonRoot.children[0].key != "k" {
		t.Errorf("key = %q", m.jsonRoot.children[0].key)
	}
}

func TestJSONFocusedYank(t *testing.T) {
	m := &DashboardModel{}
	m.buildJSONTree(LogEntry{RawLine: `{"a":{"b":1}}`})
	// Cursor on "a" object.
	out, ok := m.jsonFocusedYank()
	if !ok {
		t.Fatal("expected yank")
	}
	want := "{\n  \"b\": 1\n}"
	if out != want {
		t.Errorf("yank = %q, want %q", out, want)
	}
}
