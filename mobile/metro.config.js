const { getDefaultConfig } = require('expo/metro-config');
const path = require('path');

const projectRoot = __dirname;
const monorepoRoot = path.resolve(projectRoot, '..');

const config = getDefaultConfig(projectRoot);

// Add the monorepo root to watchFolders so Metro can resolve shared modules
config.watchFolders = [
  path.resolve(monorepoRoot, 'shared'),
];

// Configure the resolver to find modules in the monorepo's shared folder
config.resolver.nodeModulesPaths = [
  path.resolve(projectRoot, 'node_modules'),
  path.resolve(monorepoRoot, 'node_modules'),
];

module.exports = config;
