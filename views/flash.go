package views

import (
	"encoding/base64"
	"net/http"
	"time"
)

const alertLevelCookie = "_bkr_alert_lvl"
const alertMessageCookie = "_bkr_alert_msg"

const AlertLevelError = "danger"
const AlertLevelSuccess = "success"

type Message struct {
	Level string
	Text  string
}

func setCookie(w http.ResponseWriter, name, value string, expires time.Time) {
	data := base64.StdEncoding.EncodeToString([]byte(value))
	c := http.Cookie{Name: name, Value: data, HttpOnly: true, Expires: expires, Path: "/"}
	http.SetCookie(w, &c)
}

func getCookie(r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	val, err := base64.URLEncoding.DecodeString(c.Value)
	message := string(val)
	if err != nil {
		return "", err
	}

	return message, nil
}

func ClearCookies(w http.ResponseWriter) {
	// we overwrite cookies so next request will consume message and the cookie won't get resent
	http.SetCookie(w, &http.Cookie{Name: alertLevelCookie, Expires: time.Now(), HttpOnly: true, Value: "", Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: alertMessageCookie, Expires: time.Now(), HttpOnly: true, Value: "", Path: "/"})
}

func SetMessage(w http.ResponseWriter, m *Message) {
	expiresAt := time.Now().Add(5 * time.Minute)
	setCookie(w, alertLevelCookie, m.Level, expiresAt)
	setCookie(w, alertMessageCookie, m.Text, expiresAt)
}

func GetMessage(w http.ResponseWriter, r *http.Request) *Message {
	level, err := getCookie(r, alertLevelCookie)
	if err != nil {
		return nil
	}
	text, err := getCookie(r, alertMessageCookie)
	if err != nil {
		return nil
	}
	// clearCookies(w)
	return &Message{Level: level, Text: text}
}

func FlashSuccess(w http.ResponseWriter, message string) {
	SetMessage(w, &Message{Level: AlertLevelSuccess, Text: message})
}

func FlashError(w http.ResponseWriter, message string) {
	SetMessage(w, &Message{Level: AlertLevelError, Text: message})
}
