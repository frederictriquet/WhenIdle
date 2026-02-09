# Pattern : Pluggable Strategy via Function Injection

## Problème
Comment rendre un composant testable et extensible sans la lourdeur d'une interface formelle en Go ?

## Contexte
- Besoin de supporter plusieurs implémentations d'un algorithme (ex: détection d'inactivité CPU vs User Input)
- Tests unitaires nécessitent des mocks simples
- Go ne nécessite pas toujours des interfaces pour l'abstraction
- Performance critique (pas de virtual dispatch overhead souhaité)

## Solution
Injecter une fonction de stratégie directement dans la structure, sans interface.

## Code

### Structure avec stratégie pluggable
```go
type Monitor struct {
    config    Config
    // ... autres champs
    checkIdle func() (idle bool, detail string) // stratégie injectable
}

func NewMonitor(config Config, ...) *Monitor {
    m := &Monitor{
        config: config,
        // ...
    }

    // Sélection de la stratégie basée sur la config
    switch config.IdleMode {
    case IdleModeUserIdle:
        m.checkIdle = m.checkUserIdle
    default:
        m.checkIdle = m.checkCPUIdle
    }

    return m
}

// Implémentations de stratégies
func (m *Monitor) checkCPUIdle() (bool, string) {
    percentages, _ := cpu.Percent(0, false)
    usage := percentages[0]
    return usage < m.config.CPUThreshold, fmt.Sprintf("CPU at %.1f%%", usage)
}

func (m *Monitor) checkUserIdle() (bool, string) {
    seconds := UserIdleSeconds()
    idle := seconds >= float64(m.config.IdleDuration)
    return idle, fmt.Sprintf("User idle %.0fs", seconds)
}
```

### Tests avec mock
```go
func TestMonitorIdleDetection(t *testing.T) {
    monitor := NewMonitor(cfg, onIdle, onBusy, getState)

    // Mock simple via réaffectation de fonction
    monitor.checkIdle = func() (bool, string) {
        return true, "CPU at 20.0%"
    }

    monitor.tick(3) // utilise le mock
}
```

## Avantages
- **Simplicité** : Pas besoin de définir une interface `IdleChecker`
- **Performance** : Appel direct de fonction (pas de virtual dispatch)
- **Testabilité** : Mock en une seule ligne via réaffectation
- **Flexibilité** : Peut changer la stratégie dynamiquement si besoin
- **Lisibilité** : Signature de fonction explicite dans la struct

## Inconvénients
- Moins "idiomatique" que les interfaces en Go
- Pas de vérification de type au compile-time (mais accepteur = tests)
- Documentation moins explicite (la fonction n'a pas de nom de type)

## Quand utiliser ce pattern ?
- ✅ Stratégie interne au package
- ✅ Tests nécessitent des mocks simples
- ✅ Performance critique
- ✅ Petit nombre d'implémentations (<5)
- ❌ API publique (préférer interface)
- ❌ Besoin de reflection/introspection

## Exemples d'utilisation
- `monitor.go:21` - checkIdle func() (idle bool, detail string)
- `monitor.go:32-37` - Sélection de stratégie dans NewMonitor
- `monitor_test.go:31-33` - Mock via réaffectation

## Alternative : Interface classique
```go
type IdleChecker interface {
    CheckIdle() (bool, string)
}

type CPUChecker struct { ... }
func (c *CPUChecker) CheckIdle() (bool, string) { ... }

type UserIdleChecker struct { ... }
func (u *UserIdleChecker) CheckIdle() (bool, string) { ... }
```

**Trade-off** : Plus verbeux, mais compile-time safety et meilleure documentation.

## Voir aussi
- Strategy Pattern (GoF)
- Go Proverbs: "Accept interfaces, return structs"
- Effective Go: Function values
