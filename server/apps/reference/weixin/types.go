package weixin

// Code2SessionRequest 小程序登录请求
type Code2SessionRequest struct {
	Code string `json:"code" binding:"required"`
}

// Code2SessionResponse 小程序登录响应
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
}

// PhoneNumberRequest 获取手机号请求
type PhoneNumberRequest struct {
	Code string `json:"code" binding:"required"`
}

// PhoneNumberResponse 获取手机号响应
type PhoneNumberResponse struct {
	PhoneInfo PhoneInfo `json:"phone_info"`
}

type PhoneInfo struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
	Watermark       struct {
		Timestamp int64  `json:"timestamp"`
		AppID     string `json:"appid"`
	} `json:"watermark"`
}

// LoginResult 登录结果
type LoginResult struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	User         LoginUser `json:"user"`
	IsNewUser    bool      `json:"is_new_user"`
}

type LoginUser struct {
	ID       uint   `json:"id"`
	OpenID   string `json:"openid"`
	UnionID  string `json:"unionid,omitempty"`
	Phone    string `json:"phone,omitempty"`
	TenantID uint   `json:"tenant_id"`
	Code     string `json:"code"`
	Role     string `json:"role"`
	Status   int16  `json:"status"`
}
