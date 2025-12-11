package users

type UserServiceInterface interface {
	GetUser() string
}

type UserService struct {
	repo string
}

func NewUserService(s string) *UserService {
	return &UserService{
		repo: s,
	}
}

func (u *UserService) GetUser() string {
	return "Hello there"
}
