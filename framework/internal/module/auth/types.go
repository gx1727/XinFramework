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

type registerRequest struct {
	Account  string `json:"account" binding:"required"` // 手机号或邮箱
	Password string `json:"password" binding:"required,min=6,max=32"`
	TenantID uint   `json:"tenant_id" binding:"required"` // 租户ID
	RealName string `json:"real_name"`                    // 真实姓名（可选）
}

type registerResult struct {
	Token string
	User  struct {
		ID       uint
		TenantID uint
		Code     string
		Role     string
	}
}
