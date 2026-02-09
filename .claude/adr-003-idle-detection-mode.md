# ADR-003 : Support de Deux Modes de Détection d'Inactivité

## Statut
Accepté

## Date
2026-02-09

## Contexte
WhenIdle 1.1 supportait uniquement la détection d'inactivité basée sur le CPU (CPU < threshold).

**Problème** : Cette approche ne convient pas pour tous les cas d'usage :
- Les tâches d'analyse/backup doivent se lancer **quand l'utilisateur est absent**, pas seulement quand le CPU est bas
- Un ordinateur peut avoir un CPU idle mais l'utilisateur est actif (navigation web, lecture)
- Inversement, un CPU peut être chargé (compilation en arrière-plan) mais l'utilisateur est inactif

**Besoin** : Ajouter un mode "User Idle" qui détecte l'absence d'événements clavier/souris.

## Décision
Implémenter **deux modes de détection** via une stratégie pluggable :
1. **Mode CPU** (existant) : CPU usage < threshold
2. **Mode User Idle** (nouveau) : Aucun événement clavier/souris pendant N secondes

## Options Considérées

### Option A : CGEventSourceSecondsSinceLastEventType (CGo, CoreGraphics)
- **Score** : 118/150
- **Avantages** :
  - API native macOS Quartz Event Services
  - Aucune permission d'accessibilité requise
  - Simple appel CGo, pas de dépendance externe
  - Retourne directement les secondes depuis dernier événement
  - Polling-compatible (s'intègre dans notre architecture existante)
- **Inconvénients** :
  - Nécessite CGo (déjà requis pour Cocoa/Fyne)
  - macOS uniquement (stub no-op pour autres OS)
- **Choix** : ✅ **RETENU**

### Option B : gohook (event-driven)
- **Score** : 52/150
- **Avantages** :
  - Pure Go (pas de CGo)
  - Cross-platform
- **Inconvénients** :
  - Event-driven (incompatible avec notre architecture polling)
  - Dépendance externe avec peu de maintenance
  - Nécessite refactoring significatif du monitor
  - Peut nécessiter permissions accessibilité
- **Choix** : ❌ Rejeté

### Option C : robotgo
- **Score** : 40/150
- **Avantages** :
  - Framework complet (input simulation, etc.)
- **Inconvénients** :
  - Dépendance lourde (~50 MB)
  - Overkill pour notre besoin (on ne fait pas de simulation)
  - Mauvaise réputation (nombreux issues GitHub)
  - Nécessite permissions accessibilité
- **Choix** : ❌ Rejeté

## Justification
**Option A** est la meilleure solution car :
1. **Intégration naturelle** : S'intègre dans notre boucle de polling existante
2. **Zéro permission** : Pas besoin de demander l'accès accessibilité à l'utilisateur
3. **Minimale** : Un seul fichier CGo (`user_idle_darwin.go`), ~20 lignes
4. **Performance** : Appel système direct, pas de overhead
5. **Déjà CGo** : Le projet utilise déjà CGo pour Fyne et Cocoa

## Architecture Technique

### Refactoring `CPUMonitor` → `Monitor`
```go
type Monitor struct {
    checkIdle func() (idle bool, detail string) // Stratégie pluggable
}

func NewMonitor(config Config, ...) *Monitor {
    switch config.IdleMode {
    case IdleModeUserIdle:
        m.checkIdle = m.checkUserIdle
    default:
        m.checkIdle = m.checkCPUIdle
    }
    return m
}
```

**Trade-off** : Function injection au lieu d'interface formelle.
- ✅ Simplicité, testabilité (mock trivial)
- ❌ Moins idiomatique Go (mais acceptable pour stratégie interne)

### Configuration
```go
type IdleMode string

const (
    IdleModeCPU      IdleMode = "cpu"        // default
    IdleModeUserIdle IdleMode = "user_idle"  // nouveau
)
```

**Validation stricte** : Le mode est validé dans `Config.Validate()` pour éviter les valeurs invalides.

### GUI
Radio button Fyne pour sélectionner le mode :
```go
idleModeRadio := widget.NewRadioGroup(
    []string{"CPU Activity", "User Activity"},
    nil
)
```

## Conséquences

### Positives
- ✅ Cas d'usage étendus (backup nocturne, tâches pendant absence utilisateur)
- ✅ Rétro-compatible : Mode CPU reste le défaut
- ✅ Testable : Stratégie injectable = mocks faciles
- ✅ Extensible : Ajouter un 3ème mode est trivial

### Négatives
- ❌ Complexité accrue : 2 modes = 2 chemins de code à maintenir
- ❌ Documentation : Nécessite expliquer les différences CPU vs User Idle
- ❌ macOS seulement : Mode User Idle ne fonctionne pas sur Linux/Windows (stub retourne 0)

### Impact sur les tests
- Passage de 24 à 33 tests (+9 tests)
- Couverture baisse à 43% (CGo non couvert par `go test`)

## Comportement

### Mode CPU (existant)
```
checks_needed = idle_duration / check_interval
→ Accumule N checks consécutifs avec CPU < threshold
→ Exemple : 120s / 5s = 24 checks avant de lancer
```

### Mode User Idle (nouveau)
```
checks_needed = 1
→ CGEventSourceSecondsSinceLastEventType() retourne secondes directement
→ Si >= idle_duration, lance immédiatement
```

**Important** : En mode User Idle, `cpu_threshold` est ignoré.

## Migration
Pas de breaking change :
- Ancien config sans `idle_mode` → défaut = `"cpu"`
- Comportement identique à v1.1

## Références
- Apple Developer: [Quartz Event Services](https://developer.apple.com/documentation/coregraphics/quartz_event_services)
- `CGEventSourceSecondsSinceLastEventType` documentation
- ADR-002 : Fyne GUI (contexte CGo existant)

## Fichiers modifiés
- `config.go` - Type IdleMode, validation
- `monitor.go` - Refactoring CPUMonitor → Monitor, stratégie
- `user_idle_darwin.go` - CGo wrapper CoreGraphics (nouveau)
- `user_idle_other.go` - Stub non-Darwin (nouveau)
- `gui.go` - Radio button idle mode
- Tests : `config_test.go`, `monitor_test.go`
