package auth

type ExchangeRequest struct {
	Ticket string `json:"ticket" binding:"required" example:"pQx7...base64url..."`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required" example:"eyJhbGci..."`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}
