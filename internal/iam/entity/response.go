package entity

type AuthResponse struct {
	Token string              `json:"token"`
	User  UserPayloadResponse `json:"user"`
}

type UserPayloadResponse struct {
	ID    uint   `json:"-"`
	UUID  string `json:"uuid"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
