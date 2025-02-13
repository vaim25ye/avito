package model

type User struct {
	UserID   int    `db:"user_id" json:"user_id"`
	Name     string `db:"name"    json:"name"`
	Password string `db:"password"`
	Balance  int    `db:"balance" json:"balance"`
}

type Merch struct {
	MerchID int    `db:"merch_id" json:"merch_id"`
	Type    string `db:"type"     json:"type"`
	Price   int    `db:"price"    json:"price"`
}

type Purchase struct {
	PurchaseID int `db:"purchase_id" json:"purchase_id"`
	UserID     int `db:"user_id"     json:"user_id"`
	MerchID    int `db:"merch_id"    json:"merch_id"`
	Amount     int `db:"amount"      json:"amount"`
}

type Operation struct {
	OperationID int `db:"operation_id" json:"operation_id"`
	FromUser    int `db:"fromUser"     json:"fromUser"`
	ToUser      int `db:"toUser"       json:"toUser"`
	Amount      int `db:"amount"       json:"amount"`
}
