package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/indiejustice/google-analytics/src" // fork of github.com/OzqurYalcin/google-analytics/config
	"github.com/indiejustice/redirection-tracking/pkg/client_cookie"
)

type Configuration struct {
	Port       string           `json:"port"`
	GAID       string           `json:"ga_id"`
	CookieName string           `json:"cookie_name"`
	Events     map[string]Event `json:"events"`
	Debug      bool             `json:"debug"`
}
type Event struct {
	URL      string `json:"url"`
	Category string `json:"category"`
	Action   string `json:"action"`
	Label    string `json:"label"`
	Value    string `json:"value"`
}

var config Configuration

var logError *log.Logger
var logInfo *log.Logger
var logDebug *log.Logger
var clientCookie *client_cookie.ClientCookie

func main() {
	file, err := ioutil.ReadFile("./config/config.json")
	panicError(err)

	json.Unmarshal(file, &config)

	logError = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	logInfo = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	if config.Debug {
		logDebug = log.New(os.Stdout, "DEBUG\t", log.Ldate|log.Ltime)
	} else {
		logDebug = log.New(ioutil.Discard, "", 0)
	}

	clientCookie = &client_cookie.ClientCookie{Name: config.CookieName}

	mux := http.NewServeMux()
	mux.HandleFunc("/", redirectionHandler)

	logInfo.Println("Redirection Server running on port:", config.Port)
	panic(http.ListenAndServe(":"+config.Port, mux))
}
func redirectionHandler(w http.ResponseWriter, r *http.Request) {
	var cid string

	w.Header().Set("Cache-Control", "no-store")

	cid, w = clientCookie.GetClientID(w, r)

	uriSegments := strings.Split(r.RequestURI, "/")

	var eventLabel string
	validLabel := regexp.MustCompile(`^[a-z0-9_]+$`)

	if len(uriSegments) > 2 && validLabel.MatchString(uriSegments[2]) {
		eventLabel = uriSegments[2]
		logDebug.Println("Event Label:", eventLabel)
	}

	if event, ok := config.Events[uriSegments[1]]; ok {

		api := new(ga.API)
		api.Lock()
		defer api.Unlock()
		api.UserAgent = r.UserAgent()
		api.ContentType = "application/x-www-form-urlencoded"

		client := new(ga.Client)
		client.ProtocolVersion = "1"
		client.TrackingID = config.GAID
		client.HitType = "event"
		client.ClientID = cid
		client.EventCategory = event.Category
		client.EventAction = event.Action
		client.EventLabel = eventLabel

		api.Send(client)

		http.Redirect(w, r, event.URL, 301)
		return
	}

	returnCode404(w, r)
}
func returnCode404(w http.ResponseWriter, r *http.Request) {
	// see http://golang.org/pkg/net/http/#pkg-constants
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Page Not Found"))
}
func logAndExit(message string) {
	logError.Println(message)
	os.Exit(1)
}
func panicError(err error) {
	if err != nil {
		logAndExit(err.Error())
	}
}
