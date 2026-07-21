package di

import (
	"errors"

	"github.com/revora/revora/internal/config"
	"github.com/revora/revora/internal/crypto"
	"github.com/revora/revora/internal/github"
	"github.com/revora/revora/internal/patch"
	"github.com/revora/revora/internal/platform"
	"go.uber.org/zap"
)

type Container struct {
	Config           *config.Manager
	GitHubClient     github.Client
	Logger           *zap.Logger
	PlatformDetector platform.Detector
	KeyManager       crypto.KeyManager
	PatchEngine      patch.Engine
}

func NewContainer(logger *zap.Logger, cfg *config.Manager) (*Container, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	var ghClient github.Client = nil
	if cfg.GitHubToken != "" {
		var err error
		ghClient, err = github.NewRealClient(cfg.GitHubToken)
		if err != nil {
			return nil, err
		}
	}

	detector := platform.NewDetector()
	keyMgr := crypto.NewKeyManager(cfg.ProjectDir)

	patchPath := cfg.PatchEngine
	if patchPath == "" {
		patchPath = "revora-patch"
	}
	patchEngine := patch.NewLocalEngine(patchPath)

	return &Container{
		Config:           cfg,
		GitHubClient:     ghClient,
		Logger:           logger,
		PlatformDetector: detector,
		KeyManager:       keyMgr,
		PatchEngine:      patchEngine,
	}, nil
}
