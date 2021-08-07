package handler

import (
	"errors"
	"github.com/IBM/world-wire/auth-service-go/idtoken"
	"github.com/IBM/world-wire/auth-service-go/permission"
	"github.com/IBM/world-wire/auth-service-go/totp"
	"github.com/IBM/world-wire/utility/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strings"
	"time"
)

type IRoutePermissions struct {
	SuperPermissions       []string
	ParticipantPermissions []string
}

var routePermissions = map[string]IRoutePermissions{
	"/permissions/participant": {
		SuperPermissions:       []string{"admin", "manager"},
		ParticipantPermissions: []string{"admin"},
	},
	"/permissions/super": {
		SuperPermissions:       []string{"admin"},
		ParticipantPermissions: []string{},
	},
	"/jwt/request": {
		SuperPermissions:       []string{"admin", "manager"},
		ParticipantPermissions: []string{"admin", "manager"},
	},
	"/jwt/revoke": {
		SuperPermissions:       []string{"admin", "manager"},
		ParticipantPermissions: []string{"admin", "manager"},
	},
	"/jwt/generate": {
		SuperPermissions:       []string{"admin", "manager"},
		ParticipantPermissions: []string{"admin", "manager"},
	},
	"/jwt/reject": {
		SuperPermissions:       []string{"admin", "manager"},
		ParticipantPermissions: []string{"admin", "manager"},
	},
	"/jwt/verify": {
		SuperPermissions:       []string{"admin", "manager"},
		ParticipantPermissions: []string{"admin", "manager"},
	},
	"/jwt/approve": {
		SuperPermissions:       []string{"admin", "manager"},
		ParticipantPermissions: []string{"admin"},
	},
}

// This middleware will check if the provided ID token contained the user ID and will use this ID to query the record
// in the DB to check if this user exist. Later on, set the email in the header for later use.
func (op *AuthOperations) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LOGGER.Debugf("----- Middleware: authenticate portal user")

		idToken := r.Header.Get("x-fid")
		LOGGER.Debugf("ID token is: %s", idToken)

		if idToken == "" {
			LOGGER.Error("unauthorized")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("unauthorized"))
			return
		}

		// get the user id from the x-fid
		claims, err := idtoken.Parse(idToken)
		if err != nil {
			LOGGER.Error("Could not Decode FID %s ", err.Error())
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1011", errors.New("fID could not be parsed"))
			return
		}
		userID := claims.UID
		LOGGER.Debugf("User ID in the id token is: %s", userID)

		if userID == "" {
			LOGGER.Error("unauthorized: empty user ID")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("unauthorized: empty user ID"))
			return
		}

		r.Header.Set("uid", userID)

		var user permission.User
		id, _ := primitive.ObjectIDFromHex(userID)
		collection, ctx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
		err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
		if err != nil {
			LOGGER.Errorf("No user found for this user id: %s; Error: %s", userID, err.Error())
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("no user found for this user id: " + userID))
			return
		}
		LOGGER.Debugf("%+v", user)

		r.Header.Set("email", user.Profile.Email)

		next.ServeHTTP(w, r)
	})
}

// This middleware will extract the request path and check if this user has the permission to access this endpoint
func (op *AuthOperations) CheckPermissions(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LOGGER.Debugf("----- Middleware: check permissions")

		LOGGER.Debugf("request URL path: %s", r.URL.Path)
		requiredPermissionsForAccessingRoute := routePermissions[r.URL.Path]
		LOGGER.Debugf("Route permission: %+v", requiredPermissionsForAccessingRoute)

		// default access init to false
		allowAccess := false

		uid := r.Header.Get("uid")
		email := r.Header.Get("email")
		LOGGER.Debugf("User ID in the header is: %s", uid)
		LOGGER.Debugf("Email in the header is: %s", email)

		if uid == "" && email == "" {
			LOGGER.Errorf("Missing user ID and email in the header, request rejected")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("missing user ID and email in the header, request rejected"))
			return
		}

		if uid == "" {
		// get uid by the signed in IBMId user
			LOGGER.Warningf("User ID not set in the header, get the user ID from DB using the email")
			var user permission.User
			collection, ctx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
			err := collection.FindOne(ctx, bson.M{"profile": bson.M{"email": strings.ToLower(email)}}).Decode(&user)
			if err != nil {
				LOGGER.Errorf("No user found for this user email: %s", err.Error())
				response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("no user found for this user email"))
				return
			}
			LOGGER.Debugf("%+v", user)

			uid = user.UID.Hex()
		}

		LOGGER.Debugf("Length of super permission: %d", len(requiredPermissionsForAccessingRoute.SuperPermissions))
		if len(requiredPermissionsForAccessingRoute.SuperPermissions) > 0 {
			// get super Permissions
			users, ok := op.setUserSuperPermissions(uid)
			if !ok {
				LOGGER.Warningf("No super permission record found for this user id: %s", uid)
			} else {
				role := ""
				if users.SuperPermissions.Role.Admin {
					role = "admin"
				} else if users.SuperPermissions.Role.Manager {
					role = "manager"
				} else {
					role = "viewer"
				}

				for _, r := range requiredPermissionsForAccessingRoute.SuperPermissions {
					LOGGER.Debugf("role: %s, r: %s", role, r)
					if role == r {
						allowAccess = true
						break
					}
				}
			}
		}

		// if super permissions have already permitted access to this route then skip checking participant permissions
		if allowAccess == false {

			// user was not granted super permissions to access this route, so check if they have permissable participant permissions

			LOGGER.Debugf("Length of participant permission: %d", len(requiredPermissionsForAccessingRoute.ParticipantPermissions))
			// check if the route allows users with participant permissions to access route
			if len(requiredPermissionsForAccessingRoute.ParticipantPermissions) > 0 {

				iid := r.Header.Get("x-iid")
				// iid header is required to determine if the user has access rights for a specific institution "participant" permissions
				// iid = institution id which is required
				LOGGER.Debugf("Institution ID: %s", iid)
				if iid == "" {
					LOGGER.Errorf("Unknown institution id")
					response.NotifyWWError(w, r, http.StatusBadRequest, "AUTH-1001", errors.New("unknown institution id"))
					return
				}
				// set iid on request (consistent with how developer token will set iid on req)

				// get participant permissions
				participantPermissions, ok := op.setUserParticipantPermissions(uid, iid)
				if !ok {
					LOGGER.Warningf("No participant permission record found for this user id: %s", uid)
				} else {
					role := ""
					if participantPermissions.Roles.Admin {
						role = "admin"
					} else if participantPermissions.Roles.Manager {
						role = "manager"
					} else {
						role = "viewer"
					}

					for _, r := range requiredPermissionsForAccessingRoute.ParticipantPermissions {
						if role == r {
							allowAccess = true
							break
						}
					}
				}
			}
		}

		LOGGER.Debugf("*** allow access status: %t ***", allowAccess)

		// evaluate if the user has access rights to proceed
		if allowAccess {
			// user has access rights to access route
			next.ServeHTTP(w, r)
		} else {
			LOGGER.Errorf("No user permissions to manage this institution")
			response.NotifyWWError(w, r, http.StatusForbidden, "AUTH-1001", errors.New("no user permissions to manage this institution"))
			return
		}
	})
}

func (op *AuthOperations) setUserSuperPermissions(uid string) (permission.User, bool) {
	var p permission.User
	collection, ctx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
	id, _ := primitive.ObjectIDFromHex(uid)
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		LOGGER.Errorf("Error during user's super permission query: %s", err.Error())
		return permission.User{}, false
	}
	LOGGER.Debugf("%+v", p)

	return p, true
}

func (op *AuthOperations) setUserParticipantPermissions(uid, iid string) (permission.ParticipantPermission, bool) {
	var p permission.ParticipantPermission
	collection, ctx := op.session.GetSpecificCollection(PortalDBName, ParticipantPermissionsCollection)
	err := collection.FindOne(ctx, bson.M{"user_id": uid, "institution_id": iid}).Decode(&p)
	if err != nil {
		LOGGER.Errorf("Error during participant permission query: %s", err.Error())
		return permission.ParticipantPermission{}, false
	}
	LOGGER.Debugf("%+v", p)

	return p, true
}

func (op *AuthOperations) CheckTOTPMiddleWareIBMIdUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LOGGER.Debugf("----- Middleware: check TOTP IBM ID User")

		// check if request contains user information from database user session authentication
		accountName := r.Header.Get("email")
		cacheKey := ""
		if accountName == "" {
			for t := time.Now().Unix(); t > time.Now().Unix() - 90; t-- {
				value, found := op.c.Get(string(t))
				if found {
					accountName = value.(string)
					cacheKey = string(t)
					op.c.Delete(cacheKey)
					break
				}
			}
		}

		if cacheKey == "" {
			LOGGER.Errorf("unauthorized, missing email")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("unauthorized, missing email"))
			return
		}
		LOGGER.Debugf("Email: %s", accountName)
		r.Header.Set("email", accountName)

		token := r.Header.Get("x-verify-code")
		LOGGER.Debugf("Token: %s", token)
		if token == "" {
			// missing header
			LOGGER.Errorf("unauthorized, missing token")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("unauthorized, missing token"))
			return
		}

		var user permission.User
		userCollection, userCtx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
		err := userCollection.FindOne(userCtx, bson.M{"profile": bson.M{"email": strings.ToLower(accountName)}}).Decode(&user)
		if err != nil {
			LOGGER.Errorf("Error during user query: %s", err.Error())
			response.NotifyWWError(w, r, http.StatusInternalServerError, "AUTH-1001", errors.New("no user record found"))
			return
		}

		var tUser totp.User
		collection, ctx := op.session.GetSpecificCollection(PortalDBName, TOTPCollection)
		err = collection.FindOne(ctx, bson.M{"uid": user.UID.Hex()}).Decode(&tUser)
		if err != nil {
			LOGGER.Errorf("Error during totp query: %s", err.Error())
			response.NotifyWWError(w, r, http.StatusInternalServerError, "AUTH-1001", errors.New("no totp record found"))
			return
		}

		tokenBody := totp.TokenBody{
			Token: token,
		}

		if totp.Check(tUser, tokenBody) {
			LOGGER.Debugf("totp two-factor authentication success")
			next.ServeHTTP(w, r)
		} else {
			LOGGER.Errorf("totp two-factor authentication failed")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("totp two-factor authentication failed"))
			return
		}
	})
}

func (op *AuthOperations) CheckTOTPMiddleWarePortalUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LOGGER.Debugf("----- Middleware: check TOTP Portal User")

		// check if request contains user information from database user session authentication
		accountName := r.Header.Get("email")
		token := r.Header.Get("x-verify-code")

		LOGGER.Debugf("Email: %s", accountName)
		LOGGER.Debugf("Token: %s", token)

		if token == "" {
			// missing header
			LOGGER.Errorf("unauthorized: missing token")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("unauthorized: missing token"))
			return
		}

		var user permission.User
		userCollection, userCtx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
		err := userCollection.FindOne(userCtx, bson.M{"profile": bson.M{"email": strings.ToLower(accountName)}}).Decode(&user)
		if err != nil {
			LOGGER.Errorf("Error during user query: %s", err.Error())
			response.NotifyWWError(w, r, http.StatusInternalServerError, "AUTH-1001", errors.New("no user record found"))
			return
		}

		var tUser totp.User
		collection, ctx := op.session.GetSpecificCollection(PortalDBName, TOTPCollection)
		err = collection.FindOne(ctx, bson.M{"uid": user.UID.Hex()}).Decode(&tUser)
		if err != nil {
			LOGGER.Errorf("Error during totp query: %s", err.Error())
			response.NotifyWWError(w, r, http.StatusInternalServerError, "AUTH-1001", errors.New("no totp record found"))
			return
		}

		tokenBody := totp.TokenBody{
			Token: token,
		}

		if totp.Check(tUser, tokenBody) {
			LOGGER.Debugf("totp two-factor authentication success")
			next.ServeHTTP(w, r)
		} else {
			LOGGER.Errorf("totp two-factor authentication failed")
			response.NotifyWWError(w, r, http.StatusUnauthorized, "AUTH-1001", errors.New("totp two-factor authentication failed"))
			return
		}
	})
}