package service

import (
	"context"
	"errors"
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

func newTestService(t *testing.T) (types.Service, *mocks.MockStorage) {
	t.Helper()
	ctrl := gomock.NewController(t)
	st := mocks.NewMockStorage(ctrl)
	return New(st), st
}

func TestWalletService_Create_Success(t *testing.T) {
	svc, st := newTestService(t)
	want := &domain.Wallet{}

	st.EXPECT().
		CreateWallet(gomock.Any(), storagetypes.CreateWalletInput{Currency: "RUB", Balance: 0}).
		Return(want, nil)

	got, err := svc.Create(context.Background(), types.CreateInput{Currency: "RUB"})

	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestWalletService_Get_NilID(t *testing.T) {
	svc, _ := newTestService(t)

	got, err := svc.Get(context.Background(), uuid.Nil)

	require.ErrorIs(t, err, types.ErrInvalidWalletID)
	assert.Nil(t, got)
}

func TestWalletService_Get_Success(t *testing.T) {
	svc, st := newTestService(t)
	id := uuid.New()
	want := &domain.Wallet{}

	st.EXPECT().GetWallet(gomock.Any(), id).Return(want, nil)

	got, err := svc.Get(context.Background(), id)

	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestWalletService_Get_Errors(t *testing.T) {
	tests := []struct {
		name       string
		storageErr error
		wantErr    error
		notErr     error
	}{
		{
			name:       "not found",
			storageErr: storagetypes.ErrNotFound,
			wantErr:    types.ErrWalletNotFound,
		},
		{
			name:       "wrapped not found",
			storageErr: errors.Join(errors.New("query failed"), storagetypes.ErrNotFound),
			wantErr:    types.ErrWalletNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, st := newTestService(t)
			id := uuid.New()

			st.EXPECT().GetWallet(gomock.Any(), id).Return(nil, tt.storageErr)

			got, err := svc.Get(context.Background(), id)

			require.ErrorIs(t, err, tt.wantErr)
			if tt.notErr != nil {
				assert.NotErrorIs(t, err, tt.notErr)
			}
			assert.Nil(t, got)
		})
	}
}
