package exporter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kerlenton/mcpsnoop/internal/proxy"
)

func writeEnv(t *testing.T, path string, env proxy.Envelope) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		t.Fatal(err)
	}
}

func sampleLog(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "session.jsonl")
	t0 := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	writeEnv(t, path, proxy.Envelope{
		SessionID: "s1", ServerLabel: "demo", Seq: 1, TS: t0,
		Direction: proxy.ClientToServer, Raw: json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"text":"hi"}}}`),
	})
	writeEnv(t, path, proxy.Envelope{
		SessionID: "s1", ServerLabel: "demo", Seq: 2, TS: t0.Add(25 * time.Millisecond),
		Direction: proxy.ServerToClient, Raw: json.RawMessage(`{"jsonrpc":"2.0","id":1,"result":{"content":[],"isError":false}}`),
	})
	return path
}

func TestBuildCorrelatedExport(t *testing.T) {
	st, id, err := LoadFile(sampleLog(t))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Build(st, id)
	if err != nil {
		t.Fatal(err)
	}
	if out.Session.ID != "s1" || out.Session.Requests != 1 || out.Session.Responses != 1 {
		t.Fatalf("bad summary: %+v", out.Session)
	}
	if len(out.Calls) != 1 || out.Calls[0].ToolName != "echo" || out.Calls[0].DurationMS == nil {
		t.Fatalf("bad calls: %+v", out.Calls)
	}
	if len(out.Events) != 2 || out.Events[1].CallIndex == nil || *out.Events[1].CallIndex != 0 {
		t.Fatalf("bad event correlation: %+v", out.Events)
	}
}

func TestWriteFormats(t *testing.T) {
	st, id, err := LoadFile(sampleLog(t))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Build(st, id)
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []Format{FormatJSON, FormatHTML, FormatText} {
		var buf bytes.Buffer
		if err := Write(&buf, out, Options{Format: format}); err != nil {
			t.Fatalf("%s write failed: %v", format, err)
		}
		got := buf.String()
		if !strings.Contains(got, "echo") {
			t.Fatalf("%s export missing tool name:\n%s", format, got)
		}
	}
}
