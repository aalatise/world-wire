// The file contents for the current environment will overwrite these during build.
// The build system defaults to the dev environment which uses `environment.ts`, but if you do
// `ng build --env=prod` then `environment.prod.ts` will be used instead.
// The list of which env maps to which file can be found in `.angular-cli.json`.

// Next site
export const environment = {
  production: false,
  firebase: {
    apiKey: 'AIzaSyC0NKtCv93fjV3SfpxGC10uIw55QlXNG88',
    authDomain: 'dev-2-c8774.firebaseapp.com',
    databaseURL: 'https://dev-2-c8774.firebaseio.com',
    projectId: 'dev-2-c8774',
    storageBucket: '',
    messagingSenderId: '969579303659',
    appId: '1:969579303659:web:9a7337ce8ce058bc'
  },
  apiRootUrl: 'https://auth.foc-dap-world-wire-2a0beb393d3242574412e5315d3d4662-0006.jp-tok.containers.appdomain.cloud',
  supported_env: {
    text: 'Development',
    name: 'dev',
    val: 'eksdev',
    envApiRoot: 'https://ww.foc-dap-world-wire-2a0beb393d3242574412e5315d3d4662-0006.jp-tok.containers.appdomain.cloud',
    envGlobalRoot: 'https://ww.foc-dap-world-wire-2a0beb393d3242574412e5315d3d4662-0006.jp-tok.containers.appdomain.cloud/global',
    envDeploymentRoot: 'https://ww.foc-dap-world-wire-2a0beb393d3242574412e5315d3d4662-0006.jp-tok.containers.appdomain.cloud/admin/automate/v1/deploy/participant',
  },
  inactivityTimeout: 100, // increase so that view doesn't timeout during development
  portalApiUrl: 'https://ww.foc-dap-world-wire-2a0beb393d3242574412e5315d3d4662-0006.jp-tok.containers.appdomain.cloud/admin/api/v1/portal',
  adminApiUrl: 'https://ww.foc-dap-world-wire-2a0beb393d3242574412e5315d3d4662-0006.jp-tok.containers.appdomain.cloud/global/admin/v1/admin',
};
