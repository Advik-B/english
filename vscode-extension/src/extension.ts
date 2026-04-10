import * as child_process from 'node:child_process';
import * as fs from 'node:fs';
import * as https from 'node:https';
import * as os from 'node:os';
import * as path from 'node:path';
import * as vscode from 'vscode';
import { LanguageClient, LanguageClientOptions, ServerOptions, TransportKind } from 'vscode-languageclient/node';
import { CatHighlightController } from './catHighlight';

const ENGLISH_GITHUB_ARCHIVE_URL = 'https://api.github.com/repos/Advik-B/english/tarball/main';

let client: LanguageClient | undefined;
let catHighlightController: CatHighlightController | undefined;
let languageServerCommandPath = 'english';

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

async function runCommand(
  command: string,
  args: string[],
  outputChannel: vscode.OutputChannel,
  options?: child_process.SpawnOptions
): Promise<boolean> {
  return new Promise(resolve => {
    outputChannel.show(true);
    outputChannel.appendLine(`Running: ${command} ${args.join(' ')}`);
    const proc = child_process.spawn(command, args, { env: process.env, ...options });
    proc.stdout?.on('data', (data: Buffer) => outputChannel.append(data.toString()));
    proc.stderr?.on('data', (data: Buffer) => outputChannel.append(data.toString()));
    proc.on('close', code => {
      if (code === 0) {
        resolve(true);
      } else {
        outputChannel.appendLine(`${command} exited with code ${code}.`);
        resolve(false);
      }
    });
    proc.on('error', err => {
      outputChannel.appendLine(`Failed to run ${command}: ${err.message}`);
      resolve(false);
    });
  });
}

function downloadToFile(url: string, destination: string, outputChannel: vscode.OutputChannel): Promise<void> {
  return new Promise((resolve, reject) => {
    const request = (nextUrl: string) => {
      outputChannel.appendLine(`Downloading: ${nextUrl}`);
      const req = https.get(
        nextUrl,
        {
          headers: {
            'User-Agent': 'english-vscode-extension',
            Accept: 'application/vnd.github+json'
          }
        },
        response => {
          const status = response.statusCode ?? 0;
          if (status >= 300 && status < 400 && response.headers.location) {
            response.resume();
            request(response.headers.location);
            return;
          }
          if (status < 200 || status >= 300) {
            response.resume();
            reject(new Error(`HTTP ${status} while downloading source archive`));
            return;
          }
          const file = fs.createWriteStream(destination);
          let settled = false;
          const fail = (err: Error) => {
            if (settled) {
              return;
            }
            settled = true;
            file.destroy();
            reject(err);
          };
          file.once('error', fail);
          response.once('error', fail);
          file.once('finish', () => {
            if (settled) {
              return;
            }
            settled = true;
            file.close(err => (err ? reject(err) : resolve()));
          });
          response.pipe(file);
        }
      );
      req.on('error', err => reject(err));
    };
    request(url);
  });
}

async function buildEnglishFromGithubArchive(
  context: vscode.ExtensionContext,
  outputChannel: vscode.OutputChannel
): Promise<string | undefined> {
  if (!isAvailable('tar')) {
    const guidance =
      process.platform === 'win32'
        ? 'Install tar (for example via Git for Windows/bsdtar) or set english.languageServer.path manually.'
        : 'Install tar from your system package manager or set english.languageServer.path manually.';
    outputChannel.appendLine(`The "tar" command was not found on PATH; cannot extract GitHub source archive. ${guidance}`);
    return undefined;
  }

  const tmpRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'english-src-'));
  const archivePath = path.join(tmpRoot, 'english.tar.gz');
  const sourceDir = path.join(tmpRoot, 'source');
  const binDir = path.join(context.globalStorageUri.fsPath, 'bin');
  const binaryName = process.platform === 'win32' ? 'english.exe' : 'english';
  const binaryPath = path.join(binDir, binaryName);

  try {
    fs.mkdirSync(sourceDir, { recursive: true });
    fs.mkdirSync(binDir, { recursive: true });
    await downloadToFile(ENGLISH_GITHUB_ARCHIVE_URL, archivePath, outputChannel);
    const extracted = await runCommand(
      'tar',
      ['-xzf', archivePath, '-C', sourceDir, '--strip-components=1'],
      outputChannel
    );
    if (!extracted) {
      return undefined;
    }
    const built = await runCommand('go', ['build', '-o', binaryPath, '.'], outputChannel, { cwd: sourceDir });
    if (!built) {
      outputChannel.appendLine(
        'Failed to build english from source. Verify your Go toolchain and review build output above for details.'
      );
      return undefined;
    }
    outputChannel.appendLine(`english compiler built successfully: ${binaryPath}`);
    return binaryPath;
  } catch (err) {
    outputChannel.appendLine(
      `Failed to build from GitHub archive: ${err instanceof Error ? err.message : String(err)}`
    );
    return undefined;
  } finally {
    fs.rmSync(tmpRoot, { recursive: true, force: true });
  }
}

async function ensureEnglishInstalled(
  context: vscode.ExtensionContext,
  outputChannel: vscode.OutputChannel
): Promise<string> {
  const config = vscode.workspace.getConfiguration('english.languageServer');
  const command = config.get<string>('path', 'english');

  if (isAvailable(command)) {
    return command;
  }

  if (!isAvailable('go')) {
    void vscode.window.showWarningMessage(
      `The English compiler ("${command}") was not found, and "go" is not on your PATH. ` +
      'Install Go (https://go.dev/dl/) and reload VS Code, or set english.languageServer.path to the compiler binary.'
    );
    return command;
  }

  const builtBinary = await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Notification,
      title: 'Building English compiler from GitHub source…',
      cancellable: false
    },
    () => buildEnglishFromGithubArchive(context, outputChannel)
  );

  if (!builtBinary) {
    void vscode.window.showErrorMessage(
      'Failed to build the English compiler automatically. ' +
      `See the "English Language Server" output channel for details.`
    );
    return command;
  }

  return builtBinary;
}

function createServerOptions(commandOverride?: string): ServerOptions {
  const config = vscode.workspace.getConfiguration('english.languageServer');
  const command = commandOverride ?? config.get<string>('path', 'english');
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

  languageServerCommandPath = await ensureEnglishInstalled(context, outputChannel);

  const serverOptions = createServerOptions(languageServerCommandPath);
  const clientOptions = createClientOptions(outputChannel);

  client = new LanguageClient('englishLanguageServer', 'English Language Server', serverOptions, clientOptions);
  context.subscriptions.push(client);
  void client.start().catch(err => {
    outputChannel.appendLine(`Failed to start language client: ${err instanceof Error ? err.message : String(err)}`);
  });
}

export async function deactivate(): Promise<void> {
  if (catHighlightController) {
    catHighlightController.dispose();
    catHighlightController = undefined;
  }
  if (!client) {
    return;
  }
  try {
    await client.stop();
  } catch {
    // VS Code may dispose extension host while the client is still starting.
  } finally {
    client = undefined;
  }
}
