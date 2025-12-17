package models

type ProviderUser struct {
	ProviderID string
	Email      string
	Name       string
	Provider   string
}

type User struct {
	user_id int
}
