#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');

const binaryName = process.platform === 'win32' ? 'faucet-terminal.exe' : 'faucet-terminal';
const binaryPath = path.join(__dirname, binaryName);

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  env: process.env
});

child.on('exit', (code) => {
  process.exit(code || 0);
});

child.on('error', (err) => {
  console.error('failed to start:', err.message);
  process.exit(1);
});
