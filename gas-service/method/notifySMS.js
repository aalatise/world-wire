

const Database = require('../method/Database')
const environment = require('../environment/env')
const LOGGER = require('../method/logger')
const log = new LOGGER('AWS SNS')
const topicsTablename = process.env[environment.ENV_KEY_DYNAMODB_GROUPS_TABLE_NAME]
module.exports = async function (pkey,topicName) {
    
    // let groupName = await AWS.queryDataGroupID(accountsTableName, pkey)
    let arn = await Database.getTopicArn(topicsTablename, topicName)
    if (arn!=null){
        let result =await Database.sendSMS('please top up to : ' + pkey, arn)
        log.info('Send SNS', result)    

    }
}

