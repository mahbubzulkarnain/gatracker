package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/appengine"

	uuid "github.com/gofrs/uuid"
)

var gaPropertyID = mustGetenv("GA_TRACKING_ID")

func mustGetenv(k string) string {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("loading from os")
	}

	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s environment variable not set.", k)
	}
	return v
}

func main() {
	http.HandleFunc("/", handle)

	appengine.Main()
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if err := trackEvent(r, "Example", "Test action", "label", nil); err != nil {
		log.Fatal(fmt.Fprintf(w, "Event did not track: %v", err))
		return
	}
	log.Fatal(fmt.Fprint(w, "Event tracked."))
}

func trackEvent(r *http.Request, category, action, label string, value *uint) error {
	if gaPropertyID == "" {
		return errors.New("analytics: GA_TRACKING_ID environment variable is missing")
	}
	if category == "" || action == "" {
		return errors.New("analytics: category and action are required")
	}

	v := url.Values{
		"v":   {"1"},
		"tid": {gaPropertyID},
		// Anonymously identifies a particular user. See the parameter guide for
		// details:
		// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#cid
		//
		// Depending on your application, this might want to be associated with the
		// user in a cookie.
		"cid": {uuid.Must(uuid.NewV4()).String()},
		"t":   {"event"},
		"ec":  {category},
		"ea":  {action},
		"ua":  {r.UserAgent()},
	}

	if label != "" {
		v.Set("el", label)
	}

	if value != nil {
		v.Set("ev", fmt.Sprintf("%d", *value))
	}

	if remoteIP, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
		v.Set("uip", remoteIP)
	}

	// NOTE: Google Analytics returns a 200, even if the request is malformed.
	_, err := http.PostForm("https://www.google-analytics.com/collect", v)
	return err
}
