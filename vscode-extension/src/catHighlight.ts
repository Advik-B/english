import * as vscode from 'vscode';

type TokenKind =
  | 'keyword'
  | 'controlFlow'
  | 'declaration'
  | 'string'
  | 'number'
  | 'bool'
  | 'null'
  | 'identifier'
  | 'operator'
  | 'comparison'
  | 'possessive'
  | 'punctuation'
  | 'comment'
  | 'invalid';

type Span = { start: number; end: number };

const comparisonPhrases = [
  'is less than or equal to',
  'is greater than or equal to',
  'is not equal to',
  'is less than',
  'is greater than',
  'is equal to',
  "isn't true",
  'isn’t true',
  "isn't false",
  'isn’t false',
  'is true',
  'is false',
  'is nothing',
  'is something'
];

const controlFlowPhrases: Array<{ phrase: string; kind: TokenKind }> = [
  { phrase: 'on error', kind: 'controlFlow' },
  { phrase: 'thats it', kind: 'controlFlow' },
  { phrase: "that's it", kind: 'controlFlow' }
];

const declarationPhrases: Array<{ phrase: string; kind: TokenKind }> = [
  { phrase: 'declare function', kind: 'declaration' }
];

const controlFlowKeywords = new Set([
  'if', 'then', 'otherwise', 'repeat', 'while', 'forever', 'for', 'each', 'do',
  'break', 'continue', 'skip', 'return', 'sleep', 'out', 'loop', 'times'
]);

const declarationKeywords = new Set([
  'declare', 'let', 'function', 'set', 'call', 'calling', 'takes', 'please'
]);

const keywordWords = new Set([
  'always', 'to', 'be', 'import', 'everything', 'all', 'safely', 'as', 'structure',
  'struct', 'fields', 'field', 'instance', 'new', 'raise', 'toggle', 'ask', 'array',
  'lookup', 'table', 'range', 'reference', 'copy', 'swap', 'casted', 'cast', 'type',
  'from', 'default', 'with', 'without', 'and', 'or', 'not', 'in'
]);

const boolWords = new Set(['true', 'false']);
const nullWords = new Set(['nothing', 'none', 'null']);

function isWordChar(ch: string): boolean {
  return /[A-Za-z0-9_]/.test(ch);
}

function isWordBoundary(text: string, index: number): boolean {
  if (index < 0 || index >= text.length) {
    return true;
  }
  return !isWordChar(text[index]);
}

function matchPhrase(text: string, index: number, phrase: string): boolean {
  const slice = text.slice(index, index + phrase.length);
  if (slice.toLowerCase() !== phrase) {
    return false;
  }
  return isWordBoundary(text, index - 1) && isWordBoundary(text, index + phrase.length);
}

function addSpan(spans: Record<TokenKind, Span[]>, kind: TokenKind, start: number, end: number): void {
  spans[kind].push({ start, end });
}

export function tokenizeForCatHighlight(text: string): Record<TokenKind, Span[]> {
  const spans: Record<TokenKind, Span[]> = {
    keyword: [],
    controlFlow: [],
    declaration: [],
    string: [],
    number: [],
    bool: [],
    null: [],
    identifier: [],
    operator: [],
    comparison: [],
    possessive: [],
    punctuation: [],
    comment: [],
    invalid: []
  };

  let i = 0;
  while (i < text.length) {
    const ch = text[i];

    if (ch === '\n' || ch === '\r' || ch === ' ' || ch === '\t') {
      i++;
      continue;
    }

    if (ch === '#') {
      const start = i;
      while (i < text.length && text[i] !== '\n') {
        i++;
      }
      addSpan(spans, 'comment', start, i);
      continue;
    }

    if (text.startsWith("'s", i) && isWordBoundary(text, i + 2)) {
      addSpan(spans, 'possessive', i, i + 2);
      i += 2;
      continue;
    }

    if (ch === '"' || ch === '\'') {
      const quote = ch;
      const start = i;
      i++;
      while (i < text.length) {
        if (text[i] === '\\') {
          i += 2;
          continue;
        }
        if (text[i] === quote) {
          i++;
          break;
        }
        i++;
      }
      addSpan(spans, 'string', start, i);
      continue;
    }

    let matched = false;
    for (const phrase of comparisonPhrases) {
      if (matchPhrase(text, i, phrase)) {
        addSpan(spans, 'comparison', i, i + phrase.length);
        i += phrase.length;
        matched = true;
        break;
      }
    }
    if (matched) {
      continue;
    }

    for (const entry of declarationPhrases) {
      if (matchPhrase(text, i, entry.phrase)) {
        addSpan(spans, entry.kind, i, i + entry.phrase.length);
        i += entry.phrase.length;
        matched = true;
        break;
      }
    }
    if (matched) {
      continue;
    }

    for (const entry of controlFlowPhrases) {
      if (matchPhrase(text, i, entry.phrase)) {
        addSpan(spans, entry.kind, i, i + entry.phrase.length);
        i += entry.phrase.length;
        matched = true;
        break;
      }
    }
    if (matched) {
      continue;
    }

    if (text.startsWith('..', i)) {
      addSpan(spans, 'operator', i, i + 2);
      i += 2;
      continue;
    }

    if ('+-*/='.includes(ch)) {
      addSpan(spans, 'operator', i, i + 1);
      i++;
      continue;
    }

    if ('.,:()[]'.includes(ch)) {
      addSpan(spans, 'punctuation', i, i + 1);
      i++;
      continue;
    }

    const numberMatch = /^\d+(?:\.\d+)?\b/.exec(text.slice(i));
    if (numberMatch) {
      const len = numberMatch[0].length;
      addSpan(spans, 'number', i, i + len);
      i += len;
      continue;
    }

    const wordMatch = /^[A-Za-z_][A-Za-z0-9_]*/.exec(text.slice(i));
    if (wordMatch) {
      const word = wordMatch[0];
      const lower = word.toLowerCase();
      const end = i + word.length;
      if (boolWords.has(lower)) {
        addSpan(spans, 'bool', i, end);
      } else if (nullWords.has(lower)) {
        addSpan(spans, 'null', i, end);
      } else if (controlFlowKeywords.has(lower)) {
        addSpan(spans, 'controlFlow', i, end);
      } else if (declarationKeywords.has(lower)) {
        addSpan(spans, 'declaration', i, end);
      } else if (keywordWords.has(lower)) {
        addSpan(spans, 'keyword', i, end);
      } else {
        addSpan(spans, 'identifier', i, end);
      }
      i = end;
      continue;
    }

    addSpan(spans, 'invalid', i, i + 1);
    i++;
  }

  return spans;
}

function toRanges(document: vscode.TextDocument, spans: Span[]): vscode.Range[] {
  return spans.map(span => new vscode.Range(document.positionAt(span.start), document.positionAt(span.end)));
}

export class CatHighlightController implements vscode.Disposable {
  private readonly disposables: vscode.Disposable[] = [];
  private readonly pending = new Map<string, NodeJS.Timeout>();
  private readonly decorations: Record<TokenKind, vscode.TextEditorDecorationType>;

  constructor() {
    this.decorations = {
      keyword: vscode.window.createTextEditorDecorationType({ color: '#BD93F9' }),
      controlFlow: vscode.window.createTextEditorDecorationType({ color: '#FF79C6' }),
      declaration: vscode.window.createTextEditorDecorationType({ color: '#8BE9FD' }),
      string: vscode.window.createTextEditorDecorationType({ color: '#50FA7B' }),
      number: vscode.window.createTextEditorDecorationType({ color: '#F1FA8C' }),
      bool: vscode.window.createTextEditorDecorationType({ color: '#FFB86C' }),
      null: vscode.window.createTextEditorDecorationType({ color: '#6272A4' }),
      identifier: vscode.window.createTextEditorDecorationType({ color: '#F8F8F2' }),
      operator: vscode.window.createTextEditorDecorationType({ color: '#FF5555' }),
      // Keep this gold to match the compiler's `cat` command comparison styling exactly.
      comparison: vscode.window.createTextEditorDecorationType({ color: '#FFD700' }),
      possessive: vscode.window.createTextEditorDecorationType({ color: '#8BE9FD' }),
      punctuation: vscode.window.createTextEditorDecorationType({ color: '#6272A4' }),
      comment: vscode.window.createTextEditorDecorationType({ color: '#6272A4', fontStyle: 'italic' }),
      invalid: vscode.window.createTextEditorDecorationType({ color: '#44475A' })
    };

    this.disposables.push(
      ...Object.values(this.decorations),
      vscode.window.onDidChangeActiveTextEditor(editor => {
        if (editor) {
          this.refreshEditor(editor);
        }
      }),
      vscode.window.onDidChangeVisibleTextEditors(editors => {
        for (const editor of editors) {
          this.refreshEditor(editor);
        }
      }),
      vscode.workspace.onDidOpenTextDocument(document => {
        this.refreshVisibleForDocument(document);
      }),
      vscode.workspace.onDidChangeTextDocument(event => {
        this.scheduleRefresh(event.document);
      }),
      vscode.workspace.onDidCloseTextDocument(document => {
        const key = document.uri.toString();
        const timeout = this.pending.get(key);
        if (timeout) {
          clearTimeout(timeout);
          this.pending.delete(key);
        }
      })
    );

    for (const editor of vscode.window.visibleTextEditors) {
      this.refreshEditor(editor);
    }
  }

  private isEnglishDocument(document: vscode.TextDocument): boolean {
    return document.languageId === 'english';
  }

  private refreshVisibleForDocument(document: vscode.TextDocument): void {
    for (const editor of vscode.window.visibleTextEditors) {
      if (editor.document.uri.toString() === document.uri.toString()) {
        this.refreshEditor(editor);
      }
    }
  }

  private scheduleRefresh(document: vscode.TextDocument): void {
    if (!this.isEnglishDocument(document)) {
      return;
    }
    const key = document.uri.toString();
    const existing = this.pending.get(key);
    if (existing) {
      clearTimeout(existing);
    }
    const timeout = setTimeout(() => {
      this.pending.delete(key);
      this.refreshVisibleForDocument(document);
    }, 75);
    this.pending.set(key, timeout);
  }

  private clearAll(editor: vscode.TextEditor): void {
    for (const decoration of Object.values(this.decorations)) {
      editor.setDecorations(decoration, []);
    }
  }

  private refreshEditor(editor: vscode.TextEditor): void {
    if (!this.isEnglishDocument(editor.document)) {
      this.clearAll(editor);
      return;
    }

    const text = editor.document.getText();
    const tokens = tokenizeForCatHighlight(text);
    for (const [kind, decoration] of Object.entries(this.decorations) as Array<[TokenKind, vscode.TextEditorDecorationType]>) {
      editor.setDecorations(decoration, toRanges(editor.document, tokens[kind]));
    }
  }

  dispose(): void {
    for (const timeout of this.pending.values()) {
      clearTimeout(timeout);
    }
    this.pending.clear();
    for (const disposable of this.disposables) {
      disposable.dispose();
    }
  }
}
