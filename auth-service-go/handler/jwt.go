package handler

import (
	"encoding/json"
	"errors"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/jwt"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/utility/stringutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strings"
	"time"
)

// #3 Third step for JWT creation
func (op *AuthOperations) HandleJWTGenerate(w http.ResponseWriter, r *http.Request) {
	var input jwt.General
	//fid := r.Header.Get("x-fid")
	iid := r.Header.Get("x-iid")
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		LOGGER.Warningf("Error while validating token body :  %v", err)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	LOGGER.Debugf("%s, %s", iid, input.JTI)

	var token jwt.Info
	infoCollection, ctx := op.session.GetSpecificCollection(AuthDBName, JWTInfoCollection)
	err = infoCollection.FindOne(ctx,
		bson.M{
			"institution": iid,
			"jti":         input.JTI,
		}).Decode(&token)
	if err != nil {
		LOGGER.Errorf("Error getting JWT info from query: %s", err.Error())
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}
	LOGGER.Debugf("%+v", token)

	// if token exists
	if &token == nil {
		LOGGER.Errorf("Token info not found")
		jwt.ResponseError(w, http.StatusInternalServerError, errors.New("token info not found"))
		return
	}

	// only generate if stage is currently approved
	if token.Stage != jwt.Approved {
		LOGGER.Errorf("Token is not currently approved")
		jwt.ResponseError(w, http.StatusForbidden, errors.New("token is not currently approved"))
		return
	}

	email := r.Header.Get("email")
	// check to make sure the authenticated user is the same user who requested the token
	if email == "" || email != token.CreatedBy {
		LOGGER.Errorf("User who requested the token must be the same user to generate the token")
		jwt.ResponseError(w, http.StatusForbidden, errors.New("user who requested the token must be the same user to generate the token"))
		return
	}

	// ensure that the approved request includes a jti
	if token.JTI != input.JTI {
		LOGGER.Errorf("Unknown token id")
		jwt.ResponseError(w, http.StatusForbidden, errors.New("unknown token id"))
		return
	}

	// update token info
	token.Stage = jwt.Ready

	// set default expiration time
	//initExp := "15m" //os.Getenv("initial_mins") + "m"
	//if initExp == "" {
	//	initExp = "1h"
	//}

	// generate the token with payload and claims
	// initialize to expire in n1 hrs and not before n2 seconds from now
	//encodedToken := jwt.GenerateToken(payload, initExp, "0s")
	tokenSecret := stringutil.RandStringRunes(64, false)

	keyID := primitive.NewObjectIDFromTimestamp(time.Now())
	jwtSecure := jwt.IJWTSecure{
		ID:     keyID,
		Secret: tokenSecret,
		JTI:    input.JTI,
		Number: 0,
	}

	secureCollection, secureCtx := op.session.GetSpecificCollection(AuthDBName, JWTSecureCollection)
	_, err = secureCollection.InsertOne(secureCtx, jwtSecure)
	if err != nil {
		LOGGER.Errorf("Insert JWT secure failed:  %+v", err)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	// convert the interface type ID to string
	LOGGER.Debugf("New generate ID: %s" , keyID.Hex())

	count := 0
	// define payload
	payload := jwt.CreateClaims(token, count, iid, keyID.Hex())
	payload.ExpiresAt = time.Now().Add(time.Minute * 60).Unix()
	payload.NotBefore = time.Now().Unix()

	encodedToken, _ := jwt.CreateAndSign(payload, tokenSecret, keyID.Hex())

	// save updated token info
	updateResult, updateInfoErr := infoCollection.UpdateOne(ctx, bson.M{"institution": iid, "jti": input.JTI}, bson.M{"$set": &token})
	if updateInfoErr != nil || updateResult.MatchedCount < 1{
		LOGGER.Errorf("Error update token info: %+v", updateInfoErr)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	LOGGER.Debugf("Successfully generate JWT token")
	jwt.ResponseSuccess(w, encodedToken)
	return
}

// #1 First step for JWT creation
func (op *AuthOperations) HandleJWTRequest(w http.ResponseWriter, r *http.Request) {
	var input jwt.Info
	//fid := r.Header.Get("x-fid")
	iid := r.Header.Get("x-iid")

	if iid == "" {
		LOGGER.Errorf("Missing institution ID: %s")
		jwt.ResponseError(w, http.StatusBadRequest, errors.New("missing institution ID"))
		return
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		LOGGER.Warningf("Error while validating token body :  %v", err)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	email := r.Header.Get("email")
	if email == "" {
		LOGGER.Errorf("Missing requester's email: %s")
		jwt.ResponseError(w, http.StatusBadRequest, errors.New("missing requester's email"))
		return
	}

	// Get the participant information from the institutions collection
	var institution jwt.Institution
	id, _ := primitive.ObjectIDFromHex(iid)
	participantCollection, ctx := op.session.GetSpecificCollection(PortalDBName, InstitutionsCollection)
	err = participantCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&institution)
	if err != nil {
		LOGGER.Debugf("Error during get institutions query: %s", err.Error())
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	LOGGER.Debugf("%+v", institution)

	// check if the participantId is associated with the institutionId
	institutionOwnsParticipantId := jwt.InstitutionOwnsParticipantId(input.Aud, institution)
	if !institutionOwnsParticipantId {
		LOGGER.Errorf("Participant ID must be associated with institution")
		jwt.ResponseError(w, http.StatusBadRequest, errors.New("participantId must be associated with institution"))
		return
	}

	input.Institution = iid
	input.Stage = jwt.Review
	input.Active = false
	input.CreatedAt = time.Now().Unix()
	input.CreatedBy = email
	jti := stringutil.RandStringRunes(26, false)
	input.JTI = jti
	input.JTI = jti
	input.Sub = iid

	infoCollection, infoCtx := op.session.GetSpecificCollection(AuthDBName, JWTInfoCollection)
	_, err = infoCollection.InsertOne(infoCtx, input)
	if err != nil {
		LOGGER.Errorf("Insert JWT info failed:  %+v", err)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	LOGGER.Debugf("Successfully requested token. Your new jti: " + input.JTI)
	jwt.ResponseSuccess(w, input.JTI)
	return
}

func (op *AuthOperations) HandleJWTVerify(w http.ResponseWriter, r *http.Request) {
	var input jwt.IVerifyCompare
	//fid := r.Header.Get("x-fid")
	//iid := r.Header.Get("x-iid")
	auth := r.Header.Get("Authorization")

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		LOGGER.Warningf("Error while validating token body :  %v", err)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	if auth == "" {
		LOGGER.Errorf("No authentication token provided")
		jwt.ResponseError(w, http.StatusUnauthorized, errors.New("no authentication token provided"))
		return
	}

	if input.JTI == "" {
		LOGGER.Errorf("Missing JTI in the parameters")
		jwt.ResponseError(w, http.StatusUnauthorized, errors.New("missing jti in the parameters"))
		return
	}

	bearer := strings.Split(auth, " ")
	encodedToken := bearer[1]

	// get db /jwt_secure data and pepper secret
	var secure jwt.IJWTSecure
	secureCollection, ctx := op.session.GetSpecificCollection(AuthDBName, JWTSecureCollection)
	err = secureCollection.FindOne(ctx,
		bson.M{
			"jti": input.JTI,
		}).Decode(&secure)
	if err != nil {
		LOGGER.Debugf("Error during get JWT secure query: %s", err.Error())
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	// if both secrets are available then continue
	if secure.Secret == "" {
		LOGGER.Errorf("Unable to get token secret")
		jwt.ResponseError(w, http.StatusForbidden, errors.New("unable to get token secret"))
		return
	}

	// decode token
	jwtClaim, ok := jwt.Verify(encodedToken, secure.Secret)
	if  !ok {
		jwt.ResponseError(w, http.StatusUnauthorized, errors.New("token verification failed"))
		return
	}

	// default ip to actual request ip
	LOGGER.Debugf("Remote address: %s", r.RemoteAddr)
	compareIncomingIp := r.RemoteAddr //req.connection.remoteAddress;
	if input.IP != "" {
		// if IP is provided in body, then compare the encoded token
		// ip to the one in the request body
		compareIncomingIp = input.IP
	}

	LOGGER.Debugf("Incoming IP: %s", compareIncomingIp)

	// run custom validation on developer token
	pass, msg := jwt.VerifyWWTokenCustom(*jwtClaim, secure.Number, secure.JTI , compareIncomingIp, input.Endpoint, input.Account)
	if !pass {
		LOGGER.Errorf("Failed to pass one (or more) of the many token validation checks: " + msg)
		jwt.ResponseError(w, http.StatusForbidden, errors.New("failed to pass one (or more) of the many token validation checks: " + msg))
		return
	}

	LOGGER.Debugf("Success! Token is valid for the supplied body parameters.")
	jwt.ResponseSuccess(w, "Success! Token is valid for the supplied body parameters.")
	return
}

func (op *AuthOperations) HandleJWTRefresh(w http.ResponseWriter, r *http.Request) {
	var input jwt.General
	iid := r.Header.Get("x-iid")

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		LOGGER.Warningf("Error while validating token body :  %v", err)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		LOGGER.Errorf("No authentication token provided")
		jwt.ResponseError(w, http.StatusUnauthorized, errors.New("no authentication token provided"))
		return
	}

	bearer := strings.Split(auth, " ")
	encodedToken := bearer[1]
	//
	//payload := strings.Split(encodedToken, ".")[1]
	//// decode token will error out if jwt is expired or
	//decodedPayload, _ := base64.StdEncoding.DecodeString(payload)
	//LOGGER.Debugf("%+v", string(decodedPayload))
	////decodedPayload = []byte(string(decodedPayload)+"}")
	//
	//var jwtClaim jwt.IJWTTokenClaim
	//json.Unmarshal(decodedPayload, &jwtClaim)
	//LOGGER.Debugf("%+v", jwtClaim)

	var token jwt.Info
	infoCollection, ctx := op.session.GetSpecificCollection(AuthDBName, JWTInfoCollection)
	dbErr := infoCollection.FindOne(ctx,
		bson.M{
			"institution": iid,
			"jti":         input.JTI,
		}).Decode(&token)
	if dbErr != nil {
		LOGGER.Errorf("Error getting JWT info from query: %s", dbErr.Error())
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}
	LOGGER.Debugf("%+v", token)

	if token.JTI == "" {
		LOGGER.Errorf("Unable to find token info")
		jwt.ResponseError(w, http.StatusBadRequest, errors.New("unable to find token info"))
		return
	}

	if token.Stage != jwt.Ready && token.Stage != jwt.Initialized {
		LOGGER.Errorf("Token is not currently under review")
		jwt.ResponseError(w, http.StatusUnauthorized, errors.New("token is not currently under review"))
		return
	}

	// ================= Update general information about the token visible to the ui =================

	// update token info
	token.Stage = jwt.Initialized
	token.RefreshedAt = time.Now().Unix()
	token.Active = true

	// ================= Generate the new refresh token  =================
	jwtClaim, keyID, parseErr := jwt.Parse(encodedToken)
	if parseErr != nil {
		LOGGER.Errorf("Failed to parse token claim")
		jwt.ResponseError(w, http.StatusInternalServerError, errors.New("failed to parse token claim"))
		return
	}

	// increment high watermark
	jwtClaim.Number = jwtClaim.Number + 1

	// now + n minutes
	jwtClaim.ExpiresAt = time.Now().Add(time.Minute * 15).Unix()  //+ (60000 * Number(this.env.refresh_mins));

	// cannot use until n milliseconds earlier (allows a buffer for un-synced clocks)
	jwtClaim.NotBefore = time.Now().Add(time.Millisecond * -500).Unix()

	//  ================= Write updated data to mongo  =================

	// delete out "old" secure token data
	id, _ := primitive.ObjectIDFromHex(keyID)
	secureCollection, secureCtx := op.session.GetSpecificCollection(AuthDBName, JWTSecureCollection)
	_, deleteErr := secureCollection.DeleteOne(secureCtx, bson.M{"_id": id})
	if deleteErr != nil {
		LOGGER.Errorf("Error while delete JWT secure: %s", deleteErr.Error())
		jwt.ResponseError(w, http.StatusInternalServerError, deleteErr)
		return
	}

	// update secure "new" token data and generate newly refreshed encoded token
	tokenSecret := stringutil.RandStringRunes(64, false)

	newKeyID := primitive.NewObjectIDFromTimestamp(time.Now())
	jwtSecure := jwt.IJWTSecure{
		ID:     newKeyID,
		Secret: tokenSecret,
		JTI:    jwtClaim.Id,
		Number: jwtClaim.Number,
	}

	_, err = secureCollection.InsertOne(secureCtx, jwtSecure)
	if err != nil {
		LOGGER.Errorf("Insert JWT secure failed:  %+v", err)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	// convert the interface type ID to string
	LOGGER.Debugf("refreshed ID: %s" , newKeyID.Hex())

	signedEncodedToken, _ := jwt.CreateAndSign(*jwtClaim, tokenSecret, newKeyID.Hex())

	// save updated token info
	updateResult, updateInfoErr := infoCollection.UpdateOne(ctx, bson.M{"institution": iid, "jti": token.JTI}, bson.M{"$set": &token})
	if updateInfoErr != nil || updateResult.MatchedCount < 1 {
		LOGGER.Errorf("Error update token info: %+v", updateInfoErr)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	LOGGER.Debugf("Successfully refresh JWT token")
	jwt.ResponseSuccess(w, signedEncodedToken)
	return

}

func (op *AuthOperations) HandleJWTRevoke(w http.ResponseWriter, r *http.Request) {
	var input jwt.General
	//fid := r.Header.Get("x-fid")
	iid := r.Header.Get("x-iid")

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		LOGGER.Warningf("Error while validating token body :  %v", err)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	email := r.Header.Get("email")
	if email == "" {
		LOGGER.Errorf("Missing requester's email: %s")
		jwt.ResponseError(w, http.StatusBadRequest, errors.New("missing requester's email"))
		return
	}

	var token jwt.Info
	infoCollection, ctx := op.session.GetSpecificCollection(AuthDBName, JWTInfoCollection)
	err = infoCollection.FindOne(ctx,
		bson.M{
			"institution": iid,
			"jti":         input.JTI,
		}).Decode(&token)
	if err != nil {
		LOGGER.Errorf("Error getting JWT info from query: %s", err.Error())
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}
	LOGGER.Debugf("%+v", token)

	if &token == nil {
		LOGGER.Errorf("Token id info not found")
		jwt.ResponseError(w, http.StatusBadRequest, errors.New("token id info not found"))
		return
	}

	// delete out secure token data
	secureCollection, secureCtx := op.session.GetSpecificCollection(AuthDBName, JWTSecureCollection)
	_, deleteErr := secureCollection.DeleteOne(secureCtx, bson.M{"jti": input.JTI})
	if deleteErr != nil {
		LOGGER.Errorf("Error while delete JWT secure: %s", deleteErr.Error())
		jwt.ResponseError(w, http.StatusInternalServerError, deleteErr)
		return
	}

	// update token info
	token.Stage = jwt.Revoked
	token.RevokedAt = time.Now().Unix()
	token.RevokedBy = email

	// save updated token info
	updateResult, updateInfoErr := infoCollection.UpdateOne(ctx, bson.M{"institution": iid, "jti": input.JTI}, bson.M{"$set": &token})
	if updateInfoErr != nil || updateResult.MatchedCount < 1 {
		LOGGER.Errorf("Error update token info: %+v", updateInfoErr)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	LOGGER.Debugf("Successfully revoked JWT token")
	jwt.ResponseSuccess(w,"Successfully revoked JWT token")
	return
}

// #2 Second step for JWT creation
func (op *AuthOperations) HandleJWTApprove(w http.ResponseWriter, r *http.Request) {
	var input jwt.General
	//fid := r.Header.Get("x-fid")
	iid := r.Header.Get("x-iid")

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		LOGGER.Warningf("Error while validating token body :  %v", err)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	var token jwt.Info
	infoCollection, ctx := op.session.GetSpecificCollection(AuthDBName, JWTInfoCollection)
	err = infoCollection.FindOne(ctx,
		bson.M{
			"institution": iid,
			"jti":         input.JTI,
		}).Decode(&token)
	if err != nil {
		LOGGER.Errorf("Error getting JWT info from query: %s", err.Error())
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}
	LOGGER.Debugf("%+v", token)

	if &token == nil {
		LOGGER.Errorf("Token id info not found")
		jwt.ResponseError(w, http.StatusBadRequest, errors.New("token id info not found"))
		return
	}

	// only approve if stage is currently request
	if token.Stage != jwt.Review {
		LOGGER.Errorf("Token is not currently under review")
		jwt.ResponseError(w, http.StatusForbidden, errors.New("token is not currently under review"))
		return
	}

	email := r.Header.Get("email")
	// check to make sure the approver is not the same user (by email)
	if email == "" || email == token.CreatedBy {
		LOGGER.Errorf("Same user who created the token request cannot also approve")
		jwt.ResponseError(w, http.StatusForbidden, errors.New("same user who created the token request cannot also approve"))
		return
	}

	if token.JTI != input.JTI {
		LOGGER.Errorf("Unknown token id")
		jwt.ResponseError(w, http.StatusForbidden, errors.New("unknown token id"))
		return
	}

	// update token info
	token.Stage = jwt.Approved
	token.ApprovedAt = time.Now().Unix()
	token.ApprovedBy = email

	// save updated token info
	updateResult, updateInfoErr := infoCollection.UpdateOne(ctx,
		bson.M{
			"institution": iid,
			"jti": input.JTI,},
			bson.M{
				"$set": &token,
			})
	if updateInfoErr != nil || updateResult.MatchedCount < 1 {
		LOGGER.Errorf("Error update token info: %+v", updateInfoErr)
		jwt.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	LOGGER.Debugf("Successfully approved JWT token")
	jwt.ResponseSuccess(w,"Successfully approved JWT token")
	return
}
