package load

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSortsKeysAndRefusesSymlink(t *testing.T) {
	items, err := Parse("tools.json", []byte(`{"mcpServers":{"b":{"command":"b"},"a":{"command":"a"}},"tools":[{"name":"search","input_schema":{"z":1,"a":2}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("%d items", len(items))
	}
	if items[0].Name != "search" || string(items[0].Schema) != `{"input_schema":{"a":2,"z":1},"name":"search"}` {
		t.Fatalf("tool: %+v %s", items[0], items[0].Schema)
	}
	if items[1].Name != "a" || items[2].Name != "b" {
		t.Fatalf("server order: %+v", items)
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "t.json")
	if err := os.WriteFile(target, []byte(`[]`), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(link); err == nil {
		t.Fatal("symlink should be refused")
	}
}
