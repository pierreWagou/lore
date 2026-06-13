#!/usr/bin/env node
// postinstall: downloads the lore binary for the current platform from GitHub Releases.

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const pkg = require("./package.json");
const VERSION = pkg.version;

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const platform = PLATFORM_MAP[process.platform];
const arch = ARCH_MAP[process.arch];

if (!platform || !arch) {
  console.error(
    `lore: unsupported platform ${process.platform}/${process.arch}`
  );
  process.exit(1);
}

const ext = platform === "windows" ? ".exe" : "";
const binaryName = `lore-${platform}-${arch}${ext}`;
const destDir = path.join(__dirname, "bin");
const destPath = path.join(destDir, `lore${ext}`);

const url = `https://github.com/pierreWagou/lore/releases/download/v${VERSION}/${binaryName}`;

fs.mkdirSync(destDir, { recursive: true });

console.log(`lore: downloading ${binaryName} v${VERSION}...`);

download(url, destPath, (err) => {
  if (err) {
    console.error(`lore: download failed: ${err.message}`);
    console.error(`      Manual install: go install github.com/pierreWagou/lore@v${VERSION}`);
    process.exit(1);
  }
  fs.chmodSync(destPath, 0o755);
  console.log(`lore: installed to ${destPath}`);
});

function download(url, dest, cb) {
  const file = fs.createWriteStream(dest);
  const get = (u) => {
    https.get(u, (res) => {
      if (res.statusCode === 301 || res.statusCode === 302) {
        return get(res.headers.location);
      }
      if (res.statusCode !== 200) {
        return cb(new Error(`HTTP ${res.statusCode} for ${u}`));
      }
      res.pipe(file);
      file.on("finish", () => file.close(cb));
    }).on("error", (err) => {
      fs.unlink(dest, () => {});
      cb(err);
    });
  };
  get(url);
}
