#!/usr/bin/env node
'use strict';

const { spawn } = require('node:child_process');
const { join } = require('node:path');

function binaryPath(platform = process.platform, arch = process.arch) {
  const supported = ['darwin-x64', 'darwin-arm64', 'linux-x64', 'linux-arm64', 'win32-x64'];
  const target = `${platform}-${arch}`;
  if (!supported.includes(target)) {
    throw new Error(`Chunkify does not support ${target}. Supported platforms: ${supported.join(', ')}.`);
  }
  return join(__dirname, '..', 'native', target, platform === 'win32' ? 'chunkify.exe' : 'chunkify');
}

function run() {
  let binary;
  try {
    binary = binaryPath();
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
    return;
  }

  const child = spawn(binary, process.argv.slice(2), { stdio: 'inherit' });
  // Forward signals sent to the launcher so long-running commands can clean up.
  const handlers = new Map();
  for (const signal of ['SIGINT', 'SIGTERM']) {
    const handler = () => child.kill(signal);
    handlers.set(signal, handler);
    process.on(signal, handler);
  }
  const cleanup = () => {
    for (const [signal, handler] of handlers) process.removeListener(signal, handler);
  };
  child.on('error', (error) => {
    cleanup();
    console.error(`Could not start Chunkify: ${error.message}`);
    console.error('Reinstall with npm install -g @chunkify/cli@latest.');
    process.exitCode = 1;
  });
  child.on('exit', (code, signal) => {
    cleanup();
    if (signal) {
      process.kill(process.pid, signal);
    } else {
      process.exitCode = code;
    }
  });
}

if (require.main === module) run();

module.exports = { binaryPath };
