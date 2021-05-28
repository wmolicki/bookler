package handlers

import (
	stdCtx "context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"

	"github.com/wmolicki/bookler/context"
	"github.com/wmolicki/bookler/helpers"
	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/random"
)

func NewConfig() *oauth2.Config {
	conf := &oauth2.Config{
		ClientID:     "810036611838-k3ur24fbnamqvlu4stsorm47v2onlv0k.apps.googleusercontent.com",
		ClientSecret: "9PD3Fvk7qlraXgKM1T9eQQ0X",
		Scopes:       []string{"openid", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://oauth2.googleapis.com/token",
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		},
		RedirectURL: "http://localhost:3333/oauth/google/callback",
	}
	return conf
}

func NewOauthHandler(config *oauth2.Config, us *models.UserService) *OauthHandler {
	return &OauthHandler{config: config, us: us}
}

type OauthHandler struct {
	config *oauth2.Config
	us     *models.UserService
}

type tokenSignInData struct {
	IDToken string `json:"id_token"`
}

func (o *OauthHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	if user == nil {
		// TODO: probably an error or malicious action
		http.Redirect(w, r, "/", http.StatusFound)
	}

	cookie := http.Cookie{Name: models.AuthCookieName, Value: "", Expires: time.Now(), HttpOnly: true}
	http.SetCookie(w, &cookie)

	random_token, _ := random.RememberToken()
	user.RememberToken = random_token
	o.us.Update(user)

	http.Redirect(w, r, "/", http.StatusFound)
}

// TokenSignIn receives the id_token from google sign in, verifies it, will create user (if not exists) and user session
func (o *OauthHandler) TokenSignIn(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var data tokenSignInData
	err := decoder.Decode(&data)
	helpers.Must(err)

	ctx := stdCtx.TODO()
	payload, err := idtoken.Validate(ctx, data.IDToken, o.config.ClientID)
	// TODO: we should check expiry ourselves
	if err != nil {
		// Not a Google ID token.
		http.Error(w, fmt.Sprintf("bad stuff happened mate: %v", err), http.StatusInternalServerError)
		return
	}

	name := fmt.Sprintf("%v", payload.Claims["name"])
	profile_image_url := fmt.Sprintf("%v", payload.Claims["picture"])
	email := fmt.Sprintf("%v", payload.Claims["email"])

	var user *models.User
	user, err = o.us.ByEmail(email)

	switch err {
	case models.ErrorEntityNotFound:
		user, err = o.us.Create(email, name, profile_image_url)
		helpers.Must(err)
	case nil:

	default:
		panic(err)
	}

	token, err := o.us.SignIn(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad stuff happened mate: %v", err), http.StatusInternalServerError)
		return
	}

	signInCookie := http.Cookie{Name: models.AuthCookieName, Value: token, Path: "/", HttpOnly: true}
	http.SetCookie(w, &signInCookie)
	w.WriteHeader(http.StatusOK)
	return
}

func (o *OauthHandler) SetCookieRedirect(w http.ResponseWriter, r *http.Request) {
	state := csrf.Token(r)
	url := o.config.AuthCodeURL(state)

	cookie := http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, url, http.StatusFound)

}

func (o *OauthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.RequestURI)
	code := r.FormValue("code")
	state := r.FormValue("state")

	cookie, err := r.Cookie("oidc_state")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if cookie.Value != state {
		http.Error(w, "invalid state provided", http.StatusBadRequest)
		return
	}

	ctx := stdCtx.TODO()
	token, err := o.config.Exchange(ctx, code)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: clean this up
	fmt.Fprintf(w, "code: ", code, " state: ", state, " token: %v+", token)
}
