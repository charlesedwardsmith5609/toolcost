// Package load reads tool and MCP server definitions.
package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

const MaxBytes = 1 << 20

// Item is one tool or one MCP server definition.
type Item struct {
	Name   string
	Kind   string
	Schema []byte
}

// ReadFile loads path. Symlinks and oversized files are errors.
func ReadFile(path string) ([]Item, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%s is a symlink", path)
	}
	if info.Size() > MaxBytes {
		return nil, fmt.Errorf("%s is larger than %d bytes", path, MaxBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(path, data)
}

// Parse accepts an object with tools and mcpServers, or a bare array of tools.
func Parse(path string, data []byte) ([]Item, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var root any
	if err := dec.Decode(&root); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	switch value := root.(type) {
	case []any:
		return itemsFromTools("tool", value)
	case map[string]any:
		var out []Item
		if raw, ok := value["tools"]; ok {
			tools, ok := raw.([]any)
			if !ok {
				return nil, fmt.Errorf("%s: tools must be an array", path)
			}
			parsed, err := itemsFromTools("tool", tools)
			if err != nil {
				return nil, err
			}
			out = append(out, parsed...)
		}
		if raw, ok := value["mcpServers"]; ok {
			servers, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s: mcpServers must be an object", path)
			}
			names := make([]string, 0, len(servers))
			for name := range servers {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				body, err := canonical(servers[name])
				if err != nil {
					return nil, err
				}
				out = append(out, Item{Name: name, Kind: "mcp", Schema: body})
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%s: expected an object or an array", path)
	}
}

func itemsFromTools(kind string, tools []any) ([]Item, error) {
	var out []Item
	for i, tool := range tools {
		body, err := canonical(tool)
		if err != nil {
			return nil, err
		}
		name := fmt.Sprintf("tool[%d]", i)
		if obj, ok := tool.(map[string]any); ok {
			if raw, ok := obj["name"].(string); ok && raw != "" {
				name = raw
			}
		}
		out = append(out, Item{Name: name, Kind: kind, Schema: body})
	}
	return out, nil
}

func canonical(value any) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeCanonical(buf *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			encoded, err := json.Marshal(key)
			if err != nil {
				return err
			}
			buf.Write(encoded)
			buf.WriteByte(':')
			if err := writeCanonical(buf, typed[key]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	case []any:
		buf.WriteByte('[')
		for i, item := range typed {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeCanonical(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case json.Number:
		buf.WriteString(typed.String())
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return err
		}
		buf.Write(encoded)
	}
	return nil
}
