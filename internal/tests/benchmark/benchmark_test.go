package benchmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
	"walletApp/internal/api/rest"
	"walletApp/internal/config"
	"walletApp/internal/domain"
	"walletApp/internal/service"
	"walletApp/internal/storage"
	"walletApp/internal/storage/psql"
	"walletApp/internal/tests"
	"walletApp/internal/tests/container"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	resttypes "walletApp/internal/api/rest/types"
)

const benchmarkPoolSize = 32

type testApp struct {
	pool   *pgxpool.Pool
	server *httptest.Server
	client *http.Client
}

func BenchmarkOperations(b *testing.B) {
	cfg := container.StartPostgres(b, benchmarkPoolSize)
	app := newApp(b, cfg)

	for _, clients := range []int{1, 10, 50, 100} {
		b.Run(fmt.Sprintf("clients=%d", clients), func(b *testing.B) {
			app.run(b, clients)
		})
	}
}

func newApp(b *testing.B, cfg config.Database) *testApp {
	b.Helper()

	pool, err := psql.InitDB(b.Context(), cfg)
	require.NoError(b, err)
	b.Cleanup(pool.Close)

	log := slog.New(tests.NewTestHandler(b))
	walletStorage := storage.New(pool)
	require.NoError(b, psql.RunMigrations(b.Context(), cfg))

	walletService := service.New(walletStorage)
	router := chi.NewRouter()
	rest.New(log, walletService).SetupRouter(router)

	server := httptest.NewServer(router)
	b.Cleanup(server.Close)

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     time.Minute,
	}
	b.Cleanup(transport.CloseIdleConnections)

	return &testApp{
		pool:   pool,
		server: server,
		client: &http.Client{Transport: transport, Timeout: 30 * time.Second},
	}
}

func (a *testApp) run(b *testing.B, clients int) {
	b.Helper()
	b.StopTimer()

	walletID := a.createWallet(b)
	assert.Equal(b, int64(0), a.getWalletBalance(b, walletID))

	wg := sync.WaitGroup{}
	startCh := make(chan struct{})

	for range clients {
		wg.Add(1)
		go func() {
			defer wg.Done()

			<-startCh
			for n := 0; n < b.N; n++ {
				operationType := domain.OperationTypeDeposit
				if n%2 != 0 {
					operationType = domain.OperationTypeWithdraw
				}

				a.applyOperation(b, walletID, operationType, 1)
			}
		}()
	}

	b.ResetTimer()
	b.StartTimer()
	close(startCh)
	wg.Wait()
	b.StopTimer()

	totalOps := float64(b.N * clients)
	b.ReportMetric(totalOps/b.Elapsed().Seconds(), "req/s")
}

func (a *testApp) createWallet(b *testing.B) string {
	b.Helper()

	var wallet resttypes.WalletResponse
	input := resttypes.CreateWalletRequest{Currency: "RUB"}
	a.doRequest(b, http.MethodPost, "/api/v1/wallet", input, &wallet, http.StatusCreated)

	return wallet.WalletID
}

func (a *testApp) getWalletBalance(b *testing.B, walletID string) int64 {
	b.Helper()

	var wallet resttypes.WalletResponse
	a.doRequest(b, http.MethodGet, "/api/v1/wallets/"+walletID, nil, &wallet, http.StatusOK)

	return wallet.Balance
}

func (a *testApp) applyOperation(b *testing.B, walletID string, operationType domain.OperationType, amount int64) {
	var operation resttypes.OperationResponse
	input := resttypes.CreateOperationRequest{
		WalletID:      uuid.MustParse(walletID),
		OperationType: string(operationType),
		Amount:        amount,
	}
	a.doRequest(b, http.MethodPost, "/api/v1/wallets", input, &operation, http.StatusOK)

	require.Equal(b, domain.OperationStatusApplied, domain.OperationStatus(operation.Status))
}

func (a *testApp) doRequest(b *testing.B, method, path string, input, output any, expectedStatus int) {
	var requestBody []byte
	if input != nil {
		var err error
		requestBody, err = json.Marshal(input)
		require.NoError(b, err)
	}

	request, err := http.NewRequestWithContext(b.Context(), method, a.server.URL+path, bytes.NewReader(requestBody))
	require.NoError(b, err)

	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Idempotency-Key", uuid.New().String())

	response, err := a.client.Do(request)
	require.NoError(b, err)

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(b, err)
	require.NoError(b, response.Body.Close())

	require.Equal(b, expectedStatus, response.StatusCode, "doRequest: unexpected HTTP method=%s, status=%d, path=%s, expected=%d, body=%s", method, response.StatusCode, path, expectedStatus, responseBody)
	require.NoError(b, json.Unmarshal(responseBody, output))
}
