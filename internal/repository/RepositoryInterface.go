package repository

import (
	"context"

	"github.com/vaim25ye/avito/internal/model"
)

type Repo interface {
	CreateUser(ctx context.Context, name, password string, balance int) (model.User, error)
	GetUserByID(ctx context.Context, userID int) (model.User, error)
	Transfer(ctx context.Context, fromUser, toUser, amount int) error
	PurchaseMerch(ctx context.Context, userID, merchID, count int) error
	LoadAllUserData(ctx context.Context) ([]model.UserInfo, error)
	// ... если нужны ещё методы
}
