const fs = require('node:fs');

module.exports = {
  token: fs.readFileSync('/run/secrets/renovate_token', 'utf8').trim(),
  onboarding: false,
  requireConfig: 'required',
  autodiscover: false,
};
