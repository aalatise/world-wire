
const Database = require("../method/Database")
const accountController = require("../method/account")
const LOGGER = require("../method/logger")
const environment = require('../environment/env')
const log = new LOGGER('Monitor BL')

/**
 * check lock per monitoringTime
 * if lock timestamp + expireTime > now
 * unlock
 */

module.exports = 
/**
 * 
 * @param {Integer} timeout 
 * @param {Integer} expireTime 
 * @param {Array} lockAccounts 
 * @param {Array} unlockAccounts 
 * @param {String} accountstablename 
 */
function monitor(timeout, expireTime, lockAccounts, unlockAccounts, accountstablename, db) {


    setTimeout(function () {

        // log.data(lockAccounts,'')
        /**
         * if lockArray has something then do chck
         */
        let counter = 0;
        if (lockAccounts.length > 0) {
            lockAccounts.forEach(async function(account){
                
                let now = new Date().getTime()
                /**
                 * if expire then unlock
                 */
                
                if (account.lockTimestamp + parseInt(expireTime)*1000 < now) {
                    /**
                     * updateDB
                     */

                    let filter = {pkey:account.pkey};
                    let query = {$set:{"accountStatus": true, "lockTimestamp": null}}
                    let result = await db.collection( process.env[environment.ENV_KEY_COLLECTION_NAME] ).updateOne(filter, query);
                    if (result.modifiedCount != 1){
                        log.error("monitor", `account ${account.pkey} update failed`)
                        throw ({
                            statusCode: 500,
                            Message: {
                                title: "Internal server error",
                                failure_reason: "DB update failed",
                                statusCode: 500
                            }
                        })
                    }

                    /**
                     * update memory
                     */
                    let unlocked = accountController.unlock(lockAccounts, unlockAccounts, account.pkey)
                    //lockAccounts.splice(accountInunlockArrayIndex, 1);
                    //unlockAccounts.push(account.pkey)
                    if (unlocked){
                        counter++
                    }
                    //log.info('Updated','Successful');
                }
            });
        }
        if (counter){
            log.info("Monitor", `--- Released ${counter} gas accounts ---`)
        }
        /**
         * monitor
         */
        //log.data("Lock Accounts",JSON.stringify(lockAccounts))
        //log.data("Unlock Accounts",JSON.stringify(unlockAccounts))

        monitor(timeout, expireTime, lockAccounts, unlockAccounts, accountstablename, db)
    }, timeout);
}