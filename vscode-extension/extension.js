const path = require('path');
const vscode = require('vscode');
const { LanguageClient, TransportKind } = require('vscode-languageclient/node');

let client;

function createServerOptions() {
  const config = vscode.workspace.getConfiguration('english.languageServer');
  const command = config.get('path', 'english');
  const args = config.get('args', ['lsp']);

  return {
    run: { command, args, transport: TransportKind.stdio },
    debug: { command, args, transport: TransportKind.stdio }
  };
}

function createClientOptions(context) {
  return {
    documentSelector: [{ scheme: 'file', language: 'english' }],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher('**/*.abc')
    },
    outputChannel: vscode.window.createOutputChannel('English Language Server')
  };
}

function activate(context) {
  const serverOptions = createServerOptions();
  const clientOptions = createClientOptions(context);

  client = new LanguageClient(
    'englishLanguageServer',
    'English Language Server',
    serverOptions,
    clientOptions
  );

  context.subscriptions.push(client.start());
}

async function deactivate() {
  if (!client) {
    return undefined;
  }
  return client.stop();
}

module.exports = {
  activate,
  deactivate
};
