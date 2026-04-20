package auth

type loginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	TenantID uint   `json:"tenant_id"`
}

type loginResult struct {
	Token string
	User  struct {
		ID       uint
		TenantID uint
		Code     string
		Role     string
	}
}
