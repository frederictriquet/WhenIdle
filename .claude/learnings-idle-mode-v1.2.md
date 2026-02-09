# Learnings : Idle Detection Mode (v1.2.0)

## Contexte
Feature complète ajoutant un mode de détection d'inactivité basé sur l'absence d'événements clavier/souris, en complément du mode CPU existant.

**Durée** : ~1 session
**Workflow** : `/analyze` → `/explore-options` → `/architecture` → `/implement` → `/test-write` → `/test-run` → `/code-review` → `/document` → `/capitalize`

---

## 🎯 Objectif Atteint
✅ Support de deux modes : CPU idle et User idle (clavier/souris)
✅ Aucune permission d'accessibilité requise
✅ Radio button dans la GUI pour sélectionner le mode
✅ Tests complets (33 tests, 43% couverture)
✅ Documentation complète (README, CHANGELOG, commentaires)

---

## 🔑 Décisions Techniques

### 1. CGEventSourceSecondsSinceLastEventType (Option A)
**Pourquoi** : API native macOS Quartz Event Services, zéro permission, polling-compatible.

**Alternatives rejetées** :
- gohook : Event-driven incompatible avec notre architecture
- robotgo : Overkill, dépendance lourde, mauvaise réputation

**Fichier** : `user_idle_darwin.go` (~20 lignes CGo)

### 2. Stratégie Pluggable via Function Injection
**Pourquoi** : Simplicité et testabilité sans lourdeur d'interface.

```go
type Monitor struct {
    checkIdle func() (idle bool, detail string) // Injecté à la construction
}
```

**Trade-off** : Moins idiomatique Go, mais acceptable pour stratégie interne.
**Pattern documenté** : `pattern-pluggable-strategy.md`

### 3. Refactoring CPUMonitor → Monitor
**Pourquoi** : Nom plus générique pour supporter les deux modes.

**Impact** :
- Renommage dans `main.go`, `gui.go`, `monitor_test.go`
- `getCPU` → `checkIdle` (signature change)

### 4. checksNeeded = 1 en mode User Idle
**Pourquoi** : `CGEventSourceSecondsSinceLastEventType` retourne déjà les secondes cumulées, pas besoin d'accumuler N checks.

**Différence** :
- Mode CPU : Accumule 24 checks (120s / 5s)
- Mode User Idle : 1 seul check suffit

---

## 🐛 Bugs Résolus Pendant l'Implémentation

### Bug : Singleton Windows (GUI)
**Symptôme** : Cliquer "View Logs" ouvre une nouvelle fenêtre à chaque fois.

**Cause** : `showLogsWindow()` créait systématiquement une nouvelle window.

**Fix** :
```go
type GUI struct {
    logsWindow   fyne.Window // Référence persistante
    configWindow fyne.Window
}

func (g *GUI) showLogsWindow() {
    if g.logsWindow != nil {
        g.logsWindow.Show()
        g.logsWindow.RequestFocus()
        return
    }
    // Créer seulement si nécessaire
}
```

**Leçon** : Pattern singleton pour les fenêtres GUI Fyne.

### Bug : PATH pour Launch Agent
**Symptôme** : `make: npm: No such file or directory` quand la tâche s'exécute via Launch Agent.

**Cause** : Launch Agent fournit un PATH minimal (`/usr/bin:/bin:/usr/sbin:/sbin`). nvm installé dans `~/.nvm` (via zsh) n'est pas dans ce PATH.

**Fix initial raté** : `/bin/bash -l -c 'echo $PATH'` → ne trouve pas nvm car configuré dans `.zshrc`, pas `.bash_profile`.

**Fix final** :
```go
func resolveLoginEnv() []string {
    shell := os.Getenv("SHELL")
    if shell == "" {
        shell = "/bin/zsh" // Défaut macOS
    }
    out, _ := exec.Command(shell, "-l", "-c", "echo $PATH").Output()
    // Remplace PATH dans os.Environ()
}
```

**Leçon** : Toujours utiliser le shell de l'utilisateur (`$SHELL`), pas hardcoder `/bin/bash`.

---

## 🧪 Tests

### Coverage
- **Avant** : 24 tests, ~65% couverture
- **Après** : 33 tests, ~43% couverture (baisse due au CGo non couvert)

### Tests ajoutés
1. `TestIdleModeDefault` - Vérifie défaut = `"cpu"`
2. `TestIdleModeUserIdle` - Vérifie parsing `"user_idle"`
3. `TestIdleModeInvalid` - Vérifie rejet mode invalide
4. `TestMonitorUserIdleMode` - Détection idle en mode User Idle
5. `TestMonitorUserIdleBusy` - Détection busy en mode User Idle

### Tests mis à jour
Tous les tests de `monitor_test.go` mis à jour pour utiliser `checkIdle` mock au lieu de `getCPU`.

**Leçon** : La stratégie pluggable rend les tests triviaux (1 ligne de mock).

---

## 📝 Documentation

### Fichiers mis à jour
- **README.md** :
  - Section "Deux modes de détection"
  - Nouveau tableau Options avec `idle_mode` et `restart`
  - Exemple 4 : backup nocturne en mode User Idle
  - Mise à jour des stats (33 tests, 43% couverture)
- **CHANGELOG.md** :
  - Version 1.2.0 documentée (Added, Changed, Fixed)
- **config.example.json** :
  - Ajout `idle_mode: "cpu"` et `restart: false`
- **Commentaires inline** :
  - Godoc complet sur `UserIdleSeconds()`
  - Commentaires sur `IdleMode`, `Monitor`, `checkUserIdle()`

**Leçon** : Documenter au fur et à mesure, pas à la fin.

---

## 🚀 Bonnes Pratiques

### 1. Workflow Complet
Respect strict de `/analyze` → `/explore-options` → `/architecture` → `/implement` → `/test-write` → `/test-run` → `/code-review` → `/document` → `/capitalize`.

**Bénéfice** : Pas de surprise, pas de refactoring inattendu.

### 2. Build Tag Explicite
```go
//go:build darwin
```
Même si `_darwin.go` suffit, l'expliciter améliore la lisibilité.

**Code review remarque** : Ajouté suite à review.

### 3. Validation Centralisée
`Config.Validate()` est l'unique source de vérité pour la validation.
La GUI appelle `Validate()` au lieu de dupliquer la logique.

**Leçon** : Single source of truth évite les désynchronisations.

---

## 🔧 Refactoring Effectués

### 1. CPUMonitor → Monitor
**Fichiers** : `monitor.go`, `main.go`, `gui.go`, `monitor_test.go`

### 2. getCPU → checkIdle
**Pourquoi** : Signature générique pour CPU et User Idle.

**Avant** :
```go
getCPU func() (float64, error)
```

**Après** :
```go
checkIdle func() (idle bool, detail string)
```

### 3. Réduction du bruit dans les logs
**Changement** : Suppression des logs CPU répétitifs quand `state == Running`.

**Code** :
```go
if state != Running {
    log.Printf("[INFO] %s - idle for %d/%d checks", detail, ...)
}
```

---

## 🤔 Décisions à Revisiter

### 1. Mode User Idle sur non-Darwin
Actuellement, `user_idle_other.go` retourne toujours 0 → système jamais idle.

**Options futures** :
- Implémenter pour Linux (X11 XScreenSaverQueryInfo)
- Implémenter pour Windows (GetLastInputInfo)
- Documenter clairement : "macOS only"

### 2. Couverture de tests CGo
CGo n'est pas couvert par `go test -cover`. Coverage passe de 65% à 43%.

**Options** :
- Accepter (CGo simple, testé manuellement)
- Tests d'intégration macOS-only
- Mock CGo (complexe)

---

## 🎓 Leçons Générales

### 1. CGo n'est pas si effrayant
Un wrapper de 20 lignes suffit pour appeler une API macOS.

**Template** :
```go
//go:build darwin
package main

/*
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>

double myFunc() {
    return CGSomeFunction(...);
}
*/
import "C"

func MyFunc() float64 {
    return float64(C.myFunc())
}
```

### 2. Stratégie via function injection
Alternative valide aux interfaces Go pour les stratégies internes.

**Quand l'utiliser** :
- ✅ Stratégie interne au package
- ✅ Tests avec mocks simples
- ✅ Performance critique
- ❌ API publique

### 3. Singleton pattern pour Fyne windows
Fyne ne gère pas automatiquement la réutilisation de windows.

**Pattern** :
```go
if g.window != nil {
    g.window.Show()
    return
}
g.window = g.app.NewWindow(...)
g.window.SetCloseIntercept(func() {
    g.window = nil // Cleanup
    g.window.Close()
})
```

### 4. Launch Agent PATH gotchas
Les Launch Agents héritent d'un PATH minimal. **Toujours** résoudre via le shell utilisateur.

---

## 📊 Métriques

| Métrique | Avant | Après |
|----------|-------|-------|
| Fichiers Go | 13 | 15 (+2: user_idle) |
| Tests | 24 | 33 (+9) |
| Couverture | ~65% | ~43% (CGo non couvert) |
| Modes supportés | 1 (CPU) | 2 (CPU + User) |
| Permissions requises | 0 | 0 |

---

## 🔗 Fichiers de Capitalisation Créés

1. `pattern-pluggable-strategy.md` - Pattern de stratégie via function injection
2. `adr-003-idle-detection-mode.md` - ADR complet sur la décision
3. `learnings-idle-mode-v1.2.md` - Ce fichier (leçons apprises)

---

## ✅ Checklist de Complétion

- [x] Feature complètement implémentée
- [x] Tests écrits et passants (33/33)
- [x] Code review approuvé
- [x] Documentation complète (README, CHANGELOG, inline)
- [x] Capitalisation SERENA
- [ ] Roadmap mise à jour (prochaine étape)
- [ ] Pre-merge checklist (après roadmap)

---

**Date** : 2026-02-09
**Version** : 1.2.0
**Branch** : `feature/idle-detection-mode`
