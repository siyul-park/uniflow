package system

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/secret"
)

const (
	CodeCreateSecrets = "secrets.create"
	CodeReadSecrets   = "secrets.read"
	CodeUpdateSecrets = "secrets.update"
	CodeDeleteSecrets = "secrets.delete"
)

func CreateSecrets(s secret.Store) func(context.Context, []*secret.Secret) ([]*secret.Secret, error) {
	return func(ctx context.Context, secrets []*secret.Secret) ([]*secret.Secret, error) {
		if _, err := s.Store(ctx, secrets...); err != nil {
			return nil, err
		}
		return s.Load(ctx, secrets...)

	}
}

func ReadSecrets(s secret.Store) func(context.Context, []*secret.Secret) ([]*secret.Secret, error) {
	return func(ctx context.Context, secrets []*secret.Secret) ([]*secret.Secret, error) {
		return s.Load(ctx, secrets...)
	}
}

func UpdateSecrets(s secret.Store) func(context.Context, []*secret.Secret) ([]*secret.Secret, error) {
	return func(ctx context.Context, secrets []*secret.Secret) ([]*secret.Secret, error) {
		if _, err := s.Swap(ctx, secrets...); err != nil {
			return nil, err
		}
		return s.Load(ctx, secrets...)
	}
}

func DeleteSecrets(s secret.Store) func(context.Context, []*secret.Secret) ([]*secret.Secret, error) {
	return func(ctx context.Context, secrets []*secret.Secret) ([]*secret.Secret, error) {
		exists, err := s.Load(ctx, secrets...)
		if err != nil {
			return nil, err
		}
		if _, err := s.Delete(ctx, exists...); err != nil {
			return nil, err
		}
		return exists, nil
	}
}
