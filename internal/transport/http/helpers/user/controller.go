package user

import (
	"errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/pkg/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"net/http"
	"strconv"
)

func GetUserProfile(w http.ResponseWriter, r *http.Request, userService Service) {
	//Get the user id from the context
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiErr := apierror.Build(apierror.InternalServerError, "The user validation flow has failed",
			apierror.WithTarget("Authorization"),
			apierror.WithInnerError(apierror.InnerError{
				Code: "MissingUserIdContext",
			}),
		)
		apierror.Send(w, http.StatusInternalServerError, apiErr)
		return
	}

	// Get the user id from the path url
	targetId := r.PathValue("id")

	// Convert and validate the incoming id in the path url.
	id, err := strconv.Atoi(targetId)
	if err != nil {
		apiErr := apierror.Build(apierror.BadRequest, "Invalid id type",
			apierror.WithTarget("userId"),
			apierror.WithInnerError(apierror.InnerError{
				Code: "InvalidIdUsedInThePath",
			}),
		)
		apierror.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	// Check if the user is authorized to check this resource
	if authenticatedId != targetId {
		apiErr := apierror.Build(apierror.ForbiddenRequest, "User is not authorized to check this resource",
			apierror.WithInnerError(apierror.InnerError{
				Code: "AuthenticatedIdNotMatchingTargetId",
			}),
		)
		apierror.Send(w, http.StatusForbidden, apiErr)
		return
	}

	user, err := userService.GetUser(id, r.Context())
	if err != nil {
		if errors.Is(err, UserNotFoundErr) {
			logger.Info("Id was not found in the database", "Error", err.Error())
			apiErr := apierror.Build(apierror.NotFoundRequest, "No user was found",
				apierror.WithTarget("user"),
				apierror.WithInnerError(apierror.InnerError{
					Code: "NoUserWasFoundWithThatId",
				}),
			)
			apierror.Send(w, http.StatusNotFound, apiErr)
			return
		} else {
			logger.Error("Unexpected error", "Error", err.Error())
			apiErr := apierror.Build(apierror.InternalServerError, "Unexpected error while trying to process the request",
				apierror.WithTarget("user"),
				apierror.WithInnerError(apierror.InnerError{
					Code: "GetUserFailed",
				}),
			)
			apierror.Send(w, http.StatusInternalServerError, apiErr)
			return
		}

	}
	apiresponse.Send(w, http.StatusOK, user)
}
