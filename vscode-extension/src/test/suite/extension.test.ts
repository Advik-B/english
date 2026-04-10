import * as assert from 'node:assert';
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as vscode from 'vscode';
import { tokenizeForCatHighlight } from '../../catHighlight';

suite('English Extension (Real VS Code)', () => {
  async function activateEnglishExtension(): Promise<vscode.Extension<unknown>> {
    const binaryPath = process.env.ENGLISH_BINARY_PATH;
    assert.ok(binaryPath, 'ENGLISH_BINARY_PATH must be set for tests');
    await vscode.workspace.getConfiguration('english.languageServer').update(
      'path',
      binaryPath,
      vscode.ConfigurationTarget.Global
    );
    await vscode.workspace.getConfiguration('english.languageServer').update(
      'args',
      ['lsp'],
      vscode.ConfigurationTarget.Global
    );

    const extension = vscode.extensions.getExtension('advik-b.english-language');
    assert.ok(extension, 'extension should be discoverable by id');
    await extension?.activate();
    assert.ok(extension?.isActive, 'extension should activate inside real VS Code');
    return extension!;
  }

  test('registers english language and activates extension', async () => {
    await activateEnglishExtension();

    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(workspaceFolder, 'workspace folder should be available for tests');
    const sampleDocUri = vscode.Uri.joinPath(workspaceFolder!.uri, 'sample.abc');
    const document = await vscode.workspace.openTextDocument(sampleDocUri);
    await vscode.window.showTextDocument(document);
    assert.strictEqual(document.languageId, 'english');
  });

  test('keeps runtime dependencies packagable for VSIX', async () => {
    const extension = await activateEnglishExtension();
    const ignorePath = path.join(extension.extensionPath, '.vscodeignore');
    const ignore = fs.readFileSync(ignorePath, 'utf8');
    assert.ok(!ignore.includes('node_modules/**'), 'node_modules must not be excluded from VSIX');
  });

  test('tokenizes compiler-cat categories consistently', async () => {
    await activateEnglishExtension();
    const sample = [
      'declare function greet takes person',
      'set person\'s age to be 42',
      'if person is greater than or equal to 18 then',
      '  # comment',
      '  call greet with "Advik"',
      'otherwise',
      '  return nothing'
    ].join('\n');
    const spans = tokenizeForCatHighlight(sample);

    const hasTokenText = (kind: keyof typeof spans, expected: string): boolean =>
      spans[kind].some(span => sample.slice(span.start, span.end).toLowerCase() === expected.toLowerCase());

    assert.ok(hasTokenText('declaration', 'declare function'));
    assert.ok(hasTokenText('possessive', "'s"));
    assert.ok(hasTokenText('comparison', 'is greater than or equal to'));
    assert.ok(hasTokenText('number', '42'));
    assert.ok(hasTokenText('string', '"Advik"'));
    assert.ok(hasTokenText('comment', '# comment'));
    assert.ok(hasTokenText('null', 'nothing'));
  });

  test('survives rapid real-editor edits in english files', async () => {
    await activateEnglishExtension();
    const document = await vscode.workspace.openTextDocument({
      language: 'english',
      content: 'declare x to be 1\n'
    });
    const editor = await vscode.window.showTextDocument(document);

    for (let i = 0; i < 25; i++) {
      await editor.edit(editBuilder => {
        editBuilder.insert(new vscode.Position(i + 1, 0), `set x to be ${i + 2} # comment ${i}\n`);
      });
    }
    await new Promise(resolve => setTimeout(resolve, 300));

    const extension = vscode.extensions.getExtension('advik-b.english-language');
    assert.ok(extension?.isActive, 'extension should remain active after heavy editor updates');
  });
});
