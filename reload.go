package main

import (
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type ReloadableHandler struct {
	mu      sync.RWMutex
	handler http.Handler
	config  *ReloadableConfig
}

type ReloadableConfig struct {
	ConfigFile string
	lastMod    time.Time
}

func NewReloadableHandler(configFile string) (*ReloadableHandler, error) {
	rh := &ReloadableHandler{
		config: &ReloadableConfig{
			ConfigFile: configFile,
		},
	}

	// Load initial configuration
	if err := rh.reload(); err != nil {
		return nil, err
	}

	// Watch for file changes
	go rh.watch()

	return rh, nil
}

func (rh *ReloadableHandler) reload() error {
	data, err := os.ReadFile(rh.config.ConfigFile)
	if err != nil {
		return err
	}

	// Check if file was modified
	info, err := os.Stat(rh.config.ConfigFile)
	if err != nil {
		return err
	}

	if !info.ModTime().After(rh.config.lastMod) {
		return nil // No changes
	}

	rh.config.lastMod = info.ModTime()

	// Parse YAML configuration
	var cfg sitesCfg
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	if len(cfg) == 0 {
		return nil
	}

	// Create new handler
	handler := NewHostDispatchingHandler()
	for i, site := range cfg {
		if err := site.validateWithHost(); err != nil {
			return err
		}
		handler.HandleHost(site.Host, createSiteHandler(site))
	}

	// Swap handler atomically
	rh.mu.Lock()
	rh.handler = handler
	rh.mu.Unlock()

	return nil
}

func (rh *ReloadableHandler) watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	if err := watcher.Add(rh.config.ConfigFile); err != nil {
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Small delay to avoid multiple reloads
				time.Sleep(100 * time.Millisecond)
				if err := rh.reload(); err != nil {
					// Log error but continue
					continue
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			_ = err
		}
	}
}

func (rh *ReloadableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rh.mu.RLock()
	handler := rh.handler
	rh.mu.RUnlock()

	if handler == nil {
		http.Error(w, "Configuration not loaded", http.StatusServiceUnavailable)
		return
	}

	handler.ServeHTTP(w, r)
}

