package totp

type Response struct {
	Success    bool   `json:"success"`
	Registered bool   `json:"registered"`
	Msg        string `json:"msg"`
	Data       QRCode `json:"data"`
}

type QRCode struct {
	// The base64 encoded representation the the QR Code.
	QRCodeURI string `json:"qrcodeURI"`
	// A user friendly name for the registration
	AccountName string `json:"accountName"`
}

type TokenBody struct {
	Token string `json:"token"`
}

type RegistrationData struct {
	Email string `json:"email"`
}

type User struct {
	UID        string `json:"uid" bspn:"uid"`
	Secret     string `json:"secret" bson:"secret"`
	Registered bool   `json:"registered" bson:"registered"`
}