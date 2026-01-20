package cmd

import (
	"english/lsp"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kujtimiihoxha/vimtea"
	"github.com/spf13/cobra"
)

var ideCmd = &cobra.Command{
	Use:   "ide [directory]",
	Short: "Start the TUI-based IDE",
	Long: `Start the TUI-based IDE with vim-like keybindings.

Features:
  - File browser for navigation
  - Vim-like text editor with syntax highlighting
  - Code outline showing functions and variables
  - Console/IO pane for program output
  - Full LSP integration with autocomplete
  - Diagnostics and error reporting`,
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}
		StartIDE(dir)
	},
}

func init() {
	rootCmd.AddCommand(ideCmd)
}

// Pane represents which pane is currently focused
type Pane int

const (
	FileBrowserPane Pane = iota
	EditorPane
	OutlinePane
	ConsolePane
)

// ideModel represents the IDE state
type ideModel struct {
	// Window dimensions
	width  int
	height int

	// Current focused pane
	focusedPane Pane

	// Panes
	filePicker   filepicker.Model
	editor       vimtea.Editor
	outline      list.Model
	console      viewport.Model
	
	// LSP client
	analyzer     *lsp.Analyzer
	currentFile  string
	diagnostics  []lsp.Diagnostic

	// State
	quitting     bool
	ready        bool
	showHelp     bool
}

// Keybindings
type keyMap struct {
	Quit          key.Binding
	TogglePane    key.Binding
	NextPane      key.Binding
	PrevPane      key.Binding
	Help          key.Binding
	RunFile       key.Binding
	SaveFile      key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c/esc", "quit"),
	),
	TogglePane: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "cycle panes"),
	),
	NextPane: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next pane"),
	),
	PrevPane: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous pane"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	RunFile: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "run file"),
	),
	SaveFile: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save file"),
	),
}

// Styles
var (
	focusedBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	unfocusedBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	ideTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	ideHelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	diagnosticErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	diagnosticWarningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)
)

// OutlineItem represents an item in the code outline
type outlineItem struct {
	name     string
	itemType string
	line     int
}

func (i outlineItem) Title() string       { return i.name }
func (i outlineItem) Description() string { return fmt.Sprintf("%s (line %d)", i.itemType, i.line) }
func (i outlineItem) FilterValue() string { return i.name }

// initialIDEModel creates the initial IDE model
func initialIDEModel(dir string) ideModel {
	// File picker
	fp := filepicker.New()
	fp.CurrentDirectory = dir
	fp.AllowedTypes = []string{".abc", ".101", ".go", ".txt", ".md"}
	fp.DirAllowed = true
	fp.FileAllowed = true

	// Editor
	editor := vimtea.NewEditor()
	
	// Outline list
	items := []list.Item{}
	delegate := list.NewDefaultDelegate()
	outlineList := list.New(items, delegate, 0, 0)
	outlineList.Title = "Code Outline"
	outlineList.SetShowStatusBar(false)
	outlineList.SetFilteringEnabled(false)

	// Console viewport
	consoleVp := viewport.New(0, 0)
	consoleVp.SetContent("Console Output\n\nPress Ctrl+R to run the current file.\n")

	// LSP analyzer
	analyzer := lsp.NewAnalyzer()

	return ideModel{
		filePicker:  fp,
		editor:      editor,
		outline:     outlineList,
		console:     consoleVp,
		analyzer:    analyzer,
		focusedPane: FileBrowserPane,
		ready:       false,
		showHelp:    false,
	}
}

func (m ideModel) Init() tea.Cmd {
	return tea.Batch(
		m.filePicker.Init(),
		tea.EnterAltScreen,
	)
}

func (m ideModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Calculate pane sizes
		// Left column: 25% for file browser
		// Middle column: 50% for editor
		// Right column: 25% split between outline and console
		leftWidth := m.width / 4
		middleWidth := m.width / 2
		rightWidth := m.width - leftWidth - middleWidth

		// Heights
		mainHeight := m.height - 4 // Reserve space for title and status
		outlineHeight := mainHeight / 2
		consoleHeight := mainHeight - outlineHeight

		// Update component sizes
		m.outline.SetSize(rightWidth-4, outlineHeight-3)
		m.console.Width = rightWidth - 4
		m.console.Height = consoleHeight - 3

		return m, nil

	case tea.KeyMsg:
		// Global keybindings
		if key.Matches(msg, keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}

		if key.Matches(msg, keys.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}

		if key.Matches(msg, keys.TogglePane) || key.Matches(msg, keys.NextPane) {
			m.focusedPane = (m.focusedPane + 1) % 4
			return m, nil
		}

		if key.Matches(msg, keys.PrevPane) {
			m.focusedPane = (m.focusedPane + 3) % 4
			return m, nil
		}

		if key.Matches(msg, keys.RunFile) && m.currentFile != "" {
			return m, m.runCurrentFile()
		}

		if key.Matches(msg, keys.SaveFile) && m.currentFile != "" {
			return m, m.saveCurrentFile()
		}

		// Delegate to focused pane
		switch m.focusedPane {
		case FileBrowserPane:
			m.filePicker, cmd = m.filePicker.Update(msg)
			cmds = append(cmds, cmd)

			// Check if a file was selected
			if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
				return m, m.openFile(path)
			}

		case EditorPane:
			if m.editor != nil {
				var model tea.Model
				model, cmd = m.editor.Update(msg)
				m.editor = model.(vimtea.Editor)
				cmds = append(cmds, cmd)
			}

		case OutlinePane:
			m.outline, cmd = m.outline.Update(msg)
			cmds = append(cmds, cmd)

		case ConsolePane:
			m.console, cmd = m.console.Update(msg)
			cmds = append(cmds, cmd)
		}

	default:
		// Update all components
		if m.focusedPane == FileBrowserPane {
			m.filePicker, cmd = m.filePicker.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.focusedPane == EditorPane && m.editor != nil {
			var model tea.Model
			model, cmd = m.editor.Update(msg)
			m.editor = model.(vimtea.Editor)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m ideModel) View() string {
	if !m.ready {
		return "\n  Initializing IDE..."
	}

	if m.quitting {
		return "Goodbye!\n"
	}

	// Calculate pane dimensions
	leftWidth := m.width / 4
	middleWidth := m.width / 2
	rightWidth := m.width - leftWidth - middleWidth
	mainHeight := m.height - 4

	// Title bar
	title := ideTitleStyle.Render("English Language IDE")
	statusBar := m.renderStatusBar()

	// Render panes
	fileBrowserView := m.renderFileBrowser(leftWidth, mainHeight)
	editorView := m.renderEditor(middleWidth, mainHeight)
	rightColumn := m.renderRightColumn(rightWidth, mainHeight)

	// Combine panes horizontally
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		fileBrowserView,
		editorView,
		rightColumn,
	)

	// Help text
	helpText := ""
	if m.showHelp {
		helpText = m.renderHelp()
	}

	// Combine all sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		mainView,
		statusBar,
		helpText,
	)
}

func (m ideModel) renderFileBrowser(width, height int) string {
	style := unfocusedBorderStyle
	if m.focusedPane == FileBrowserPane {
		style = focusedBorderStyle
	}

	title := " File Browser "
	content := m.filePicker.View()

	// Truncate if needed
	lines := strings.Split(content, "\n")
	if len(lines) > height-3 {
		lines = lines[:height-3]
	}
	content = strings.Join(lines, "\n")

	return style.
		Width(width-2).
		Height(height-1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m ideModel) renderEditor(width, height int) string {
	style := unfocusedBorderStyle
	if m.focusedPane == EditorPane {
		style = focusedBorderStyle
	}

	title := " Editor "
	if m.currentFile != "" {
		title = fmt.Sprintf(" Editor - %s ", filepath.Base(m.currentFile))
	}

	content := ""
	if m.editor != nil {
		content = m.editor.View()
	} else {
		content = "No file open\n\nPress Tab to switch to file browser and select a file."
	}

	// Add diagnostics if any
	if len(m.diagnostics) > 0 {
		diagnosticsView := "\n\n" + m.renderDiagnostics()
		content += diagnosticsView
	}

	return style.
		Width(width-2).
		Height(height-1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m ideModel) renderRightColumn(width, height int) string {
	outlineHeight := height / 2
	consoleHeight := height - outlineHeight

	outlineView := m.renderOutline(width, outlineHeight)
	consoleView := m.renderConsole(width, consoleHeight)

	return lipgloss.JoinVertical(lipgloss.Left, outlineView, consoleView)
}

func (m ideModel) renderOutline(width, height int) string {
	style := unfocusedBorderStyle
	if m.focusedPane == OutlinePane {
		style = focusedBorderStyle
	}

	title := " Code Outline "
	content := m.outline.View()

	return style.
		Width(width-2).
		Height(height-1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m ideModel) renderConsole(width, height int) string {
	style := unfocusedBorderStyle
	if m.focusedPane == ConsolePane {
		style = focusedBorderStyle
	}

	title := " Console "
	content := m.console.View()

	return style.
		Width(width-2).
		Height(height-1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m ideModel) renderStatusBar() string {
	leftStatus := ""
	switch m.focusedPane {
	case FileBrowserPane:
		leftStatus = "File Browser"
	case EditorPane:
		leftStatus = "Editor"
	case OutlinePane:
		leftStatus = "Outline"
	case ConsolePane:
		leftStatus = "Console"
	}

	rightStatus := "Press ? for help"
	if m.currentFile != "" {
		rightStatus = fmt.Sprintf("%s | %s", filepath.Base(m.currentFile), rightStatus)
	}

	statusWidth := m.width - len(leftStatus) - len(rightStatus) - 2
	if statusWidth < 0 {
		statusWidth = 0
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf(" %s%s%s ", leftStatus, strings.Repeat(" ", statusWidth), rightStatus))
}

func (m ideModel) renderDiagnostics() string {
	var diagnostics []string
	for _, diag := range m.diagnostics {
		style := diagnosticErrorStyle
		if diag.Severity == lsp.DiagnosticSeverityWarning {
			style = diagnosticWarningStyle
		}
		diagnostics = append(diagnostics, style.Render(fmt.Sprintf(
			"Line %d: %s", diag.Range.Start.Line+1, diag.Message,
		)))
	}
	return strings.Join(diagnostics, "\n")
}

func (m ideModel) renderHelp() string {
	help := `
Keybindings:
  Tab/Shift+Tab  - Switch between panes
  Ctrl+W         - Cycle through panes
  Ctrl+R         - Run current file
  Ctrl+S         - Save current file
  Ctrl+C / Esc   - Quit
  ?              - Toggle this help

Editor (Vim mode):
  i              - Insert mode
  Esc            - Normal mode
  v              - Visual mode
  :              - Command mode
  h/j/k/l        - Navigation
  dd             - Delete line
  yy             - Yank line
  p              - Paste
  u              - Undo
  Ctrl+R         - Redo
`
	return ideHelpStyle.Render(help)
}

// Commands for file operations
func (m *ideModel) openFile(path string) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return fileOpenError{err}
		}

		// Create new editor with content
		m.currentFile = path
		m.editor = vimtea.NewEditor(
			vimtea.WithContent(string(content)),
			vimtea.WithFileName(filepath.Base(path)),
		)

		// Analyze file with LSP
		if strings.HasSuffix(path, ".abc") {
			m.analyzeFile(string(content))
		}

		return fileOpenSuccess{path}
	}
}

func (m *ideModel) analyzeFile(content string) {
	// Create a document for the analyzer
	doc := lsp.NewDocument(m.currentFile, "english", 1, content)
	
	// Analyze the program
	result := m.analyzer.Analyze(doc)

	// Update outline
	var items []list.Item
	for _, symbol := range result.Symbols {
		itemType := "Variable"
		if symbol.Type == lsp.SymbolTypeFunction {
			itemType = "Function"
		}
		items = append(items, outlineItem{
			name:     symbol.Name,
			itemType: itemType,
			line:     symbol.Range.Start.Line + 1,
		})
	}
	m.outline.SetItems(items)

	// Update diagnostics
	m.diagnostics = result.Diagnostics
}

func (m *ideModel) saveCurrentFile() tea.Cmd {
	return func() tea.Msg {
		if m.currentFile == "" {
			return fileSaveError{fmt.Errorf("no file open")}
		}

		// Get content from buffer
		content := m.editor.GetBuffer().Text()
		err := os.WriteFile(m.currentFile, []byte(content), 0644)
		if err != nil {
			return fileSaveError{err}
		}

		return fileSaveSuccess{m.currentFile}
	}
}

func (m *ideModel) runCurrentFile() tea.Cmd {
	return func() tea.Msg {
		if m.currentFile == "" {
			return fileRunError{fmt.Errorf("no file open")}
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run the file
		RunFile(m.currentFile)

		// Restore stdout and read output
		w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)

		return fileRunSuccess{string(output)}
	}
}

// Messages
type fileOpenSuccess struct{ path string }
type fileOpenError struct{ err error }
type fileSaveSuccess struct{ path string }
type fileSaveError struct{ err error }
type fileRunSuccess struct{ output string }
type fileRunError struct{ err error }

// StartIDE starts the TUI-based IDE
func StartIDE(dir string) {
	p := tea.NewProgram(
		initialIDEModel(dir),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running IDE: %v\n", err)
		os.Exit(1)
	}
}
