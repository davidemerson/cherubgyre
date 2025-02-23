package interfaces

import "net/http"

type ClerkHandler interface {
	HandleCreateUserEvent(w http.ResponseWriter, r *http.Request)
	HandleUpdateUserEvent(w http.ResponseWriter, r *http.Request)
}
