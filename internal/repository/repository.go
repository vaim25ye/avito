package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/vaim25ye/avito/internal/model"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// Настройки пула соединений
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

// CreateUser - создание нового пользователя
func (r *Repository) CreateUser(ctx context.Context, name, password string, balance int) (model.User, error) {
	query := `INSERT INTO "user"(name, password, balance)
              VALUES ($1, $2, $3)
              RETURNING user_id`
	var id int
	err := r.db.QueryRowContext(ctx, query, name, password, balance).Scan(&id)
	if err != nil {
		return model.User{}, err
	}
	return model.User{
		UserID:   id,
		Name:     name,
		Password: password,
		Balance:  balance,
	}, nil
}

// GetUserByID - получение пользователя по ID
func (r *Repository) GetUserByID(ctx context.Context, userID int) (model.User, error) {
	query := `SELECT user_id, name, password, balance
              FROM "user"
              WHERE user_id = $1`
	var u model.User
	err := r.db.QueryRowContext(ctx, query, userID).
		Scan(&u.UserID, &u.Name, &u.Password, &u.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, fmt.Errorf("user not found")
		}
		return model.User{}, err
	}
	return u, nil
}

// Transfer - перевод денег между пользователями (с транзакцией)
func (r *Repository) Transfer(ctx context.Context, fromUserID, toUserID, amount int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 1. Проверяем баланс fromUser
	var balance int
	sel := `SELECT balance FROM "user" WHERE user_id=$1 FOR UPDATE`
	if err = tx.QueryRowContext(ctx, sel, fromUserID).Scan(&balance); err != nil {
		return err
	}
	if balance < amount {
		return fmt.Errorf("not enough funds: have %d, need %d", balance, amount)
	}

	// 2. Списываем
	updFrom := `UPDATE "user" SET balance=balance-$1 WHERE user_id=$2`
	_, err = tx.ExecContext(ctx, updFrom, amount, fromUserID)
	if err != nil {
		return err
	}

	// 3. Зачисляем
	updTo := `UPDATE "user" SET balance=balance+$1 WHERE user_id=$2`
	_, err = tx.ExecContext(ctx, updTo, amount, toUserID)
	if err != nil {
		return err
	}

	// 4. Пишем в operation
	insOp := `INSERT INTO operation(fromUser, toUser, amount) VALUES($1, $2, $3)`
	_, err = tx.ExecContext(ctx, insOp, fromUserID, toUserID, amount)
	if err != nil {
		return err
	}

	return nil
}

// PurchaseMerch - покупка мерча (списываем баланс + создаём запись в purchase)
func (r *Repository) PurchaseMerch(ctx context.Context, userID, merchID, count int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 1. Получить price мерча
	var price int
	qPrice := `SELECT price FROM merch WHERE merch_id=$1`
	if err = tx.QueryRowContext(ctx, qPrice, merchID).Scan(&price); err != nil {
		return err
	}
	total := price * count

	// 2. Проверить баланс
	var balance int
	qBalance := `SELECT balance FROM "user" WHERE user_id=$1 FOR UPDATE`
	if err = tx.QueryRowContext(ctx, qBalance, userID).Scan(&balance); err != nil {
		return err
	}
	if balance < total {
		return fmt.Errorf("not enough funds: have %d, need %d", balance, total)
	}

	// 3. Списать
	upd := `UPDATE "user" SET balance=balance-$1 WHERE user_id=$2`
	_, err = tx.ExecContext(ctx, upd, total, userID)
	if err != nil {
		return err
	}

	// 4. Добавить запись в purchase
	ins := `INSERT INTO purchase(user_id, merch_id, amount) VALUES($1, $2, $3)`
	_, err = tx.ExecContext(ctx, ins, userID, merchID, count)
	if err != nil {
		return err
	}

	return nil
}
