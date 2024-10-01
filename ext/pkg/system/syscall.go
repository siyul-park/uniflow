package system

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
)

const (
	CodeCreateNodes = "nodes.create"
	CodeReadNodes   = "nodes.read"
	CodeUpdateNodes = "nodes.update"
	CodeDeleteNodes = "nodes.delete"

	CodeCreateSecrets = "secrets.create"
	CodeReadSecrets   = "secrets.read"
	CodeUpdateSecrets = "secrets.update"
	CodeDeleteSecrets = "secrets.delete"
)

func CreateNodes(s spec.Store) func(context.Context, []spec.Spec) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []spec.Spec) ([]spec.Spec, error) {
		if _, err := s.Store(ctx, specs...); err != nil {
			return nil, err
		}
		return s.Load(ctx, specs...)

	}
}

func ReadNodes(s spec.Store) func(context.Context, []spec.Spec) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []spec.Spec) ([]spec.Spec, error) {
		return s.Load(ctx, specs...)
	}
}

func UpdateNodes(s spec.Store) func(context.Context, []spec.Spec) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []spec.Spec) ([]spec.Spec, error) {
		if _, err := s.Swap(ctx, specs...); err != nil {
			return nil, err
		}
		return s.Load(ctx, specs...)
	}
}

func DeleteNodes(s spec.Store) func(context.Context, []spec.Spec) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []spec.Spec) ([]spec.Spec, error) {
		ok, err := s.Load(ctx, specs...)
		if err != nil {
			return nil, err
		}
		if _, err := s.Delete(ctx, ok...); err != nil {
			return nil, err
		}
		return ok, nil
	}
}

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
		ok, err := s.Load(ctx, secrets...)
		if err != nil {
			return nil, err
		}
		if _, err := s.Delete(ctx, ok...); err != nil {
			return nil, err
		}
		return ok, nil
	}
}
