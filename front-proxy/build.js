const fs = require('fs');
const path = require('path');

// Log step for debugging
console.log('Starting build process...');

try {
  // Read config and worker files
  const configPath = path.join(__dirname, 'config.json');
  const workerPath = path.join(__dirname, 'worker.js');

  console.log(`Reading config from: ${configPath}`);
  console.log(`Reading worker from: ${workerPath}`);

  // Check if files exist
  if (!fs.existsSync(configPath)) {
    throw new Error(`Config file not found at: ${configPath}`);
  }
  if (!fs.existsSync(workerPath)) {
    throw new Error(`Worker file not found at: ${workerPath}`);
  }

  // Read files
  const configContent = fs.readFileSync(configPath, 'utf8');
  console.log(`Config file read (${configContent.length} bytes)`);
  
  let config;
  try {
    config = JSON.parse(configContent);
    console.log('Config JSON parsed successfully');
  } catch (e) {
    throw new Error(`Error parsing config.json: ${e.message}`);
  }
  
  let workerCode = fs.readFileSync(workerPath, 'utf8');
  console.log(`Worker file read (${workerCode.length} bytes)`);

  // Replace config in worker code
  const configReplaceRegex = /const CONFIG = \{[\s\S]*?\};/;
  
  if (!configReplaceRegex.test(workerCode)) {
    throw new Error('Could not find CONFIG object in worker.js to replace');
  }
  
  const configString = `const CONFIG = ${JSON.stringify(config, null, 2)};`;
  console.log('Prepared CONFIG replacement string');

  // Replace the config part of the worker code
  workerCode = workerCode.replace(configReplaceRegex, configString);
  console.log('Replaced CONFIG in worker code');

  // Write to dist folder
  const distDir = path.join(__dirname, 'dist');
  if (!fs.existsSync(distDir)) {
    console.log(`Creating dist directory: ${distDir}`);
    fs.mkdirSync(distDir);
  }

  const outputPath = path.join(distDir, 'worker.js');
  fs.writeFileSync(outputPath, workerCode);
  console.log(`Wrote output to: ${outputPath}`);

  console.log('✅ Build complete. Configuration injected into worker.');
} catch (error) {
  console.error('❌ Build failed with error:', error);
  process.exit(1);
} 