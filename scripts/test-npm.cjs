'use strict';

const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const { mkdtempSync, rmSync } = require('node:fs');
const { tmpdir } = require('node:os');
const { join } = require('node:path');
const { version } = require('../package.json');

const directory = mkdtempSync(join(tmpdir(), 'chunkify npm '));
const root = join(__dirname, '..');
const npm = (args, cwd = directory) => execFileSync(process.execPath, [process.env.npm_execpath, ...args], {
  cwd, encoding: 'utf8', stdio: ['ignore', 'pipe', 'inherit'],
});

try {
  const [packed] = JSON.parse(npm(['pack', '--json', '--pack-destination', directory], root));
  const files = packed.files.map((file) => file.path);
  for (const target of ['darwin-x64', 'darwin-arm64', 'linux-x64', 'linux-arm64', 'win32-x64']) {
    assert.ok(files.includes(`npm/native/${target}/chunkify${target === 'win32-x64' ? '.exe' : ''}`), `Missing binary for ${target}`);
  }
  assert.ok(files.includes('LICENSE'));
  assert.ok(!files.some((file) => file.startsWith('scripts/') || file.startsWith('npm/test/') || file.endsWith('.go')));
  const tarball = join(directory, packed.filename);
  const prefix = join(directory, 'global');
  npm(['install', '--global', '--prefix', prefix, '--ignore-scripts', '--no-audit', '--no-fund', tarball]);

  const globalBin = process.platform === 'win32' ? join(prefix, 'chunkify.cmd') : join(prefix, 'bin', 'chunkify');
  const globalRun = (args) => process.platform === 'win32'
    ? execFileSync(process.env.ComSpec || 'cmd.exe', ['/d', '/s', '/c', `""${globalBin}" ${args.join(' ')}"`], { cwd: directory, encoding: 'utf8' })
    : execFileSync(globalBin, args, { cwd: directory, encoding: 'utf8' });
  assert.equal(globalRun(['version']).trim().split(/\r?\n/).at(-1), `Chunkify version v${version}`);
  assert.match(globalRun(['update']), /npm install -g @chunkify\/cli@latest/);
  assert.match(globalRun(['--help']), /chunkify/);

  // npm exec is the command behind npx. An empty offline cache prevents a
  // pre-existing global install or registry package from masking a broken pack.
  const output = npm(['exec', '--offline', '--yes', '--ignore-scripts', '--cache', join(directory, 'cache'), '--package', tarball, '--', 'chunkify', 'version']);
  assert.equal(output.trim().split(/\r?\n/).at(-1), `Chunkify version v${version}`);
  console.log(`npm global install and npx passed. Package size: ${(packed.size / 1024 / 1024).toFixed(1)} MiB.`);
} finally {
  rmSync(directory, { recursive: true, force: true });
}
