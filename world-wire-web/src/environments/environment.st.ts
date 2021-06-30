export const environment = {
  production: true,
  firebase: {
      projectId: 'st21-251107',
      appId: '1:676790897826:web:49bd2be203d4aac4',
      databaseURL: 'https://st21-251107.firebaseio.com',
      storageBucket: 'st21-251107.appspot.com',
      apiKey: 'AIzaSyBUQDcmQQDvwHtRTcILn8-H_J5AaIeONXI',
      authDomain: 'st21-251107.firebaseapp.com',
      messagingSenderId: '676790897826'
  },
  apiRootUrl: 'https://auth.worldwire-st.io',
  supported_env: {
      text: 'Staging',
      name: 'st',
      val: 'st',
      envApiRoot: 'worldwire-st.io/local/api',
      envGlobalRoot: 'worldwire-st.io/global',
  },
  inactivityTimeout: 15
};
