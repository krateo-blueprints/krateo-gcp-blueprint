package main

import "strings"

// keptScalarKeys are the JSON-Schema keywords we carry through from the Config Connector CRD
// spec into the blueprint's values.schema.json. Everything else -- notably oneOf/anyOf/allOf/not
// and any x-kubernetes-* extension -- is dropped, because Krateo's crdgen cannot express them
// and they would break Composition CRD generation.
//
// Config Connector uses those combinators far more than ACK does, but almost entirely for two
// mechanical idioms (see idiomNote): a reference that takes either `external` or `name`, and a
// secret that takes either `value` or `valueFrom`. Dropping the combinator loses the machine-
// enforced mutual exclusion, so we fold it into the field description instead -- the portal form
// still tells the user the rule, and the Config Connector webhook remains the real enforcer.
var keptScalarKeys = []string{
	"description", "type", "format", "enum", "default", "nullable",
	"pattern", "minLength", "maxLength",
	"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf",
	"minItems", "maxItems", "uniqueItems",
	"minProperties", "maxProperties",
}

// idiomNote returns the sentence describing the mutual-exclusion rule a dropped combinator
// encoded, or "" when the node carries no combinator we recognise.
func idiomNote(node map[string]interface{}) string {
	if _, ok := node["oneOf"]; !ok {
		return ""
	}
	props, _ := node["properties"].(map[string]interface{})
	if props == nil {
		return ""
	}
	_, hasExternal := props["external"]
	_, hasName := props["name"]
	if hasExternal && hasName {
		return "Set exactly one of `external` (the resource's GCP URI or ID) or `name` " +
			"(plus optional `namespace`) referencing another Config Connector resource in the cluster."
	}
	_, hasValue := props["value"]
	_, hasValueFrom := props["valueFrom"]
	if hasValue && hasValueFrom {
		return "Set exactly one of `value` (a literal) or `valueFrom` (a reference to a Secret key)."
	}
	return ""
}

// appendSentence joins a note onto an existing description without duplicating it.
func appendSentence(desc, note string) string {
	if note == "" {
		return desc
	}
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return note
	}
	if strings.Contains(desc, note) {
		return desc
	}
	if !strings.HasSuffix(desc, ".") {
		desc += "."
	}
	return desc + " " + note
}

// curate returns a crdgen-safe copy of an openAPIV3Schema node.
func curate(node map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}

	for _, k := range keptScalarKeys {
		if v, ok := node[k]; ok {
			out[k] = v
		}
	}

	// Preserve the meaning of a dropped oneOf we recognise, as prose on the field.
	if note := idiomNote(node); note != "" {
		desc, _ := out["description"].(string)
		out["description"] = appendSentence(desc, note)
	}

	// required: keep only string entries.
	if req, ok := node["required"].([]interface{}); ok {
		var keep []interface{}
		for _, r := range req {
			if s, ok := r.(string); ok {
				keep = append(keep, s)
			}
		}
		if len(keep) > 0 {
			out["required"] = keep
		}
	}

	// properties: recurse.
	if props, ok := node["properties"].(map[string]interface{}); ok {
		cp := map[string]interface{}{}
		for name, sub := range props {
			if subMap, ok := sub.(map[string]interface{}); ok {
				cp[name] = curate(subMap)
			}
		}
		if len(cp) > 0 {
			out["properties"] = cp
		}
	}

	// items: recurse (single schema form).
	if items, ok := node["items"].(map[string]interface{}); ok {
		out["items"] = curate(items)
	}

	// additionalProperties: bool passthrough, or recurse when it is a schema (map types).
	switch ap := node["additionalProperties"].(type) {
	case bool:
		out["additionalProperties"] = ap
	case map[string]interface{}:
		out["additionalProperties"] = curate(ap)
	}

	// A node that declared x-kubernetes-preserve-unknown-fields but no concrete type/properties
	// becomes a free-form object so the form still renders something usable.
	if _, hasType := out["type"]; !hasType {
		if _, hasProps := out["properties"]; !hasProps {
			if preserve, _ := node["x-kubernetes-preserve-unknown-fields"].(bool); preserve {
				out["type"] = "object"
			}
		}
	}

	return out
}

const projectIDDescription = "GCP project ID for this resource. Rendered as the " +
	"cnrm.cloud.google.com/project-id annotation. Leave empty to inherit the project from the " +
	"namespace (Config Connector falls back to the namespace's project-id annotation, then to the " +
	"namespace name). Leave empty when this resource's spec already carries a projectRef."

// buildValuesSchema projects the curated Config Connector spec into the blueprint's top-level
// values.schema.json, injecting the Krateo-wiring `projectId` field.
func buildValuesSchema(spec map[string]interface{}, title, desc string) map[string]interface{} {
	curated := curate(spec)

	props, _ := curated["properties"].(map[string]interface{})
	if props == nil {
		props = map[string]interface{}{}
	}
	props["projectId"] = map[string]interface{}{
		"type":        "string",
		"title":       "GCP project ID",
		"description": projectIDDescription,
		"default":     "",
	}

	// NOTE: do not set additionalProperties:false at the root. Krateo/Helm inject a `global`
	// values key when installing the chart, and a strict root would reject it (chart-inspector
	// fails with "additional properties 'global' not allowed").
	out := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema",
		"type":        "object",
		"title":       title,
		"description": desc,
		"properties":  props,
	}
	if req, ok := curated["required"]; ok {
		out["required"] = req
	}
	return out
}

// minimalValid produces the smallest value satisfying a curated schema's required/type/enum
// constraints -- used to seed values.yaml so `helm lint`/`helm template` pass without overrides.
func minimalValid(schema map[string]interface{}) interface{} {
	if enum, ok := schema["enum"].([]interface{}); ok && len(enum) > 0 {
		return enum[0]
	}
	if def, ok := schema["default"]; ok {
		return def
	}

	typ, _ := schema["type"].(string)
	switch typ {
	case "string":
		// A restrictive `pattern` makes the generic placeholder invalid, and values.schema.json
		// is enforced by `helm template`, so derive a matching sample when one is declared.
		if pat, ok := schema["pattern"].(string); ok && pat != "" {
			minLen := 0
			if m, ok := toFloat(schema["minLength"]); ok {
				minLen = int(m)
			}
			if sample := sampleForPattern(pat, minLen); sample != "" {
				return sample
			}
		}
		return "example"
	case "integer", "number":
		if m, ok := toFloat(schema["minimum"]); ok {
			return m
		}
		return 0
	case "boolean":
		return false
	case "array":
		if mi, ok := toFloat(schema["minItems"]); ok && mi > 0 {
			if items, ok := schema["items"].(map[string]interface{}); ok {
				return []interface{}{minimalValid(items)}
			}
		}
		return []interface{}{}
	case "object":
		obj := map[string]interface{}{}
		props, _ := schema["properties"].(map[string]interface{})
		for _, name := range requiredNames(schema) {
			if sub, ok := props[name].(map[string]interface{}); ok {
				obj[name] = minimalValid(sub)
			}
		}
		return obj
	default:
		// Unknown/typeless node: emit an empty object so the field exists but stays inert.
		return map[string]interface{}{}
	}
}

func requiredNames(schema map[string]interface{}) []string {
	var out []string
	if req, ok := schema["required"].([]interface{}); ok {
		for _, r := range req {
			if s, ok := r.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	}
	return 0, false
}
