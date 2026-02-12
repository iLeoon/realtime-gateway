package user

type User struct {
	UserId   int    `json:"id"`
	Email    string `json:"email"`
	UserName string `json:"userName"`
}
