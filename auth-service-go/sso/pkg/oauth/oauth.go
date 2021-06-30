package oauth

import (
	"golang.org/x/oauth2"
	"os"
)

var (
	IBMIdOAuthConfig *oauth2.Config
	OAuthStateMap  map[string]OAuthMapItem
)

type OAuthMapItem struct {
	State string
	OriginalURL string
}

func init() {
	IBMIdOAuthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"openid"},
		Endpoint:     oauth2.Endpoint{
			AuthURL:   os.Getenv("AUTH_URL"),
			TokenURL:  os.Getenv("TOKEN_URL"),
			AuthStyle: 1,
		},
	}

	OAuthStateMap = make(map[string]OAuthMapItem)
}
