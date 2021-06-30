const signXDR = require("./signXDR.js")
const executeXDR = require("./executeXDR.js")
const stellar = require("../method/stellar.js")
const accountController = require("../method/account")
const LOGGER = require('../method/logger')
const log = new LOGGER('Sign and Execute XDR')

/**
 * 1. signedXDR
 * 2. execute signedXDR
 */
module.exports =
    /**
     * 
     * @param {String} signedXDRin 
     * @param {Array} lockAccounts 
     * @param {Array} unlockAccounts 
     * @param {String} accountstablename 
     */
    function (signedXDRin, lockAccounts, unlockAccounts, accountstablename, db) {

        let lockedAccount
        return stellar.newTransaction(signedXDRin)
            .then(result => {
                lockedAccount = result.source
                return signXDR(result, accountstablename, db)
            }).then(signedXDR => {
                return executeXDR(signedXDR, lockAccounts, unlockAccounts, accountstablename, db)
            }).catch(e => {
                let unlocked = accountController.unlock(lockAccounts, unlockAccounts, lockedAccount)
                if (!unlocked) {
                    log.warn("WARN", "Account is already unlocked")
                }
                throw e
            })

    }