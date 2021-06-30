let StellarSdk = require('stellar-sdk');
let environment = require('../environment/env')
const LOGGER = require('../method/logger')
const log = new LOGGER('Stellar Call')

//StellarSdk.Network.use(new StellarSdk.Network(process.env[environment.ENV_KEY_STELLAR_NETWORK]));
let server = new StellarSdk.Server(process.env[environment.ENV_KEY_HORIZON_CLIENT_URL], { allowHttp: true });

module.exports = {

    /* get account Info. from stellar */
    /**
     * 
     * @param {Object} pkey 
     */
    getBalance: function(account) {
        return server.accounts()
            .accountId(account.toString())
            .call()
            .then(function(accountResult) {
                //log.logger('stellar-API - server.accounts()', "asset_type= " + accountResult.balances[0].asset_type + ',balance = ' + accountResult.balances[0].balance)
               return accountResult.balances
            })
            .catch(function(err) {
                log.error('stellar-API - server.accounts()', err)
                throw err
            })
    },
    transactionBuilder:
    /**
     * 
     * @param {String} source 
     */
        function(source) {
        return new Promise(function(res, rej) {
            log.logger('SDK - TransactionBuilder(source)', 'Created')
            res(new StellarSdk.TransactionBuilder(source))
        }).catch(function(err) {

            log.logger('SDK - TransactionBuilder(source)', JSON.stringify(err))
            rej(err)
        })
    },
    getAccountSequenceNumber:
    /**
     * 
     * @param {String} account 
     */
        function(account) {
        return new Promise(function(res, rej) {
            server.loadAccount(account)
                .then(function(account) {
                    res(account.sequence)
                }).catch(function(err) {
                    log.error('stellar-API - server.accounts()', "Retreiving seq number for gas account failed: "+JSON.stringify(err))
                    rej(err);
                })
        })
    },

    getBuilderAccount: function(pkey, sequenceNumber) {
        return new Promise(function(res, rej) {
            try {
                res(new StellarSdk.Account(pkey, sequenceNumber.toString()))
            } catch (err) {
                log.error('SDK - getBuilderAccount', err)
                throw (err)
            }
        })
    },
    submitTransaction:
    /**
     * 
     * @param {Object} transaction 
     */
        function(transaction) {
        return new Promise(function(res, rej) {
            server.submitTransaction(transaction)
                .then(function(transactionResult) {
                    log.logger('stellar-API -  server.submitTransaction(transaction)', transactionResult)
                    res(JSON.stringify(transactionResult, null, 2))
                })
                .catch(function(err) {
                    log.error('stellar-API -  server.submitTransaction(transaction)', JSON.stringify(err))
                    rej(err)
                });
        })

    },
    getAsset:
    /**
     * 
     * @param {Object} asset 
     */
        function(asset) {
        return new Promise(function(res, rej) {
            if (asset.code == '') {
                log.logger('SDK -  StellarSdk.Asset.native()', StellarSdk.Asset.native())
                res(StellarSdk.Asset.native())
            } else {
                log.logger('SDK -  StellarSdk.Asset(asset.code, asset.issuer)', new StellarSdk.Asset(asset.code, asset.issuer))
                res(new StellarSdk.Asset(asset.code, asset.issuer))
            }
        })
    },
    addPaymentOperation:
    /**
     * 
     * @param {Object} transaction 
     * @param {String} source 
     * @param {String} destination 
     * @param {Object} asset 
     * @param Float*} amount 
     */
        function(transaction, source, destination, asset, amount) {
        return new Promise(function(res, rej) {
            try {
                log.logger('stellar-API - addOperation - StellarSdk.Operation.payment', 'Create')
                res(transaction
                    .addOperation(StellarSdk.Operation.payment({
                        source: source,
                        destination: destination,
                        // asset: StellarSdk.Asset.native(),
                        asset: asset,
                        amount: amount
                    }))
                    // .setTimeout(StellarSdk.TimeoutInfinite)
                )
            } catch (err) {
                log.error('SDK - addOperation', JSON.stringify(err))
                rej(err)
            }
        })
    },
    addSignerOperation:
    /**
     * 
     * @param {Object} transaction 
     * @param {String} destination 
     */
        function(transaction, destination) {
        return new Promise(function(res, rej) {
            try {
                log.logger('stellar-API - addOperation - StellarSdk.Operation.setOptions', 'Create')
                res(transaction
                    .addOperation(StellarSdk.Operation.setOptions({
                        source: destination,
                        signer: {
                            ed25519PublicKey: destination,
                            weight: 4
                        }
                    }))
                )
            } catch (err) {
                log.error('SDK - addOperation', JSON.stringify(err))
                rej(err)
            }

        })
    },
    addSetWeightOperation:
    /**
     * 
     * @param {Object} transaction 
     * @param {String} destination 
     */
        function(transaction, destination) {
        return new Promise(function(res, rej) {
            try {
                log.logger('SDK - addOperation - StellarSdk.Operation.setOptions', 'Create')
                res(transaction
                    .addOperation(StellarSdk.Operation.setOptions({
                        source: destination,
                        masterWeight: 1,
                        lowThreshold: 1,
                        medThreshold: 2,
                        highThreshold: 2
                    })))
            } catch (err) {
                log.error('SDK - addOperation', JSON.stringify(err))
                rej(err)
            }
        })
    },
    buildTransaction:
    /**
     * 
     * @param {Object} transaction 
     */
        function(transaction) {
        return new Promise(function(res, rej) {
            try {
                log.logger('Transaction Build', 'Promise Object')
                res(transaction.build())
            } catch (err) {
                log.error('Transaction Build', JSON.stringify(err))
                rej(err)
            }
        })
    },
    signTx:
    /**
     * 
     * @param {Object} transaction 
     * @param String*} secret 
     */
        function(transaction, secret) {
            try {
                transaction.sign(StellarSdk.Keypair.fromSecret(secret))
                //log.logger('stellar-API - sign', JSON.stringify(transaction))
                return transaction
            } catch (err) {
                log.error('stellar-API - sign', err)
                throw err
            }
    },
    decode:
    /**
     * 
     * @param {Object} transaction 
     */
        function(transaction) {
        return new Promise(function(res, rej) {
            try {
                res(StellarSdk.xdr.TransactionEnvelope.fromXDR(transaction, 'base64'))
            } catch (err) {
                log.error('SDK - .xdr.TransactionEnvelope.fromXDR', JSON.stringify(err))
                rej(err)
            }
        })
    },
    newTransaction:
    /**
     * 
     * @param {String} signedXDR 
     */
        function(signedXDR) {
        return new Promise(function(res, rej) {
            try {
                res(new StellarSdk.Transaction(signedXDR, process.env[environment.ENV_KEY_STELLAR_NETWORK]))
            } catch (err) {
                rej(err)
            }

        })
    },
    submitTransaction:
    /**
     * 
     * @param {Object} transaction 
     */
        function(transaction) {
        return new Promise(function(res, rej) {
                server.submitTransaction(transaction)
                .then(tx=>{
                    let r = {
                        title: "Transaction successful",
                        hash: tx.hash,
                        ledger: tx.ledger,
                        statusCode: 200
                    }
                    log.logger('stellar-API - submitTransaction', `Stellar tx has been submitted successfully, tx hash: ${tx.hash}`)
                    res(r)
                })
                .catch(err=>{
                    let errMsg = {
                        title: "Transaction Failed",
                        failure_reason: JSON.stringify(err),
                        statusCode: 500
                    }
                    log.error('stellar-API - submitTransaction Fail', err)
                    rej(errMsg)
                })
        })
    }

}