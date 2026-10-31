package di

import (
	"context"
	"database/sql"
	"errors"

	"github.com/labstack/echo/v4"

	controller "github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/controller/http"
	gateway "github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/gateway/http"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/presenter"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository/postgres"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/config"
	echoserver "github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/echo"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/jwks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/jwt"
	infrapostgres "github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/postgres"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/rsakey"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/scheduler"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/sealer"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/store"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/activationnotification"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/authentication"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/entitlement"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/peersigningkey"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/signingkey"
)

const (
	JobEnsureSigningKeys              = "ensure_signing_keys"
	JobDeleteExpiredSigningKeys       = "delete_expired_signing_keys"
	JobRefreshPeerSigningKeys         = "refresh_peer_signing_keys"
	JobDeliverActivationNotifications = "deliver_activation_notifications"
	JobTrackSigningKeyStatuses        = "track_signing_key_statuses"
)

// Overrides replace dependencies of the container, for tests.
type Container struct {
	DB        *sql.DB
	Echo      *echo.Echo
	Scheduler *scheduler.Runner
	Signer    *signingkey.Signer
}

func New(cfg *config.Config, opts ...Option) (*Container, error) {
	db, err := infrapostgres.Open(cfg.DatabaseURL, cfg.DatabaseMaxConnections)
	if err != nil {
		return nil, err
	}
	container, err := NewWithDB(cfg, db, opts...)
	if err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return container, nil
}

func NewWithDB(cfg *config.Config, db *sql.DB, opts ...Option) (*Container, error) {
	o := newOptions(opts)
	logger := o.logger
	systemClock := o.clock
	httpClient := o.httpClient
	uuidGenerator := systemclock.NewUUIDGenerator()

	privateKeySealer, err := sealer.NewAESGCM(cfg.KeyEncryptionKey)
	if err != nil {
		return nil, err
	}
	keyPairGenerator, err := rsakey.NewGenerator(rsakey.WithBits(cfg.KeyRSABits))
	if err != nil {
		return nil, err
	}

	queries := store.New(db)
	signingKeyRepository := postgres.NewSigningKeyRepository(db, privateKeySealer, postgres.WithQuerier(queries))
	entitlementRepository := postgres.NewEntitlementRepository(db, postgres.WithQuerier(queries))
	activationNotificationRepository := postgres.NewActivationNotificationRepository(db, postgres.WithQuerier(queries))
	jobLock := postgres.NewJobLockRepository(db)
	databaseProbe := postgres.NewDatabaseProbeRepository(db, postgres.WithQuerier(queries))

	peerKeySet := jwks.NewRemoteKeySet(cfg.PeerJWKSURL,
		jwks.WithHTTPClient(httpClient),
		jwks.WithRefreshOnUnknownKeyID(cfg.PeerJWKSRefreshOnUnknownKeyID),
		jwks.WithRefreshMinimumInterval(cfg.PeerJWKSRefreshMinInterval),
		jwks.WithClock(systemClock),
		jwks.WithLogger(logger),
	)
	tokenIssuer := jwt.NewIssuer(signingKeyRepository)
	tokenVerifier := jwt.NewVerifier(peerKeySet, cfg.PeerKeyAlgorithm, jwt.WithLogger(logger))
	signer := signingkey.NewSigner(signingKeyRepository, tokenIssuer, cfg.TokenIssuer, cfg.TokenTTL, signingkey.WithClock(systemClock), signingkey.WithUUIDGenerator(uuidGenerator))
	activationNotificationSender := gateway.NewActivationNotificationClient(cfg.ActivationNotificationURL, gateway.WithHTTPClient(httpClient))

	ensureSigningKeysUC := signingkey.NewEnsureSigningKeysInteractor(
		signingKeyRepository,
		signingKeyRepository,
		keyPairGenerator,
		cfg.KeySchedule,
		cfg.KeyAlgorithm,
		signingkey.WithClock(systemClock),
		signingkey.WithUUIDGenerator(uuidGenerator),
		signingkey.WithLogger(logger),
	)
	deleteExpiredSigningKeysUC := signingkey.NewDeleteExpiredSigningKeysInteractor(signingKeyRepository, signingkey.WithClock(systemClock), signingkey.WithLogger(logger))
	trackSigningKeyStatusesUC := signingkey.NewTrackSigningKeyStatusesInteractor(signingKeyRepository, signingkey.WithClock(systemClock), signingkey.WithLogger(logger))
	listPublishedSigningKeysUC := signingkey.NewListPublishedSigningKeysInteractor(signingKeyRepository, signingkey.WithClock(systemClock))
	listSigningKeysUC := signingkey.NewListSigningKeysInteractor(signingKeyRepository, signingkey.WithClock(systemClock))
	rotateSigningKeyUC := signingkey.NewRotateSigningKeyInteractor(
		jobLock,
		signingKeyRepository,
		signingKeyRepository,
		keyPairGenerator,
		JobEnsureSigningKeys,
		cfg.KeySchedule,
		cfg.KeyAlgorithm,
		signingkey.WithClock(systemClock),
		signingkey.WithUUIDGenerator(uuidGenerator),
		signingkey.WithLogger(logger),
	)
	listPeerSigningKeysUC := peersigningkey.NewListPeerSigningKeysInteractor(peerKeySet)
	refreshPeerSigningKeysUC := peersigningkey.NewRefreshPeerSigningKeysInteractor(peerKeySet)
	authenticatePeerRequestUC := authentication.NewAuthenticatePeerRequestInteractor(tokenVerifier, cfg.PeerIssuer, cfg.TokenLeeway, authentication.WithClock(systemClock))
	upsertEntitlementUC := entitlement.NewUpsertEntitlementInteractor(entitlementRepository, cfg.PeerProvider, entitlement.WithClock(systemClock), entitlement.WithUUIDGenerator(uuidGenerator), entitlement.WithLogger(logger))
	getEntitlementUC := entitlement.NewGetEntitlementInteractor(entitlementRepository, cfg.PeerProvider)
	listEntitlementsUC := entitlement.NewListEntitlementsInteractor(entitlementRepository, cfg.PeerProvider)
	terminateEntitlementUC := entitlement.NewTerminateEntitlementInteractor(entitlementRepository, cfg.PeerProvider, entitlement.WithClock(systemClock), entitlement.WithUUIDGenerator(uuidGenerator), entitlement.WithLogger(logger))
	deliverActivationNotificationsUC := activationnotification.NewDeliverActivationNotificationsInteractor(
		activationNotificationRepository,
		activationNotificationRepository,
		activationNotificationSender,
		signer,
		cfg.ActivationNotificationAudience,
		cfg.ActivationNotificationBatchSize,
		cfg.ActivationNotificationMaximumAttempts,
		activationnotification.WithClock(systemClock),
		activationnotification.WithLogger(logger),
	)

	var routerOptions []controller.Option
	if cfg.EnableAPIDocs {
		routerOptions = append(routerOptions, controller.WithDocs(controller.NewDocsController(o.openAPISpec)))
	}
	router := controller.NewRouter(
		controller.NewHealthController(databaseProbe, presenter.NewHealthPresenter()),
		controller.NewSigningKeyController(listPublishedSigningKeysUC, listSigningKeysUC, rotateSigningKeyUC, presenter.NewSigningKeyPresenter(), cfg.PeerJWKSPollInterval),
		controller.NewPeerSigningKeyController(listPeerSigningKeysUC, presenter.NewPeerSigningKeyPresenter()),
		controller.NewEntitlementController(upsertEntitlementUC, getEntitlementUC, listEntitlementsUC, terminateEntitlementUC, presenter.NewEntitlementPresenter()),
		controller.NewPeerAuthMiddleware(authenticatePeerRequestUC, controller.WithLogger(logger)),
		controller.Audiences{EntitlementSet: cfg.EntitlementSetAudience, EntitlementGet: cfg.EntitlementGetAudience},
		routerOptions...,
	)
	server := echoserver.New(router, presenter.NewErrorPresenter(presenter.WithLogger(logger)),
		echoserver.WithRequestTimeout(cfg.RequestTimeout), echoserver.WithLogger(logger))

	runner := scheduler.NewRunner(jobLock, []scheduler.Job{
		{Name: JobEnsureSigningKeys, Interval: cfg.KeyRotationCheckInterval, Exclusive: true, Run: ensureSigningKeysUC.Execute},
		{Name: JobDeleteExpiredSigningKeys, Interval: cfg.KeyRotationCheckInterval, Exclusive: true, Run: deleteExpiredSigningKeysUC.Execute},
		{Name: JobRefreshPeerSigningKeys, Interval: cfg.PeerJWKSPollInterval, Run: refreshPeerSigningKeysUC.Execute},
		{Name: JobDeliverActivationNotifications, Interval: cfg.ActivationNotificationInterval, Exclusive: true, Run: deliverActivationNotificationsUC.Execute},
		{Name: JobTrackSigningKeyStatuses, Interval: cfg.KeyRotationCheckInterval, Run: trackSigningKeyStatusesUC.Execute},
	}, scheduler.WithLogger(logger))

	return &Container{DB: db, Echo: server, Scheduler: runner, Signer: signer}, nil
}

func (c *Container) StartBackground(ctx context.Context) {
	c.Scheduler.Start(ctx)
}

func (c *Container) Close() error {
	return c.DB.Close()
}
