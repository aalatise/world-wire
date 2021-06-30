
const stellar = require("../method/stellar")
const environment = require('../environment/env')
const accountController = require("../method/account")
const LOGGER = require('../method/logger')
const log = new LOGGER('Lock Accounts')
/**
    * 1.pop from unlockAccounts ($account)
    * 2.get timestamp
    * 3.update to dynamoDB ,
    * [ update account from table where pkey = account.pkey and status = unlock (status,timestamp) ]
    * 3.add timestamp to $account 
    * 4.push $account to lockArray (lockAccounts.push($account))
    * 5.get sequence number $account.seqNum
    * 5.return $account.seqNum,$account.pkey , $lockAccounts , $unlockAccounts
    */




module.exports =

    /**
     * 
     * @param {Array} lockAccounts 
     * @param {Array} unlockAccounts 
     * @param {String} accountstablename 
     */
    async function (lockAccounts, unlockAccounts, db) {

        let account = {}
        try {
            /**
            * 1. get unlocked account
            */

            account.pkey = accountController.lock(lockAccounts, unlockAccounts)
            if (!account.pkey) {
                throw ({
                    statusCode: 503,
                    Message: {
                        title: "Cannot retrieve gas account at the moment",
                        failure_reason: "No available gas account at the moment",
                        statusCode: 503
                    }
                })
            }
        } catch (err) {
            if (typeof account.pkey !== "undefined") {
                accountController.unlock(lockAccounts, unlockAccounts, account.pkey)
            }

            throw err
        }

        /**
         * 2.get timestamp
         */
        let timestamp = new Date().getTime()
        /**
         * 3.update to dynamoDB ,
         * [ update account from table where pkey = account.pkey and status = unlock (status,timestamp) ]
         */
        account.accountStatus = false

        let filter = {"pkey": account.pkey}
        let query = {$set:{"accountStatus": account.accountStatus, "lockTimestamp": timestamp}}
        let result = await db.collection(process.env[environment.ENV_KEY_COLLECTION_NAME]).updateOne(filter, query)
        if (result.modifiedCount != 1){
            log.error("Unlock", `account ${account.pkey} update failed`)
            throw ({
                statusCode: 500,
                Message: {
                    title: "Internal server error",
                    failure_reason: "DB update failed",
                    statusCode: 500
                }
            })
        }

        return new Promise((res, rej) => {
            stellar.getAccountSequenceNumber(account.pkey)
                .then(seq => {
                    account.sequenceNumber = seq
                    if (!account.sequenceNumber) {
                        throw ({
                            statusCode: 500,
                            Message: {
                                title: "Cannot retrieve gas account at the moment",
                                failure_reason: "Cannot retrieve sequence number of the gas account",
                                statusCode: 500
                            }
                        })
                    }
                    res(account)
                })
                .catch(err => {
                    log.error("err", JSON.stringify(err))
                    rej(err)
                })
        })
    };