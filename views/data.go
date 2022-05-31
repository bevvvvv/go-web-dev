package views

import (
	"go-web-dev/models"
	"log"
	"net/http"
	"time"
)

const (
	AlertLevelError   = "danger"
	AlertLevelWarning = "warning"
	AlertLevelInfo    = "info"
	AlertLevelSuccess = "success"

	AlertMessageGeneric = "Something went wrong. Please try again and contact us if problem persists."

	alertCookieTTLMin = 5
)

type Alert struct {
	Level   string
	Message string
}

type Data struct {
	Alert *Alert
	User  *models.User
	Yield interface{}
}

func (data *Data) SetAlert(err error) {
	if publicErr, ok := err.(PublicError); ok {
		data.Alert = &Alert{
			Level:   AlertLevelError,
			Message: publicErr.Public(),
		}
	} else {
		log.Println(err)
		data.Alert = &Alert{
			Level:   AlertLevelError,
			Message: AlertMessageGeneric,
		}
	}
}

func (data *Data) AlertError(msg string) {
	data.Alert = &Alert{
		Level:   AlertLevelError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}

func persistAlert(w http.ResponseWriter, alert Alert) {
	expiresAt := time.Now().Add(alertCookieTTLMin * time.Minute)
	level := http.Cookie{
		Name:     "alert_level",
		Value:    alert.Level,
		Expires:  expiresAt,
		HttpOnly: true,
	}
	message := http.Cookie{
		Name:     "alert_message",
		Value:    alert.Message,
		Expires:  expiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, &level)
	http.SetCookie(w, &message)
}

func clearAlert(w http.ResponseWriter) {
	expiresAt := time.Now()
	level := http.Cookie{
		Name:     "alert_level",
		Value:    "",
		Expires:  expiresAt,
		HttpOnly: true,
	}
	message := http.Cookie{
		Name:     "alert_message",
		Value:    "",
		Expires:  expiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, &level)
	http.SetCookie(w, &message)
}

func getAlert(r *http.Request) *Alert {
	levelCookie, err := r.Cookie("alert_level")
	if err != nil {
		return nil
	}
	messageCookie, err := r.Cookie("alert_message")
	if err != nil {
		return nil
	}
	if levelCookie.Value == "" {
		return nil
	}
	return &Alert{
		Level:   levelCookie.Value,
		Message: messageCookie.Value,
	}
}

func RedirectAlert(w http.ResponseWriter, r *http.Request, urlStr string, code int, alert Alert) {
	persistAlert(w, alert)
	http.Redirect(w, r, urlStr, code)
}
