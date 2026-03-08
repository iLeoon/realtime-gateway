package auth

type ProviderIdentity struct {
	ProviderID string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"displayName"`
	Provider   string `json:"provider"`
	PictureURL string `json:"picture_url"`
}

type User struct {
	UserID   string
	Email    string
	UserName string
}
