package main

import "testing"

func TestNormalizeSeverity(t *testing.T) {
	cases := map[string]string{
		"info":        "INFO",
		"INF":         "INFO",
		"information": "INFO",
		"warn":        "WARN",
		"warning":     "WARN",
		"WRN":         "WARN",
		"err":         "ERROR",
		"ERRO":        "ERROR",
		"error":       "ERROR",
		"critical":    "FATAL",
		"crit":        "FATAL",
		"panic":       "FATAL",
		"fatal":       "FATAL",
		"trace":       "TRACE",
		"debug":       "DEBUG",
		"":            "INFO",
		"garbage":     "INFO",
		"  Info  ":    "INFO",
	}
	for in, want := range cases {
		if got := normalizeSeverity(in); got != want {
			t.Errorf("normalizeSeverity(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExtractSeverityFromText(t *testing.T) {
	cases := map[string]string{
		"2025-01-01 INFO server started":     "INFO",
		"2025-01-01 [warn] disk almost full": "WARN",
		"WARNING: deprecated flag":           "WARN",
		"plain text with no level":           "INFO",
		"ts=... level=error msg=boom":        "ERROR",
		"CRITICAL failure in subsystem":      "FATAL",
		"trace span emitted":                 "TRACE",
		"msg at debug level":                 "DEBUG",
	}
	for in, want := range cases {
		if got := extractSeverityFromText(in); got != want {
			t.Errorf("extractSeverityFromText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGetVersionInfo(t *testing.T) {
	v, c := GetVersionInfo()
	if v == "" || c == "" {
		t.Fatalf("GetVersionInfo returned empty: version=%q commit=%q", v, c)
	}
}
