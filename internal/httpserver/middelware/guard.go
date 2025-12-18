package middelware

import (
	"fmt"
	"net/http"
)

func Guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Before")
		next.ServeHTTP(w, r)
		fmt.Println("This is my handler")
	})
}
