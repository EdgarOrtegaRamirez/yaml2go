package main

import (
	"bytes"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConvertYAML(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantContains []string
	}{
		{
			name:         "simple string field",
			input:        "name: John\nage: 30\n",
			wantContains: []string{"type Name struct", "Name string", "Age int64"},
		},
		{
			name:         "nested object",
			input:        "database:\n  host: localhost\n  port: 5432\n",
			wantContains: []string{"type Database struct", "Host string", "Port int64"},
		},
		{
			name:         "array field",
			input:        "tags:\n  - go\n  - rust\n  - python\n",
			wantContains: []string{"Tags []string"},
		},
		{
			name:         "boolean field",
			input:        "enabled: true\ndebug: false\n",
			wantContains: []string{"Enabled bool", "Debug bool"},
		},
		{
			name:         "float field",
			input:        "ratio: 0.75\n",
			wantContains: []string{"Ratio float64"},
		},
		{
			name:         "json and yaml tags",
			input:        "name: John\n",
			wantContains: []string{"json:\"name\" yaml:\"name\""},
		},
		{
			name:         "camelCase naming",
			input:        "host_name: localhost\n",
			wantContains: []string{"HostName string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				PackageName:      "models",
				NamingConvention: "pascal",
				IncludeTags:      true,
				Tags:             []string{"json", "yaml"},
			}

			var out bytes.Buffer
			err := convertYAML([]byte(tt.input), cfg, &out)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result := out.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("output does not contain %q\nGot:\n%s", want, result)
				}
			}
		})
	}
}

func TestConvertYAML_CamelCaseNaming(t *testing.T) {
	input := "host_name: localhost\nport_number: 8080\n"
	cfg := Config{
		PackageName:      "models",
		NamingConvention: "camel",
		IncludeTags:      true,
		Tags:             []string{"json", "yaml"},
	}

	var out bytes.Buffer
	err := convertYAML([]byte(input), cfg, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "hostName") {
		t.Errorf("output does not contain camelCase field 'hostName'\nGot:\n%s", result)
	}
	if !strings.Contains(result, "portNumber") {
		t.Errorf("output does not contain camelCase field 'portNumber'\nGot:\n%s", result)
	}
}

func TestConvertYAML_NoTags(t *testing.T) {
	input := "name: John\n"
	cfg := Config{
		PackageName: "models",
		IncludeTags: false,
	}

	var out bytes.Buffer
	err := convertYAML([]byte(input), cfg, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if strings.Contains(result, "`") {
		t.Errorf("output should not contain tags but got:\n%s", result)
	}
}

func TestSplitIntoParts(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"hello_world", []string{"hello", "world"}},
		{"host-name", []string{"host", "name"}},
		{"camelCase", []string{"camel", "Case"}},
		{"PascalCase", []string{"Pascal", "Case"}},
		{"123-numbers", []string{"123", "numbers"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitIntoParts(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitIntoParts(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestToFieldName(t *testing.T) {
	tests := []struct {
		convention string
		input      string
		want       string
	}{
		{"pascal", "host_name", "HostName"},
		{"camel", "host_name", "hostName"},
		{"snake", "host_name", "host_name"},
	}

	for _, tt := range tests {
		t.Run(tt.convention+"_"+tt.input, func(t *testing.T) {
			got := toFieldName(tt.input, tt.convention)
			if got != tt.want {
				t.Errorf("toFieldName(%q, %q) = %q, want %q", tt.input, tt.convention, got, tt.want)
			}
		})
	}
}

func TestInferScalarType(t *testing.T) {
	tests := []struct {
		tag  string
		val  string
		want string
	}{
		{"!!bool", "true", "bool"},
		{"!!int", "42", "int64"},
		{"!!float", "3.14", "float64"},
		{"!!str", "hello", "string"},
		{"!!str", "yes", "bool"},
		{"!!str", "no", "bool"},
		{"!!null", "", "string"},
		{"!!map", "", "map[string]interface{}"},
		{"!!seq", "", "[]interface{}"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			node := &yaml.Node{Kind: yaml.ScalarNode, Tag: tt.tag, Value: tt.val}
			got := inferScalarType(node, nil)
			if got != tt.want {
				t.Errorf("inferScalarType(%q, %q) = %q, want %q", tt.tag, tt.val, got, tt.want)
			}
		})
	}
}

func TestBuildTag(t *testing.T) {
	tests := []struct {
		name  string
		tags  []string
		input string
		want  string
	}{
		{"json only", []string{"json"}, "userName", `json:"username"`},
		{"json and yaml", []string{"json", "yaml"}, "hostName", `json:"hostname" yaml:"hostname"`},
		{"empty", []string{}, "name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTag(tt.input, tt.tags)
			if got != tt.want {
				t.Errorf("buildTag(%q, %v) = %q, want %q", tt.input, tt.tags, got, tt.want)
			}
		})
	}
}

func TestIsBoolPattern(t *testing.T) {
	if !isBoolPattern("true") || !isBoolPattern("false") || !isBoolPattern("yes") || !isBoolPattern("no") {
		t.Error("isBoolPattern should accept true/false/yes/no")
	}
	if isBoolPattern("maybe") {
		t.Error("isBoolPattern should reject 'maybe'")
	}
}

func TestInlineStruct(t *testing.T) {
	fields := []Field{
		{Name: "Name", Type: "string"},
		{Name: "Age", Type: "int64"},
	}

	result := inlineStruct(fields)
	if !strings.Contains(result, "struct {") {
		t.Errorf("inlineStruct should contain 'struct {': %s", result)
	}
	if !strings.Contains(result, "Name string") {
		t.Errorf("inlineStruct should contain field name: %s", result)
	}
}

func TestEmptyInput(t *testing.T) {
	cfg := Config{PackageName: "models"}
	var out bytes.Buffer
	err := convertYAML([]byte(""), cfg, &out)
	if err != nil {
		t.Errorf("empty input should not error, got: %v", err)
	}
}
