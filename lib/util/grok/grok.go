package grok

import "github.com/Jeffail/benthos/lib/util/grok/regexp"

// Config is used to pass a set of configuration values to the grok.New function.
type Config struct {
	SkipDefaultPatterns bool
	RemoveEmptyValues   bool
	Patterns            map[string]string
}

// Grok holds a cache of known pattern substitions and acts as a builder for
// compiled grok patterns. All pattern substitutions must be passed at creation
// time and cannot be changed during runtime.
type Grok struct {
	patterns    patternMap
	removeEmpty bool
}

// New returns a Grok object that caches a given set of patterns and creates
// compiled grok patterns based on the passed configuration settings.
// You can use multiple grok objects that act independently.
func New(config Config) (*Grok, error) {
	patterns := patternMap{}

	if !config.SkipDefaultPatterns {
		// Add default patterns first so they can be referenced later
		if err := patterns.addList(DefaultPatterns); err != nil {
			return nil, err
		}
	}

	// Add passed patterns
	if err := patterns.addList(config.Patterns); err != nil {
		return nil, err
	}

	return &Grok{
		patterns:    patterns,
		removeEmpty: config.RemoveEmptyValues,
	}, nil
}

// Compile precompiles a given grok expression. This function should be used
// when a grok expression is used more than once.
func (grok Grok) Compile(pattern string) (*CompiledGrok, error) {
	grokPattern, err := newPattern(pattern, grok.patterns)
	if err != nil {
		return nil, err
	}

	compiled, err := regexp.Compile(grokPattern.expression)
	if err != nil {
		return nil, err
	}

	return &CompiledGrok{
		regexp:      compiled,
		typeHints:   grokPattern.typeHints,
		removeEmpty: grok.removeEmpty,
	}, nil
}

// ParseTyped processes the given data and returns a map containing the values
// of all named fields converted to their corresponding types. If no typehint is
// given, the value will be converted to string.
// The given pattern is compiled on every call to this function.
// If you want to call this function more than once consider using Compile.
func (grok Grok) ParseTyped(pattern string, data []byte) (map[string]interface{}, error) {
	complied, err := grok.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return complied.ParseTyped(data)
}
