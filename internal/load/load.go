// Package load reads tool and MCP server definitions.
package load

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

const MaxBytes = 1 << 20

// Item is one tool or one MCP server definition.
type Item struct {
	Name   string
	Kind   string
	Schema []byte
}

// MaxDepth is the deepest container nesting Parse accepts.
const MaxDepth = 64

// ReadFile loads path. It opens the file once, refuses symlinks and anything that is not a
// regular file (FIFOs, devices, directories), and reads at most MaxBytes.
func ReadFile(path string) ([]Item, error) {
	f, err := openNoFollow(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	data, err := io.ReadAll(io.LimitReader(f, MaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxBytes {
		return nil, fmt.Errorf("%s is larger than %d bytes", path, MaxBytes)
	}
	return Parse(path, data)
}

// checkJSON walks the tokens once. It rejects duplicate keys, nesting deeper than MaxDepth,
// and any value after the first document.
func checkJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := walk(dec, 1); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return errors.New("unexpected data after the first JSON value")
		}
		return err
	}
	return nil
}

func walk(dec *json.Decoder, depth int) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	if depth > MaxDepth {
		return fmt.Errorf("nesting deeper than %d", MaxDepth)
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return err
			}
			key, _ := kt.(string)
			if seen[key] {
				return fmt.Errorf("duplicate key %q", key)
			}
			seen[key] = true
			if err := walk(dec, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for dec.More() {
			if err := walk(dec, depth+1); err != nil {
				return err
			}
		}
	}
	_, err = dec.Token() // closing delimiter
	return err
}

// Parse accepts an object with tools and/or mcpServers, or a bare array of tools.
// Any other shape, unknown top-level keys, duplicate keys, trailing values and deep nesting
// are errors.
func Parse(path string, data []byte) ([]Item, error) {
	if err := checkJSON(data); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
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
		for key := range value {
			if key != "tools" && key != "mcpServers" {
				return nil, fmt.Errorf("%s: unknown top-level key %q", path, key)
			}
		}
		if _, ok := value["tools"]; !ok {
			if _, ok := value["mcpServers"]; !ok {
				return nil, fmt.Errorf("%s: expected tools or mcpServers", path)
			}
		}
		out := []Item{}
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
