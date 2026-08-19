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
	rdb *redis.Client
}

func NewRepository(rdb *redis.Client) *repository {
	return &repository{
		rdb: rdb,
	}
}

func (repo *repository) SaveState(ctx context.Context, state string, ttl time.Duration) error {
	return repo.rdb.Set(ctx, (oauthPrefix + state), "1", ttl).Err()
}

func (repo *repository) ConsumeState(ctx context.Context, state string) error {
	key := oauthPrefix + state

	exists, err := repo.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}

	if exists == 0 {
		return ErrNotFound
	}

	return repo.rdb.Del(ctx, key).Err()
}

func (repo *repository) SaveTicket(ctx context.Context, ticket string, credential Credential, ttl time.Duration) error {
	payload, err := json.Marshal(credential)
	if err != nil {
		return err
	}

	return repo.rdb.Set(ctx, (oauthPrefix + ticket), payload, ttl).Err()
}

func (repo *repository) ConsumeTicket(ctx context.Context, ticket string) (Credential, error) {
	key := oauthPrefix + ticket

	payload, err := repo.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Credential{}, ErrNotFound
		}

		return Credential{}, err
	}

	repo.rdb.Del(ctx, key)

	var credential Credential
	if err := json.Unmarshal(payload, &credential); err != nil {
		return Credential{}, err
	}

	return credential, nil
}
