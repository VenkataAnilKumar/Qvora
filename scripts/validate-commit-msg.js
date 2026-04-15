#!/usr/bin/env node

const fs = require("node:fs");

const messageFile = process.argv[2];
if (!messageFile) {
  console.error("ERROR: Missing commit message file path");
  process.exit(1);
}

let raw;
try {
  raw = fs.readFileSync(messageFile, "utf8");
} catch (err) {
  console.error(`ERROR: Unable to read commit message file: ${messageFile}`);
  console.error(String(err));
  process.exit(1);
}

const message =
  raw
    .split(/\r?\n/)
    .find((line) => line.trim().length > 0)
    ?.trim() ?? "";

const pattern =
  /^(feat|fix|chore|docs|style|refactor|test|perf|ci|build|revert)(\([^)]+\))?: .{1,100}$/;

if (!pattern.test(message)) {
  console.error("ERROR: Commit message must follow Conventional Commits format");
  console.error("  e.g. feat(web): add video generation form");
  process.exit(1);
}
