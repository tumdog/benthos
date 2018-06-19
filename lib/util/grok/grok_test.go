package grok

import (
	"reflect"
	"testing"

	"github.com/Jeffail/benthos/lib/util/grok/patterns"
)

func TestNew(t *testing.T) {
	g, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if len(g.patterns) == 0 {
		t.Errorf("Expected more than %v patterns", len(g.patterns))
	}

	g, err = New(Config{SkipDefaultPatterns: true})
	if err != nil {
		t.Fatal(err)
	}
	if exp, act := 0, len(g.patterns); exp != act {
		t.Errorf("%v != %v", act, exp)
	}

	g, err = New(Config{Patterns: patterns.AWS})
	if err != nil {
		t.Fatal(err)
	}

	g, err = New(Config{Patterns: patterns.Grok})
	if err != nil {
		t.Fatal(err)
	}

	g, err = New(Config{Patterns: patterns.Firewalls})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDayCompile(t *testing.T) {
	g, err := New(Config{Patterns: map[string]string{
		"DAY": "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)",
	}})
	if err != nil {
		t.Fatal(err)
	}

	_, err = g.Compile("%{DAY}")
	if err != nil {
		t.Fatal(err)
	}
}

func TestErrorCompile(t *testing.T) {
	g, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = g.Compile("(")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestParseTypedWithDefaultCaptureMode(t *testing.T) {
	g, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}

	act, err := g.ParseTyped("%{IPV4:ip:string} %{NUMBER:status:int} %{NUMBER:duration:float}", []byte(`127.0.0.1 200 0.8`))
	if err != nil {
		t.Fatal(err)
	}

	exp := map[string]interface{}{
		"ip":       "127.0.0.1",
		"status":   200,
		"duration": 0.8,
	}
	if !reflect.DeepEqual(exp, act) {
		t.Errorf("Unexpected result: %s != %s", act, exp)
	}
}

func TestParseTypedWithNoTypeInfo(t *testing.T) {
	g, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}

	act, err := g.ParseTyped("%{COMMONAPACHELOG}", []byte(`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`))
	if err != nil {
		t.Fatal(err)
	}
	exp := map[string]interface{}{
		"timestamp":   "23/Apr/2014:22:58:32 +0200",
		"request":     "/index.php",
		"response":    "404",
		"auth":        "-",
		"ident":       "-",
		"verb":        "GET",
		"httpversion": "1.1",
		"rawrequest":  "",
		"bytes":       "207",
		"clientip":    "127.0.0.1",
	}
	if !reflect.DeepEqual(exp, act) {
		t.Errorf("Unexpected result: %s != %s", act, exp)
	}
}

func TestParseTypedWithIntegerTypeCoercion(t *testing.T) {
	g, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}

	act, err := g.ParseTyped("%{WORD:coerced:int}", []byte(`5.75`))
	if err != nil {
		t.Fatal(err)
	}
	exp := map[string]interface{}{
		"coerced": 5,
	}
	if !reflect.DeepEqual(exp, act) {
		t.Errorf("Unexpected result: %s != %s", act, exp)
	}
}

func TestParseTypedWithUnknownType(t *testing.T) {
	g, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = g.ParseTyped("%{WORD:word:unknown}", []byte(`hello`))
	if err == nil {
		t.Error("Expected err")
	}
}

func TestParseTypedErrorCaptureUnknowPattern(t *testing.T) {
	g, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = g.ParseTyped("%{UNKNOWPATTERN}", []byte(""))
	if err == nil {
		t.Error("Expected err")
	}
}

func TestParseTypedWithTypedParents(t *testing.T) {
	g, err := New(Config{
		Patterns: map[string]string{
			"TESTCOMMON": `%{IPORHOST:clientip} %{USER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest})" %{NUMBER:response} (?:%{NUMBER:bytes:int}|-)`,
		}})
	if err != nil {
		t.Fatal(err)
	}

	act, err := g.ParseTyped("%{TESTCOMMON}", []byte(`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`))
	if err != nil {
		t.Fatal(err)
	}
	exp := map[string]interface{}{
		"verb":        "GET",
		"httpversion": "1.1",
		"response":    "404",
		"bytes":       207,
		"clientip":    "127.0.0.1",
		"auth":        "-",
		"request":     "/index.php",
		"rawrequest":  "",
		"ident":       "-",
		"timestamp":   "23/Apr/2014:22:58:32 +0200",
	}
	if !reflect.DeepEqual(exp, act) {
		t.Errorf("Unexpected result: %s != %s", act, exp)
	}
}

func TestParseTypedWithSemanticHomonyms(t *testing.T) {
	g, err := New(Config{
		SkipDefaultPatterns: true,
		Patterns: map[string]string{
			"BASE10NUM": `([+-]?(?:[0-9]+(?:\.[0-9]+)?)|\.[0-9]+)`,
			"NUMBER":    `(?:%{BASE10NUM})`,
			"MYNUM":     `%{NUMBER:bytes:int}`,
			"MYSTR":     `%{NUMBER:bytes:string}`,
		}})

	if err != nil {
		t.Fatal(err)
	}

	act, err := g.ParseTyped("%{MYNUM}", []byte(`207`))
	if err != nil {
		t.Fatal(err)
	}
	exp := map[string]interface{}{
		"bytes": 207,
	}
	if !reflect.DeepEqual(exp, act) {
		t.Errorf("Unexpected result: %s != %s", act, exp)
	}

	act, err = g.ParseTyped("%{MYSTR}", []byte(`207`))
	if err != nil {
		t.Fatal(err)
	}
	exp = map[string]interface{}{
		"bytes": "207",
	}
	if !reflect.DeepEqual(exp, act) {
		t.Errorf("Unexpected result: %s != %s", act, exp)
	}
}

var resultNew *Grok

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var g *Grok
	// run the check function b.N times
	for n := 0; n < b.N; n++ {
		g, _ = New(Config{})
	}
	resultNew = g
}

func BenchmarkCapturesTypedReal(b *testing.B) {
	g, _ := New(Config{})
	b.ReportAllocs()
	b.ResetTimer()
	// run the check function b.N times
	c, _ := g.Compile(`%{IPORHOST:clientip} %{USER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion:int})?|%{DATA:rawrequest})" %{NUMBER:response:int} (?:%{NUMBER:bytes:int}|-)`)
	for n := 0; n < b.N; n++ {
		c.ParseTyped([]byte(`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`))
	}
}
