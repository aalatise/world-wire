package sso

type General struct {
	Email string `json:"email"`
}

type User struct {
	UID   string `json:"uid"`
	Email string `json:"email"`
}

type RedirectData struct {
	Data string `json:"data"`
}
