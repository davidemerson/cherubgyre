package errorhandler

import (
	"errors"
	"net/http"

	"github.com/dev3mike/go-api-swagger-boilerplate/cmd/server/logger"
	internalerrors "github.com/dev3mike/go-api-swagger-boilerplate/pkg/constants/internal_errors"
	apierror "github.com/dev3mike/go-api-swagger-boilerplate/pkg/utils/api_error"
)

func HandleInternalError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, internalerrors.ErrNotFound) {
		apierror.NotFoundError(w, r)
		return
	}

	if errors.Is(err, internalerrors.ErrBadRequest) {
		apierror.BadRequestError(w, r)
		return
	}

	if errors.Is(err, internalerrors.ErrUnauthorized) {
		apierror.UnauthorizedError(w, r)
		return
	}

	// Fallback to internal server error
	logger.Logger.Error("Internal server error: %v", err)
	apierror.InternalServerError(w, r)
}
