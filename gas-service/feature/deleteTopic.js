
const Database = require('../method/Database')
const LOGGER = require('../method/logger')
const log = new LOGGER('Delete Call')

module.exports =
  /**
   * 
   * @param {String} signedXDRin 
   * @param {String} accountstablename 
   */
  async function (groupstablename, TopicName,TopicArn) {
    
    let result = await Database.deleteTopic(TopicArn)
    let param = params = {
      Key: {
          "TopicName": {
              S: TopicName
          },
          "TopicArn": {
              S: TopicArn
          }
      },
      TableName: groupstablename
  };
    await Database.deleteItem(param)
    return result
  }
