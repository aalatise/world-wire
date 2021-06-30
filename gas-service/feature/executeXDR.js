const stellar = require("../method/stellar.js")
const accountController = require("../method/account")
const LOGGER = require('../method/logger')
const environment = require('../environment/env')

const log = new LOGGER('Execute Tx')

/**
 * 1. decode signedXDRin , get source account 
 * 2. check whether account is in $lockAccounts , if account is not in $lockAccounts , return can not execute 
 * 3. unlock account
 * 4. execute signedXDR
 * 5. return result ,$lockAccounts,$unlockAccounts
 */
module.exports =
    /**
     * 
     * @param {String} signedXDRin 
     * @param {String} lockAccounts 
     * @param {String} unlockAccounts 
     * @param {String} accountstablename 
     */

    function (signedXDRin, lockAccounts, unlockAccounts, accountstablename, db) {
        let account = {}
        /**
         * 1. decode signedXDRin , get source account 
         */
        return new Promise((res, rej) => {
            stellar.newTransaction(signedXDRin)
                .then(async function(transaction){
                    account.pkey = transaction.source
                    log.info("executeXDR", `executing XDR with account ${account.pkey}`)
                    log.info("executeXDR", `verifying if the account is locked`)

                    /**
                     * 2. check whether account is in $lockAccounts
                     */
                    // var accountInunlockArrayIndex = lockAccounts.indexOf(account.pkey)
                    let accountInunlockArrayIndex = lockAccounts.map(function (item) { return item.pkey; }).indexOf(account.pkey);
                    /**
                     * if account is not in $lockAccounts , return tx fail ,  
                     */

                    if (accountInunlockArrayIndex < 0) {
                        throw ({
                            statusCode: 500,
                            Message: {
                                title: "Source Account Expire",
                                failure_reason: "source account not availible",
                                statusCode: 408
                            }
                        })
                    } else {


                        let filter = {pkey:account.pkey};
                        let query = {$set:{"accountStatus": true, "lockTimestamp": null}}
                        let result = await db.collection( process.env[environment.ENV_KEY_COLLECTION_NAME] ).updateOne(filter, query);
                        if (result.modifiedCount != 1){
                            log.error("executeXDR", `account ${account.pkey} update failed`)
                            throw ({
                                statusCode: 500,
                                Message: {
                                    title: "Internal server error",
                                    failure_reason: "DB update failed",
                                    statusCode: 500
                                }
                            })
                        }

                        return transaction
                    }
                }).then(data => {
                    log.info("executeXDR", `Submitting transactions`)

                    /**
                     * 3. execute signedXDR
                     */
                    return stellar.submitTransaction(data)
                }).then(result => {
                    log.info("executeXDR", `Unlocking account`)

                    /**
                     *  unlock account
                     */
                    unlocked = accountController.unlock(lockAccounts, unlockAccounts, account.pkey)
                    if (!unlocked) {
                        throw ({
                            statusCode: 500,
                            Message: {
                                title: "Gas Account Unlock Failed",
                                failure_reason: `can not unlock gas account ${account.pkey}`,
                                statusCode: 500
                            }
                        })
                    }
                    res(result)
                }).catch(err => {
                    rej(err)
                })
        })
    }