import * as cp from 'node:child_process';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { runTests } from '@vscode/test-electron';

async function main(): Promise<void> {
  const extensionDevelopmentPath = path.resolve(__dirname, '../..');
  const extensionTestsPath = path.resolve(__dirname, './suite/index');
  const extensionRoot = extensionDevelopmentPath;
  const repoRoot = path.resolve(extensionRoot, '..');
  const binaryDir = path.join(extensionRoot, '.bin');
  const binaryPath = path.join(binaryDir, process.platform === 'win32' ? 'english.exe' : 'english');

  fs.mkdirSync(binaryDir, { recursive: true });
  cp.execSync(`go build -o "${binaryPath}" .`, {
    cwd: repoRoot,
    stdio: 'inherit'
  });

  const testWorkspace = path.join(extensionRoot, 'src', 'test', 'fixtures');

  await runTests({
    extensionDevelopmentPath,
    extensionTestsPath,
    launchArgs: [testWorkspace, '--disable-extensions'],
    extensionTestsEnv: {
      ENGLISH_BINARY_PATH: binaryPath,
      PATH: `${binaryDir}${path.delimiter}${process.env.PATH ?? ''}`
    }
  });
}

void main().catch((err) => {
  console.error('Failed to run VS Code tests');
  console.error(err);
  process.exit(1);
});
