import * as assert from 'node:assert';
import * as vscode from 'vscode';

suite('English Extension (Real VS Code)', () => {
  test('registers english language and activates extension', async () => {
    const binaryPath = process.env.ENGLISH_BINARY_PATH;
    assert.ok(binaryPath, 'ENGLISH_BINARY_PATH must be set for tests');

    await vscode.workspace.getConfiguration('english.languageServer').update(
      'path',
      binaryPath,
      vscode.ConfigurationTarget.Global
    );

    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(workspaceFolder, 'workspace folder should be available for tests');
    const sampleDocUri = vscode.Uri.joinPath(workspaceFolder!.uri, 'sample.abc');
    const document = await vscode.workspace.openTextDocument(sampleDocUri);
    await vscode.window.showTextDocument(document);

    assert.strictEqual(document.languageId, 'english');

    const extension = vscode.extensions.getExtension('advik-b.english-language');
    assert.ok(extension, 'extension should be discoverable by id');

    await extension?.activate();
    assert.ok(extension?.isActive, 'extension should activate inside real VS Code');
  });
});
