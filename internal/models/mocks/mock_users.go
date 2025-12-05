package mocks

import (
	"context"

	"github.com/cdriehuys/secret-santa/internal/models"
)

type UserModel struct {
	RegisteredUser models.NewUser
}

func (m *UserModel) Register(_ context.Context, user models.NewUser) error {
	m.RegisteredUser = user

	return nil
}
