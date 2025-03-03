package indexhandler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dev3mike/go-api-swagger-boilerplate/internal/dtos"
	errorhandler "github.com/dev3mike/go-api-swagger-boilerplate/internal/handlers/error_handler"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/interfaces"
	internalerrors "github.com/dev3mike/go-api-swagger-boilerplate/pkg/constants/internal_errors"
	"github.com/dev3mike/go-xmapper"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type IndexHandler struct {
}

func NewIndexHandler() interfaces.IndexHandler {
	return &IndexHandler{}
}

// GetRootHandler godoc
// @Summary Get a welcome message
// @Description Responds with a welcome message for the API
// @Tags Sample API
// @Produce json
// @Success 200 {object} dtos.ResponseDTO
// @Router / [get]
func (h *IndexHandler) GetRootHandler(w http.ResponseWriter, r *http.Request) {
	responseDto := dtos.ResponseDTO{
		Message: "Welcome to the API",
	}
	render.Render(w, r, &responseDto)
}

// PostRootHandler godoc
// @Summary Post a username and get a profile
// @Description Takes a username from the URL and returns a profile if the username is valid
// @Tags Sample API
// @Param username path string true "Username"
// @Produce json
// @Success 200 {object} dtos.ProfileResponseDTO
// @Failure 400 {object} apierror.ErrResponse "Bad request"
// @Router /{username} [post]
func (h *IndexHandler) PostRootHandler(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	_, err := xmapper.ValidateSingleField(username, "validators:'required,minLength:4,maxLength:12'")
	if err != nil {

		if errors.Is(err, xmapper.ErrValidation) {
			errorhandler.HandleInternalError(w, r, internalerrors.ErrBadRequest)
			return
		}

		errorhandler.HandleInternalError(w, r, err)
		return
	}

	// Read token claims
	tokenClaims := r.Context().Value(dtos.TokenClaimsKey).(*dtos.UserClaims)
	clerkUserId := tokenClaims.Id
	fmt.Println(clerkUserId)

	profileDto := dtos.ProfileResponseDTO{
		Fullname: username,
	}

	render.Render(w, r, &profileDto)
}
