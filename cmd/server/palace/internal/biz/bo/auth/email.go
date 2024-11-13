package auth

type EmailLoginParams struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

type RegisterWithEmailParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}
