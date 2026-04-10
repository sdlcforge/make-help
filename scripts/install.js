#!/usr/bin/env node

const { execFileSync } = require('child_process');
const fs = require('fs');
const https = require('https');
const path = require('path');
const { createGunzip } = require('zlib');

const pkg = require('../package.json');
const version = pkg.version;

const binDir = path.join(__dirname, '..', 'bin');
const isWindows = process.platform === 'win32';
const binaryName = isWindows ? 'make-help.exe' : 'make-help';
const output = path.join(binDir, binaryName);

// Map Node.js platform/arch to goreleaser naming
const PLATFORM_MAP = {
  darwin  : 'darwin',
  linux   : 'linux',
  win32   : 'windows',
};

const ARCH_MAP = {
  x64   : 'amd64',
  arm64 : 'arm64',
};

/**
 * Follow redirects and return the final response as a buffer.
 * @param {string} url - URL to fetch
 * @param {number} [maxRedirects=5] - Maximum number of redirects to follow
 * @returns {Promise<Buffer>} Response body
 */
function download(url, maxRedirects = 5) {
  return new Promise((resolve, reject) => {
    https.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        if (maxRedirects <= 0) {
          reject(new Error('Too many redirects'));

          return;
        }
        resolve(download(res.headers.location, maxRedirects - 1));

        return;
      }
      if (res.statusCode !== 200) {
        reject(new Error(`Download failed: HTTP ${res.statusCode}`));

        return;
      }
      const chunks = [];
      res.on('data', (chunk) => chunks.push(chunk));
      res.on('end', () => resolve(Buffer.concat(chunks)));
      res.on('error', reject);
    }).on('error', reject);
  });
}

/**
 * Extract a tar.gz buffer and write the make-help binary to the output path.
 * Uses a minimal tar parser (512-byte header blocks) to avoid external dependencies.
 * @param {Buffer} buf - The tar.gz archive content
 * @param {string} destPath - Path to write the extracted binary
 * @returns {Promise<void>}
 */
function extractTarGz(buf, destPath) {
  return new Promise((resolve, reject) => {
    const gunzip = createGunzip();
    const chunks = [];
    gunzip.on('data', (chunk) => chunks.push(chunk));
    gunzip.on('error', reject);
    gunzip.on('end', () => {
      const tar = Buffer.concat(chunks);
      let offset = 0;
      while (offset + 512 <= tar.length) {
        const header = tar.subarray(offset, offset + 512);
        // Empty block signals end of archive
        if (header.every((b) => b === 0)) break;

        const name = header.subarray(0, 100).toString('utf8').replace(/\0/g, '');
        // Parse octal size from bytes 124-136
        const sizeStr = header.subarray(124, 136).toString('utf8').replace(/\0/g, '').trim();
        const size = parseInt(sizeStr, 8) || 0;

        offset += 512; // advance past header

        if (name === binaryName || name.endsWith('/' + binaryName)) {
          const fileData = tar.subarray(offset, offset + size);
          fs.mkdirSync(path.dirname(destPath), { recursive : true });
          fs.writeFileSync(destPath, fileData, { mode : 0o755 });
          resolve();

          return;
        }

        // Advance past data blocks (rounded up to 512-byte boundary)
        offset += Math.ceil(size / 512) * 512;
      }
      reject(new Error(`Binary "${binaryName}" not found in archive`));
    });
    gunzip.end(buf);
  });
}

/**
 * Extract a zip buffer and write the make-help binary to the output path.
 * Uses a minimal zip parser (end-of-central-directory scan) to avoid external dependencies.
 * @param {Buffer} buf - The zip archive content
 * @param {string} destPath - Path to write the extracted binary
 */
function extractZip(buf, destPath) {
  // Find end of central directory record (last 22+ bytes of file)
  let eocdOffset = -1;
  for (let i = buf.length - 22; i >= 0; i--) {
    if (buf.readUInt32LE(i) === 0x06054b50) {
      eocdOffset = i;
      break;
    }
  }
  if (eocdOffset === -1) throw new Error('Invalid zip archive');

  const centralDirOffset = buf.readUInt32LE(eocdOffset + 16);
  let offset = centralDirOffset;

  // Walk central directory entries
  while (offset < eocdOffset && buf.readUInt32LE(offset) === 0x02014b50) {
    const nameLen = buf.readUInt16LE(offset + 28);
    const extraLen = buf.readUInt16LE(offset + 30);
    const commentLen = buf.readUInt16LE(offset + 32);
    const localHeaderOffset = buf.readUInt32LE(offset + 42);
    const name = buf.subarray(offset + 46, offset + 46 + nameLen).toString('utf8');

    if (name === binaryName || name.endsWith('/' + binaryName)) {
      // Read from local file header
      const localNameLen = buf.readUInt16LE(localHeaderOffset + 26);
      const localExtraLen = buf.readUInt16LE(localHeaderOffset + 28);
      const compressedSize = buf.readUInt32LE(localHeaderOffset + 18);
      const compressionMethod = buf.readUInt16LE(localHeaderOffset + 8);
      const dataOffset = localHeaderOffset + 30 + localNameLen + localExtraLen;

      if (compressionMethod !== 0) {
        throw new Error('Zip entry is compressed; only stored entries are supported');
      }

      const fileData = buf.subarray(dataOffset, dataOffset + compressedSize);
      fs.mkdirSync(path.dirname(destPath), { recursive : true });
      fs.writeFileSync(destPath, fileData, { mode : 0o755 });

      return;
    }

    offset += 46 + nameLen + extraLen + commentLen;
  }
  throw new Error(`Binary "${binaryName}" not found in zip archive`);
}

/**
 * Try to build from source using Go (fallback when binary download fails).
 * @returns {boolean} True if build succeeded
 */
function tryGoBuild() {
  try {
    execFileSync('go', ['version'], { stdio : 'ignore' });
  }
  catch {
    return false;
  }

  console.log('Attempting to build from source with Go...');
  const ldflags = `-s -w -X github.com/sdlcforge/make-help/internal/version.Version=${version}`;

  try {
    execFileSync('go', ['build', '-ldflags', ldflags, '-o', output, './cmd/make-help'], {
      stdio : 'inherit',
      cwd   : path.join(__dirname, '..'),
    });
    console.log(`Successfully built ${output}`);

    return true;
  }
  catch {
    return false;
  }
}

/**
 * Install by downloading a pre-built binary from GitHub releases.
 */
async function installFromDownload() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    throw new Error(`Unsupported platform: ${process.platform}/${process.arch}`);
  }

  const ext = isWindows ? 'zip' : 'tar.gz';
  const filename = `make-help_${version}_${platform}_${arch}.${ext}`;
  const url = `https://github.com/sdlcforge/make-help/releases/download/v${version}/${filename}`;

  console.log(`Installing make-help v${version} (${platform}/${arch})...`);

  const buf = await download(url);

  if (isWindows) {
    extractZip(buf, output);
  }
  else {
    await extractTarGz(buf, output);
  }

  console.log(`Successfully installed ${output}`);
}

/**
 * Install by building from Go source. Requires Go on PATH.
 */
function installFromSource() {
  if (!tryGoBuild()) {
    console.error('Build from source failed.');
    console.error('Make sure Go is installed and in your PATH.');
    console.error('Install Go from: https://go.dev/dl/');
    process.exit(1);
  }
}

async function main() {
  const args = process.argv.slice(2);
  const forceBuild = args.includes('--build');

  if (forceBuild) {
    installFromSource();

    return;
  }

  try {
    await installFromDownload();
  }
  catch (err) {
    console.error(`Download failed: ${err.message}`);
    console.error('Falling back to source build...');

    if (!tryGoBuild()) {
      console.error(`\nCould not install make-help v${version}.`);
      console.error('To install manually:');
      console.error(`  Download from: https://github.com/sdlcforge/make-help/releases/tag/v${version}`);
      console.error('  Or install Go (https://go.dev/dl/) and run: npm rebuild @sdlcforge/make-help');
      process.exit(1);
    }
  }
}

main();
