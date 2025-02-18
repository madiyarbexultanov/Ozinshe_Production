package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/spf13/viper"
)

var GoogleOAuthConfig = &oauth2.Config{
    ClientID:     viper.GetString("GOOGLE_CLIENT_ID"),
    ClientSecret: viper.GetString("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  viper.GetString("GOOGLE_REDIRECT_URL"),
    Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
    Endpoint:     google.Endpoint,
}