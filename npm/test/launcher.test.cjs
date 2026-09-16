'use strict';

const assert = require('node:assert/strict');
const { spawn, spawnSync } = require('node:child_process');
const { copyFileSync, mkdirSync, mkdtempSync, rmSync, symlinkSync, writeFileSync } = require('node:fs');
const { tmpdir } = require('node:os');
const { dirname, join } = require('node:path');
const { after, test } = require('node:test');
const { binaryPath } = require('../bin/chunkify.cjs');

test('selects each supported native executable', () => {
  for (const target of ['darwin-x64', 'darwin-arm64', 'linux-x64', 'linux-arm64', 'win32-x64']) {
    const [platform, arch] = target.split('-');
    assert.equal(binaryPath(platform, arch), join(__dirname, '..', 'native', target, platform === 'win32' ? 'chunkify.exe' : 'chunkify'));
  }
  assert.throws(() => binaryPath('win32', 'arm64'), /does not support win32-arm64/);
  assert.throws(() => binaryPath('freebsd', 'x64'), /does not support freebsd-x64/);
});

const directory = mkdtempSync(join(tmpdir(), 'chunkify launcher '));
after(() => rmSync(directory, { recursive: true, force: true }));
const launcher = join(directory, 'bin', 'chunkify.cjs');
mkdirSync(dirname(launcher));
copyFileSync(join(__dirname, '..', 'bin', 'chunkify.cjs'), launcher);
const native = join(directory, 'native', `${process.platform}-${process.arch}`, process.platform === 'win32' ? 'chunkify.exe' : 'chunkify');
mkdirSync(dirname(native), { recursive: true });
// Use Node as a native executable fixture to observe exactly what the launcher passes.
if (process.platform === 'win32') copyFileSync(process.execPath, native);
else symlinkSync(process.execPath, native);
const fixture = join(directory, 'fixture.cjs');
writeFileSync(fixture, `
const fs = require('node:fs');
console.log(JSON.stringify({ args: process.argv.slice(2), cwd: process.cwd(), env: process.env.CHUNKIFY_TEST, input: fs.readFileSync(0, 'utf8') }));
console.error('fixture stderr');
process.exitCode = 17;
`);

test('preserves arguments, streams, environment, working directory and exit status', () => {
  const args = ['a file.mp4', '$(echo nope)', '--profile', 'staging'];
  const result = spawnSync(process.execPath, [launcher, fixture, ...args], {
    cwd: directory,
    env: { ...process.env, CHUNKIFY_TEST: 'inherited' },
    input: 'standard input',
    encoding: 'utf8',
  });
  assert.equal(result.status, 17, result.stderr);
  assert.deepEqual(JSON.parse(result.stdout), { args, cwd: require('node:fs').realpathSync(directory), env: 'inherited', input: 'standard input' });
  assert.match(result.stderr, /fixture stderr/);
});

test('forwards termination so the native process can clean up', { skip: process.platform === 'win32', timeout: 10000 }, async () => {
  const script = join(directory, 'signal.cjs');
  writeFileSync(script, "process.on('SIGTERM', () => process.exit(23)); console.log('ready'); setInterval(() => {}, 1000);");
  const child = spawn(process.execPath, [launcher, script], { stdio: ['ignore', 'pipe', 'pipe'] });
  try {
    const exited = new Promise((resolve, reject) => {
      child.once('error', reject);
      child.once('exit', (code, signal) => resolve({ code, signal }));
    });
    child.stdout.once('data', () => child.kill('SIGTERM'));
    assert.deepEqual(await exited, { code: 23, signal: null });
  } finally {
    child.kill('SIGKILL');
  }
});

test('reports a missing executable with reinstall instructions', () => {
  const missing = join(directory, 'missing', 'bin', 'chunkify.cjs');
  mkdirSync(dirname(missing), { recursive: true });
  copyFileSync(launcher, missing);
  const result = spawnSync(process.execPath, [missing], { encoding: 'utf8' });
  assert.equal(result.status, 1);
  assert.match(result.stderr, /Could not start Chunkify/);
  assert.match(result.stderr, /npm install -g @chunkify\/cli@latest/);
});
