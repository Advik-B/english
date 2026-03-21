package repl

import (
	"fmt"
	"strings"

	"github.com/Advik-B/english/help"
	"github.com/Advik-B/english/highlight"
)

// printHelp writes help information to the REPL output writer.
// If args is empty, shows the general help summary.
// If args contains a search query, performs fuzzy search and displays matching topics.
func (r *REPL) printHelp(args ...string) {
	// If no search query provided, show general help
	if len(args) == 0 || args[0] == "" {
		r.printGeneralHelp()
		return
	}

	// Perform fuzzy search
	query := strings.Join(args, " ")
	results := r.helpRegistry.Search(query)

	if len(results) == 0 {
		fmt.Fprintf(r.out, "No help topics found for '%s'.\n", query)
		fmt.Fprintln(r.out, "Type 'help' for general help.")
		return
	}

	// Show the best match in detail if it's a very good match
	if results[0].Score >= 700 {
		r.printDetailedHelp(results[0].Entry)
		// Show other related topics if available
		if len(results) > 1 && len(results) <= 5 {
			fmt.Fprintln(r.out, "\nRelated topics:")
			for i := 1; i < len(results) && i < 5; i++ {
				fmt.Fprintf(r.out, "  - %s: %s\n",
					results[i].Entry.Name,
					results[i].Entry.Description)
			}
		}
	} else {
		// Show multiple search results
		fmt.Fprintf(r.out, "Search results for '%s':\n", query)
		limit := 10
		if len(results) < limit {
			limit = len(results)
		}
		for i := 0; i < limit; i++ {
			entry := results[i].Entry
			fmt.Fprintf(r.out, "  %s [%s]\n", entry.Name, entry.Category)
			fmt.Fprintf(r.out, "    %s\n", entry.Description)
		}
		fmt.Fprintln(r.out, "\nType 'help <topic>' for detailed information.")
	}
}

// printGeneralHelp displays the general help summary.
func (r *REPL) printGeneralHelp() {
	fmt.Fprintln(r.out, "English REPL")
	fmt.Fprintln(r.out, strings.Repeat("─", 50))
	fmt.Fprintln(r.out, "Commands:")
	fmt.Fprintln(r.out, "  exit / quit     Exit the REPL")
	fmt.Fprintln(r.out, "  help            Show this help message")
	fmt.Fprintln(r.out, "  help <topic>    Search help for a specific topic")
	fmt.Fprintln(r.out, "")
	fmt.Fprintln(r.out, "Language tips:")
	fmt.Fprintln(r.out, "  · Statements must end with a period (.)")
	fmt.Fprintln(r.out, "  · Multi-line blocks open with 'do the following:'")
	fmt.Fprintln(r.out, "    and close with 'thats it.'")
	fmt.Fprintln(r.out, "  · If/else: 'If …, then … otherwise … thats it.'")
	fmt.Fprintln(r.out, "  · Loops:   'repeat the following while …: … thats it.'")
	fmt.Fprintln(r.out, "")
	fmt.Fprintln(r.out, "Quick examples:")
	fmt.Fprintln(r.out, "  >>> Declare x to be 5.")
	fmt.Fprintln(r.out, "  >>> Print the value of x.")
	fmt.Fprintln(r.out, "  5")
	fmt.Fprintln(r.out, "  >>> For each n in [1, 2, 3], do the following:")
	fmt.Fprintln(r.out, "  ...     Print the value of n.")
	fmt.Fprintln(r.out, "  ... thats it.")
	fmt.Fprintln(r.out, "")
	fmt.Fprintln(r.out, "Available categories:")
	categories := r.helpRegistry.AllCategories()
	for _, cat := range categories {
		entries := r.helpRegistry.EntriesByCategory(cat)
		fmt.Fprintf(r.out, "  %s (%d topics)\n", cat, len(entries))
	}
	fmt.Fprintln(r.out, "")
	fmt.Fprintln(r.out, "Try: help print, help loop, help if, help list")
}

// printDetailedHelp displays detailed information about a specific help entry.
func (r *REPL) printDetailedHelp(entry *help.HelpEntry) {
	fmt.Fprintf(r.out, "%s [%s]\n", entry.Name, entry.Category)
	fmt.Fprintln(r.out, strings.Repeat("─", 50))
	fmt.Fprintln(r.out, entry.Description)

	if entry.LongDesc != "" {
		fmt.Fprintln(r.out, "")
		fmt.Fprintln(r.out, entry.LongDesc)
	}

	if len(entry.Examples) > 0 {
		fmt.Fprintln(r.out, "")
		fmt.Fprintln(r.out, "Examples:")
		for _, example := range entry.Examples {
			// Apply syntax highlighting to examples
			highlighted := highlight.Highlight(example, r.useColor)
			fmt.Fprintf(r.out, "  %s\n", highlighted)
		}
	}

	if len(entry.Aliases) > 0 {
		fmt.Fprintln(r.out, "")
		fmt.Fprintf(r.out, "Aliases: %s\n", strings.Join(entry.Aliases, ", "))
	}

	if len(entry.SeeAlso) > 0 {
		fmt.Fprintln(r.out, "")
		fmt.Fprintf(r.out, "See also: %s\n", strings.Join(entry.SeeAlso, ", "))
	}
}

