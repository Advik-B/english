package lsp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Advik-B/english/help"
	"github.com/Advik-B/english/stdlib"
	"github.com/Advik-B/english/token"
	"github.com/Advik-B/english/tokeniser"
)

// This file is what the editor knows about the language, and all of it is
// derived rather than restated.
//
// The analyser used to carry two hand-written lists: 23 keyword completions
// and 14 keyword hover entries. They disagreed with each other — one had
// "always" and the other did not, one had the comparison phrases and the
// other did not — and both were missing everything added since they were
// written: try, raise, structure, import, array, lookup table, range, cast,
// continue, sleep and the politeness prefixes. Neither knew about a single one
// of the 84 standard-library functions, though the help registry documents
// every one of them with a description and examples.

// keywordSnippet is a template offered for a multi-word construct, where the
// shape is worth filling in rather than just the word.
type keywordSnippet struct {
	label   string
	detail  string
	snippet string
}

// keywordSnippets are the constructs whose shape is worth completing. Every
// other keyword is offered as a plain word, derived from the lexer's table, so
// a keyword added later is offered without being listed here.
var keywordSnippets = []keywordSnippet{
	{"Declare", "Declare a variable", "Declare ${1:name} to be ${2:value}."},
	{"Declare function", "Declare a function", "Declare function ${1:name} that takes ${2:args} and gives back a ${3:number}, and does the following:\n\t${4:statements}\nthats it."},
	{"Declare structure", "Declare a struct", "Declare ${1:Name} as a structure with the following fields:\n\t${2:field} is a ${3:number}.\nthats it."},
	{"Set", "Assign a value", "Set ${1:name} to be ${2:value}."},
	{"Print", "Print a value with a newline", "Print ${1:value}."},
	{"Write", "Print a value without a newline", "Write ${1:value}."},
	{"If", "Conditional statement", "If ${1:condition}, then\n\t${2:statements}\nthats it."},
	{"Otherwise", "Else clause", "otherwise\n\t${1:statements}"},
	{"Repeat", "Counted loop", "Repeat the following ${1:count} times:\n\t${2:statements}\nthats it."},
	{"Repeat while", "Conditional loop", "Repeat the following while ${1:condition}:\n\t${2:statements}\nthats it."},
	{"For each", "Iterate a collection", "For each ${1:item} in ${2:collection}, do the following:\n\t${3:statements}\nthats it."},
	{"Call", "Call a function", "Call ${1:function} with ${2:args}."},
	{"Return", "Return from a function", "Return ${1:value}."},
	{"Try", "Handle an error", "Try doing the following:\n\t${1:statements}\non error:\n\t${2:handler}\nthats it."},
	{"Raise", "Raise an error", "Raise ${1:message} as ${2:ErrorType}."},
	{"Import", "Import another file", "Import \"${1:file.abc}\"."},
	{"Ask", "Read a line of input", "Ask ${1:prompt} as ${2:name}."},
	{"the result of calling", "Use a function's result", "the result of calling ${1:function} with ${2:args}"},
	{"the item at position", "Read a collection item", "the item at position ${1:index} in ${2:collection}"},
	{"the length of", "Count items or characters", "the length of ${1:collection}"},
	{"the remainder of", "Remainder after division", "the remainder of ${1:a} divided by ${2:b}"},
	{"a new instance of", "Build a struct value", "a new instance of ${1:Name} with the following fields:\n\t${2:field} is ${3:value}.\nthats it."},
	{"cast to", "Convert a value", "cast to ${1:number}"},
}

// derivedKeywords are every keyword and multi-word operator the lexer knows,
// minus the ones a snippet already covers.
var derivedKeywords = func() []string {
	covered := make(map[string]bool, len(keywordSnippets))
	for _, s := range keywordSnippets {
		covered[strings.ToLower(s.label)] = true
	}

	var out []string
	for t := token.Type(0); t <= token.COMMENT; t++ {
		word, ok := tokeniser.Spelling(t)
		if !ok || covered[word] {
			continue
		}
		out = append(out, word)
	}
	sort.Strings(out)
	return out
}()

// getKeywordCompletions offers the snippets and then every other keyword.
func (a *Analyzer) getKeywordCompletions(prefix string) []CompletionItem {
	items := make([]CompletionItem, 0, len(keywordSnippets)+len(derivedKeywords))

	for _, kw := range keywordSnippets {
		if !matchesPrefix(kw.label, prefix) {
			continue
		}
		items = append(items, CompletionItem{
			Label:            kw.label,
			Kind:             CompletionItemKindKeyword,
			Detail:           kw.detail,
			InsertText:       kw.snippet,
			InsertTextFormat: InsertTextFormatSnippet,
		})
	}

	for _, word := range derivedKeywords {
		if !matchesPrefix(word, prefix) {
			continue
		}
		item := CompletionItem{
			Label:  word,
			Kind:   CompletionItemKindKeyword,
			Detail: "keyword",
		}
		if entry := a.knowledge.Lookup(word); entry != nil {
			item.Detail = entry.Description
		}
		items = append(items, item)
	}

	return items
}

// getBuiltinCompletions offers the standard library, which the editor
// previously knew nothing about.
func (a *Analyzer) getBuiltinCompletions(prefix string) []CompletionItem {
	items := make([]CompletionItem, 0)
	for _, sig := range stdlib.Signatures() {
		if !matchesPrefix(sig.Name, prefix) {
			continue
		}
		detail := stdlib.Usage(sig)
		documentation := ""
		if entry := a.knowledge.Lookup(sig.Name); entry != nil {
			documentation = builtinDocumentation(entry, sig)
		}
		items = append(items, CompletionItem{
			Label:         sig.Name,
			Kind:          CompletionItemKindFunction,
			Detail:        detail,
			Documentation: documentation,
			InsertText:    sig.Name,
		})
	}
	return items
}

// builtinDocumentation renders a built-in's help entry as markdown.
func builtinDocumentation(entry *help.HelpEntry, sig stdlib.Signature) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**%s**\n\n`%s`\n\n%s", sig.Name, stdlib.Usage(sig), entry.Description)
	if entry.LongDesc != "" {
		b.WriteString("\n\n" + entry.LongDesc)
	}
	if len(entry.Examples) > 0 {
		b.WriteString("\n\n```english\n")
		for _, example := range entry.Examples {
			b.WriteString(example + "\n")
		}
		b.WriteString("```")
	}
	return b.String()
}

// getKeywordDocumentation returns hover text for a word, from the same help
// registry that backs "english help-topic".
//
// The analyser used to hold its own map of 14 entries, so hovering anything
// else — including every standard-library function — showed nothing.
func (a *Analyzer) getKeywordDocumentation(word string) string {
	entry := a.knowledge.Lookup(strings.ToLower(word))
	if entry == nil {
		return ""
	}
	if sig, ok := stdlib.Lookup(entry.Name); ok {
		return builtinDocumentation(entry, sig)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "**%s**\n\n%s", entry.Name, entry.Description)
	if entry.LongDesc != "" {
		b.WriteString("\n\n" + entry.LongDesc)
	}
	if len(entry.Examples) > 0 {
		b.WriteString("\n\n```english\n")
		for _, example := range entry.Examples {
			b.WriteString(example + "\n")
		}
		b.WriteString("```")
	}
	return b.String()
}

// matchesPrefix reports whether a label should be offered for a typed prefix.
func matchesPrefix(label, prefix string) bool {
	return prefix == "" || strings.HasPrefix(strings.ToLower(label), prefix)
}
