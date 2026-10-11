#!/bin/sh
# Run on a Docker host: sh generate-env.sh > .env
# Keep this output private; paste it into Arcane's Environment editor.
set -eu
cd "$(dirname "$0")"
docker run --rm -i -v "$PWD/.env.example:/template:ro" node:22-alpine node - <<'NODE'
const fs = require('node:fs');
const crypto = require('node:crypto');
let env = fs.readFileSync('/template', 'utf8');
env = env.replace(/^([A-Z0-9_]+)=$/gm, (line, name) => {
  if (name.startsWith('FLUXER_VAPID_')) return line;
  const value = name.endsWith('_BASE64')
    ? crypto.randomBytes(32).toString('base64')
    : crypto.randomBytes(name === 'FLUXER_ADMIN_SECRET_KEY_BASE' ? 64 : 32).toString('hex');
  return `${name}=${value}`;
});
const vapid = crypto.createECDH('prime256v1');
vapid.generateKeys();
env = env.replace(/^FLUXER_VAPID_PUBLIC_KEY=$/m, `FLUXER_VAPID_PUBLIC_KEY=${vapid.getPublicKey().toString('base64url')}`);
env = env.replace(/^FLUXER_VAPID_PRIVATE_KEY=$/m, `FLUXER_VAPID_PRIVATE_KEY=${vapid.getPrivateKey().toString('base64url')}`);
process.stdout.write(env);
NODE
