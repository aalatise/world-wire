const environment = require('../environment/env')

let istoken = environment.ENV_KEY_ISTOKEN

module.exports = (options, id, WLorQuote) => {
    if (environment.ENV_KEY_ISTOKEN == "true" || WLorQuote == true) {
        switch (id) {
            case "ENV_KEY_PARTICIPANT_1_ID":
                options["headers"].Authorization = 'Bearer ' + environment["ENV_KEY_PARTICIPANT1_JWT_TOKEN"]
                break;
            case "ENV_KEY_PARTICIPANT_2_ID":
                options["headers"].Authorization = 'Bearer ' + environment["ENV_KEY_PARTICIPANT2_JWT_TOKEN"]
                break;
            case "ENV_KEY_ANCHOR_ID":
                options["headers"].Authorization = 'Bearer ' + environment["ENV_KEY_ANCHOR_JWT_TOKEN"]
                break;
            default:
                break;
        }
        return options
    } else { return (options) }

}