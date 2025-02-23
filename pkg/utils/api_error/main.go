package apierror

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	Message   string `json:"message"`         // user-level status message
	ErrorCode string `json:"code"`            // user-level error code
	ErrorText string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

var (
	errBadRequest          = &ErrResponse{HTTPStatusCode: 400, Message: "Bad request", ErrorCode: "bad_request"}
	errUnauthorized        = &ErrResponse{HTTPStatusCode: 401, Message: "Unauthorized", ErrorCode: "unauthorized"}
	errNotFound            = &ErrResponse{HTTPStatusCode: 404, Message: "Resource not found.", ErrorCode: "not_found"}
	errInternalServerError = &ErrResponse{HTTPStatusCode: 500, Message: "Internal Server Error", ErrorCode: "internal_server_error"}
)

func BadRequestError(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, errBadRequest)
}

func UnauthorizedError(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, errUnauthorized)
}

func NotFoundError(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, errNotFound)
}

func InternalServerError(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, errInternalServerError)
}
