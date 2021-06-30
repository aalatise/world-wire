
const Database = require('../method/Database')
const LOGGER = require('../method/logger')
const log = new LOGGER('Delete Call')

module.exports =
  /**
   * 
   * @param {String} signedXDRin 
   * @param {String} accountstablename 
   */
  async function (contactstablename, topicName, email) {
    if (topicName == undefined) {
      throw ({
        statusCode: 400,
        Message: {
          ErrorMsg: ' missing parameter : topicName'
        }
      })
    }
    if (email == undefined) {
      throw ({
        statusCode: 400,
        Message: {
          ErrorMsg: ' missing parameter : email'
        }
      })
    }
    let SubscriptionArn = await Database.getSubscriptionArn(contactstablename, topicName, email)
    console.log(SubscriptionArn);
    if (SubscriptionArn == null) {
      throw ({
        statusCode: 400,
        Message: {
          ErrorMsg: email + ' in ' + topicName + ' was not exist'
        }
      })
    }
    else {
      let result = await Database.unsubscribe(SubscriptionArn)
      let  params = {
        Key: {
            "topicName": {
                S: topicName
            },
            "email": {
                S: email
            }
        },
        TableName: contactstablename
    };
    await Database.deleteItem(params)
      return result
    }
  }
