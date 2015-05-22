package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
)

const icon = "https://raw.githubusercontent.com/mailgun/media/master/Mailgun_Icon.png"

// Payload is the data provided to Slack Incoming Webhooks.
type Payload struct {
	Text      string `json:"text"`
	Username  string `json:"username"`
	Channel   string `json:"channel,omitempty"`
	IconURL   string `json:"icon_url,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", receiveWebhook).
		Methods("POST")

	http.ListenAndServe("0.0.0.0:8000", r)
}

func receiveWebhook(w http.ResponseWriter, r *http.Request) {
	if isInvalidSignature(r) {
		// TODO: Validate webhook signature.
	}

	from := r.FormValue("from")
	body := r.FormValue("stripped-text")

	postToSlack(from, body)
}

func isInvalidSignature(r *http.Request) bool {
	// TODO: Prevent replays by only accepting tokens once.
	// TODO: Prevent timing attacks by limiting accepted timestamps.

	token := r.FormValue("token")
	timestamp := r.FormValue("timestamp")
	signature := r.FormValue("signature")

	key := os.Getenv("MAILGUN_KEY")
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(token + timestamp))
	expectedSignature := mac.Sum(nil)

	return !hmac.Equal([]byte(signature), expectedSignature)
}

func postToSlack(from, body string) error {
	message := fmt.Sprintf("%s says, \"%s\"\n", from, body)
	payload := Payload{Username: "slackmail", IconURL: icon, Text: message}
	json, err := json.Marshal(&payload)

	if err != nil {
		return err
	}

	hook := os.Getenv("SLACK_WEBHOOK_URL")
	form := url.Values{"payload": {string(json)}}
	resp, err := http.PostForm(hook, form)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	return nil
}
