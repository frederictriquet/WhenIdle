# Apprentissages : GUI System Tray v1.1.0

Date : 2026-02-08
Projet : WhenIdle
Branche : `feature/gui-system-tray`

---

## ADR-002 : Choix de Fyne pour le System Tray macOS

### Contexte
WhenIdle nécessitait une interface GUI dans la barre de menu macOS (system tray) pour permettre une configuration sans édition manuelle du JSON.

### Options Considérées

1. **Fyne.io v2.7.2** (choisi)
   - Score : 124/155
   - Avantages : API system tray native depuis v2.2+, cross-platform, documentation complète
   - Inconvénients : Binary plus lourd (~32 MB), nécessite CGO

2. **Wails v3**
   - Score : 104/155
   - Avantages : Web UI moderne, hot-reload
   - Inconvénients : Pas de system tray natif, webview nécessaire

3. **DarwinKit/Cocoa**
   - Score : 83/155
   - Avantages : Cocoa pur, léger
   - Inconvénients : macOS-only, bindings complexes

4. **getlantern/systray**
   - Score : 111/155
   - Avantages : Minimal, léger
   - Inconvénients : Pas de fenêtres, menu basique seulement

### Décision
Fyne.io pour son API système tray complète, sa capacité à créer des fenêtres (config, logs), et son support multi-plateforme.

### Conséquences
- Binary passe de 3 MB à 32 MB (CGO + Fyne)
- Nécessite `CGO_ENABLED=1` pour compiler
- Excellente UX avec fenêtres natives et icônes dynamiques

---

## Pattern : LogBuffer Ring Buffer Thread-Safe

### Problème
Capturer les logs stdout dans une GUI nécessite :
- Thread safety (plusieurs goroutines écrivent)
- Limitation mémoire (éviter accumulation infinie)
- Notifications pour mise à jour UI en temps réel

### Solution
Ring buffer implémentant `io.Writer` avec callback `onChange`.

### Code
```go
type LogBuffer struct {
    mu       sync.Mutex
    lines    []string
    maxLines int
    onChange func()
}

func (b *LogBuffer) Write(p []byte) (n int, err error) {
    b.mu.Lock()

    lines := bytes.Split(p, []byte("\n"))
    for i, line := range lines {
        if len(line) == 0 && i == len(lines)-1 {
            continue
        }
        b.lines = append(b.lines, string(line))
        if len(b.lines) > b.maxLines {
            b.lines = b.lines[len(b.lines)-b.maxLines:]
        }
    }

    onChange := b.onChange
    b.mu.Unlock()

    if onChange != nil {
        onChange() // Synchrone après unlock
    }

    return len(p), nil
}
```

### Avantages
- Thread-safe via mutex
- Pas de goroutine storm (onChange synchrone après unlock)
- Ring buffer = mémoire bornée (500 lignes max)
- Copie defensive dans `Lines()` évite races

### Leçon Apprise
**ERREUR INITIALE** : `onChange` lancé dans une goroutine (`go onChange()`).
**PROBLÈME** : Goroutine storm quand logs fréquents (monitor actif).
**FIX** : Appel synchrone après `Unlock()`, sans goroutine.

### Utilisation
```go
logBuf := NewLogBuffer(500)
multiWriter := io.MultiWriter(os.Stdout, logBuf)
log.SetOutput(multiWriter)

logBuf.SetOnChange(func() {
    fyne.Do(func() { /* update UI */ })
})
```

---

## Pattern : HideFromDock via CGo NSApp

### Problème
Fyne apps apparaissent par défaut dans le Dock macOS. Pour une app menu bar only, le `LSUIElement=true` dans Info.plist ne suffit pas car Fyne force sa propre activation policy.

### Solution
CGo pour appeler `NSApp.setActivationPolicy(.accessory)` via dispatch_async.

### Code
```go
// dock_darwin.go
/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

void hideFromDock() {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    });
}
*/
import "C"

func HideFromDock() {
    C.hideFromDock()
}
```

### Timing Critique
Appeler **après** que Fyne soit initialisé, sinon NSApp n'existe pas :
```go
gui.app.Lifecycle().SetOnStarted(func() {
    HideFromDock()
})
```

### Multi-plateforme
```go
// dock_other.go
//go:build !darwin

func HideFromDock() {} // no-op
```

### Alternatives Échouées
1. ❌ Info.plist seul → Fyne override
2. ❌ Appel avant `app.New()` → NSApp nil
3. ✅ Lifecycle hook `SetOnStarted` → Fonctionne

---

## Pattern : Icônes Dynamiques pour System Tray

### Problème
Indiquer visuellement l'état de l'app (disabled/idle/running/paused) sans dépendre de fichiers statiques.

### Solution
Génération programmatique de cercles PNG colorés via `image` package.

### Code
```go
func generateCircleIcon(name string, c color.Color) fyne.Resource {
    img := image.NewNRGBA(image.Rect(0, 0, 22, 22))

    cx, cy := 11, 11
    radius := 9

    for y := 0; y < 22; y++ {
        for x := 0; x < 22; x++ {
            dx := x - cx
            dy := y - cy
            if dx*dx+dy*dy <= radius*radius {
                img.Set(x, y, c)
            }
        }
    }

    var buf bytes.Buffer
    png.Encode(&buf, img)

    return fyne.NewStaticResource("tray-"+name+".png", buf.Bytes())
}
```

### États
- Gris (`R:140, G:140, B:140`) : Monitoring désactivé
- Vert (`R:80, G:180, B:80`) : Idle, monitoring actif
- Bleu (`R:50, G:140, B:240`) : Tâche running
- Orange (`R:240, G:180, B:40`) : Tâche paused

### Avantages
- Pas de fichiers PNG à maintenir
- Noms distincts évitent confusion cache Fyne
- Taille : 22x22 (standard macOS menu bar)

---

## Bug : Fenêtre Fantôme avec ShowAndRun()

### Symptômes
Une petite fenêtre 1x1 invisible apparaissait en plus de la fenêtre de logs.

### Cause
`window.ShowAndRun()` crée une fenêtre cachée pour démarrer l'event loop Fyne. Cette fenêtre devient parfois visible.

### Solution
Remplacer `ShowAndRun()` par `app.Run()` qui lance l'event loop sans nécessiter de fenêtre visible.

```go
// ❌ Avant
w := gui.app.NewWindow("WhenIdle")
w.Resize(fyne.NewSize(1, 1))
w.ShowAndRun()

// ✅ Après
gui.app.Run() // Pas de fenêtre nécessaire
```

### Leçon
Pour des apps system tray sans fenêtre principale, utiliser `app.Run()` directement.

---

## Bug : Logs Illisibles (Gris sur Gris)

### Symptômes
Fenêtre de logs affichait du texte gris clair sur fond gris.

### Cause
`widget.NewMultiLineEntry().Disable()` grey out le texte.

### Solution
Utiliser `widget.NewRichTextWithText()` qui est read-only par nature avec couleurs normales.

```go
// ❌ Avant
logText := widget.NewMultiLineEntry()
logText.Disable()

// ✅ Après
logText := widget.NewRichTextWithText(logs)
```

### Update Live
```go
logText.Segments = []widget.RichTextSegment{
    &widget.TextSegment{Text: strings.Join(lines, "\n")},
}
logText.Refresh()
```

**NE PAS** utiliser `ParseMarkdown()` car il fusionne les newlines simples.

---

## Bug : Lignes de Logs Concaténées

### Symptômes
Les logs s'affichaient sur une seule ligne au lieu de lignes séparées.

### Cause
`logText.ParseMarkdown()` traite les single newlines comme des espaces (format Markdown).

### Solution
Mise à jour directe des segments avec `TextSegment` :

```go
logText.Segments = []widget.RichTextSegment{
    &widget.TextSegment{Text: strings.Join(g.logBuf.Lines(), "\n")},
}
logText.Refresh()
```

---

## Standard : Validation Unique Source of Truth

### Règle
La validation des données doit avoir une **unique source de vérité**. Ne pas dupliquer la logique de validation.

### Justification
Double validation (GUI + `Config.Validate()`) peut diverger et créer des bugs.

### Exemple

#### ❌ Incorrect
```go
// GUI
threshold, err := strconv.ParseFloat(text, 64)
if err != nil || threshold <= 0 || threshold > 100 {
    return error
}

// Config.Validate()
if c.CPUThreshold <= 0 || c.CPUThreshold > 100 {
    return error
}
```

#### ✅ Correct
```go
// GUI - Parse seulement
threshold, err := strconv.ParseFloat(text, 64)
if err != nil {
    return fmt.Errorf("parse error: %w", err)
}

// Config.Validate() - Unique validation des ranges
if c.CPUThreshold <= 0 || c.CPUThreshold > 100 {
    return error
}
```

### Application
GUI parse les valeurs, `Config.Validate()` vérifie les contraintes métier.

---

## Pattern : Callbacks Wrappés pour UI Sync

### Problème
Quand le monitor détecte idle/busy, l'icône du tray doit se mettre à jour mais :
- Le runner change son état (via OnIdle/OnBusy)
- La GUI doit lire ce nouvel état
- L'update UI doit être thread-safe (Fyne main thread)

### Solution
Wrapper les callbacks pour ajouter l'update UI après l'action.

```go
onIdle := func() {
    g.runner.OnIdle()  // Change l'état
    fyne.Do(func() {   // Thread-safe UI update
        g.updateTrayIcon()
    })
}

onBusy := func() {
    g.runner.OnBusy()
    fyne.Do(func() {
        g.updateTrayIcon()
    })
}

g.monitor = NewCPUMonitor(cfg, onIdle, onBusy, g.runner.State)
```

### Avantages
- État du runner toujours à jour avant l'UI refresh
- `fyne.Do()` garantit l'exécution sur le main thread
- Icône se met à jour en temps réel (vert → bleu → orange)

---

## Métriques de Développement

### Durée
- Analyse → Architecture : ~2h
- Implémentation initiale : ~4h
- Debug itératif (6 bugs) : ~2h
- Code review + fixes : ~1h
- Documentation : ~30min
- **Total : ~9-10h**

### Complexité
- Fichiers créés : 6 (gui.go, logbuf.go, icon.go, dock_darwin.go, dock_other.go, logbuf_test.go)
- Lignes ajoutées : +905
- Tests : 24 (dont 7 nouveaux pour LogBuffer)
- Couverture : 65%

### Bugs Rencontrés
1. Launch Agent pointait vers ancien binary
2. Duplicate "Quit" menu items
3. Logs gris illisibles
4. Fenêtre fantôme
5. Logs lignes concaténées
6. Icône dans le Dock au lieu du menu bar

### Patterns Réutilisables
- LogBuffer ring buffer → Applicable à toute app avec logs UI
- HideFromDock CGo → Pattern pour apps menu bar macOS
- Icônes programmatiques → Éviter fichiers statiques
- Validation unique source → S'applique à tout formulaire

---

## Dépendances Ajoutées

```go
require (
    fyne.io/fyne/v2 v2.7.2
    // + ~30 dépendances transitives
)
```

### Compilation
```bash
CGO_ENABLED=1 go build -o whenidle .
```

Build time : ~30s sur macOS (Darwin 23.5.0)
Binary size : 32 MB (vs 3 MB CLI-only)

---

## Points d'Attention pour Futures Features

### Performance
- LogBuffer limité à 500 lignes : si logs très verbeux, peut perdre historique
- onChange synchrone : si callback lent, bloque l'écriture (mais évite goroutine storm)

### Portabilité
- `dock_darwin.go` + `dock_other.go` : pattern à réutiliser pour code platform-specific
- Fyne cross-platform mais bundle `.app` macOS-only

### UX
- Icône 22x22 : respecter les guidelines macOS (Retina-ready)
- Menu simple 4 items : ajouter features progressivement sans surcharger

### Tests
- GUI non testée unitairement (Fyne difficilement mockable)
- LogBuffer testé exhaustivement (thread safety critique)
- Focus tests sur la logique métier (runner, monitor)

---

## Ressources Utiles

- [Fyne System Tray Docs](https://docs.fyne.io/extend/systray)
- [macOS NSApplicationActivationPolicy](https://developer.apple.com/documentation/appkit/nsapplication/1428621-activationpolicy)
- [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)

---

## Conclusion

Implémentation GUI réussie avec Fyne, malgré quelques challenges macOS-specific (Dock hiding, lifecycle timing). Les patterns ring buffer et CGo integration sont réutilisables pour d'autres projets Go/GUI sur macOS.

**Prochain objectif** : Merge vers `main` après roadmap update.
