package jwt

import (
	"crypto/subtle"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

var TimeFunc = time.Now

type General struct {
	JTI   string `json:"jti"`
}

type Institution struct {
	ID    primitive.ObjectID             `json:"_id" bson:"_id"`
	Info  InstitutionInfo                `json:"info" bson:"info"`
	Nodes []InstitutionNode              `json:"nodes" bson:"nodes"`
}

type InstitutionInfo struct {
	Address1      string `json:"address1" bson:"address1"`
	Address2      string `json:"address2" bson:"address2"`
	City          string `json:"city" bson:"city"`
	Country       string `json:"country" bson:"country"`
	GeoLat        string `json:"geo_lat" bson:"geo_lat"`
	GeoLon        string `json:"geo_lon" bson:"geo_lon"`
	InstitutionId string `json:"institutionId" bson:"institutionId"`
	Kind          string `json:"kind" bson:"kind"`
	LogoUrl       string `json:"logo_url" bson:"logo_url"`
	Name          string `json:"name" bson:"name"`
	SiteUrl       string `json:"site_url" bson:"site_url"`
	Slug          string `json:"slug" bson:"slug"`
	State         string `json:"state" bson:"state"`
	Status        string `json:"status" bson:"status"`
	Zip           string `json:"zip" bson:"zip"`
}

type InstitutionNode struct {
	ApprovalIds   []string `json:"approvalIds" bson:"approvalIds"`
	BIC           string   `json:"bic" bson:"bic"`
	CountryCode   string   `json:"countryCode" bson:"countryCode"`
	Initialized   bool     `json:"initialized" bson:"initialized"`
	InstitutionId string   `json:"institutionId" bson:"institutionId"`
	ParticipantId string   `json:"participantId" bson:"participantId"`
	Role          string   `json:"role" bson:"role"`
	Status        []string `json:"status" bson:"status"`
	Version       string   `json:"version" bson:"version,omitempty"`
}

type InstitutionNodeUser struct {
	Profile Profile `json:"profile" bson:"profile"`
	Roles   Role    `json:"roles" bson:"roles"`
}

type Profile struct {
	Email string `json:"email" bson:"email"`
}

type Role struct {
	Admin bool `json:"admin" bson:"admin"`
}

type IVerifyCompare struct {
	Endpoint string
	IP       string
	Account  string
	JTI      string
}

type Stage string
const(
	Review Stage = "review"
	Approved = "approved"
	Ready = "ready"
	Initialized = "initialized"
	Revoked = "revoked"
)

type Info struct {
	Acc         []string `json:"acc" bson:"acc"`
	Active      bool     `json:"active" bson:"active"`
	ApprovedAt  int64    `json:"approvedAt" bson:"approvedAt"`
	ApprovedBy  string   `json:"approvedBy" bson:"approvedBy"`
	Aud         string   `json:"aud" bson:"aud"`
	CreatedAt   int64    `json:"createdAt" bson:"createdAt"`
	CreatedBy   string   `json:"createdBy" bson:"createdBy"`
	Description string   `json:"description" bson:"description"`
	Enp         []string `json:"enp" bson:"enp"`
	Env         string   `json:"env" bson:"env"`
	IPs         []string `json:"ips" bson:"ips"`
	JTI         string   `json:"jti" bson:"jti"`
	Stage       Stage    `json:"stage" bson:"stage"`
	Sub         string   `json:"sub" bson:"sub"`
	RevokedAt   int64    `json:"revokedAt" bson:"revokedAt"`
	RevokedBy   string   `json:"revokedBy" bson:"revokedBy"`
	RefreshedAt int64    `json:"refreshedAt" bson:"refreshedAt"`
	Ver         string   `json:"ver" bson:"ver"`
	Institution string   `json:"institution" bson:"institution"`
}

//type IRandomPepperObj struct {
//	// the old prefix key value for signing tokens
//	O int
//	// c = the current prefix key to use for signing tokens
//	// the current prefix that should be used to generated new pepper values
//	// format - prefix should be a single character a-z; convention = "{1-1:XXXXXRandomStringHereXXXXX}"
//	C int
//	// v = array of random values must be less than {4096 characters for aws secrets manager to store serialized data}
//	// old values should be provided in body and appended to new values array for lookup of new and old values
//	V map[string]string
//}

type IJWTSecure struct {
	ID     primitive.ObjectID `json:"_id" bson:"_id"`
	Secret string             `json:"secret" bson:"secret"`
	JTI    string             `json:"jti" bson:"jti"`
	Number int                `json:"number" bson:"number"`
}

type IJWTTokenClaim struct {
	jwt.StandardClaims
	Account     []string `json:"acc"`
	Version     string   `json:"ver"`
	IPs         []string `json:"ips"`
	Environment string   `json:"env"`
	Endpoints   []string `json:"enp"`
	Number      int      `json:"n"`
}

// Validates time based claims "exp, iat, nbf".
// There is no accounting for clock skew.
// As well, if any of the above claims are not in the token, it will still
// be considered a valid claim.
func (c IJWTTokenClaim) Valid() error {
	vErr := new(ValidationError)
	now := TimeFunc().Unix()

	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if c.VerifyExpiresAt(now, false) == false {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt, 0))
		vErr.Inner = fmt.Errorf("token is expired by %v", delta)
		vErr.Errors |= ValidationErrorExpired
	}

	if c.VerifyIssuedAt(now, false) == false {
		vErr.Inner = fmt.Errorf("Token used before issued")
		vErr.Errors |= ValidationErrorIssuedAt
	}

	if c.VerifyNotBefore(now, false) == false {
		vErr.Inner = fmt.Errorf("token is not valid yet")
		vErr.Errors |= ValidationErrorNotValidYet
	}

	if vErr.valid() {
		return nil
	}

	return vErr
}

// Compares the aud claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c *IJWTTokenClaim) VerifyAudience(cmp string, req bool) bool {
	return verifyAud(c.Audience, cmp, req)
}

// Compares the exp claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c *IJWTTokenClaim) VerifyExpiresAt(cmp int64, req bool) bool {
	return verifyExp(c.ExpiresAt, cmp, req)
}

// Compares the iat claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c *IJWTTokenClaim) VerifyIssuedAt(cmp int64, req bool) bool {
	return verifyIat(c.IssuedAt, cmp, req)
}

//// Compares the iss claim against cmp.
//// If required is false, this method will return true if the value matches or is unset
//func (c *IJWTTokenClaim) VerifyIssuer(cmp string, req bool) bool {
//	return verifyIss(c.Issuer, cmp, req)
//}

// Compares the nbf claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c *IJWTTokenClaim) VerifyNotBefore(cmp int64, req bool) bool {
	return verifyNbf(c.NotBefore, cmp, req)
}

// ----- helpers

func verifyAud(aud string, cmp string, required bool) bool {
	if aud == "" {
		return !required
	}
	if subtle.ConstantTimeCompare([]byte(aud), []byte(cmp)) != 0 {
		return true
	} else {
		return false
	}
}

func verifyExp(exp int64, now int64, required bool) bool {
	if exp == 0 {
		return !required
	}
	return now <= exp
}

func verifyIat(iat int64, now int64, required bool) bool {
	if iat == 0 {
		return !required
	}
	return now >= iat
}

func verifyIss(iss string, cmp string, required bool) bool {
	if iss == "" {
		return !required
	}
	if subtle.ConstantTimeCompare([]byte(iss), []byte(cmp)) != 0 {
		return true
	} else {
		return false
	}
}

func verifyNbf(nbf int64, now int64, required bool) bool {
	if nbf == 0 {
		return !required
	}
	return now >= nbf
}