package user

type loginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	TenantID uint   `json:"tenant_id"`
}

type loginResult struct {
	Token        string
	RefreshToken string
	User         struct {
		ID       uint
		TenantID uint
		Code     string
		Role     string
	}
}

type registerRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	TenantID uint   `json:"tenant_id" binding:"required"`
	RealName string `json:"real_name"`
}

type registerResult struct {
	Token        string
	RefreshToken string
	User         struct {
		ID       uint
		TenantID uint
		Code     string
		Role     string
	}
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshResult struct {
	Token        string
	RefreshToken string
}
