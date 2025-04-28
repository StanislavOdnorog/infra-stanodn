const fs = require('fs');
const path = require('path');

// Read config and worker files
const configPath = path.join(__dirname, 'config.json');
const workerPath = path.join(__dirname, 'worker.js');

// Read files
const config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
let workerCode = fs.readFileSync(workerPath, 'utf8');

// Replace config in worker code
const configReplaceRegex = /const CONFIG = \{[\s\S]*?\};/;
const configString = `const CONFIG = ${JSON.stringify(config, null, 2)};`;

// Replace the config part of the worker code
workerCode = workerCode.replace(configReplaceRegex, configString);

// Write to dist folder
const distDir = path.join(__dirname, 'dist');
if (!fs.existsSync(distDir)) {
  fs.mkdirSync(distDir);
}

fs.writeFileSync(path.join(distDir, 'worker.js'), workerCode);

console.log('✅ Build complete. Configuration injected into worker.'); 