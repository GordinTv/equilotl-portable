'use strict';
// Only virtual paths requested by Equicord are translated. No ASAR is created.
const Module = require('module');
const path = require('path');
const fs = require('fs');
const root = __dirname;
const state = path.join(root, '.equilotl-portable');
const crypto = require('crypto');
const manifest = JSON.parse(fs.readFileSync(path.join(state, 'state.json'), 'utf8'));
function verifiedJSON(file, hash) {
  const bytes = fs.readFileSync(path.join(state, file));
  if (manifest.schema !== 1 || crypto.createHash('sha256').update(bytes).digest('hex') !== hash)
    throw new Error('Equilotl portable backup integrity failure: ' + file);
  return JSON.parse(bytes.toString('utf8'));
}
const original = verifiedJSON('original.json', manifest.originalSHA256);
const config = verifiedJSON('loader.json', manifest.configSHA256);
const entry = path.resolve(root, original.main);
const alias = path.join(root, '..', '_app.asar');
const guard = Symbol.for('equilotl.portable.loaded');
if (global[guard]) throw new Error('Equicord portable loader invoked twice');
global[guard] = true;
process.env.DISABLE_UPDATER_AUTO_PATCHING = '1';
const resolve = Module._resolveFilename;
Module._resolveFilename = function(request, parent, ...args) {
  if (request === path.join(alias, 'package.json')) {
    return path.join(state, 'original.json');
  }
  if (request === path.join(alias, original.main)) {
    // AntiLag computes resourcesPath from require.main.filename.
    require.main.filename = entry;
    require.main.paths = Module._nodeModulePaths(path.dirname(entry));
    Module._resolveFilename = resolve;
    return resolve.call(this, entry, parent, ...args);
  }
  return resolve.call(this, request, parent, ...args);
};
const { app } = require('electron');
const setAppPath = app.setAppPath;
app.setAppPath = function(value) {
  if (value === alias) {
    app.setAppPath = setAppPath;
    return setAppPath.call(this, root);
  }
  return setAppPath.call(this, value);
};
// Equicord loads the original entry itself, after installing its hooks.
require(config.equicord);
