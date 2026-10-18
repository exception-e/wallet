package storage

import (
	"context"
	"errors"
	"fmt"
	"walletApp/internal/domain"
	"walletApp/internal/storage/psql"
	"walletApp/internal/storage/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ types.Storage = (*WalletStorage)(nil)

type WalletStorage struct {
	pool *pgxpool.Pool
}

const (
	walletsTableName    = "wallets"
	operationsTableName = "operations"
	walletColumns       = "id, balance, currency, created_at, updated_at"
	operationColumns    = "id, wallet_id, idempotency_key, type, amount, status, rejection_code, created_at"
)

func (w *WalletStorage) CreateWallet(ctx context.Context, input types.CreateWalletInput) (*domain.Wallet, error) {
	query, args, err := sq.Insert(walletsTableName).
		Columns("currency", "balance").
		Values(input.Currency, input.Balance).
		Suffix("RETURNING " + walletColumns).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("walletRepository.Create (build query): %w", err)
	}

	wallet, err := scanWallet(psql.QuerierFromContext(ctx, w.pool).QueryRow(ctx, query, args...))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "wallets_pkey" {
			err = fmt.Errorf("%w: %w", types.ErrAlreadyExists, err)
		}
		return nil, fmt.Errorf("creating wallet (currency=%s): %w", input.Currency, err)
	}

	return wallet, nil
}

func (w *WalletStorage) RunInTx(ctx context.Context, fn types.TxFunc) error {
	err := pgx.BeginTxFunc(ctx, w.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		return fn(psql.WithTx(ctx, tx))
	})
	if err != nil {
		return fmt.Errorf("baseRepository.RunInTx: %w", err)
	}

	return nil
}

func (w *WalletStorage) GetWallet(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	query, args, err := sq.Select(walletColumns).
		From(walletsTableName).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("walletRepository.Get (build query, walletID=%s): %w", id, err)
	}

	wallet, err := scanWallet(psql.QuerierFromContext(ctx, w.pool).QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("walletRepository.Get (walletID=%s): %w", id, types.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("walletRepository.Get (walletID=%s): %w", id, err)
	}

	return wallet, nil
}

func (w *WalletStorage) UpdateWallet(ctx context.Context, input types.UpdateInput) (*domain.Wallet, error) {
	builder := sq.Update(walletsTableName).
		Set("updated_at", sq.Expr("clock_timestamp()")).
		Where(sq.Eq{"id": input.Id}).
		Suffix("RETURNING " + walletColumns).
		PlaceholderFormat(sq.Dollar)
	if input.Balance != nil {
		builder = builder.Set("balance", *input.Balance)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("walletRepository.Update (build query, walletID=%s): %w", input.Id, err)
	}

	wallet, err := scanWallet(psql.QuerierFromContext(ctx, w.pool).QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("walletRepository.Update (walletID=%s): %w", input.Id, types.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("walletRepository.Update (walletID=%s): %w", input.Id, err)
	}

	return wallet, nil
}

func (w *WalletStorage) LockWallet(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	if psql.TxFromContext(ctx) == nil {
		return nil, fmt.Errorf("walletRepository.Lock (walletID=%s): %w", id, types.ErrTransactionRequired)
	}

	query, args, err := sq.Select(walletColumns).
		From(walletsTableName).
		Where(sq.Eq{"id": id}).
		Suffix("FOR UPDATE").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("walletRepository.Lock (build query, walletID=%s): %w", id, err)
	}

	wallet, err := scanWallet(psql.QuerierFromContext(ctx, w.pool).QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("walletRepository.Lock (walletID=%s): %w", id, types.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("walletRepository.Lock (walletID=%s): %w", id, err)
	}

	return wallet, nil
}

func (w *WalletStorage) GetOperationByIdempotencyKey(ctx context.Context, input types.GetOperationByIdempotencyKeyInput) (*domain.Operation, error) {
	query, args, err := sq.Select(operationColumns).
		From(operationsTableName).
		Where(sq.Eq{"wallet_id": input.WalletID, "idempotency_key": input.IdempotencyKey}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("operationRepository.GetByIdempotencyKey (build query, walletID=%s): %w", input.WalletID, err)
	}

	operation, err := scanOperation(psql.QuerierFromContext(ctx, w.pool).QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("operationRepository.GetByIdempotencyKey (walletID=%s, idempotencyKey=%q): %w", input.WalletID, input.IdempotencyKey, types.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("operationRepository.GetByIdempotencyKey (walletID=%s, idempotencyKey=%q): %w", input.WalletID, input.IdempotencyKey, err)
	}

	return operation, nil
}

func (w *WalletStorage) CreateOperation(ctx context.Context, input types.CreateOperationInput) (*domain.Operation, error) {
	query, args, err := sq.Insert(operationsTableName).
		Columns("wallet_id", "idempotency_key", "type", "amount", "status", "rejection_code").
		Values(input.WalletID, input.IdempotencyKey, input.Type, input.Amount, input.Status, input.RejectionCode).
		Suffix("ON CONFLICT ON CONSTRAINT wallet_operations_idempotency_unique DO NOTHING").
		Suffix("RETURNING " + operationColumns).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("operationRepository.Create (build query, walletID=%s): %w", input.WalletID, err)
	}

	operation, err := scanOperation(psql.QuerierFromContext(ctx, w.pool).QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("operationRepository.Create (walletID=%s, idempotencyKey=%q): %w", input.WalletID, input.IdempotencyKey, types.ErrAlreadyExists)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "wallet_operations_wallet_id_fkey" {
			err = fmt.Errorf("%w: %w", types.ErrNotFound, err)
		}
		return nil, fmt.Errorf("operationRepository.Create (walletID=%s, idempotencyKey=%q): %w", input.WalletID, input.IdempotencyKey, err)
	}

	return operation, nil
}

func New(pool *pgxpool.Pool) types.Storage {
	return &WalletStorage{
		pool: pool,
	}
}

func scanWallet(row pgx.Row) (*domain.Wallet, error) {
	var wallet domain.Wallet
	err := row.Scan(
		&wallet.ID,
		&wallet.Balance,
		&wallet.Currency,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("wallet.scanWallet: %w", err)
	}

	return &wallet, nil
}

func scanOperation(row pgx.Row) (*domain.Operation, error) {
	var operation domain.Operation
	err := row.Scan(
		&operation.ID,
		&operation.WalletID,
		&operation.IdempotencyKey,
		&operation.Type,
		&operation.Amount,
		&operation.Status,
		&operation.RejectionCode,
		&operation.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("operation.scanOperation: %w", err)
	}

	return &operation, nil
}
