package auth

import "time"

type Credential struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	BearerType   string    `json:"bearerType"`
	ExpiresAt    time.Time `json:"expiresAt"`
}
