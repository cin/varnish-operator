const [major] = process.versions.node.split(".").map(Number);
if (major < 20) {
  console.error(
    `Node.js 20+ is required (current: ${process.version}).\n` +
      "With nvm: nvm install && nvm use   (reads docs/.nvmrc)\n" +
      "With Homebrew: brew install node@20 && brew link node@20 --force --overwrite"
  );
  process.exit(1);
}
