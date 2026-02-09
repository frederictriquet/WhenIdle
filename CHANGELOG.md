# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2026-02-09

### Added
- **Mode User Idle** : Détection d'inactivité basée sur l'absence d'événements clavier/souris
  - Utilise `CGEventSourceSecondsSinceLastEventType` (macOS Quartz Event Services)
  - Aucune permission d'accessibilité requise
  - Nouvelle option de configuration : `idle_mode` (`"cpu"` ou `"user_idle"`)
- **Radio button dans la GUI** pour sélectionner le mode de détection (CPU Activity / User Activity)
- **Option `restart`** : Relance automatiquement la tâche après sa complétion
- **Réduction du bruit dans les logs** : Suppression des messages CPU répétitifs lorsque la tâche tourne
- Fichier `user_idle_darwin.go` : Wrapper CGo pour CoreGraphics
- Fichier `user_idle_other.go` : Stub pour plateformes non-macOS

### Changed
- Refactoring : `CPUMonitor` renommé en `Monitor` (plus générique)
- Architecture : Stratégie pluggable via `checkIdle func() (idle bool, detail string)`
- En mode `user_idle`, `checksNeeded = 1` car la durée est mesurée directement par l'API système
- Fenêtres de configuration et logs réutilisées (singleton pattern) au lieu de dupliquer
- 33 tests au lieu de 24 (ajout tests `IdleMode` et User Idle)
- Couverture de tests : ~43% (ajout code CGo non couvert)

### Fixed
- **Singleton windows** : Les fenêtres "Configure Task" et "View Logs" ne se dupliquent plus
- **PATH pour Launch Agent** : Résolution via le shell de l'utilisateur (`$SHELL`) au lieu de `/bin/bash`
  - Permet de trouver `npm`, `node` (nvm), et autres outils installés via homebrew/nvm

### Technical Details
- Nouveau type `IdleMode` avec validation stricte
- Build constraint `//go:build darwin` pour `user_idle_darwin.go`
- LDFLAGS : `-framework CoreGraphics` ajouté pour User Idle
- Mode CPU préserve le comportement existant (rétro-compatible)

## [1.1.0] - 2026-02-08

### Added
- **Interface graphique (GUI)** dans la barre de menu macOS
  - Icône dans le system tray avec 4 états visuels (gris/vert/bleu/orange)
  - Menu pour activer/désactiver le monitoring
  - Fenêtre de configuration interactive avec validation en temps réel
  - Visualisation des logs en temps réel (dernières 500 lignes)
- **LogBuffer** : Ring buffer thread-safe pour capturer les logs
- **SaveConfig()** : Sauvegarde de la configuration depuis la GUI
- Support multi-plateforme : stub `HideFromDock()` pour non-Darwin
- Bundle `.app` macOS avec `LSUIElement=true` (caché du Dock)
- Flag `--gui` pour lancer le mode graphique
- Icônes générées programmatiquement (pas de fichiers statiques)

### Changed
- Installation via `install.sh` crée maintenant un bundle `.app` en plus du CLI
- Binary passe de ~3.4 MB à ~32 MB (inclut Fyne framework + CGO)
- Launch Agent lance maintenant le mode GUI par défaut
- 24 tests au lieu de 13 (ajout tests LogBuffer et SaveConfig)
- Couverture de tests augmentée à ~65%

### Fixed
- LogBuffer : appel onChange synchrone au lieu d'une goroutine par write (évite goroutine storm)
- Validation de config : source unique via `Validate()` au lieu de duplication dans la GUI
- Noms de resources distincts pour les icônes Fyne (évite confusion en cache)

### Technical Details
- Dépendance ajoutée : `fyne.io/fyne/v2 v2.7.2`
- CGO requis pour compilation (Objective-C pour Cocoa)
- Mode headless (CLI) toujours disponible avec `--config`

## [1.0.0] - 2026-02-07

### Added
- Initial release of WhenIdle daemon for macOS
- CPU monitoring with configurable threshold and idle duration
- Automatic task launching during idle periods
- Task suspension/resume with SIGSTOP/SIGCONT when CPU becomes active/idle
- Process group signaling for full process tree management
- Graceful shutdown with SIGTERM → SIGKILL timeout (5 seconds)
- JSON configuration file support
- Launch Agent macOS integration for auto-start at login
- Installation and uninstallation scripts
- Comprehensive test suite (13 unit tests, 60% coverage)
- Complete documentation (README, examples, FAQ)

### Technical Details
- Built with Go 1.18+
- Single dependency: gopsutil v4 for CPU monitoring
- Binary size: ~3.4 MB
- Memory footprint: ~5-10 MB at runtime
- Non-blocking CPU measurement (0 second interval)

[Unreleased]: https://github.com/user/whenidle/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/user/whenidle/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/user/whenidle/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/user/whenidle/releases/tag/v1.0.0
