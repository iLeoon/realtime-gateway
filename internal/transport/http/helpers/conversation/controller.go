package conversation

import (
	"encoding/json"
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/validation"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func Create(w http.ResponseWriter, r *http.Request, convService Service) {
	var reqBody CreateConversationReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.Error("Error on decoding the request body", "Error", err)
		apiErr := apierror.Build(apierror.BadRequest, "Invalid Json format",
			apierror.WithTarget("ConversationRequest"),
			apierror.WithInnerError(apierror.InnerError{
				Code: "InvalidFieldDataTypeOrInvalidJSONFormat",
			}),
		)
		apierror.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	if err := validation.Validate(reqBody); err != nil {
		apiErr := apierror.Build(apierror.BadRequest, "Invalid arguments",
			apierror.WithTarget("Conversation request"),
			apierror.WithDetails(err),
		)
		apierror.Send(w, http.StatusBadRequest, apiErr)
	}

}

func List(w http.ResponseWriter, r *http.Request, convService Service) {

}

func GetConversation(w http.ResponseWriter, r *http.Request, convService Service) {

}
