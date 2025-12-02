package payload

type Recharge struct {
	Payload
	Status   *string  `json:"status"`
	UserID   *int32   `json:"user_id"`
	OrderID  *string  `json:"order_id"`
	Amount   *float64 `json:"amount"`
	Currency *string  `json:"currency"`
	Reason   *string  `json:"reason"`
}
