const LOGGER = require('../method/logger')
const lockLogger = new LOGGER('Locking Account')
const unlockLogger = new LOGGER('Unlocking Account')

module.exports = {

    /* insert account into lock/unlock array */
    /**
     * 
     * @param {Object} pkey 
     */
    lock: function(lockAccounts, unlockAccounts) {

        lockLogger.debug("Info", `Locking account`)

        if (unlockAccounts.length <= 0){
            lockLogger.error("Warning", "No available unlocked accounts at the moment")
            return ""
        }

        let unlockAccount = unlockAccounts.shift()
        lockLogger.debug("Info", `Locking account ${unlockAccount}`)

        let locked = lockAccounts.some(item => item.pkey === unlockAccount)
        if (locked){
            lockLogger.error("Error", `the target account: ${unlockAccount} is already locked`)
            return ""
        }

        let unlockedIndex = unlockAccounts.indexOf(unlockAccount)
        if (unlockedIndex !== -1){
            lockLogger.error("Err", `the target account: ${unlockAccount} is not unlocked`)
            lockLogger.error("Err", `${unlockAccounts}`)
            return ""
        }
        
        let lockedAccountDetail = {
            pkey: unlockAccount,
            lockTimestamp: new Date().getTime()
        }
        lockAccounts.push(lockedAccountDetail)
        lockLogger.debug("Info", `${unlockAccount} locked successfully`)

        return unlockAccount
    },
        /* remove account from lock/unlock array */
    /**
     * 
     * @param {Object} pkey 
     */
    unlock: function(lockAccounts, unlockAccounts, target) {

        if (typeof target === "undefined" || target === ""){
            unlockLogger.error("Error", "incoming target account is empty")
            return false
        }

        unlockLogger.debug("Info", `Unlocking account ${target}`)

        if (lockAccounts.length <= 0){
            unlockLogger.error("Warning", "There is nothing to unlock at the moment")
            return false
        }

        let unlockedIndex = unlockAccounts.indexOf(target)
        if (unlockedIndex !== -1){
            unlockLogger.error("Err", `the target account: ${target} is already unlocked`)
            return false
        }

        let lockedIndex = lockAccounts.map(function (item) { return item.pkey; }).indexOf(target);
        if (lockedIndex === -1){
            unlockLogger.error("Error", `the target account: ${target} is not locked`)
            return false
        }
        
        unlockLogger.debug("Info", `Timestamp to be unlocked ${lockAccounts[lockedIndex].lockTimestamp}. Current timestamp ${new Date().getTime()}.`)
        lockAccounts.splice(lockedIndex, 1);
        unlockAccounts.push(target)
        unlockLogger.debug("Info", `${target} unlocked successfully`)
        return true
    }

}