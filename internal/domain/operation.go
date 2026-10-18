package domain

import (
	"time"

	"github.com/google/uuid"
)

type OperationType string

const (
	OperationTypeDeposit  OperationType = "DEPOSIT"
	OperationTypeWithdraw OperationType = "WITHDRAW"
)

type OperationStatus string

const (
	OperationStatusApplied  OperationStatus = "APPLIED"
	OperationStatusRejected OperationStatus = "REJECTED"
)

type RejectionCode string

const (
	RejectionCodeInsufficientFunds RejectionCode = "INSUFFICIENT_FUNDS"
	RejectionCodeBalanceOverflow   RejectionCode = "BALANCE_OVERFLOW"
)

type Operation struct {
	ID             uuid.UUID
	WalletID       uuid.UUID
	IdempotencyKey string
	Type           OperationType
	Balance        int64
	Status         OperationStatus
	BalanceAfter   *int64
	RejectionCode  *RejectionCode
	CreatedAt      time.Time
}
