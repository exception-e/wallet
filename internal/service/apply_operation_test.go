package service

import (
	"context"
	"testing"
	"walletApp/internal/domain"
	"walletApp/internal/service/mocks"
	"walletApp/internal/service/types"
	storagetypes "walletApp/internal/storage/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestApplyOperation_Rejected_InsufficientFunds(t *testing.T) {
	for _, tt := range []struct {
		name    string
		balance int64
		amount  int64
	}{
		{"balance is zero", 0, 1},
		{"one unit short", 1000, 1001},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svc, st := newTestService(t)
			in := newInput(domain.OperationTypeWithdraw, tt.amount)
			code := domain.RejectionCodeInsufficientFunds
			expected := createInputFor(in, domain.OperationStatusRejected)
			expected.RejectionCode = &code
			want := operationFor(in, domain.OperationStatusRejected)

			expectTx(st)
			st.EXPECT().LockWallet(gomock.Any(), in.WalletID).Return(&domain.Wallet{Balance: tt.balance}, nil)
			st.EXPECT().CreateOperation(gomock.Any(), expected).Return(want, nil)
			// UpdateWallet не ожидается: при отказе баланс менять нельзя

			got, err := svc.ApplyOperation(context.Background(), in)

			require.NoError(t, err)
			assert.Same(t, want, got)
		})
	}
}

func TestApplyOperation_Applied(t *testing.T) {
	tests := []struct {
		name        string
		opType      domain.OperationType
		balance     int64
		amount      int64
		wantBalance int64
		key         string
	}{
		{"deposit", domain.OperationTypeDeposit, 1000, 500, 1500, "key-1"},
		{"deposit to zero wallet", domain.OperationTypeDeposit, 0, 1, 1, "key-1"},
		{"withdraw", domain.OperationTypeWithdraw, 1000, 400, 600, "key-1"},
		{"withdraw all", domain.OperationTypeWithdraw, 1000, 1000, 0, "key-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, st := newTestService(t)
			in := newInput(tt.opType, tt.amount)
			in.IdempotencyKey = tt.key
			want := operationFor(in, domain.OperationStatusApplied)
			newBalance := tt.wantBalance

			gomock.InOrder(
				expectTx(st),
				st.EXPECT().LockWallet(gomock.Any(), in.WalletID).
					Return(&domain.Wallet{Balance: tt.balance}, nil),
				st.EXPECT().CreateOperation(gomock.Any(), createInputFor(in, domain.OperationStatusApplied)).
					Return(want, nil),
				st.EXPECT().UpdateWallet(gomock.Any(), storagetypes.UpdateInput{Id: in.WalletID, Balance: &newBalance}).
					Return(&domain.Wallet{Balance: newBalance}, nil),
			)

			got, err := svc.ApplyOperation(context.Background(), in)

			require.NoError(t, err)
			assert.Same(t, want, got)
		})
	}
}

func newInput(opType domain.OperationType, amount int64) types.ApplyOperationInput {
	return types.ApplyOperationInput{
		WalletID:       uuid.New(),
		IdempotencyKey: "key-1",
		Type:           opType,
		Amount:         amount,
	}
}
func expectTx(st *mocks.MockStorage) *gomock.Call {
	return st.EXPECT().
		RunInTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
}

func createInputFor(in types.ApplyOperationInput, status domain.OperationStatus) storagetypes.CreateOperationInput {
	return storagetypes.CreateOperationInput{
		WalletID:       in.WalletID,
		IdempotencyKey: in.IdempotencyKey,
		Type:           in.Type,
		Amount:         in.Amount,
		Status:         status,
	}
}

func operationFor(in types.ApplyOperationInput, status domain.OperationStatus) *domain.Operation {
	return &domain.Operation{Type: in.Type, Amount: in.Amount, Status: status}
}
