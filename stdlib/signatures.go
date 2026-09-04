package stdlib

import (
	"fmt"
	"sort"

	vm "github.com/Advik-B/english/astvm"
)

// Signature describes a built-in function's calling convention.
//
// This table is the single source of truth for which built-ins exist and how
// many arguments they take.  Registration (Register), dispatch (Eval) and
// arity checking all derive from it, so a built-in can never be registered
// with one arity and implemented with another.
type Signature struct {
	// Name is the function name as written in English source.
	Name string
	// Params are the parameter names in order, used for registration metadata
	// and error messages.
	Params []string
	// Optional is the number of trailing parameters that may be omitted.
	Optional int
	// Module selects the implementation dispatcher.
	Module string
}

// MinArgs is the smallest legal argument count.
func (s Signature) MinArgs() int { return len(s.Params) - s.Optional }

// MaxArgs is the largest legal argument count.
func (s Signature) MaxArgs() int { return len(s.Params) }

// sig is a small constructor keeping the table below readable.
func sig(module, name string, params ...string) Signature {
	return Signature{Name: name, Params: params, Module: module}
}

// optSig is like sig but marks the last opt parameters as optional.
func optSig(module, name string, opt int, params ...string) Signature {
	return Signature{Name: name, Params: params, Optional: opt, Module: module}
}

// signatureList declares every built-in exactly once.
var signatureList = func() []Signature {
	var out []Signature

	// ── Math ──────────────────────────────────────────────────────────────
	for _, n := range []string{
		"sqrt", "abs", "floor", "ceil", "round",
		"sin", "cos", "tan", "log", "log10", "log2", "exp",
		"is_nan", "is_infinite",
	} {
		out = append(out, sig("math", n, "x"))
	}
	out = append(out,
		sig("math", "pow", "base", "exponent"),
		sig("math", "min", "a", "b"),
		sig("math", "max", "a", "b"),
		sig("math", "random"),
		sig("math", "random_between", "min", "max"),
	)

	// ── Strings ───────────────────────────────────────────────────────────
	for _, n := range []string{
		"uppercase", "lowercase", "casefold", "trim", "to_number", "to_string",
		"is_empty", "title", "capitalize", "swapcase", "trim_left", "trim_right",
		"is_digit", "is_alpha", "is_alnum", "is_space", "is_upper", "is_lower",
	} {
		out = append(out, sig("string", n, "text"))
	}
	out = append(out,
		sig("string", "split", "text", "separator"),
		sig("string", "join", "list", "separator"),
		sig("string", "replace", "text", "old", "new"),
		sig("string", "contains", "text", "substring"),
		sig("string", "starts_with", "text", "prefix"),
		sig("string", "ends_with", "text", "suffix"),
		sig("string", "index_of", "text", "search"),
		sig("string", "substring", "text", "start", "length"),
		sig("string", "str_repeat", "text", "n"),
		sig("string", "count_occurrences", "text", "substring"),
		optSig("string", "pad_left", 1, "text", "width", "char"),
		optSig("string", "pad_right", 1, "text", "width", "char"),
		optSig("string", "center", 1, "text", "width", "char"),
		sig("string", "zfill", "text", "width"),
	)

	// ── Number ────────────────────────────────────────────────────────────
	out = append(out,
		sig("number", "is_integer", "x"),
		sig("number", "clamp", "x", "min", "max"),
		sig("number", "sign", "x"),
	)

	// ── List ──────────────────────────────────────────────────────────────
	for _, n := range []string{
		"sort", "reverse", "sum", "unique", "first", "last", "flatten", "count",
		"average", "min_value", "max_value", "any_true", "all_true", "product",
		"sorted_desc",
	} {
		out = append(out, sig("list", n, "list"))
	}
	out = append(out,
		sig("list", "append", "list", "item"),
		sig("list", "remove", "list", "index"),
		sig("list", "insert", "list", "index", "item"),
		sig("list", "slice", "list", "start", "end"),
		sig("list", "zip_with", "list", "other"),
	)

	// ── I/O ───────────────────────────────────────────────────────────────
	out = append(out, optSig("io", "ask", 1, "prompt"))

	// ── Lookup table ──────────────────────────────────────────────────────
	out = append(out,
		sig("lookup", "keys", "table"),
		sig("lookup", "values", "table"),
		sig("lookup", "table_remove", "table", "key"),
		sig("lookup", "table_has", "table", "key"),
		sig("lookup", "merge", "table", "other"),
		sig("lookup", "get_or_default", "table", "key", "default"),
	)

	// ── Time ──────────────────────────────────────────────────────────────
	out = append(out,
		sig("time", "current_time"),
		sig("time", "elapsed_time"),
		sig("time", "sleep", "seconds"),
	)

	return out
}()

// signatures indexes signatureList by function name.
var signatures = func() map[string]Signature {
	m := make(map[string]Signature, len(signatureList))
	for _, s := range signatureList {
		if _, dup := m[s.Name]; dup {
			panic("stdlib: duplicate builtin signature for " + s.Name)
		}
		m[s.Name] = s
	}
	return m
}()

// Lookup returns the signature of a built-in, and whether it exists.
func Lookup(name string) (Signature, bool) {
	s, ok := signatures[name]
	return s, ok
}

// Names returns every built-in name, sorted. Used by tooling that needs the
// authoritative list (completions, help, transpiler coverage tests).
func Names() []string {
	out := make([]string, 0, len(signatures))
	for n := range signatures {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// checkArity validates an argument count against a built-in's signature.
// It is called before dispatch so that no implementation ever indexes past the
// end of args — which previously produced an uncaught Go panic.
func checkArity(name string, got int) error {
	s, ok := signatures[name]
	if !ok {
		return nil // unknown name; the dispatcher reports it
	}
	minA, maxA := s.MinArgs(), s.MaxArgs()
	if got >= minA && got <= maxA {
		return nil
	}
	return vm.NewRuntimeError(fmt.Sprintf(
		"%s expects %s, got %s\n  Usage: %s",
		name, plural(minA, maxA), countWord(got), usage(s),
	))
}

// plural renders an expected-argument-count range in English.
func plural(minA, maxA int) string {
	switch {
	case minA == maxA:
		return countWord(minA)
	case maxA == minA+1:
		return fmt.Sprintf("%s or %s", countWord(minA), countWord(maxA))
	default:
		return fmt.Sprintf("between %s and %s", countWord(minA), countWord(maxA))
	}
}

func countWord(n int) string {
	switch n {
	case 0:
		return "no arguments"
	case 1:
		return "1 argument"
	default:
		return fmt.Sprintf("%d arguments", n)
	}
}

// usage renders a call template, e.g. `pad_left(text, width[, char])`.
func usage(s Signature) string {
	out := s.Name + "("
	for i, p := range s.Params {
		optional := i >= s.MinArgs()
		switch {
		case i == 0 && optional:
			out += "["
		case i == 0:
			// first required parameter needs no separator
		case optional:
			out += "[, "
		default:
			out += ", "
		}
		out += p
	}
	if s.Optional > 0 {
		for i := 0; i < s.Optional; i++ {
			out += "]"
		}
	}
	return out + ")"
}
