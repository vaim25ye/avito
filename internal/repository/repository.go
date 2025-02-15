package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
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
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

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

func (r *Repository) GetUserByID(ctx context.Context, userID int) (model.User, error) {
	query := `SELECT user_id, name, password, balance
              FROM "user" WHERE user_id = $1`
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

	// Проверить баланс
	var balance int
	sel := `SELECT balance FROM "user" WHERE user_id=$1 FOR UPDATE`
	if err = tx.QueryRowContext(ctx, sel, fromUserID).Scan(&balance); err != nil {
		return err
	}
	if balance < amount {
		return fmt.Errorf("not enough funds: have %d, need %d", balance, amount)
	}

	// Списать
	updFrom := `UPDATE "user" SET balance=balance-$1 WHERE user_id=$2`
	_, err = tx.ExecContext(ctx, updFrom, amount, fromUserID)
	if err != nil {
		return err
	}

	// Зачислить
	updTo := `UPDATE "user" SET balance=balance+$1 WHERE user_id=$2`
	_, err = tx.ExecContext(ctx, updTo, amount, toUserID)
	if err != nil {
		return err
	}

	// Запись в operation
	insOp := `INSERT INTO operation(fromUser, toUser, amount) VALUES($1, $2, $3)`
	_, err = tx.ExecContext(ctx, insOp, fromUserID, toUserID, amount)
	if err != nil {
		return err
	}

	return nil
}

// PurchaseMerch - покупка мерча (транзакция)
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

	// Узнать price
	var price int
	qPrice := `SELECT price FROM merch WHERE merch_id=$1`
	if err = tx.QueryRowContext(ctx, qPrice, merchID).Scan(&price); err != nil {
		return err
	}
	total := price * count

	// Проверить баланс
	var balance int
	qBalance := `SELECT balance FROM "user" WHERE user_id=$1 FOR UPDATE`
	if err = tx.QueryRowContext(ctx, qBalance, userID).Scan(&balance); err != nil {
		return err
	}
	if balance < total {
		return fmt.Errorf("not enough funds: have %d, need %d", balance, total)
	}

	// Списать
	upd := `UPDATE "user" SET balance=balance-$1 WHERE user_id=$2`
	_, err = tx.ExecContext(ctx, upd, total, userID)
	if err != nil {
		return err
	}

	// Вставить запись в purchase
	ins := `INSERT INTO purchase(user_id, merch_id, amount) VALUES($1, $2, $3)`
	_, err = tx.ExecContext(ctx, ins, userID, merchID, count)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) LoadAllUserData(ctx context.Context) ([]model.UserInfo, error) {
	users, err := r.loadAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	ops, err := r.loadAllOperations(ctx)
	if err != nil {
		return nil, err
	}

	pur, err := r.loadAllPurchases(ctx)
	if err != nil {
		return nil, err
	}

	// Собираем map[userID] -> UserInfo
	m := make(map[int]model.UserInfo)
	for _, u := range users {
		m[u.UserID] = model.UserInfo{
			User:       u,
			Operations: []model.Operation{},
			Purchases:  []model.Purchase{},
		}
	}

	for _, o := range ops {
		// fromUser
		if ui, ok := m[o.FromUser]; ok {
			ui.Operations = append(ui.Operations, o)
			m[o.FromUser] = ui
		}
		if ui, ok := m[o.ToUser]; ok {
			ui.Operations = append(ui.Operations, o)
			m[o.ToUser] = ui
		}
	}

	// Привязываем покупки
	for _, p := range pur {
		if ui, ok := m[p.UserID]; ok {
			ui.Purchases = append(ui.Purchases, p)
			m[p.UserID] = ui
		}
	}

	// Собираем результат
	var result []model.UserInfo
	for _, info := range m {
		result = append(result, info)
	}

	return result, nil
}

func (r *Repository) loadAllUsers(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT user_id, name, password, balance FROM "user"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.User
	for rows.Next() {
		var u model.User
		if err = rows.Scan(&u.UserID, &u.Name, &u.Password, &u.Balance); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	return res, rows.Err()
}

func (r *Repository) loadAllOperations(ctx context.Context) ([]model.Operation, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT operation_id, fromUser, toUser, amount FROM operation`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []model.Operation
	for rows.Next() {
		var o model.Operation
		if err = rows.Scan(&o.OperationID, &o.FromUser, &o.ToUser, &o.Amount); err != nil {
			return nil, err
		}
		ops = append(ops, o)
	}
	return ops, rows.Err()
}

func (r *Repository) loadAllPurchases(ctx context.Context) ([]model.Purchase, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT purchase_id, user_id, merch_id, amount FROM purchase`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pur []model.Purchase
	for rows.Next() {
		var p model.Purchase
		if err = rows.Scan(&p.PurchaseID, &p.UserID, &p.MerchID, &p.Amount); err != nil {
			return nil, err
		}
		pur = append(pur, p)
	}
	return pur, rows.Err()
}
