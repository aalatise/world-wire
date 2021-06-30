package permission

import "go.mongodb.org/mongo-driver/bson/primitive"

type Role string
const (
	Admin Role = "admin"
	Manager = "manager"
	Viewer = "viewer"
)

type General struct {
	Institution string `json:"institutionId"`
	Email       string `json:"email"`
	Role        Role   `json:"role"`
}

type User struct {
	UID              primitive.ObjectID `json:"_id" bson:"_id"`
	Profile          Profile            `json:"profile" bson:"profile"`
	SuperPermissions SuperPermission    `json:"super_permission" bson:"super_permission"`
}

type Profile struct {
	Email string `json:"email" bson:"email"`
}

type ParticipantPermission struct {
	UID           string `json:"user_id" bson:"user_id"`
	InstitutionID string `json:"institution_id" bson:"institution_id"`
	Roles         Roles  `json:"roles" bson:"roles"`
}

type SuperPermission struct {
	Role Roles `json:"roles" bson:"roles"`
}

type Data struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Roles       Roles  `json:"roles"`
	Slug        string `json:"slug"`
}

type Roles struct {
	Admin   bool `json:"admin" bson:"admin"`
	Manager bool `json:"manager" bson:"manager"`
	Viewer  bool `json:"viewer" bson:"viewer"`
}