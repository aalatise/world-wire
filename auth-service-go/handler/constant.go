package handler

const (
	// Mongo database name
	AuthDBName = "auth"
	PortalDBName = "portal-db"

	// Mongo collection name
	UserCollection = "users"
	JWTInfoCollection = "jwt_info"
	JWTSecureCollection = "jwt_secure"
	InstitutionsCollection = "institutions"
	ParticipantPermissionsCollection = "participant_permissions"
	ParticipantApprovalsCollection = "participant_approvals"
	SuperApprovalsCollection = "super_approvals"
	IDTokenSecureCollection = "id_token_secure"
	TOTPCollection = "totp"
)
