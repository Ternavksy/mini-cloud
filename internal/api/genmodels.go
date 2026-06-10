//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type openAPI struct {
	Components components `yaml:"components"`
}

type components struct {
	Schemas map[string]schema `yaml:"schemas"`
}

type schema struct {
	Ref        string            `yaml:"$ref"`
	Type       string            `yaml:"type"`
	Format     string            `yaml:"format"`
	Properties map[string]schema `yaml:"properties"`
	Items      *schema           `yaml:"items"`
}

func main() {
	data, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var spec openAPI
	if err := yaml.Unmarshal(data, &spec); err != nil {
		log.Fatal(err)
	}

	var out bytes.Buffer
	out.WriteString("// Code generated from api/openapi.yaml; DO NOT EDIT.\n\n")
	out.WriteString("package api\n\n")
	out.WriteString("import \"time\"\n\n")

	names := make([]string, 0, len(spec.Components.Schemas))
	for name := range spec.Components.Schemas {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		schema := spec.Components.Schemas[name]
		if schema.Type != "object" {
			continue
		}
		writeStruct(&out, name, schema)
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		log.Fatalf("format generated source: %v\n%s", err, out.String())
	}
	if err := os.WriteFile("models.gen.go", formatted, 0644); err != nil {
		log.Fatal(err)
	}
}

func writeStruct(out *bytes.Buffer, name string, s schema) {
	fmt.Fprintf(out, "type %s struct {\n", name)

	keys := make([]string, 0, len(s.Properties))
	for key := range s.Properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		property := s.Properties[key]
		fmt.Fprintf(out, "\t%s %s `json:\"%s\"`\n", fieldName(key), goType(property), key)
	}
	out.WriteString("}\n\n")
}

func goType(s schema) string {
	if s.Ref != "" {
		return refType(s.Ref)
	}
	if s.Type == "array" && s.Items != nil {
		return "[]" + goType(*s.Items)
	}
	if s.Type == "string" && s.Format == "date-time" {
		return "time.Time"
	}
	if s.Type == "string" {
		return "string"
	}
	if s.Type == "object" {
		return "map[string]any"
	}
	return "any"
}

func refType(ref string) string {
	const prefix = "#/components/schemas/"
	return strings.TrimPrefix(ref, prefix)
}

func fieldName(name string) string {
	parts := strings.Split(name, "_")
	for i, part := range parts {
		if part == "id" {
			parts[i] = "ID"
			continue
		}
		if part == "ids" {
			parts[i] = "IDs"
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}
