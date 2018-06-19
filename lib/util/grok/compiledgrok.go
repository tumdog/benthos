package grok

import (
	"fmt"
	"strconv"

	"github.com/Jeffail/benthos/lib/util/grok/regexp"
)

// CompiledGrok represents a compiled Grok expression.
// Use Grok.Compile to generate a CompiledGrok object.
type CompiledGrok struct {
	regexp      regexp.Compiled
	typeHints   typeHintByKey
	removeEmpty bool
}

type typeHintByKey map[string]string

// ParseTyped processes the given data and returns a map containing the values
// of all named fields converted to their corresponding types. If no typehint is
// given, the value will be converted to string.
func (compiled CompiledGrok) ParseTyped(data []byte) (map[string]interface{}, error) {
	captures := make(map[string]interface{})

	if matches := compiled.regexp.FindSubmatch(data); len(matches) > 0 {
		for idx, key := range compiled.regexp.SubexpNames() {
			match := matches[idx]
			if compiled.omitField(key, match) {
				continue
			}

			if val, err := compiled.typeCast(string(match), key); err == nil {
				captures[key] = val
			} else {
				return nil, err
			}
		}
	}

	return captures, nil
}

// omitField return true if the field is to be omitted
func (compiled CompiledGrok) omitField(key string, match []byte) bool {
	return len(key) == 0 || compiled.removeEmpty && len(match) == 0
}

// typeCast casts a field based on a typehint
func (compiled CompiledGrok) typeCast(match, key string) (interface{}, error) {
	typeName, hasTypeHint := compiled.typeHints[key]
	if !hasTypeHint {
		return match, nil
	}

	switch typeName {
	case "int":
		return strconv.Atoi(match)

	case "float":
		return strconv.ParseFloat(match, 64)

	case "string":
		return match, nil

	default:
		return nil, fmt.Errorf("ERROR the value %s cannot be converted to %s. Must be int, float, string or empty", match, typeName)
	}
}
