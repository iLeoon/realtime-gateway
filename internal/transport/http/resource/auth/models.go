package auth

type ProviderIdentity struct {
	ProviderID string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"displayName"`
	Provider   string `json:"provider"`
}

type User struct {
	UserID   string `json:"id"`
	Email    string `json:"email"`
	UserName string `json:"displayName"`
}
