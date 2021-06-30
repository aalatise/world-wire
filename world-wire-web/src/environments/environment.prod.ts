export const environment = {
  production: true,
  firebase: {
    projectId: 'prod-251807',
    appId: '1:711269174375:web:8d10352546aa4878',
    databaseURL: 'https://prod-251807.firebaseio.com',
    storageBucket: 'prod-251807.appspot.com',
    apiKey: 'AIzaSyCtTtvUUmFFJczHOc2UcOt38wvxpHckwTs',
    authDomain: 'prod-251807.firebaseapp.com',
    messagingSenderId: '711269174375'
  },
  apiRootUrl: 'https://auth.worldwire.io',
  supported_env: {
    text: 'Production',
    name: 'prod',
    val: 'prod',
    envApiRoot: 'worldwire.io/local/api',
    envGlobalRoot: 'worldwire.io/global',
  },
  inactivityTimeout: 15
};
