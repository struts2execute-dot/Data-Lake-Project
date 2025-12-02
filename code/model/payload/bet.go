package payload

type Bet struct {
	Payload
	UserID   *int32   `json:"user_id"`
	OrderID  *string  `json:"order_id"`
	Amount   *float64 `json:"amount"`
	Currency *string  `json:"currency"`
	GameID   *int32   `json:"game_id"`
}
