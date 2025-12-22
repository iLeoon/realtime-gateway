package users

import (
	"encoding/json"
	"net/http"

	"github.com/iLeoon/realtime-gateway/pkg/ctx"
)

func GetUser(w http.ResponseWriter, r *http.Request) {

	userID, _ := ctx.GetUserIDCtx(r.Context())
	json.NewEncoder(w).Encode(map[string]any{
		"key":    "Hi",
		"userID": userID,
	})

}
