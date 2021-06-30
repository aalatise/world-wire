export const environment = {
    production: true,
    firebase: // Copy and paste this into your JavaScript code to initialize the Firebase SDK.
    // You will also need to load the Firebase SDK.
    // See https://firebase.google.com/docs/web/setup for more details.
    {
        projectId: 'pen1-260919',
        appId: '1:16829926048:web:e20c62dc69415015a597c5',
        databaseURL: 'https://pen1-260919.firebaseio.com',
        storageBucket: 'pen1-260919.appspot.com',
        apiKey: 'AIzaSyBsATNXiBkbekDe2haJ8CDdygY2bYAP1zg',
        authDomain: 'pen1-260919.firebaseapp.com',
        messagingSenderId: '16829926048'
    },
    apiRootUrl: 'https://auth.worldwire-pen.io',
    supported_env: {
        text: 'Pentesting1',
        val: 'pen1',
        envApiRoot: 'worldwire-pen.io/local/api',
        envGlobalRoot: 'worldwire-pen.io/global',
    },
    inactivityTimeout: 15
};
