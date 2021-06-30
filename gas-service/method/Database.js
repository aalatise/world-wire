
const AWS = require('aws-sdk')
const MongoClient = require('mongodb').MongoClient;
const LOGGER = require('./logger');
const environment = require('../environment/env')
let log = new LOGGER('AWS Call')
AWS.config.update({
    region: process.env[environment.ENV_KEY_DYNAMODB_REGION],
    accessKeyId: process.env[environment.ENV_KEY_DYNAMODB_ACCESSKEYID],
    secretAccessKey: process.env[environment.ENV_KEY_DYNAMODB_SECRECT_ACCESS_KEY]
});
var dynamodb = new AWS.DynamoDB();
var docClient = new AWS.DynamoDB.DocumentClient();

// Read the certificate authority
let buff = new Buffer(process.env[environment.ENV_KEY_MONGO_CONNECTION_CERT], 'base64');
let ca = [buff.toString('ascii')];

const dbUser = process.env[environment.ENV_KEY_DB_USER]
const dbPassword = process.env[environment.ENV_KEY_DB_PWD]
const mongoHosts = process.env[environment.ENV_KEY_MONGO_ID]
const connectionString = `mongodb://${dbUser}:${dbPassword}@${mongoHosts}/ibmclouddb?authSource=admin&replicaSet=replset&ssl=true`

let Mongoclient = new MongoClient(connectionString, {
  sslValidate:true,
  sslCA:ca,
  useUnifiedTopology: true
});

// for connection reuse/pooling
var db;


module.exports = {

    createTable:
        /**
         * 
         * @param {Object} create table content 
         */
        function (params) {
            return new Promise(function (res, rej) {
                dynamodb.createTable(params, function (err, data) {
                    if (err) {
                        res(err)
                    } else {
                        res(data)
                    }
                });


            })
        },

    deleteTable:
        /**
         * 
         * @param {String} tablename 
         */
        function (accountstablename) {
            return new Promise(function (res, rej) {

                var params = {
                    TableName: accountstablename
                };
                dynamodb.deleteTable(params, function (err, data) {
                    if (err) {
                        rej(err)
                    } else {
                        res(data)
                    }
                });
            })
        },

    createItem:
        function (tablename, item) {
            return new Promise(function (res, rej) {

                var params = {
                    TableName: tablename,
                    Item: item
                };

                docClient.put(params, function (err, data) {
                    if (err) {
                        throw err
                    } else {
                        res(item)
                    }
                });


            })
        },
    deleteItem:
        function (params) {
            return new Promise(function (res, rej) {
                dynamodb.deleteItem(params, function (err, data) {
                    console.log(err);
                    console.log(data);
                                        
                    if (err) {
                        throw err
                    }
                    else {
                        res(data)
                    }
                });

            })
        },
    updateItem:
        function (params) {
            return new Promise(function (res, rej) {
                docClient.update(params, function (err, data) {
                    if (err) {
                        throw ({
                            statusCode: 500,
                            Message: {
                                ErrorMsg: err
                            }
                        })
                    } else {
                        res(data)
                    }
                });

            })
        },
    getAccountsInfo:
        async function (db, unlock) {

            return new Promise(async function (res, rej) {
                try {
                    let response = [];
                    let cursor = await db.collection( process.env[environment.ENV_KEY_COLLECTION_NAME] ).find({"accountStatus": unlock});
                    cursor.each(function(err, item) {
                        // If the item is null then the cursor is exhausted/empty and closed
                        if(item == null){
                            res(response)
                            return
                        }
                        if (!unlock){
                            let obj = {
                                pkey: item.pkey,
                                lockTimestamp: item.lockTimestamp
                            }
                            response.push(obj)
                        }else{
                            response.push(item.pkey)
                        }
                    });
                } catch (err) {
                    rej({
                        statusCode: 500,
                        Message: {
                            ErrorMsg: err
                        }
                    })
                }

            })

        },

    queryData:
    function (params) {
        return new Promise(function (res, rej) {
            docClient.query(params, function (err, data) {
                if (err) {
                    res(err)
                } else {
                    if (data.Items.length == 0) {
                        res(null)
                    }
                    else {
                        res(data.Items[0])
                    }
                }
            });

        })
    },
    // queryDataSecret:
    //     function (accountstablename, pkey) {
    //         return new Promise(function (res, rej) {
    //             var params = {
    //                 TableName: accountstablename,
    //                 KeyConditionExpression: "#pk = :pkey",
    //                 ExpressionAttributeNames: {
    //                     "#pk": "pkey"
    //                 },
    //                 ExpressionAttributeValues: {
    //                     ":pkey": pkey
    //                 }
    //             };

    //             docClient.query(params, function (err, data) {
    //                 if (err) {
    //                     res(err)
    //                 } else {
    //                     if (data.Items.length == 0) {
    //                         res(null)
    //                     }
    //                     else {
    //                         res(data.Items[0].secret)
    //                     }
    //                 }
    //             });

    //         })
    //     },
    // queryDataGroupID:
    //     function (accountstablename, pkey) {
    //         return new Promise(function (res, rej) {

    //             var params = {
    //                 TableName: accountstablename,
    //                 KeyConditionExpression: "#pk = :pkey",
    //                 ExpressionAttributeNames: {
    //                     "#pk": "pkey"
    //                 },
    //                 ExpressionAttributeValues: {
    //                     ":pkey": pkey
    //                 }
    //             };

    //             docClient.query(params, function (err, data) {
    //                 if (err) {
    //                     res(err)
    //                 } else {
    //                     res(data.Items[0].groupName)
    //                 }
    //             });

    //         })
    //     },
    getDataEmail:
        /**
         * 
         * @param {String} tablename 
         * @param {Bool} status 
         * @param {Array} lockAccounts 
         */
        function (contactsTableName, topicName) {
            return new Promise(function (res, rej) {
                let Array = []

                var params = {
                    TableName: contactsTableName,
                    ProjectionExpression: "#topicName,email, phoneNumber",
                    FilterExpression: "topicName = :topicName ",
                    ExpressionAttributeNames: {
                        "#topicName": "topicName"
                    },
                    ExpressionAttributeValues: {
                        ":topicName": topicName
                    }

                };
                
                docClient.scan(params, onScan);

                function onScan(err, data) {
                    if (err) {
                        rej(err)
                    } else {
                        res(data.Items)
                    }
                }

            })

        },
    getSubscriptionArn:
        function (contactstablename, topicName, email) {
            return new Promise(function (res, rej) {

                var params = {
                    TableName: contactstablename,
                    ProjectionExpression: "#topicName, #email, phoneNumber,Topicarn,SubscriptionArn",
                    FilterExpression: "topicName = :topicName and email = :email",
                    ExpressionAttributeNames: {
                        "#topicName": "topicName",
                        "#email": "email"
                    },
                    ExpressionAttributeValues: {
                        ":topicName": topicName,
                        ":email": email
                    }

                };

                docClient.scan(params, onScan);
                function onScan(err, data) {
                    if (err) {
                        throw ({
                            statusCode: 500,
                            Message: {
                                ErrorMsg: err
                            }
                        })
                    } else {
                        if (data.Items.length > 0) res(data.Items[0].SubscriptionArn)
                        else {
                            res(null)
                        }
                    }
                }

            })

        },
    getTopicArn:
        function (topicstablename, topicName) {
            return new Promise(function (res, rej) {

                var params = {
                    TableName: topicstablename,
                    ProjectionExpression: "#TopicName,TopicArn, displayName",
                    FilterExpression: "TopicName = :TopicName ",
                    ExpressionAttributeNames: {
                        "#TopicName": "TopicName"
                    },
                    ExpressionAttributeValues: {
                        ":TopicName": topicName
                    }

                };

                docClient.scan(params, onScan);

                function onScan(err, data) {
                    if (err) {
                        throw ({
                            statusCode: 500,
                            Message: {
                                ErrorMsg: err
                            }
                        })
                    } else {
                        if (data.Items.length > 0) res(data.Items[0].TopicArn)
                        else {
                            res(null)
                        }
                    }
                }

            })

        },

    getAllDatas:
        /**
         * 
         * @param {String} tablename 
         */
        function (db) {

            return new Promise(async function (res, rej) {
                try {
                    let response = [];
                    let cursor = await db.collection( process.env[environment.ENV_KEY_COLLECTION_NAME] ).find();
                    cursor.each(function(err, item) {
                        // If the item is null then the cursor is exhausted/empty and closed
                        if(item == null){
                            res(response)
                            return
                        }
                        response.push(item)
                    });
                } catch (err) {
                    rej({
                        statusCode: 500,
                        Message: {
                            ErrorMsg: err
                        }
                    })
                }

            })

        },
    unsubscribe:
        function (SubscriptionArn) {
            return new Promise(function (res, rej) {

                var params = {
                    SubscriptionArn: SubscriptionArn /* required */
                };
                var sns = new AWS.SNS();
                var request = sns.unsubscribe(params);

                console.log(params);
                request.
                    on('success', function (response) {
                        
                    }).
                    on('error', function (response) {
                        
                    }).
                    on('complete', function (response) {
                        
                        res(response.data)
                    }).
                    send();
            })
        },
    createTopic:
        function (topicName, displayName) {
            return new Promise(function (res, rej) {


                var params = {
                    Name: topicName, /* required */
                    Attributes: {
                        'DisplayName': displayName
                    }
                };
                var sns = new AWS.SNS();
                var request = sns.createTopic(params);

                request.
                    on('success', function (response) {
                    }).
                    on('error', function (response) {
                    }).
                    on('complete', function (response) {
                        res(response.data)
                    }).
                    send();
            })
        },
    deleteTopic: function (TopicArn) {

        return new Promise(function (res, rej) {
            var params = {
                TopicArn: TopicArn /* required */
            };

            var sns = new AWS.SNS();
            var request = sns.deleteTopic(params);

            request.
                on('success', function (response) {
                }).
                on('error', function (response) {

                }).
                on('complete', function (response) {

                    res(response.data)
                }).
                send();
        })
    },
    subscribeTopic:
        function (TopicArn, phoneNumber) {
            return new Promise(function (res, rej) {
                // Create publish parameters
                var params = {
                    Protocol: 'sms', /* required */
                    TopicArn: TopicArn.toString(),
                    Endpoint: phoneNumber.toString(),
                    ReturnSubscriptionArn: true
                };

                var sns = new AWS.SNS();
                let request = sns.subscribe(params);

                request.
                    on('success', function (response) {
                        // console.log("Success!");
                    }).
                    on('error', function (response) {

                        // console.log("Error!");
                    }).
                    on('complete', function (response) {
                        // console.log("Always!");
                        res(response.data)
                    }).
                    send();
            })
        },
    sendSMS:
        function (msg, topic) {

            return new Promise(function (res, rej) {
                // Create publish parameters
                var params = {
                    Message: msg, /* required */
                    TopicArn: topic.toString()
                };

                // Create promise and SNS service object
                var publishTextPromise = new AWS.SNS({ apiVersion: '2010-03-31' }).publish(params).promise();

                // Handle promise's fulfilled/rejected states
                publishTextPromise.then(
                    function (data) {
                        res(`Message : ${params.Message} send sent to the topic ${params.TopicArn}` +" MessageID is " + data.MessageId )
                    }).catch(
                        function (err) {
                            rej(err)// console.error(err, err.stack);
                        });
            })
        },
    connectToDb: async function(){
        return new Promise(async function(res, rej){
            console.log('Initializing Mongo Connection...')
            // Connect validating the returned certificates from the server
            try{
                Mongoclient = await Mongoclient.connect();
                db = Mongoclient.db(process.env[environment.ENV_KEY_DB_NAME])
                console.log('Mongo Connection Established!')
                res(db)
            }catch(e){
                console.log(e)
                rej(e)
            }
        })
    },
    getDb: function(){
        return db;
    }

}