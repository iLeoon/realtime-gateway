package user

type User struct {
	UserId   int    `json:"userId"`
	UserName string `json:"userName"`
	Email    string `json:"email"`
}
