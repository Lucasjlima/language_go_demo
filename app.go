package main

import (
	"context"
	"log"

	"language_go_demo/internal/ai"
	"language_go_demo/internal/config"
)

// App struct
type App struct {
	ctx        context.Context
	explainer  *ai.Service
	startupErr error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := config.Load()
	if err != nil {
		a.startupErr = err
		log.Printf("startup config error: %v", err)
		return
	}

	a.explainer, err = ai.NewService(cfg)
	if err != nil {
		a.startupErr = err
		log.Printf("startup ai error: %v", err)
		return
	}

	log.Printf("startup complete: Gemini service configured with model %q", cfg.GeminiModel)
}

func (a *App) ExplainText(text string) (ai.ExplanationResponse, error) {
	log.Printf("binding ExplainText received text length=%d", len(text))

	if a.startupErr != nil {
		return ai.ExplanationResponse{}, a.startupErr
	}

	if a.explainer == nil {
		return ai.ExplanationResponse{}, ai.ErrServiceUnavailable
	}

	resp, err := a.explainer.Explain(ai.ExplanationRequest{Text: text})
	if err != nil {
		log.Printf("binding ExplainText failed: %v", err)
		return ai.ExplanationResponse{}, err
	}

	log.Printf("binding ExplainText succeeded")
	return resp, nil
}
