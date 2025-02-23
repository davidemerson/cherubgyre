package interfaces

import "net/http"

type IndexHandler interface {
	GetRootHandler(w http.ResponseWriter, r *http.Request)
	PostRootHandler(w http.ResponseWriter, r *http.Request)
}
