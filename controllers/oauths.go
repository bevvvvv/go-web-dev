package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-web-dev/models"
	"io"
	"net/http"
	"time"

	fakecontext "go-web-dev/context"

	"github.com/gorilla/csrf"
	"golang.org/x/oauth2"
)

func NewOAuthController(oauthService models.OAuthService, dropboxOAuthConf *oauth2.Config) *OAuthController {
	return &OAuthController{
		oauthService:     oauthService,
		dropboxOAuthConf: dropboxOAuthConf,
	}
}

type OAuthController struct {
	oauthService     models.OAuthService
	dropboxOAuthConf *oauth2.Config
}

func (oauthController *OAuthController) DropboxConnect(w http.ResponseWriter, r *http.Request) {
	state := csrf.Token(r)

	cookie := http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	url := oauthController.dropboxOAuthConf.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (oauthController *OAuthController) DropboxCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	state := r.FormValue("state")
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if cookie == nil || cookie.Value != state {
		http.Error(w, "invalid state provided", http.StatusBadRequest)
	}
	cookie.Value = ""
	cookie.Expires = time.Now()
	http.SetCookie(w, cookie)

	code := r.FormValue("code")
	token, err := oauthController.dropboxOAuthConf.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "%+v", token)

	user := fakecontext.User(r.Context())

	existing, err := oauthController.oauthService.Find(user.ID, models.OAuthDropbox)
	if err == nil {
		oauthController.oauthService.Delete(existing.ID)
	} else if err != models.ErrNotFound {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userOAuth := models.OAuth{
		UserID:      user.ID,
		ServiceName: models.OAuthDropbox,
		Token:       *token,
	}
	err = oauthController.oauthService.Create(&userOAuth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (oauthController *OAuthController) DropboxTest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	path := r.FormValue("path")

	user := fakecontext.User(r.Context())
	userOAuth, err := oauthController.oauthService.Find(user.ID, models.OAuthDropbox)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token := userOAuth.Token

	data := struct {
		Path string `json:"path"`
	}{
		Path: path,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request, err := http.NewRequest(http.MethodPost, "https://api.dropboxapi.com/2/files/list_folder", bytes.NewReader(dataBytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request.Header.Add("Content-Type", "application/json")

	client := oauthController.dropboxOAuthConf.Client(context.TODO(), &token)
	response, err := client.Do(request)
	if err != nil {
		http.Error(w, err.Error(), response.StatusCode)
		return
	}
	defer response.Body.Close()

	io.Copy(w, response.Body)
}
