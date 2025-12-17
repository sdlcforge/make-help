#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const pkg = require('../package.json');
const version = pkg.version;

const binDir = path.join(__dirname, '..', 'bin');
const binary = process.platform === 'win32' ? 'make-help.exe' : 'make-help';
const output = path.join(binDir, binary);

// Create bin directory if it doesn't exist
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

const ldflags = `-s -w -X github.com/sdlcforge/make-help/internal/version.Version=${version}`;
const cmd = `go build -ldflags "${ldflags}" -o "${output}" ./cmd/make-help`;

console.log(`Building make-help v${version}...`);

try {
  execSync(cmd, {
    stdio: 'inherit',
    cwd: path.join(__dirname, '..')
  });
  console.log(`Successfully built ${output}`);
} catch (error) {
  console.error('Build failed. Make sure Go is installed and in your PATH.');
  console.error('Install Go from: https://go.dev/dl/');
  process.exit(1);
}
