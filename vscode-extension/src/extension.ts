import * as vscode from 'vscode';
import { LanguageClient, LanguageClientOptions, ServerOptions, TransportKind } from 'vscode-languageclient/node';

let client: LanguageClient | undefined;

function createServerOptions(): ServerOptions {
  const config = vscode.workspace.getConfiguration('english.languageServer');
  const command = config.get<string>('path', 'english');
  const args = config.get<string[]>('args', ['lsp']);

  return {
    run: { command, args, transport: TransportKind.stdio },
    debug: { command, args, transport: TransportKind.stdio }
  };
}

function createClientOptions(): LanguageClientOptions {
  return {
    documentSelector: [{ scheme: 'file', language: 'english' }],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher('**/*.abc')
    },
    outputChannel: vscode.window.createOutputChannel('English Language Server')
  };
}

export function activate(context: vscode.ExtensionContext): void {
	const serverOptions = createServerOptions();
	const clientOptions = createClientOptions();

	client = new LanguageClient('englishLanguageServer', 'English Language Server', serverOptions, clientOptions);
	context.subscriptions.push(client);
	void client.start();
}

export async function deactivate(): Promise<void> {
  if (!client) {
    return;
  }
  await client.stop();
}
