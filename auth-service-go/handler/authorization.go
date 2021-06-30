package handler

import (
	"context"
	"errors"
	"fmt"
	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/idtoken"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/jwt"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/middleware"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/middleware/token"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/permission"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/response"
	authtesting "github.ibm.com/gftn/world-wire-services/utility/testing"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

/*
* SuperAuthorization : Authorization for client portal for super users
* If JWT is not enabled, the next handler is served.
* If JWT is enabled, database ID, institution ID, permission (request/approve), requestID (if permission is approve), participantID are expected in the headers.
* Participant ID and Institution ID are no longer mandatory because at the time it wont be necessary that those are available.
* All GET requests are direct access, if there is access :- No maker checker
* All POSTS are maker-checker except payout point which needs the current security lead/team member to validate before it gets merged in
 */
func SuperAuthorization(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Record the name of this middleware function for testing purpose
	authtesting.SaveFuncName()

	if r.Header.Get("auth") != "" && r.Header.Get("auth") == "image-testing" {
		LOGGER.Debugf("Image build testing header: %s", r.Header.Get("auth"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("image build test success"))
		return
	}

	enable, _ := os.LookupEnv(global_environment.ENV_KEY_ENABLE_JWT)

	if enable != "true" {
		next.ServeHTTP(w, r)
		return
	}

	mongoOp, err := CreateAuthServiceOperations()
	if err != nil {
		LOGGER.Errorf("Failed to create Mongo DB operations: %s", err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "AUTH-1012", err)
		return
	}

	// All Headers that are expected
	idToken := r.Header.Get("X-Fid")

	if idToken == "" {
		LOGGER.Error("idToken is empty")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1012", errors.New("header missing idToken"))
		return
	}

	iID := r.Header.Get("X-Iid")

	if iID == "" {
		LOGGER.Info("X-Iid is empty, it is not mandatory for super user access.")
		// response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1003", errors.New("Header missing - Institution ID"))
		// return
	}

	uri, err := url.Parse(r.RequestURI)
	if err != nil {
		LOGGER.Error("URL parse error: ", err.Error())
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1021", errors.New("url could not be parsed"))
		return
	}
	path := uri.Path
	LOGGER.Debugf("Request path: %s", path)

	// Dealing with User ID here (Extraction from FID and checking it)
	claims, err := idtoken.Parse(idToken)
	if err != nil {
		LOGGER.Errorf("Failed to parse the id token: %s", err.Error())
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1021", err)
		return
	}
	userID := claims.UID

	xPermission := r.Header.Get("X-Permission")
	makerChecker := false

	if xPermission != "" {
		makerChecker = true
	}

	var rolesForSuperUser permission.User
	id, _ := primitive.ObjectIDFromHex(userID)
	userCollection, ctx := mongoOp.session.GetSpecificCollection(PortalDBName, UserCollection)
	err = userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&rolesForSuperUser)
	if err != nil {
		LOGGER.Debugf("Error during get user super permission query: %s", err.Error())
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1021", err)
		return
	}

	roleToPass := ""
	if rolesForSuperUser.SuperPermissions.Role.Admin {
		roleToPass = "admin"
	} else if rolesForSuperUser.SuperPermissions.Role.Manager {
		roleToPass = "manager"
	} else {
		LOGGER.Error("Participant is not a super user")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("participant is not a super user"))
		return
	}

	LOGGER.Info("Super: roleToPass: ", roleToPass)
	LOGGER.Info("Super: makerChecker: ", makerChecker)
	LOGGER.Info("Super: Method: ", r.Method)
	LOGGER.Info("Super: path", path)

	routePath, err := middleware.ExtractRoutePath(r)
	if err != nil {
		LOGGER.Error("Error while extracting route path from request")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("superAuthorization failed"))
		return
	}
	LOGGER.Infof("Extracting route path %s from request", routePath)

	authorized, _ := middleware.CheckAccess("Super_permissions", roleToPass, makerChecker, r.Method, routePath)
	LOGGER.Infof("authorized in SuperUser: %t", authorized)
	if !authorized {
		LOGGER.Error("Could not authorize endpoint")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("superAuthorization failed"))
		return
	}

	if authorized && !makerChecker {
		LOGGER.Info("Going into next handler for super auth:")
		next.ServeHTTP(w, r)
		return
	}

	if authorized && makerChecker && xPermission == "request" {
		LOGGER.Info("Before Maker Request Super")
		requestID, err := mongoOp.makerRequest("", iID, path, r.Method, "super", userID)
		LOGGER.Infof("After Maker Request Super %s", requestID)
		if err != nil {
			LOGGER.Info("Maker Request failed for Super Auth")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1002", errors.New("superAuthorization Maker failed"))
			return
		}
		response.NotifySuccess(w, r, requestID)
		return
	}

	if authorized && makerChecker && xPermission == "approve" {
		requestID := r.Header.Get("X-Request")
		LOGGER.Info("Before Checker Approve Super")
		approved, err := mongoOp.checkerApprove(requestID, iID, userID, "super")
		LOGGER.Info("After Checker Approve Super: %t", approved)
		if err != nil {
			LOGGER.Errorf("superAuthorization failed Checker approve: %s", err.Error())
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1003", errors.New("superAuthorization failed Checker approve"))
			return

		}

		if approved {
			LOGGER.Info("Going into next handler for Super auth Checker Approve")
			next.ServeHTTP(w, r)
			return
		}
	}

	// Ideally the function will never reach here, in case it does, return
	// We have managed to find ourselves a loop hole or in case of an attack good resilient logic
	LOGGER.Error("Could not authorize")
	response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("superAuthorization failed"))
	return
}

/*
* ParticipantAuthorization : Authorization for client portal
* If JWT is not enabled, the next handler is served.
* If JWT is enabled, database ID, institution ID, permission (request/approve), requestID (if permission is approve), participantID are expected in the headers.
* The error message can be relayed back with NotifyWWError but it seems sensible to log it.
 */
func ParticipantAuthorization(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Record the name of this middleware function for testing purpose
	authtesting.SaveFuncName()

	if r.Header.Get("auth") != "" && r.Header.Get("auth") == "image-testing" {
		LOGGER.Debugf("Image build testing header: %s", r.Header.Get("auth"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("image build test success"))
		return
	}

	enable, _ := os.LookupEnv(global_environment.ENV_KEY_ENABLE_JWT)
	if enable != "true" {
		next.ServeHTTP(w, r)
		return
	}

	mongoOp, err := CreateAuthServiceOperations()
	if err != nil {
		LOGGER.Errorf("Failed to create Mongo DB operations: %s", err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "AUTH-1012", err)
		return
	}

	uri, err := url.Parse(r.RequestURI)
	if err != nil {
		LOGGER.Error("URL parse error: ", err.Error())
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1021", errors.New("url could not be parsed"))
		return
	}
	path := uri.Path
	LOGGER.Info(path)

	idToken := r.Header.Get("X-Fid")
	LOGGER.Debugf("idToken is: %s", idToken)

	if r.Header.Get("Authorization") == "" && idToken == "" {
		LOGGER.Error("Token is missing")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1018", errors.New("header missing - either JWT or FID"))
		return
	}

	if r.Header.Get("Authorization") != "" {
		routePath, err := middleware.ExtractRoutePath(r)
		if err != nil {
			LOGGER.Error("Error while extracting route path from request")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("ParticipantAuthorization failed"))
			return
		}
		LOGGER.Infof("Extracting route path %s from request", routePath)
		isEndpointValid, _ := middleware.CheckAccess("Jwt", "allow", false, r.Method, routePath)
		if !isEndpointValid {
			LOGGER.Error("JWT not authorized for endpoint the current endpoint, error message")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("insufficient JWT permissions"))
			return
		}

		mongoOp.jwtAuthorization(w, r, next)
		return
	}

	// At this point of execution, neither FID nor JWT got missing

	iid := r.Header.Get("X-Iid")

	if iid == "" {
		LOGGER.Error("X-Iid is empty")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1014", errors.New("header missing - Institution ID"))
		return
	}

	claims, err := idtoken.Parse(idToken)
	if err != nil {
		LOGGER.Errorf("Failed to parse the id token: %s", err.Error())
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1021", err)
		return
	}
	userID := claims.UID

	pID := r.Header.Get("X-Pid")

	if pID == "" {
		LOGGER.Error("X-Pid is empty")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1016", errors.New("header missing pID"))
		return
	}

	participantIDFromEnv, exists := os.LookupEnv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	collection, ctx := mongoOp.session.GetSpecificCollection(PortalDBName, InstitutionsCollection)
	id, _ := primitive.ObjectIDFromHex(iid)

	if exists {
		if participantIDFromEnv != "ww" {
			if participantIDFromEnv != pID {
				LOGGER.Error("Environment Home domain name does not match the participant ID from the header")
				response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("environment Home domain name does not match the participant ID from the header"))
				return
			}

			//if err := wwdatabase.FbRef.Child("/nodes/").
			//	Get(wwdatabase.AppContext, &nodes); err != nil {
			//	LOGGER.Error("Error getting nodes info from Firebase %s", err.Error())
			//	response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("Error getting nodes from database"))
			//	return
			//
			//}

			var institutionInfo jwt.Institution
			err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&institutionInfo)
			if err != nil {
				LOGGER.Debugf("Error during get participant info query: %s", err.Error())
				return
			}

			nodes := institutionInfo.Nodes

			if len(nodes) > 0 {
				for _, n := range nodes {
					if n.InstitutionId != "" {
						if n.InstitutionId != iid {
							LOGGER.Error("Institution From node and institution ID in header does not match")
							response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("institution From node and institution ID in header does not match"))
							return
						}

					}
				}
			}
		}
	}

	var institutionCheck jwt.Institution
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&institutionCheck)
	if err != nil {
		LOGGER.Error("Error getting participant info %s", err.Error())
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("error getting participant info from"))
		return
	}

	ppCollection, ppCtx := mongoOp.session.GetSpecificCollection(PortalDBName, ParticipantPermissionsCollection)
	var participantPermissions permission.ParticipantPermission
	err = ppCollection.FindOne(ppCtx, bson.M{"user_id": userID, "institution_id": iid}).Decode(&participantPermissions)
	if err != nil {
		LOGGER.Error("Institution doesn't exist for this user")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("institution doesn't exist for this user"))
		return
	}

	roleToPass := ""
	if participantPermissions.Roles.Admin {
		roleToPass = "admin"
	} else if participantPermissions.Roles.Manager {
		roleToPass = "manager"
	} else {
		LOGGER.Error("Participant is not a super user")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("participant is not a super user"))
		return
	}

	xPermission := r.Header.Get("X-Permission")
	makerChecker := false

	if xPermission != "" {
		makerChecker = true
	}

	LOGGER.Info("Participant: roleToPass: ", roleToPass)
	LOGGER.Info("Participant: makerChecker: ", makerChecker)
	LOGGER.Info("Participant: Method: ", r.Method)
	LOGGER.Info("Participant: path", path)

	routePath, err := middleware.ExtractRoutePath(r)
	if err != nil {
		LOGGER.Error("Error while extracting route path from request")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("ParticipantAuthorization failed"))
		return
	}
	LOGGER.Infof("Extracting route path %s from request", routePath)
	authorized, _ := middleware.CheckAccess("Participant_permissions", roleToPass, makerChecker, r.Method, routePath)

	LOGGER.Info("authorized in participant", authorized)

	if !authorized {
		LOGGER.Error("Could not authorize endpoint")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("ParticipantAuthorization failed"))
		return
	}

	if authorized && !makerChecker {
		LOGGER.Info("next handler for participant")
		next.ServeHTTP(w, r)
		return
	}

	if authorized && makerChecker && xPermission == "request" {
		LOGGER.Info("Participant Auth before maker request")
		requestID, err := mongoOp.makerRequest(pID, iid, path, r.Method, "participant", userID)
		LOGGER.Info("Participant Auth after maker request", requestID)
		if err != nil {
			LOGGER.Info("Maker Request failed for Participant Auth")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1002", errors.New("ParticipantAuthorization Maker failed"))
			return
		}
		response.NotifySuccess(w, r, requestID)
		return
	}

	if authorized && makerChecker && xPermission == "approve" {
		requestID := r.Header.Get("X-Request")
		LOGGER.Info("Participant Auth before Checker approve")
		approved, err := mongoOp.checkerApprove(requestID, iid, userID, "participant")
		LOGGER.Info("Participant Auth after Checker approve")

		if err != nil {
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1003", errors.New("ParticipantAuthorization failed Checker approve"))
			return

		}

		if approved {
			LOGGER.Info("Going into next handler for Participant auth Checker Approve")
			next.ServeHTTP(w, r)
			return
		}
	}

	response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("ParticipantAuthorization failed"))
	return
}

// jwtAuthorization : authorization middleware checks against JWT token to authorize a user for the enpdoint
func (op *AuthOperations) jwtAuthorization(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	LOGGER.Info("Middleware - Running JwtAuthorization")

	// get JWT bearer token from authorization header
	encodedToken, err := request.OAuth2Extractor.ExtractToken(r)
	if err != nil {
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1017", errors.New("authentication failed, token is invalid"))
		return
	}

	// parse token and claimsAndPayload and payload
	claimsAndPayload, valid := op.ExtractJWTClaims(encodedToken, r)

	// check for err parsing token
	// valid automatically checks exp and nfb
	if !valid {
		// log error
		LOGGER.Error("Auth token is no valid")
		response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1017", errors.New("auth token is not valid"))
		return
	}

	// Parse and set context to pass some session data to the handler function call.
	// Using gorilla mux and context here to share context between middleware and handler function
	// Reference: https://stackoverflow.com/questions/41876310/negroni-passing-context-from-middleware-to-handlers
	// and https://www.nicolasmerouze.com/share-values-between-middlewares-context-golang/
	parsedToken, err := middleware.ParseContext(r, &claimsAndPayload)
	if err != nil {
		LOGGER.Error("ParseContext failed:" + err.Error())
		response.NotifyWWError(w, r, http.StatusForbidden, "AUTH-1017", errors.New("authentication failed, invalid token"))
		return
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, middleware.ContextKey, parsedToken)

	// Call the next handler, which can be another middleware in the chain, or the final 
	next.ServeHTTP(w, r.WithContext(ctx))
	return
}

// MakerRequestData is a json-serializable type.
// This is the data structure that gets committed to database.
// The ApproveUserId is added from the beginning since altering  types later is not possible.
type MakerRequestData struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	RequestUserID string             `json:"uid_request,omitempty" bson:"uid_request"`
	ApproveUserID string             `json:"uid_approve" bson:"uid_approve"`
	InstitutionID string             `json:"iid,omitempty" bson:"iid"`
	ParticipantID string             `json:"pid,omitempty" bson:"pid"`
	Status        string             `json:"status,omitempty" bson:"status"`
	Endpoint      string             `json:"endpoint,omitempty" bson:"endpoint"`
	Method        string             `json:"method,omitempty" bson:"method"`
	Timestamp     int64              `json:"timestamp_request,omitempty" bson:"timestamp_request"`
}

/*
 * MakerRequest : After authorization, write the request to database in approvals (participant_approvals/super_approvals) node
 *@param{ w: http response writer, r: http request, level: string , userID: string }
 */
func (op *AuthOperations) makerRequest(participantID string, institutionID string, path string, method string, level string, userID string) (string, error) {
	collection := &mongo.Collection{}
	ctx := context.Background()

	if level == "super" {
		collection, ctx = op.session.GetSpecificCollection(PortalDBName, SuperApprovalsCollection)
	} else if level == "participant" {
		collection, ctx = op.session.GetSpecificCollection(PortalDBName, ParticipantApprovalsCollection)
	}

	timestamp := time.Now()
	id := primitive.NewObjectIDFromTimestamp(timestamp)
	makerRequest := MakerRequestData{
		ID:            id,
		RequestUserID: userID,
		ApproveUserID: "",
		InstitutionID: institutionID,
		ParticipantID: participantID,
		Status:        "request",
		Endpoint:      path,
		Method:        method,
		Timestamp:     timestamp.Unix(),
	}

	_, err := collection.InsertOne(ctx, makerRequest)
	if err != nil {
		LOGGER.Errorf("Unable to insert record to approvals DB: %s", err.Error())
		return "", err
	}

	return id.Hex(), nil
}

// CheckerApprove : After authorization, this is used for approving the request (internal) and making changes in database approvals node.
func (op *AuthOperations) checkerApprove(requestID string, iIDFromHeader string, userID string, level string) (bool, error) {

	if requestID == "" {
		LOGGER.Info("Request ID is nil", requestID)
		return false, errors.New("RequestID must be provided in a header for checker to approve")
	}

	collection := &mongo.Collection{}
	ctx := context.Background()

	if level == "super" {
		collection, ctx = op.session.GetSpecificCollection(PortalDBName, SuperApprovalsCollection)
	} else if level == "participant" {
		collection, ctx = op.session.GetSpecificCollection(PortalDBName, ParticipantApprovalsCollection)
	}

	var requestMap MakerRequestData
	id, _ := primitive.ObjectIDFromHex(requestID)
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&requestMap)
	if err != nil {
		LOGGER.Errorf("Error getting approvals from database: %s", err.Error())
		return false, err
	}

	// get original maker of the request
	requestUser := requestMap.RequestUserID
	//requestRef := approvalRef.Child(requestID)

	if requestUser == "" {
		return false, errors.New("error getting uid_request from database record")
	}

	if requestUser == userID {
		LOGGER.Error("Approve user cannot be the same person as the creator of the request")
		return false, errors.New("approve user cannot be the same person as the creator of the request")
	}

	if level == "participant" {
		if iIDFromHeader != requestMap.InstitutionID {
			LOGGER.Info("Header Institution does not matches with the institution in the request")
		} else {
			/*
			 *1) the maker should not be allowed to approve the request
			 *2) the IID in the header should be the same as the one stored in database request object for a participant
			 *3) the status of the request in database should be "request"
			 */
			if requestUser != userID && requestMap.Status == "request" && iIDFromHeader == requestMap.InstitutionID {

				requestMap.Status = "approved"
				requestMap.ApproveUserID = userID

				updateResult, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": &requestMap})
				if err != nil || updateResult.MatchedCount < 1 {
					LOGGER.Errorf("Update participant approvals failed: %+v", err)
					return false, err
				}

				return true, nil
			}
		}
	} else {
		/*
		 *1) the maker should not be allowed to approve the request
		 *2) the IID in the header should be the same as the one stored in database (which was originally in the request)
		 *3) the status of the request in database should be "request"
		 */
		if requestUser != userID && requestMap.Status == "request" {

			requestMap.Status = "approved"
			requestMap.ApproveUserID = userID

			updateResult, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": &requestMap})
			if err != nil || updateResult.MatchedCount < 1 {
				LOGGER.Errorf("Update super approvals failed: %+v", err)
				return false, err
			}

			return true, nil
		}

	}

	return false, errors.New("reached end of function, could not authorize")
}

// ExtractJWTClaims : parses (decodes) jwt token using secret and returns claims if successful
func (op *AuthOperations) ExtractJWTClaims(tokenStr string, r *http.Request) (jwt.IJWTTokenClaim, bool) {
	var jwtSecure jwt.IJWTSecure

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	parsedToken, err := jwt_go.ParseWithClaims(tokenStr, &jwt.IJWTTokenClaim{}, func(token *jwt_go.Token) (interface{}, error) {

		// Don't forget to validate the alg is what you expect...
		// someone could inject an easier alg to solve
		if _, ok := token.Method.(*jwt_go.SigningMethodHMAC); !ok {
			LOGGER.Debugf("Unexpected signing alg: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if token.Header["alg"].(string) != "HS256" {
			// expected alg does not match alg used HS256
			LOGGER.Debugf("Unexpected signing alg: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing alg: %v", token.Header["alg"])
		}

		dbKey := token.Header["kid"].(string)
		id, _ := primitive.ObjectIDFromHex(dbKey)
		collection, ctx := op.session.GetSpecificCollection(AuthDBName, JWTSecureCollection)
		err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&jwtSecure)
		if err != nil {
			LOGGER.Errorf("Unable to get jwtSecure from DB")
			return nil, errors.New("unable to get jwtSecure from DB")
		}

		decryptionKey := jwtSecure.Secret

		return []byte(decryptionKey), nil
	})

	// error parsing token
	if err != nil {
		LOGGER.Debugf("attempt to parse JWT token failed")
		return jwt.IJWTTokenClaim{}, false
	}

	// run authorization checks on jwt token claims:
	if claims, ok := parsedToken.Claims.(*jwt.IJWTTokenClaim); ok && parsedToken.Valid {

		ipString := r.Header.Get("X-Forwarded-For") // capitalisation

		// exclude all private IP from the x-forwarded-for header
		var privateIPBlocks []*net.IPNet
		for _, cidr := range []string{
			"127.0.0.0/8",    // IPv4 loopback
			"10.0.0.0/8",     // RFC1918
			"172.16.0.0/12",  // RFC1918
			"192.168.0.0/16", // RFC1918
			"::1/128",        // IPv6 loopback
			"fe80::/10",      // IPv6 link-local
			"fc00::/7",       // IPv6 unique local addr
		} {
			_, block, _ := net.ParseCIDR(cidr)
			privateIPBlocks = append(privateIPBlocks, block)
		}

		// parse the incoming ip list as array
		ips := strings.Split(ipString, ",")
		var filteredIps []string
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			IPAddress := net.ParseIP(ip)
			isPrivateIP := false
			for _, block := range privateIPBlocks {
				if block.Contains(IPAddress) {
					isPrivateIP = true
					break
				}
			}
			if isPrivateIP {
				continue
			}
			filteredIps = append(filteredIps, ip)
		}
		LOGGER.Infof("JWT Validation: Receiving request from %+v", filteredIps)
		// check if the token contains the requested ip
		hasIP := token.HasIP(claims, filteredIps)

		// check if the requested endpoint is included in token
		path, _ := middleware.ExtractRoutePath(r)
		hasEndpoint := token.HasEndpoint(claims, path)

		// // Indeed, storing all issued JWT IDs undermines the
		// // stateless nature of using JWTs. However, the purpose
		// // of JWT IDs is to be able to revoke previously-issued
		// // JWTs. This can most easily be achieved by blacklisting
		// // instead of whitelisting. If you've included the "exp"
		// // claim (you should), then you can eventually clean up
		// // blacklisted JWTs as they expire naturally. Of course
		// // you can implement other revocation options alongside
		// // (e.g. revoke all tokens of one client based on a combination
		// // of "iat" and "aud").
		isMatchingJTI := false
		if jwtSecure.JTI == claims.Id {
			isMatchingJTI = true
		}

		//Check if env set in the token matches with current runtime micro-service env
		isMatchingENV := false
		env := os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION)

		if claims.Environment == env {
			isMatchingENV = true
		}

		// check if the token count is the same as the db count
		isOnCount := false
		if jwtSecure.Number == claims.Number {
			isOnCount = true
		}

		// check if the token is for this participant's id
		isForParticipant := token.IsForParticipant(claims)

		// check if all validatations pass
		if isOnCount == true &&
			isMatchingJTI == true &&
			hasEndpoint == true &&
			isForParticipant == true &&
			isMatchingENV == true &&
			hasIP == true {
			// token is valid return true

			return *claims, true
		}
		LOGGER.Errorf("JWT token is not valid because either isOnCount:%v, isMatchingJTI:%v, hasEndpoint:%v, isForParticipant:%v, or hasIp:%v check failed: ip =%v, isMatchingENV:%v, currentENV:%v, Tokenenv:%v", isOnCount,
			isMatchingJTI, hasEndpoint, isForParticipant, hasIP, filteredIps[0], isMatchingENV, env, claims.Environment)
	}

	// default return invalid
	LOGGER.Errorf("JWT token is not valid because either isOnCount, isMatchingJTI, hasEndpoint, isForParticipant, or hasIp check failed")
	return jwt.IJWTTokenClaim{}, false
}