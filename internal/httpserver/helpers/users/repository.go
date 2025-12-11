package users

type UserRepo struct {
	Database string
}

func NewUserRepo(s string) *UserRepo {
	return &UserRepo{
		Database: s,
	}
}
