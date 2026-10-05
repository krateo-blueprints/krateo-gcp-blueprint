package main

import (
	"regexp"
	"regexp/syntax"
	"strings"
)

// sampleForPattern returns a short string that satisfies an OpenAPI `pattern`, or "" when it
// cannot derive one. It exists so the generated values.yaml stays schema-valid: `helm template`
// enforces values.schema.json, and a seeded placeholder like "example" fails any field with a
// restrictive pattern (e.g. IAMPolicyMember.role).
//
// The derivation walks the parsed regexp AST and emits the smallest match it can, then VERIFIES
// the result against the compiled pattern before returning it -- a derived string that does not
// actually match is discarded rather than written out.
func sampleForPattern(pattern string, minLength int) string {
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return ""
	}
	var b strings.Builder
	if !emitMatch(re.Simplify(), &b, 0) {
		return ""
	}
	s := b.String()

	// Pad to minLength by repeating the last rune, only when that keeps the match.
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	for minLength > 0 && len([]rune(s)) < minLength && s != "" {
		r := []rune(s)
		candidate := s + string(r[len(r)-1])
		if !compiled.MatchString(candidate) {
			break
		}
		s = candidate
	}

	if s == "" || !compiled.MatchString(s) {
		return ""
	}
	return s
}

// emitMatch appends a minimal match for one AST node. It returns false when it meets a
// construct it cannot satisfy, so the caller can fall back instead of emitting a wrong value.
func emitMatch(re *syntax.Regexp, b *strings.Builder, depth int) bool {
	if depth > 24 {
		return false
	}
	switch re.Op {
	case syntax.OpEmptyMatch, syntax.OpBeginLine, syntax.OpEndLine,
		syntax.OpBeginText, syntax.OpEndText, syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		return true

	case syntax.OpLiteral:
		b.WriteString(string(re.Rune))
		return true

	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		b.WriteRune('a')
		return true

	case syntax.OpCharClass:
		r, ok := pickFromClass(re.Rune)
		if !ok {
			return false
		}
		b.WriteRune(r)
		return true

	case syntax.OpCapture:
		return emitMatch(re.Sub[0], b, depth+1)

	case syntax.OpConcat:
		for _, sub := range re.Sub {
			if !emitMatch(sub, b, depth+1) {
				return false
			}
		}
		return true

	case syntax.OpAlternate:
		// Try each branch; keep the first that yields something.
		for _, sub := range re.Sub {
			var tmp strings.Builder
			if emitMatch(sub, &tmp, depth+1) {
				b.WriteString(tmp.String())
				return true
			}
		}
		return false

	case syntax.OpStar, syntax.OpQuest:
		// Zero repetitions is the smallest match.
		return true

	case syntax.OpPlus:
		return emitMatch(re.Sub[0], b, depth+1)

	case syntax.OpRepeat:
		for i := 0; i < re.Min; i++ {
			if !emitMatch(re.Sub[0], b, depth+1) {
				return false
			}
		}
		return true
	}
	return false
}

// pickFromClass chooses a readable rune from a char-class range table, preferring lowercase
// letters, then digits, then whatever the class starts with.
func pickFromClass(runes []rune) (rune, bool) {
	if len(runes) == 0 || len(runes)%2 != 0 {
		return 0, false
	}
	for _, want := range []rune{'a', 'x', '0'} {
		for i := 0; i < len(runes); i += 2 {
			if want >= runes[i] && want <= runes[i+1] {
				return want, true
			}
		}
	}
	for i := 0; i < len(runes); i += 2 {
		lo := runes[i]
		// Skip control/unprintable starts when the range offers something better.
		if lo >= 0x20 && lo < 0x7f {
			return lo, true
		}
	}
	return 0, false
}
