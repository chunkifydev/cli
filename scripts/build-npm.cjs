'use strict';

const { execFileSync } = require('node:child_process');
const { chmodSync, mkdirSync, rmSync } = require('node:fs');
const { join } = require('node:path');
const { version } = require('../package.json');

const root = join(__dirname, '..');
const output = join(root, 'npm', 'native');
const targets = [
  ['darwin', 'amd64', 'darwin-x64'],
  ['darwin', 'arm64', 'darwin-arm64'],
  ['linux', 'amd64', 'linux-x64'],
  ['linux', 'arm64', 'linux-arm64'],
  ['windows', 'amd64', 'win32-x64'],
];

rmSync(output, { recursive: true, force: true });
for (const [goos, goarch, target] of targets) {
  const directory = join(output, target);
  mkdirSync(directory, { recursive: true });
  const binary = join(directory, goos === 'windows' ? 'chunkify.exe' : 'chunkify');
  console.error(`Building Chunkify v${version} for ${target}`);
  execFileSync('go', [
    'build', '-trimpath', '-ldflags',
    `-s -w -X github.com/chunkifydev/cli/pkg/version.Version=v${version} -X github.com/chunkifydev/cli/pkg/version.InstallMethod=npm`,
    '-o', binary, '.',
  ], {
    cwd: root,
    env: { ...process.env, CGO_ENABLED: '0', GOOS: goos, GOARCH: goarch },
    stdio: 'inherit',
  });
  chmodSync(binary, 0o755);
}
