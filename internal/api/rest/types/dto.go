package types

type CreateOperationRequest struct {
	OperationType string `json:"operationType" enums:"DEPOSIT,WITHDRAW"`
	Amount        int64  `json:"amount" example:"1000" minimum:"1"`
}

type WalletResponse struct {
	WalletID string `json:"walletId" format:"uuid"`
	Balance  int64  `json:"balance" example:"49000"`
	Currency string `json:"currency" example:"RUB"`
}

type OperationResponse struct {
	OperationID   string  `json:"operationId" format:"uuid"`
	WalletID      string  `json:"walletId" format:"uuid"`
	Status        string  `json:"status" enums:"APPLIED,REJECTED"`
	BalanceAfter  *int64  `json:"balanceAfter,omitempty" example:"49000"`
	RejectionCode *string `json:"rejectionCode,omitempty" enums:"INSUFFICIENT_FUNDS,BALANCE_OVERFLOW"`
}

type ErrorResponse struct {
	Code    string `json:"code" example:"WALLET_NOT_FOUND"`
	Message string `json:"message" example:"wallet not found"`
}
