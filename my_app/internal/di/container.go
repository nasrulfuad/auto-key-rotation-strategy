package di

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	controller "github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/controller/http"
	gateway "github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/gateway/http"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/presenter"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/repository/postgres"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/config"
	echoserver "github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/echo"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwks"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwt"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
	infrapostgres "github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/postgres"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/rsakey"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/scheduler"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/sealer"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/systemclock"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/store"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/activationnotification"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/authentication"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/peersigningkey"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/signingkey"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/subscription"
)

const (
	JobEnsureSigningKeys        = "ensure_signing_keys"
	JobDeleteExpiredSigningKeys = "delete_expired_signing_keys"
	JobRefreshPeerSigningKeys   = "refresh_peer_signing_keys"
	JobTrackSigningKeyStatuses  = "track_signing_key_statuses"
)

type Overrides struct {
	Clock       systemclock.Clock
	HTTPClient  *http.Client
	OpenAPISpec []byte
}

type Container struct {
	DB        *sql.DB
	Echo      *echo.Echo
	Scheduler *scheduler.Runner
	Signer    *signingkey.Signer
}

func New(cfg *config.Config, logger logging.Logger, openAPISpec []byte) (*Container, error) {
	db, err := infrapostgres.Open(cfg.DatabaseURL, cfg.DatabaseMaxConnections)
	if err != nil {
		return nil, err
	}
	container, err := NewWithDB(cfg, db, logger, Overrides{OpenAPISpec: openAPISpec})
	if err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return container, nil
}

func NewWithDB(cfg *config.Config, db *sql.DB, logger logging.Logger, overrides Overrides) (*Container, error) {
	if logger == nil {
		logger = logging.NopLogger{}
	}
	systemClock := overrides.Clock
	if systemClock == nil {
		systemClock = systemclock.New()
	}
	httpClient := overrides.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	uuidGenerator := systemclock.NewUUIDGenerator()

	privateKeySealer, err := sealer.NewAESGCM(cfg.KeyEncryptionKey)
	if err != nil {
		return nil, err
	}
	keyPairGenerator, err := rsakey.NewGenerator(cfg.KeyRSABits)
	if err != nil {
		return nil, err
	}

	queries := store.New(db)
	signingKeyRepository := postgres.NewSigningKeyRepository(db, queries, privateKeySealer)
	activationNotificationRepository := postgres.NewActivationNotificationRepository(queries)
	jobLock := postgres.NewJobLockRepository(db)
	databaseProbe := postgres.NewDatabaseProbeRepository(queries)

	peerKeySet := jwks.NewRemoteKeySet(jwks.Options{
		URL:                    cfg.PeerJWKSURL,
		Client:                 httpClient,
		RefreshOnUnknownKeyID:  cfg.PeerJWKSRefreshOnUnknownKeyID,
		RefreshMinimumInterval: cfg.PeerJWKSRefreshMinInterval,
	}, systemClock, logger)
	tokenIssuer := jwt.NewIssuer(signingKeyRepository)
	tokenVerifier := jwt.NewVerifier(peerKeySet, cfg.PeerKeyAlgorithm, logger)
	signer := signingkey.NewSigner(signingKeyRepository, tokenIssuer, systemClock, uuidGenerator, cfg.TokenIssuer, cfg.TokenTTL)
	partnerEntitlementGateway := gateway.NewPartnerEntitlementClient(cfg.PartnerBaseURL, httpClient)

	ensureSigningKeysUC := signingkey.NewEnsureSigningKeysInteractor(signingKeyRepository, signingKeyRepository, keyPairGenerator,
		systemClock, uuidGenerator, logger, cfg.KeySchedule, cfg.KeyAlgorithm)
	deleteExpiredSigningKeysUC := signingkey.NewDeleteExpiredSigningKeysInteractor(signingKeyRepository, systemClock, logger)
	trackSigningKeyStatusesUC := signingkey.NewTrackSigningKeyStatusesInteractor(signingKeyRepository, systemClock, logger)
	listPublishedSigningKeysUC := signingkey.NewListPublishedSigningKeysInteractor(signingKeyRepository, systemClock)
	listSigningKeysUC := signingkey.NewListSigningKeysInteractor(signingKeyRepository, systemClock)
	rotateSigningKeyUC := signingkey.NewRotateSigningKeyInteractor(jobLock, signingKeyRepository, signingKeyRepository, keyPairGenerator,
		systemClock, uuidGenerator, logger, JobEnsureSigningKeys, cfg.KeySchedule, cfg.KeyAlgorithm)
	listPeerSigningKeysUC := peersigningkey.NewListPeerSigningKeysInteractor(peerKeySet)
	refreshPeerSigningKeysUC := peersigningkey.NewRefreshPeerSigningKeysInteractor(peerKeySet)
	authenticatePeerRequestUC := authentication.NewAuthenticatePeerRequestInteractor(tokenVerifier, systemClock, cfg.PeerIssuer, cfg.TokenLeeway)
	subscriptionRepository := postgres.NewSubscriptionRepository(queries)
	activateSubscriptionUC := subscription.NewActivateSubscriptionInteractor(partnerEntitlementGateway, subscriptionRepository, signer,
		systemClock, uuidGenerator, logger, cfg.EntitlementSetAudience, cfg.Provider)
	terminateSubscriptionUC := subscription.NewTerminateSubscriptionInteractor(partnerEntitlementGateway, subscriptionRepository, signer,
		systemClock, uuidGenerator, logger, cfg.EntitlementSetAudience, cfg.Provider)
	listSubscriptionsUC := subscription.NewListSubscriptionsInteractor(subscriptionRepository)
	receiveActivationNotificationUC := activationnotification.NewReceiveActivationNotificationInteractor(
		activationNotificationRepository, systemClock, uuidGenerator, logger)
	listActivationNotificationsUC := activationnotification.NewListActivationNotificationsInteractor(activationNotificationRepository)

	var docsController *controller.DocsController
	if cfg.EnableAPIDocs {
		docsController = controller.NewDocsController(overrides.OpenAPISpec)
	}
	router := controller.NewRouter(
		controller.NewHealthController(databaseProbe, presenter.NewHealthPresenter()),
		controller.NewSigningKeyController(listPublishedSigningKeysUC, listSigningKeysUC, rotateSigningKeyUC, presenter.NewSigningKeyPresenter(), cfg.PeerJWKSPollInterval),
		controller.NewPeerSigningKeyController(listPeerSigningKeysUC, presenter.NewPeerSigningKeyPresenter()),
		controller.NewSubscriptionController(activateSubscriptionUC, terminateSubscriptionUC, listSubscriptionsUC,
			presenter.NewSubscriptionPresenter()),
		controller.NewActivationNotificationController(receiveActivationNotificationUC, listActivationNotificationsUC,
			presenter.NewActivationNotificationPresenter()),
		controller.NewPeerAuthMiddleware(authenticatePeerRequestUC, logger),
		docsController,
		controller.Audiences{ActivationNotification: cfg.ActivationNotificationAudience},
	)
	server := echoserver.New(echoserver.Options{RequestTimeout: cfg.RequestTimeout, Logger: logger}, router, presenter.NewErrorPresenter(logger))

	runner := scheduler.NewRunner(jobLock, logger,
		scheduler.Job{Name: JobEnsureSigningKeys, Interval: cfg.KeyRotationCheckInterval, Exclusive: true, Run: ensureSigningKeysUC.Execute},
		scheduler.Job{Name: JobDeleteExpiredSigningKeys, Interval: cfg.KeyRotationCheckInterval, Exclusive: true, Run: deleteExpiredSigningKeysUC.Execute},
		scheduler.Job{Name: JobRefreshPeerSigningKeys, Interval: cfg.PeerJWKSPollInterval, Run: refreshPeerSigningKeysUC.Execute},
		scheduler.Job{Name: JobTrackSigningKeyStatuses, Interval: cfg.KeyRotationCheckInterval, Run: trackSigningKeyStatusesUC.Execute},
	)

	return &Container{DB: db, Echo: server, Scheduler: runner, Signer: signer}, nil
}

func (c *Container) StartBackground(ctx context.Context) {
	c.Scheduler.Start(ctx)
}

func (c *Container) Close() error {
	return c.DB.Close()
}
