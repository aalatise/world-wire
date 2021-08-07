package handler

import (
	"encoding/json"
	"errors"
	"github.com/IBM/world-wire/auth-service-go/permission"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
	"strings"
	"time"
)

func (op *AuthOperations) HandlePermissionParticipantUpdate(w http.ResponseWriter, r *http.Request) {
	LOGGER.Debugf("===== Update participant permission")

	var input permission.General
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		LOGGER.Warningf("Error while validating token body :  %v", err)
		permission.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	iid := r.Header.Get("x-iid")
	email := r.Header.Get("email")
	LOGGER.Debugf("Email in the header: %s, Email in the body: %s", email, input.Email)
	// requester should not be able to delete themselves
	if email == input.Email {
		LOGGER.Errorf("user may not delete themselves")
		permission.ResponseError(w, http.StatusConflict, errors.New("user may not delete themselves"))
		return
	}

	// validate required fields are present
	if input.Email == "" || input.Institution == "" || input.Role == "" {
		LOGGER.Errorf("Unknown email, role, and/or institution")
		permission.ResponseError(w, http.StatusBadRequest, errors.New("unknown email, role, and/or institution"))
		return
	}

	// Get user info from database by email address
	var uid string
	var user permission.User
	collection, ctx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
	err = collection.FindOne(ctx, bson.M{"profile": bson.M{"email": strings.ToLower(input.Email)}}).Decode(&user)
	if err != nil {
		// no user with this email exists, so create a new user
		LOGGER.Warning("No user with this email exists, so create a new user")
		objectID := primitive.NewObjectIDFromTimestamp(time.Now())
		newUser := permission.User{
			UID:   objectID,
			Profile: permission.Profile{Email: strings.ToLower(input.Email)},
		}

		// create new user data
		_, err := collection.InsertOne(ctx, newUser)
		if err != nil {
			LOGGER.Errorf("Insert new user data failed:  %+v", err)
			permission.ResponseError(w, http.StatusInternalServerError, err)
			return
		}
		uid = objectID.Hex()
	} else {
		uid = user.UID.Hex()
	}

	LOGGER.Debugf("uid: %s", uid)

	role := permission.Roles{}
	LOGGER.Debugf("Input role: %s", input.Role)
	// Set the portal user role
	switch input.Role {
	case permission.Admin:
		role.Admin = true
	case permission.Manager:
		role.Manager = true
	case permission.Viewer:
		role.Viewer = true
	default:
		LOGGER.Errorf("Undefined portal user role: %s", input.Role)
		permission.ResponseError(w, http.StatusBadRequest, err)
		return
	}

	userPermission := permission.ParticipantPermission{
		UID:           uid,
		InstitutionID: iid,
		Roles:         role,
	}

	// set/update participant permissions
	ppCollection, ppCtx := op.session.GetSpecificCollection(PortalDBName, ParticipantPermissionsCollection)
	updateResult := ppCollection.FindOneAndUpdate(ppCtx, bson.M{"user_id": uid, "institution_id": iid}, bson.M{"$set": &userPermission})
	if updateResult.Err() != nil {
		LOGGER.Warningf("Unable to update participant permission: %+v; Insert a new record", updateResult.Err())
		_, ppInsertErr := ppCollection.InsertOne(ppCtx, userPermission)
		if ppInsertErr != nil {
			LOGGER.Errorf("Unable to insert participant permissions: %s", ppInsertErr.Error())
			permission.ResponseError(w, http.StatusInternalServerError, ppInsertErr)
			return
		}
	}

	LOGGER.Debugf("Participant permission created/update for user: %s", uid)
	permission.ResponseSuccess(w, uid)
	return
}

func (op *AuthOperations) HandlePermissionSuperUpdate(w http.ResponseWriter, r *http.Request) {
	LOGGER.Debugf("===== Update super permission")

	var input permission.General
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		LOGGER.Warningf("Error while validating token body :  %v", err)
		permission.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	// validate required fields are present
	if input.Email == "" || input.Role == "" {
		LOGGER.Errorf("Unknown email, role in the payload")
		permission.ResponseError(w, http.StatusBadRequest, errors.New("unknown email, role in the payload"))
		return
	}

	email := r.Header.Get("email")
	LOGGER.Debugf("Email in the header: %s, Email in the body: %s", email, input.Email)
	// requester should not be able to delete themselves
	if email == input.Email {
		LOGGER.Errorf("user may not delete themselves")
		permission.ResponseError(w, http.StatusConflict, errors.New("user may not delete themselves"))
		return
	}

	// check that email ending ends in an allowable super admin address
	emailEnding := strings.Split(input.Email,"@")[1]
	LOGGER.Debugf("email ending is: %s", emailEnding)
	allowableEndings := []string{"ibm.com", "us.ibm.com", "sg.ibm.com", "in.ibm.com", os.Getenv("TEMP_EMAIL")}

	found := false
	for _, e := range allowableEndings {
		if e == emailEnding {
			found = true
			break
		}
	}

	if !found {
		LOGGER.Errorf("email ending must be " + strings.Join(allowableEndings, ","))
		permission.ResponseError(w, http.StatusForbidden, errors.New("email ending must be " + strings.Join(allowableEndings, ",")))
		return
	}

	// Get user info from database
	var uid string
	var user permission.User
	collection, ctx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
	err = collection.FindOne(ctx, bson.M{"profile": bson.M{"email": strings.ToLower(input.Email)}}).Decode(&user)
	if err != nil {
		// no user with this email exists, so create a new user
		LOGGER.Warning("No user with this email exists, so create a new user")
		objectID := primitive.NewObjectIDFromTimestamp(time.Now())
		newUser := permission.User{
			UID:     objectID,
			Profile: permission.Profile{Email: strings.ToLower(input.Email)},
		}

		// create new user data
		_, err := collection.InsertOne(ctx, newUser)
		if err != nil {
			LOGGER.Errorf("Insert new user data failed:  %+v", err)
			permission.ResponseError(w, http.StatusInternalServerError, err)
			return
		}
		user = newUser
	}
	uid = user.UID.Hex()
	LOGGER.Debugf("uid: %s", uid)

	role := permission.Roles{}

	LOGGER.Debugf("Input role: %s", input.Role)
	switch input.Role {
	case permission.Admin:
		role.Admin = true
	case permission.Manager:
		role.Manager = true
	case permission.Viewer:
		role.Viewer = true
	default:
		LOGGER.Errorf("Undefined user role: %s", input.Role)
		permission.ResponseError(w, http.StatusBadRequest, err)
		return
	}

	user.SuperPermissions.Role = role

	// set/update super permissions
	id, _ := primitive.ObjectIDFromHex(uid)
	updateResult := collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": &user})
	if updateResult.Err() != nil {
		LOGGER.Warningf("Unable to update user's super permissions: %+v; Insert a new record", updateResult.Err())
		_, spInsertErr := collection.InsertOne(ctx, user)
		if spInsertErr != nil {
			LOGGER.Errorf("Unable to insert user's super permissions: %s", spInsertErr.Error())
			permission.ResponseError(w, http.StatusInternalServerError, spInsertErr)
			return
		}
	}

	LOGGER.Debugf("Super permission created/updated for user: %s", uid)
	permission.ResponseSuccess(w, uid)
	return
}