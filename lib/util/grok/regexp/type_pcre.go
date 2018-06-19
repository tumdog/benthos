// Copyright (c) 2018 Ashley Jeffs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// +build PCRE

package regexp

import (
	"errors"

	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
)

type compiledPCRE struct {
	p pcre.Regexp
	m *pcre.Matcher
}

func (c *compiledPCRE) FindSubmatch(data []byte) [][]byte {
	c.m = c.p.Matcher(data, 0)
	if !c.m.Matches() {
		return [][]byte{}
	}
	matches := make([][]byte, c.m.Groups())
	for i := range matches {
		matches[i] = c.m.Group(i + 1)
	}
	return matches
}

func (c *compiledPCRE) SubexpNames() []string {
	if c.m == nil {
		return []string{}
	}
	names := []string{}
	for i := 0; i < c.m.Groups(); i++ {
		names = append(names, c.m.GroupString(i+1))
	}
	return names
}

// Compile attempts to compile the regular expression.
func Compile(expr string) (Compiled, error) {
	p, err := pcre.Compile(expr, 0)
	if err != nil {
		return nil, errors.New(err.String())
	}
	return &compiledPCRE{p: p}, nil
}
