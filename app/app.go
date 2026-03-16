package app

import (
	"context"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"

	"github.com/alecsavvy/mojave/config"
	"github.com/alecsavvy/mojave/gen/mojave/v1/v1connect"
	"github.com/alecsavvy/mojave/store"
	"github.com/cockroachdb/pebble"
	cfg "github.com/cometbft/cometbft/config"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/rpc/client/local"
	"github.com/anacrolix/torrent"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
)

type App struct {
	logger *zap.SugaredLogger
	node   *nm.Node
	rpc    *local.Local

	store         *store.Store
	localStore    *store.LocalStore
	onGoingBlock  *pebble.Batch
	connectServer *http.Server

	tmpBucket     *blob.Bucket
	storageBucket *blob.Bucket

	torrentClient *torrent.Client

	validatorPrivKey ed25519.PrivateKey
	validatorPubKey  ed25519.PublicKey
	encryptionKey    *rsa.PrivateKey
}

// NewApp starts a node from an already-initialized config. The caller must have
// written config.toml, genesis.json, priv validator key/state, and node key to the config's RootDir.
func NewApp(mojaveCfg *config.MojaveConfig) (*App, error) {
	cmtConfig := mojaveCfg.CometConfig

	z, _ := zap.NewDevelopment()
	logger := z.Sugar()

	if err := cmtConfig.ValidateBasic(); err != nil {
		return nil, err
	}

	pv := privval.LoadFilePV(cmtConfig.PrivValidatorKeyFile(), cmtConfig.PrivValidatorStateFile())
	nodeKey, err := p2p.LoadNodeKey(cmtConfig.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	dbPath := path.Join(cmtConfig.RootDir, "pebble")
	db, err := pebble.Open(dbPath, nil)
	if err != nil {
		return nil, err
	}

	localDBPath := path.Join(cmtConfig.RootDir, "local-pebble")
	localDB, err := pebble.Open(localDBPath, nil)
	if err != nil {
		return nil, err
	}

	tmpBucket, err := blob.OpenBucket(context.Background(), mojaveCfg.FilesTmpDir)
	if err != nil {
		return nil, err
	}

	storageBucket, err := blob.OpenBucket(context.Background(), mojaveCfg.FilesStorageDir)
	if err != nil {
		return nil, err
	}

	cmtLogger := cmtlog.NewNopLogger()

	addr := pv.GetAddress().String()
	logger = logger.With("addr", addr)

	appStore := store.NewStore(db)
	localStore := store.NewLocalStore(localDB)

	// Extract validator signing key from FilePV
	cmtPrivKey, ok := pv.Key.PrivKey.(cmted25519.PrivKey)
	if !ok {
		return nil, fmt.Errorf("validator key is not ed25519")
	}
	validatorPrivKey := ed25519.PrivateKey(cmtPrivKey)
	validatorPubKey := validatorPrivKey.Public().(ed25519.PublicKey)

	// Derive RSA encryption key deterministically from validator seed (for TDF / KAS).
	// RSA.GenerateKey needs many random bytes; use a deterministic stream from the seed.
	seed := validatorPrivKey.Seed()
	encryptionKey, err := rsa.GenerateKey(&deterministicReader{seed: seed, salt: []byte("mojave-encryption-key")}, 2048)
	if err != nil {
		return nil, fmt.Errorf("derive encryption key: %w", err)
	}

	if err := localStore.SetSigningKey(validatorPubKey, validatorPrivKey); err != nil {
		return nil, fmt.Errorf("persist signing key: %w", err)
	}

	torrentDataDir := path.Join(cmtConfig.RootDir, "upload-tmp")
	torrentCfg := torrent.NewDefaultClientConfig()
	torrentCfg.DataDir = torrentDataDir
	torrentCfg.NoDHT = true

	torrentClient, err := torrent.NewClient(torrentCfg)
	if err != nil {
		return nil, fmt.Errorf("start torrent client: %w", err)
	}

	app := &App{
		logger:           logger,
		store:            appStore,
		localStore:       localStore,
		onGoingBlock:     nil,
		tmpBucket:        tmpBucket,
		storageBucket:    storageBucket,
		torrentClient:    torrentClient,
		validatorPrivKey: validatorPrivKey,
		validatorPubKey:  validatorPubKey,
		encryptionKey:    encryptionKey,
	}

	node, err := nm.NewNode(
		context.Background(),
		cmtConfig,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(cmtConfig),
		cfg.DefaultDBProvider,
		nm.DefaultMetricsProvider(cmtConfig.Instrumentation),
		cmtLogger,
	)
	if err != nil {
		return nil, err
	}

	app.node = node
	app.rpc = local.New(app.node)

	mux := http.NewServeMux()
	svcPath, svcHandler := v1connect.NewServiceHandler(app)
	mux.Handle(svcPath, svcHandler)
	app.connectServer = &http.Server{
		Addr:    mojaveCfg.ConnectRPCAddr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	return app, nil
}

func (a *App) Start() error {
	if a.node.IsRunning() {
		return nil
	}
	if err := a.node.Start(); err != nil {
		return err
	}

	ln, err := net.Listen("tcp", a.connectServer.Addr)
	if err != nil {
		return err
	}
	a.logger.Infow("ConnectRPC server listening", "addr", a.connectServer.Addr)
	go func() {
		if err := a.connectServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			a.logger.Errorw("ConnectRPC server error", "err", err)
		}
	}()

	return nil
}

// configureTorrentPeersFromValidators configures torrent peers based on the current CometBFT validator set.
func (a *App) configureTorrentPeersFromValidators(ctx context.Context) error {
	// TODO: Query the validator set from the node and map it to torrent peers.
	return nil
}

func (a *App) LatestBlockHeight(ctx context.Context) (int64, error) {
	status, err := a.rpc.Status(ctx)
	if err != nil {
		return 0, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

func (a *App) Stop() error {
	if err := a.connectServer.Shutdown(context.Background()); err != nil {
		a.logger.Warnw("ConnectRPC server shutdown error", "err", err)
	}

	if a.torrentClient != nil {
		a.torrentClient.Close()
	}

	if err := a.tmpBucket.Close(); err != nil {
		a.logger.Warnw("Tmp bucket close error", "err", err)
	}
	if err := a.storageBucket.Close(); err != nil {
		a.logger.Warnw("Storage bucket close error", "err", err)
	}

	return a.node.Stop()
}

// deterministicReader produces a deterministic stream of bytes from a seed for RSA key generation.
type deterministicReader struct {
	seed   []byte
	salt   []byte
	counter uint64
	buf    []byte
}

func (r *deterministicReader) Read(p []byte) (n int, err error) {
	for len(p) > 0 {
		if len(r.buf) == 0 {
			h := sha256.New()
			h.Write(r.seed)
			h.Write(r.salt)
			_ = binary.Write(h, binary.LittleEndian, r.counter)
			r.counter++
			r.buf = h.Sum(nil)
		}
		copied := copy(p, r.buf)
		r.buf = r.buf[copied:]
		p = p[copied:]
		n += copied
	}
	return n, nil
}

var _ io.Reader = (*deterministicReader)(nil)
