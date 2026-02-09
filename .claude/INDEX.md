# Index des Connaissances - WhenIdle

## 📋 Vue d'Ensemble
Ce répertoire contient la capitalisation des décisions, patterns et apprentissages du projet WhenIdle.

**Dernière mise à jour** : 2026-02-09

---

## 🏛️ Architecture Decision Records (ADR)

| ID | Date | Titre | Statut | Fichier |
|----|------|-------|--------|---------|
| 002 | 2026-02-08 | Interface Fyne GUI | Accepté | `adr-002-fyne-gui.md` |
| 003 | 2026-02-09 | Idle Detection Mode (CPU + User) | Accepté | `adr-003-idle-detection-mode.md` |

---

## 🎨 Patterns

| Nom | Description | Fichier |
|-----|-------------|---------|
| Pluggable Strategy | Function injection pour stratégies testables | `pattern-pluggable-strategy.md` |

---

## 📚 Learnings (Leçons Apprises)

| Version | Date | Description | Fichier |
|---------|------|-------------|---------|
| v1.1.0 | 2026-02-08 | GUI avec Fyne, system tray macOS | `learnings-gui-v1.1.md` |
| v1.2.0 | 2026-02-09 | Mode User Idle, stratégie pluggable | `learnings-idle-mode-v1.2.md` |

---

## 🔍 Index par Sujet

### macOS / Darwin
- CGo pour Cocoa (cacher du Dock) → `adr-002-fyne-gui.md`
- CGo pour CoreGraphics (User Idle) → `adr-003-idle-detection-mode.md`
- Launch Agent PATH resolution → `learnings-idle-mode-v1.2.md`

### GUI / Fyne
- System tray avec icônes dynamiques → `learnings-gui-v1.1.md`
- Singleton pattern pour windows → `learnings-idle-mode-v1.2.md`
- LogBuffer ring pour logs live → `learnings-gui-v1.1.md`

### Architecture
- Stratégie pluggable via function injection → `pattern-pluggable-strategy.md`
- Monitor refactoring (CPU → générique) → `adr-003-idle-detection-mode.md`

### Testing
- Mock via function reassignment → `pattern-pluggable-strategy.md`
- CGo coverage gotchas → `learnings-idle-mode-v1.2.md`

---

## 📊 Progression du Projet

### Versions
- **v1.0.0** (2026-02-07) : CLI daemon, CPU monitoring
- **v1.1.0** (2026-02-08) : GUI Fyne, system tray
- **v1.2.0** (2026-02-09) : Idle detection modes (CPU + User)

### Métriques Actuelles
- **Fichiers Go** : 15
- **Tests** : 33 (43% couverture)
- **Modes supportés** : 2 (CPU, User Idle)
- **Plateformes** : macOS (Darwin)

---

## 🎯 Décisions Clés à Retenir

### 1. Fyne pour la GUI (ADR-002)
- **Pourquoi** : Pure Go, system tray cross-platform
- **Trade-off** : Binary ~32 MB, nécessite CGo

### 2. CGEventSource pour User Idle (ADR-003)
- **Pourquoi** : Aucune permission requise, polling-compatible
- **Alternatives rejetées** : gohook (event-driven), robotgo (overkill)

### 3. Function Injection pour Stratégie
- **Pourquoi** : Simplicité, testabilité sans interface formelle
- **Trade-off** : Moins idiomatique Go, mais acceptable

---

## 🛠️ Patterns Réutilisables

### Pluggable Strategy
```go
type Component struct {
    strategy func() (result, detail)
}

func New(mode string) *Component {
    c := &Component{}
    switch mode {
    case "A":
        c.strategy = c.strategyA
    case "B":
        c.strategy = c.strategyB
    }
    return c
}
```

**Utilisation** : `monitor.go`, `checkIdle` function

### Singleton Window (Fyne)
```go
if g.window != nil {
    g.window.Show()
    return
}
g.window = g.app.NewWindow(...)
g.window.SetCloseIntercept(func() {
    g.window = nil
    g.window.Close()
})
```

**Utilisation** : `gui.go`, logsWindow & configWindow

### Login Shell PATH Resolution
```go
shell := os.Getenv("SHELL")
if shell == "" {
    shell = "/bin/zsh"
}
out, _ := exec.Command(shell, "-l", "-c", "echo $PATH").Output()
```

**Utilisation** : `runner.go`, resolveLoginEnv()

---

## 📖 Comment Utiliser cet Index

### Avant d'ajouter une feature
1. Consulter les ADR pour comprendre les décisions architecturales
2. Vérifier les patterns existants réutilisables
3. Lire les learnings pour éviter les pièges connus

### Après avoir terminé une feature
1. Créer un ADR si décision architecturale importante
2. Documenter les patterns réutilisables
3. Écrire un learnings-*.md avec les leçons apprises
4. Mettre à jour cet INDEX.md

### Pendant le développement
- Consulter les learnings pour les gotchas (PATH, CGo, Fyne, etc.)
- S'inspirer des patterns existants
- Référencer les ADR dans les commit messages

---

## 🔗 Liens Externes

### Documentation
- [Fyne.io Documentation](https://developer.fyne.io/)
- [Apple CoreGraphics](https://developer.apple.com/documentation/coregraphics)
- [gopsutil](https://github.com/shirou/gopsutil)

### Standards
- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
- [Effective Go](https://go.dev/doc/effective_go)

---

**Contributeurs** : Claude Opus 4.6 + Fred
**Projet** : WhenIdle - macOS Idle Task Runner
**Repository** : (à définir)
