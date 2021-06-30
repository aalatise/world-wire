const LOGGER = require('../method/logger')
const accountController = require("../method/account")
const log = new LOGGER('unlockAct')
const environment = require('../environment/env')

module.exports = 
async function(account,lockAccounts,unlockAccounts, db){

    try{

        let filter = {"pkey":account.pkey}

        let result= await db.collection(process.env[environment.ENV_KEY_COLLECTION_NAME]).findOne(filter);

        if (result == null){
            throw ({
                statusCode: 400,
                Message: {
                    title: "Account Not IBM Account",
                    failure_reason: "can not find account from DynamoDB"
                }
            })
    
        }

        let unlocked = accountController.unlock(lockAccounts, unlockAccounts, account.pkey)
        if (!unlocked){
            return false
        }
        if (unlocked){
            let query = {$set:{"accountStatus": account.accountStatus, "lockTimestamp": null}}
            let result= await db.collection(process.env[environment.ENV_KEY_COLLECTION_NAME]).updateOne(filter, query);
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
        }
        
        return true
    }catch(e){
        throw e
    }



}