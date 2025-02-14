package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vaim25ye/avito/internal/cache"
	"github.com/vaim25ye/avito/internal/model"
	"github.com/vaim25ye/avito/internal/repository"
)

type mockRepo struct {
	repository.Repository
	CreateUserFunc func(ctx context.Context, name, password string, balance int) (model.User, error)
}

func (m *mockRepo) CreateUser(ctx context.Context, name, password string, balance int) (model.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, name, password, balance)
	}
	return model.User{}, nil
}

func TestCreateUserHandler(t *testing.T) {
	mock := &mockRepo{
		CreateUserFunc: func(ctx context.Context, name, password string, balance int) (model.User, error) {
			return model.User{UserID: 123, Name: name, Password: password, Balance: balance}, nil
		},
	}
	h := NewHandler(mock, cache.NewCache())

	body := `{"name":"TestUser","password":"testpass","balance":1000}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.CreateUser(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var u model.User
	err := json.Unmarshal(rr.Body.Bytes(), &u)
	assert.NoError(t, err)
	assert.Equal(t, 123, u.UserID)
	assert.Equal(t, "TestUser", u.Name)
	assert.Equal(t, "testpass", u.Password)
	assert.Equal(t, 1000, u.Balance)
}
