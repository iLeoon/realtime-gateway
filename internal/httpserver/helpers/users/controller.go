package users

import (
	"net/http"
)

func GetUser(w http.ResponseWriter, r *http.Request, service UserServiceInterface) {

	data := service.GetUser()

	w.Write([]byte(data))
}
