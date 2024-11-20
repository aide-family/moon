package auth

// EmailLoginParams 邮箱登录参数
type EmailLoginParams struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

// RegisterWithEmailParams 邮箱注册参数
type RegisterWithEmailParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}
