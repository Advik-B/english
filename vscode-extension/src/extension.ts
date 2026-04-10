import * as child_process from 'node:child_process';
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as vscode from 'vscode';
import { LanguageClient, LanguageClientOptions, ServerOptions, TransportKind } from 'vscode-languageclient/node';
import { CatHighlightController } from './catHighlight';

const ENGLISH_MODULE = 'github.com/Advik-B/english@latest';

let client: LanguageClient | undefined;
let catHighlightController: CatHighlightController | undefined;

function isExecutable(filePath: string): boolean {
  try {
    fs.accessSync(filePath, fs.constants.X_OK);
    return true;
  } catch {
    return false;
  }
}

function resolveCommandOnPath(command: string): string | undefined {
  const pathDirs = (process.env.PATH ?? '').split(path.delimiter);
  const defaultPathExt = '.COM;.EXE;.BAT;.CMD;.VBS;.VBE;.JS;.JSE;.WSF;.WSH;.MSC';
  const rawExtensions = process.platform === 'win32' ? (process.env.PATHEXT ?? defaultPathExt).split(';') : [''];
  // If the command already carries an extension (e.g. 'english.exe'), don't append another one.
  const hasExt = process.platform === 'win32' && path.extname(command) !== '';
  const extensions = hasExt ? [''] : rawExtensions;
  for (const dir of pathDirs) {
    for (const ext of extensions) {
      const candidate = path.join(dir, command + ext);
      if (isExecutable(candidate)) {
        return candidate;
      }
    }
  }
  return undefined;
}

function isAvailable(command: string): boolean {
  if (path.isAbsolute(command)) {
    return isExecutable(command);
  }
  return resolveCommandOnPath(command) !== undefined;
}

async function runGoInstall(outputChannel: vscode.OutputChannel): Promise<boolean> {
  return new Promise(resolve => {
    outputChannel.show(true);
    outputChannel.appendLine(`Running: go install ${ENGLISH_MODULE}`);
    const proc = child_process.spawn('go', ['install', ENGLISH_MODULE], {
      env: process.env
    });
    proc.stdout.on('data', (data: Buffer) => outputChannel.append(data.toString()));
    proc.stderr.on('data', (data: Buffer) => outputChannel.append(data.toString()));
    proc.on('close', code => {
      if (code === 0) {
        outputChannel.appendLine('english compiler installed successfully.');
        resolve(true);
      } else {
        outputChannel.appendLine(`go install exited with code ${code}.`);
        resolve(false);
      }
    });
    proc.on('error', err => {
      outputChannel.appendLine(`Failed to run go install: ${err.message}`);
      resolve(false);
    });
  });
}

async function ensureEnglishInstalled(outputChannel: vscode.OutputChannel): Promise<void> {
  const config = vscode.workspace.getConfiguration('english.languageServer');
  const command = config.get<string>('path', 'english');

  if (isAvailable(command)) {
    return;
  }

  if (!isAvailable('go')) {
    void vscode.window.showWarningMessage(
      `The English compiler ("${command}") was not found, and "go" is not on your PATH. ` +
      'Install Go (https://go.dev/dl/) and reload VS Code, or set english.languageServer.path to the compiler binary.'
    );
    return;
  }

  const ok = await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Notification,
      title: 'Installing English compiler via go install…',
      cancellable: false
    },
    () => runGoInstall(outputChannel)
  );

  if (!ok) {
    void vscode.window.showErrorMessage(
      'Failed to install the English compiler automatically. ' +
      `See the "English Language Server" output channel for details.`
    );
  }
}

function createServerOptions(): ServerOptions {
  const config = vscode.workspace.getConfiguration('english.languageServer');
  const command = config.get<string>('path', 'english');
  const args = config.get<string[]>('args', ['lsp']);

  return {
    run: { command, args, transport: TransportKind.stdio },
    debug: { command, args, transport: TransportKind.stdio }
  };
}

function createClientOptions(outputChannel: vscode.OutputChannel): LanguageClientOptions {
  return {
    documentSelector: [{ scheme: 'file', language: 'english' }],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher('**/*.abc')
    },
    outputChannel
  };
}

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  const outputChannel = vscode.window.createOutputChannel('English Language Server');
  context.subscriptions.push(outputChannel);
  catHighlightController = new CatHighlightController();
  context.subscriptions.push(catHighlightController);

  await ensureEnglishInstalled(outputChannel);

  const serverOptions = createServerOptions();
  const clientOptions = createClientOptions(outputChannel);

  client = new LanguageClient('englishLanguageServer', 'English Language Server', serverOptions, clientOptions);
  context.subscriptions.push(client);
  void client.start();
}

export async function deactivate(): Promise<void> {
  if (catHighlightController) {
    catHighlightController.dispose();
    catHighlightController = undefined;
  }
  if (!client) {
    return;
  }
  await client.stop();
}
