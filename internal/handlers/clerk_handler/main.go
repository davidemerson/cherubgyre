package clerkhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dev3mike/go-api-swagger-boilerplate/internal/dtos"
	errorhandler "github.com/dev3mike/go-api-swagger-boilerplate/internal/handlers/error_handler"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/interfaces"
	clerkservice "github.com/dev3mike/go-api-swagger-boilerplate/internal/services/clerk"
	internalerrors "github.com/dev3mike/go-api-swagger-boilerplate/pkg/constants/internal_errors"
)

type ClerkHandler struct {
	clerkService *clerkservice.ClerkService
}

func NewClerkHandler(clerkService *clerkservice.ClerkService) interfaces.ClerkHandler {
	return &ClerkHandler{clerkService: clerkService}
}

// Handle processes the incoming HTTP request for creating a user event
// @Summary Create User Event
// @Description This endpoint processes a create user event and creates a user profile based on the event data.
// @Tags Clerk Webhooks
// @Accept json
// @Produce json
// @Param ClerkEventDto body dtos.ClerkEventDto true "Clerk Event Data"
// @Success 200
// @Failure 400 {object} apierror.ErrResponse "Bad request"
// @Failure 500 {object} apierror.ErrResponse "Internal server error"
// @Router /webhooks/clerk/create [post]
func (h *ClerkHandler) HandleCreateUserEvent(w http.ResponseWriter, r *http.Request) {
	eventDto := &dtos.ClerkEventDto{}
	err := json.NewDecoder(r.Body).Decode(eventDto)
	if err != nil {
		if err == io.EOF {
			errorhandler.HandleInternalError(w, r, internalerrors.ErrBadRequest)
			return
		}
		errorhandler.HandleInternalError(w, r, err)
		return
	}
	clerkUserId := eventDto.Data.ID

	fmt.Println("Clerk User ID: ", clerkUserId)

	// In here you can call the Clerk API to get the user details and create a profile
	//
	// user, error := h.clerkService.GetUser(clerkUserId)
	// if error != nil {
	// 	errorhandler.HandleInternalError(w, r, error)
	// 	return
	// }

	// fullname := fmt.Sprintf("%s %s", stringhelpers.NullableStringPointer(user.FirstName), stringhelpers.NullableStringPointer(user.LastName))

	// if err != nil {
	// 	errorhandler.HandleInternalError(w, r, err)
	// 	return
	// }

	// profileEntity := &entities.ProfileEntity{
	// 	PrivateEmail: &user.EmailAddresses[0].EmailAddress,
	// 	ProfileImage: &user.ProfileImageURL,
	// 	ExternalId:   &user.ID,
	// 	Fullname:     &fullname,
	// }
	// err = profileservice.CreateProfile(profileEntity)

	// if err != nil {
	// 	errorhandler.HandleInternalError(w, r, err)
	// 	return
	// }
}

// HandleUpdateUserEvent processes the incoming HTTP request for updating a user event
// @Summary Update User Event
// @Description This endpoint processes an update user event and updates the user profile based on the event data.
// @Tags Clerk Webhooks
// @Accept json
// @Produce json
// @Param ClerkEventDto body dtos.ClerkEventDto true "Clerk Event Data"
// @Success 200
// @Failure 400 {object} apierror.ErrResponse "Bad request"
// @Failure 500 {object} apierror.ErrResponse "Internal server error"
// @Router /webhooks/clerk/update [post]
func (h *ClerkHandler) HandleUpdateUserEvent(w http.ResponseWriter, r *http.Request) {
	eventDto := &dtos.ClerkEventDto{}
	err := json.NewDecoder(r.Body).Decode(eventDto)
	if err != nil {
		if err == io.EOF {
			errorhandler.HandleInternalError(w, r, internalerrors.ErrBadRequest)
			return
		}
		errorhandler.HandleInternalError(w, r, err)
		return
	}
	clerkUserId := eventDto.Data.ID

	fmt.Println("Clerk User ID: ", clerkUserId)

	// In here you can call the Clerk API to get the user details and update the profile
	//
	// user, error := h.clerkService.GetUser(clerkUserId)
	// if error != nil {
	// 	errorhandler.HandleInternalError(w, r, error)
	// 	return
	// }

	// fullname := fmt.Sprintf("%s %s", stringhelpers.NullableStringPointer(user.FirstName), stringhelpers.NullableStringPointer(user.LastName))

	// profileEntity := &entities.ProfileEntity{
	// 	PrivateEmail: &user.EmailAddresses[0].EmailAddress,
	// 	ProfileImage: &user.ProfileImageURL,
	// 	ExternalId:   &user.ID,
	// 	Fullname:     &fullname,
	// }
	// err = profileservice.UpdateProfilePartiallyByExternalId(clerkUserId, profileEntity)

	// if err != nil {
	// 	errorhandler.HandleInternalError(w, r, err)
	// 	return
	// }
}
