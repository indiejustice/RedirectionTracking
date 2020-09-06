package client_cookie

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ClientCookie struct {
	Name string
}

func (clientCookie *ClientCookie) GetClientID(w http.ResponseWriter, r *http.Request) (cid string, writer http.ResponseWriter) {
	cookie, _ := r.Cookie(clientCookie.Name)

	if cookie != nil && cookie.Value != "" {
		cid = cookie.Value
	} else {
		cid = uuid.New().String()

		expiration := time.Now().Add(2 * 365 * 24 * time.Hour)
		cookie := http.Cookie{Name: clientCookie.Name, Value: cid, Expires: expiration}
		http.SetCookie(w, &cookie)
	}

	writer = w

	return
}
