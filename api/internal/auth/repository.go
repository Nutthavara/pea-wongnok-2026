package auth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	oauthPrefix = "oauth_state:"
)

type repository struct {
	cache *redis.Client
}

func NewRepository(rdb *redis.Client) *repository {
	return &repository{
		cache: rdb,
	}
}

func (repo *repository) SaveState(ctx context.Context, state string, ttl time.Duration) error {
	return repo.cache.Set(ctx, (oauthPrefix + state), "1", ttl).Err()
}

func (repo *repository) ConsumeState(ctx context.Context, state string) error {
	key := oauthPrefix + state

	exists, err := repo.cache.Exists(ctx, key).Result()
	if err != nil {
		return err
	}

	if exists == 0 {
		return ErrNotFound
	}

	return repo.cache.Del(ctx, key).Err()
}

func (repo *repository) SaveTicket(ctx context.Context, ticket string, credential Credential, ttl time.Duration) error {
	payload, err := json.Marshal(credential)
	if err != nil {
		return err
	}

	return repo.cache.Set(ctx, (oauthPrefix + ticket), payload, ttl).Err()
}

func (repo *repository) ConsumeTicket(ctx context.Context, ticket string) (Credential, error) {
	key := oauthPrefix + ticket

	payload, err := repo.cache.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Credential{}, ErrNotFound
		}

		return Credential{}, err
	}

	repo.cache.Del(ctx, key)

	var credential Credential
	if err := json.Unmarshal(payload, &credential); err != nil {
		return Credential{}, err
	}

	return credential, nil
}
