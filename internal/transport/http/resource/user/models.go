package user

type User struct {
	UserId   int    `json:"id"`
	Email    string `json:"email"`
	UserName string `json:"displayName"`
	Image    string `json:"displayImage"`
}

type FriendsList struct {
	Value []User `json:"value"`
}
