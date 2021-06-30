const stellar = require("../method/stellar.js")
const Database = require('../method/Database')
const LOGGER = require('../method/logger')
const log = new LOGGER('Sign Transaction')
const environment = require('../environment/env')

/**
 * 1. decode signedXDRin , get pkey 
 * x. (not do)check whether signedXDRin, sequence number is fit dynamoDB info
 * 3. use pkey to get secret
 * 4. signed by the secret 
 * 3. return txeB64 , using $account
 * 4. 
 */
module.exports =
    /**
     * 
     * @param {Object} transactionBuilder 
     * @param {String} accountstablename 
     */
    function (transactionBuilder, accountstablename, db) {
        let account = {}

        return new Promise(async function(res, rej){
            try{
                account.pkey = transactionBuilder.source
                log.info("executeXDR", `Signing XDR with ${account.pkey}`)

                /*
                * 1. use pkey to get secret
                */
    
                let result = await db.collection( process.env[environment.ENV_KEY_COLLECTION_NAME] ).findOne({"pkey":account.pkey});
                if (result == null) {
                    throw ({
                        statusCode: 400,
                        Message: {
                            title: "Source Account Not IBM Account",
                            failure_reason: "can not find account from DynamoDB",
                            statusCode: 400
                        }
                    })
                }
                let signResult = await stellar.signTx(transactionBuilder, result.secret)
                /**
                 * 2. signed by the secret 
                 */
                if (signResult === null || typeof signResult === "undefined") {
                    throw ({
                        statusCode: 500,
                        Message: {
                            title: "Signing failed",
                            failure_reason: "Signing failed",
                            statusCode: 500
                        }
                    })
                }
    
                let signedXDR = signResult.toEnvelope().toXDR('base64')
                res(signedXDR)
            }catch(e){
                rej(e)
            }
        })
    }