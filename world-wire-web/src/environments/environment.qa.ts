export const environment = {
  production: true,
  firebase: {
      projectId: 'qatwo-251007',
      appId: '1:431657058128:web:2168714cd4046432',
      databaseURL: 'https://qatwo-251007.firebaseio.com',
      storageBucket: 'qatwo-251007.appspot.com',
      apiKey: 'AIzaSyBRhKzsxcQpgnhtdvYt0ix2qoH0t1v2dBg',
      authDomain: 'qatwo-251007.firebaseapp.com',
      messagingSenderId: '431657058128'
  },
  apiRootUrl: 'https://auth.worldwire-qa.io',
  supported_env: {
      text: 'Quality Assurance',
      name: 'qa',
      val: 'eksqa',
      envApiRoot: 'worldwire-qa.io/local/api',
      envGlobalRoot: 'worldwire-qa.io/global',
  },
  inactivityTimeout: 15
};
