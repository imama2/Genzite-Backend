package dto

type CreateTransactionResult struct {
	OrderID     string
	RedirectURL string
}

type CreateTransactionResponse struct {
	OrderID     string `json:"order_id"`
	RedirectURL string `json:"redirect_url"`
}
