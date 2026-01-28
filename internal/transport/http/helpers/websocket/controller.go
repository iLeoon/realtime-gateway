package websocket

import (
	"net/http"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/token"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/pkg/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func GenerateTicket(w http.ResponseWriter, r *http.Request, t token.Service) {
	//Get the user id from request context
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		logger.Error("The userId was not attached to the request")
		apiErr := apierror.Build(apierror.InternalServerError, "The user validation flow has failed",
			apierror.WithTarget("Authorization"),
			apierror.WithInnerError(apierror.InnerError{
				Code: "MissingUserIdContext",
			}),
		)
		apierror.Send(w, http.StatusInternalServerError, apiErr)
		return
	}

	userId, err := strconv.Atoi(authenticatedId)
	if err != nil {
		logger.Error("Error on trying to convert the contextId to a number", "Error", err)
		apiErr := apierror.Build(apierror.InternalServerError, "user id is not a valid id number", apierror.WithTarget("Id"), apierror.WithInnerError(apierror.InnerError{
			Code: "InvalidContextIdType",
		}))
		apierror.Send(w, http.StatusInternalServerError, apiErr)
		return
	}

	jwtToken, err := t.GenerateWsToken(userId)
	if err != nil {
		logger.Error("Error on generating the websocket token", "Error", err)
		apiErr := apierror.Build(apierror.InternalServerError, "Unexpected error while trying to authenticate websocket",
			apierror.WithTarget("token"),
			apierror.WithInnerError(apierror.InnerError{
				Code: "GeneratingWsTokenFailed",
			}))
		apierror.Send(w, http.StatusInternalServerError, apiErr)
		return
	}

	resp := ResponseTicket{Ticket: jwtToken}
	apiresponse.Send(w, http.StatusOK, resp)
}
