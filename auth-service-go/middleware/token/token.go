package token

import (
	"github.com/op/go-logging"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/jwt"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"os"
	"strings"
)

// LOGGER for logging
var LOGGER = logging.MustGetLogger("middleware")

// HasIP : check if the request ip address is included in token ip array
func HasIP(claims *jwt.IJWTTokenClaim, remoteAddrs []string) bool {

	// split address into ip and port
	//remoteAddrs[index], _, _ = net.SplitHostPort(ip)

	//ip = "192.1.1.1"
	//check with ip specified in token is present in the caller ip url
	ipArray := claims.IPs
	for _, remoteAddr := range remoteAddrs {
		for _, i := range ipArray {
			if remoteAddr == i {
				return true
			}
		}
	}

	LOGGER.Errorf("No matched IP found")
	return false
}

// HasEndpoint : check if requested endpoint is included in endpoint array
func HasEndpoint(claims *jwt.IJWTTokenClaim, enpTarget string) bool {

	// TODO: from chase to Nakul - need to check if the endpoint trying
	// to be accessed is allows per the permissions.json. I started this
	// check in the commented code below but the parsing and checking
	// still needs to be implemented.

	// // check if the endpoint is listed as an endpoint approved for JWT authentication:
	// // get list of endpoints allowed per
	// filename := "./auth-service/permissions.json"
	// text, _ := ioutil.ReadFile(filename)
	// var data interface{}
	// err := json.Unmarshal(text, &data)
	// fmt.Printf(data)

	// mapJwtEndpoints := authconstants.EndpointJwtPermissions()

	// // TODO: Nakul, This is an LCP problem and can be optimized to an O(n) instead of KMP + O(n).
	// for k := range mapJwtEndpoints {
	// 	if strings.Contains(enpTarget, k) {
	// 		isEndpointValid = true
	// 	}
	// }

	LOGGER.Debugf("Target: %s", enpTarget)
	// check if the claims exist on the JWT token required to access the endpoint
	enpArray := claims.Endpoints
	for _, enp := range enpArray {
		if strings.Contains(enpTarget, enp) {
			LOGGER.Debugf("Matched: %s", enp)
			return true
		}
	}

	LOGGER.Warningf("No matched endpoint found")
	return false

}

// IsValid : check token against db, isOnCount and
// func IsValid(jti string, count float64) (bool, error) {

// 	// Indeed, storing all issued JWT IDs undermines the
// 	// stateless nature of using JWTs. However, the purpose
// 	// of JWT IDs is to be able to revoke previously-issued
// 	// JWTs. This can most easily be achieved by blacklisting
// 	// instead of whitelisting. If you've included the "exp"
// 	// claim (you should), then you can eventually clean up
// 	// blacklisted JWTs as they expire naturally. Of course
// 	// you can implement other revocation options alongside
// 	// (e.g. revoke all tokens of one client based on a combination
// 	// of "iat" and "aud").
// 	isMatchingSubStr := false
// 	if jti == encodedToken[len(encodedToken)-8:] {
// 		isMatchingSubStr = true
// 	}

// 	// check if the token count is the same as the db count
// 	isOnCount := false
// 	if data["n"].(float64) == count {
// 		isOnCount = true
// 	}

// 	// check if all validatations pass
// 	if isOnCount &&
// 		isMatchingJTI {
// 		return true, nil
// 	}

// 	return false, errors.New("token is not valid")

// }

// IsForParticipant : if service is not a singleton then check if the token participant matches the participant id per the env var
func IsForParticipant(claims *jwt.IJWTTokenClaim) bool {

	// check if env is set for participantId
	// get the participant id from env var
	participantID, exists := os.LookupEnv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)

	if exists {
		// skip participant id checking if its a global service identified by "ww"; tentatively code it here.
		if participantID == "ww" {
			return true
		}
		// get audience string array
		aud := claims.Audience

		// check if the token is associated with the participants id
		if aud == participantID {
			// matches
			return true
		}

		// does not match
		return false
	}

	// return true if participantId is not present
	return true

}