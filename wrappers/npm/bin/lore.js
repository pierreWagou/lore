#!/usr/bin/env node
// Shim: resolves the downloaded binary and execs it with the given args.

const path = require("path");
const { spawnSync } = require("child_process");

const ext = process.platform === "win32" ? ".exe" : "";
const binary = path.join(__dirname, `lore${ext}`);

const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  if (result.error.code === "ENOENT") {
    console.error("lore binary not found. Try reinstalling: npm install lore-agent");
  } else {
    console.error(`lore: ${result.error.message}`);
  }
  process.exit(1);
}

process.exit(result.status ?? 0);
