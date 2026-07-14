package request

type Login struct {
	UserID   string `json:"user_id" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}
type Register struct {
	UserID    string `json:"user_id" binding:"required,min=3,max=64,alphanum"`
	Nickname  string `json:"nickname" binding:"required,max=64"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,url,max=512"`
	Password  string `json:"password" binding:"required,min=6,max=72"`
}
type UpdateProfile struct {
	Nickname string `json:"nickname" binding:"required,max=64"`
}
type ChangePassword struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6,max=72"`
	NewPassword     string `json:"new_password" binding:"required,min=6,max=72"`
}
type Refresh struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type Exchange struct {
	OptionID            uint64 `json:"option_id" binding:"required"`
	ExpectedPetalAmount uint64 `json:"expected_petal_amount" binding:"required"`
	ExpectedCoinCost    uint64 `json:"expected_coin_cost" binding:"required"`
	RequestID           string `json:"request_id" binding:"required,max=64"`
}
type Page struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (p *Page) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}
