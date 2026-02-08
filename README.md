# WhenIdle

> Daemon macOS qui surveille l'utilisation CPU et lance automatiquement des tâches intensives pendant les périodes d'inactivité

## Table des Matières

- [Description](#description)
- [Fonctionnalités](#fonctionnalités)
- [Installation](#installation)
- [Configuration](#configuration)
- [Utilisation](#utilisation)
- [Exemples](#exemples)
- [Désinstallation](#désinstallation)
- [Architecture](#architecture)
- [Développement](#développement)
- [License](#license)

## Description

WhenIdle est un daemon macOS léger qui surveille l'utilisation CPU de votre système et lance automatiquement des tâches CPU-intensives uniquement pendant les périodes d'inactivité. Cela permet d'exécuter des tâches longues (compilation, encodage, traitement de données) sans impacter vos performances lorsque vous travaillez.

### Comportement

```
CPU idle (< seuil) pendant N secondes → Lance la tâche
CPU actif (> seuil) → Suspend la tâche (SIGSTOP)
CPU redevient idle → Reprend la tâche (SIGCONT)
Arrêt du daemon → Termine proprement la tâche (SIGTERM/SIGKILL)
```

## Fonctionnalités

- ✅ **Interface graphique (GUI)** dans la barre de menu macOS (system tray)
- ✅ **Surveillance CPU en temps réel** avec seuil configurable
- ✅ **Suspension/reprise automatique** via SIGSTOP/SIGCONT
- ✅ **Process group signaling** pour gérer toute l'arborescence de processus
- ✅ **Launch Agent macOS** pour démarrage automatique au login
- ✅ **Configuration via GUI** ou fichier JSON
- ✅ **Logs en temps réel** visibles dans la GUI
- ✅ **Icônes dynamiques** indiquant l'état (idle/running/paused)
- ✅ **Mode headless (CLI)** disponible sans GUI
- ✅ **Graceful shutdown** avec timeout SIGKILL
- ✅ **Binary standalone** sans dépendances runtime (~32 MB)

## Installation

### Prérequis

- macOS 10.14+ (Darwin)
- Go 1.18+ avec CGO (pour compiler depuis les sources)

### Installation automatique

```bash
cd /path/to/WhenIdle
./install.sh
```

Cette commande :
1. Compile le binary `whenidle` avec support GUI (CGO + Fyne)
2. Crée un bundle `.app` dans `~/Applications/WhenIdle.app/`
3. Installe aussi le CLI dans `~/.local/bin/`
4. Crée une configuration par défaut dans `~/.config/whenidle/config.json`
5. Configure le Launch Agent dans `~/Library/LaunchAgents/`

**Note** : L'app apparaîtra dans votre **barre de menu** (menu bar), pas dans le Dock.

### Installation manuelle

```bash
# 1. Compiler
go build -o whenidle .

# 2. Copier le binary
mkdir -p ~/.local/bin
cp whenidle ~/.local/bin/
chmod +x ~/.local/bin/whenidle

# 3. Créer la configuration
mkdir -p ~/.config/whenidle
cp config.example.json ~/.config/whenidle/config.json

# 4. Éditer la configuration (voir section Configuration)
nano ~/.config/whenidle/config.json

# 5. Installer le Launch Agent
mkdir -p ~/Library/LaunchAgents
sed -e "s|__INSTALL_DIR__|$HOME/.local/bin|g" \
    -e "s|__CONFIG_DIR__|$HOME/.config/whenidle|g" \
    com.whenidle.agent.plist > ~/Library/LaunchAgents/com.whenidle.agent.plist
```

## Configuration

Le fichier de configuration est au format JSON.

### Emplacement

`~/.config/whenidle/config.json`

### Options

| Option | Type | Défaut | Description |
|--------|------|--------|-------------|
| `cpu_threshold` | float | 15.0 | % CPU global en dessous duquel le système est considéré idle (0-100) |
| `idle_duration` | int | 120 | Nombre de secondes consécutives d'idle avant de lancer la tâche |
| `check_interval` | int | 5 | Intervalle de vérification CPU en secondes |
| `command` | string | **requis** | Chemin complet de la commande à exécuter |
| `args` | []string | [] | Arguments à passer à la commande |
| `working_dir` | string | **requis** | Répertoire de travail pour la commande |
| `log_file` | string | "" | Fichier de log optionnel (vide = stdout) |

### Exemple de configuration

```json
{
    "cpu_threshold": 15.0,
    "idle_duration": 120,
    "check_interval": 5,
    "command": "/usr/local/bin/heavy-task",
    "args": ["--option1", "value1"],
    "working_dir": "/Users/fred/projects/heavy-task",
    "log_file": "/tmp/whenidle.log"
}
```

### Calcul de l'idle

Le nombre de vérifications nécessaires est calculé automatiquement :

```
checks_needed = idle_duration / check_interval
```

Exemple : avec `idle_duration=120` et `check_interval=5`, la tâche se lance après **24 vérifications consécutives** à CPU < seuil (soit 2 minutes).

## Utilisation

### Mode GUI (recommandé)

L'interface graphique s'affiche dans la **barre de menu** macOS (system tray).

#### Démarrer la GUI

```bash
launchctl load ~/Library/LaunchAgents/com.whenidle.agent.plist
```

Une icône apparaît dans votre barre de menu avec les couleurs suivantes :
- 🔴 **Gris** : Monitoring désactivé
- 🟢 **Vert** : Monitoring actif, système idle
- 🔵 **Bleu** : Tâche en cours d'exécution
- 🟠 **Orange** : Tâche mise en pause (CPU busy)

#### Menu de l'icône

- **Enable/Disable Monitoring** : Active ou désactive le monitoring
- **Configure Task...** : Ouvre une fenêtre pour configurer la tâche
- **View Logs...** : Affiche les logs en temps réel (dernières 500 lignes)
- **Quit** : Arrête l'application proprement

#### Configuration via GUI

1. Cliquez sur l'icône dans la barre de menu
2. Sélectionnez **"Configure Task..."**
3. Modifiez les paramètres dans le formulaire
4. Cliquez **"Save"**
5. Si le monitoring est actif, il redémarre automatiquement avec la nouvelle config

#### Visualiser les logs

1. Cliquez sur l'icône dans la barre de menu
2. Sélectionnez **"View Logs..."**
3. Les logs s'affichent en temps réel dans une fenêtre scrollable

### Mode CLI (headless)

Pour utiliser WhenIdle sans GUI (ex: via SSH) :

```bash
~/.local/bin/whenidle --config ~/.config/whenidle/config.json
```

Arrêt : `Ctrl+C`

### Commandes utiles

#### Arrêter l'application

```bash
launchctl unload ~/Library/LaunchAgents/com.whenidle.agent.plist
```

#### Redémarrer

```bash
launchctl unload ~/Library/LaunchAgents/com.whenidle.agent.plist
launchctl load ~/Library/LaunchAgents/com.whenidle.agent.plist
```

#### Vérifier le statut

```bash
launchctl list | grep whenidle
```

#### Consulter les logs (si besoin brut)

```bash
tail -f /tmp/whenidle.stdout.log
```

## Exemples

### Exemple 1 : Compilation de projet

```json
{
    "cpu_threshold": 20.0,
    "idle_duration": 60,
    "check_interval": 5,
    "command": "/usr/bin/make",
    "args": ["build"],
    "working_dir": "/Users/fred/projects/myapp"
}
```

Lance `make build` après 1 minute d'inactivité (CPU < 20%).

### Exemple 2 : Encodage vidéo

```json
{
    "cpu_threshold": 15.0,
    "idle_duration": 300,
    "check_interval": 10,
    "command": "/usr/local/bin/ffmpeg",
    "args": ["-i", "input.mp4", "-c:v", "libx264", "output.mp4"],
    "working_dir": "/Users/fred/Videos"
}
```

Lance l'encodage ffmpeg après 5 minutes d'inactivité (CPU < 15%), vérifie toutes les 10 secondes.

### Exemple 3 : Script shell personnalisé

```json
{
    "cpu_threshold": 25.0,
    "idle_duration": 120,
    "check_interval": 5,
    "command": "/bin/bash",
    "args": ["-c", "date >> /tmp/idle-task.log && sleep 30"],
    "working_dir": "/tmp"
}
```

Exécute un script shell simple qui log la date.

## Désinstallation

```bash
./uninstall.sh
```

Cette commande :
1. Arrête et décharge le Launch Agent
2. Supprime le binary `~/.local/bin/whenidle`
3. Supprime le plist `~/Library/LaunchAgents/com.whenidle.agent.plist`
4. **Préserve** la configuration `~/.config/whenidle/`

Pour supprimer également la configuration :

```bash
rm -rf ~/.config/whenidle
```

## Architecture

### Composants

#### Mode GUI
```
┌─────────────────────────────────────────────┐
│              WhenIdle GUI                    │
│                                              │
│  ┌──────────┐   ┌─────────┐                │
│  │   Fyne   │   │LogBuffer│                │
│  │System Tray◀─┤Ring (500)│◀────┐          │
│  └─────┬────┘   └─────────┘     │          │
│        │                          │          │
│        ▼                          │          │
│  ┌──────────┐   ┌──────────────┐│          │
│  │  Config   │──▶│   Monitor    ├┘          │
│  │  (JSON)   │   │  (CPU poll)  │           │
│  └──────────┘   └──────┬───────┘           │
│                         │                    │
│                         ▼                    │
│                  ┌──────────────┐           │
│                  │ Task Runner  │           │
│                  │ (exec, STOP/ │           │
│                  │  CONT, TERM) │           │
│                  └──────────────┘           │
│                                              │
└──────────────────────────────────────────────┘
```

#### Mode CLI
```
┌─────────────────────────────────────────────┐
│              WhenIdle CLI                    │
│                                              │
│  ┌──────────┐   ┌──────────────┐           │
│  │  Config   │──▶│   Monitor    │           │
│  │  (JSON)   │   │  (CPU poll)  │           │
│  └──────────┘   └──────┬───────┘           │
│                         │                    │
│                         ▼                    │
│                  ┌──────────────┐           │
│                  │ Task Runner  │           │
│                  │ (exec, STOP/ │           │
│                  │  CONT, TERM) │           │
│                  └──────────────┘           │
│                                              │
└──────────────────────────────────────────────┘
```

### Machine à états du TaskRunner

```
                    onIdle()
    ┌─────────┐ ──────────▶ ┌─────────┐
    │ Stopped │             │ Running │
    └─────────┘ ◀────────── └────┬────┘
                process exits     │ onBusy()
                                  ▼
                             ┌──────────┐
                             │  Paused  │
                             └────┬─────┘
                                  │ onIdle()
                                  ▼
                             ┌─────────┐
                             │ Running │
                             └─────────┘
```

### Fichiers

| Fichier | Rôle |
|---------|------|
| `main.go` | Point d'entrée, routing CLI/GUI |
| `gui.go` | Interface Fyne (system tray, config, logs) |
| `logbuf.go` | Ring buffer thread-safe pour les logs |
| `icon.go` | Génération programmatique des icônes tray |
| `dock_darwin.go` | CGo pour cacher l'app du Dock macOS |
| `dock_other.go` | Stub no-op pour plateformes non-Darwin |
| `config.go` | Chargement, sauvegarde et validation JSON |
| `monitor.go` | Surveillance CPU avec boucle de polling |
| `runner.go` | Gestion du processus (exec, SIGSTOP/SIGCONT/SIGTERM) |
| `*_test.go` | Tests unitaires (24 tests) |

## Développement

### Prérequis

- Go 1.18+
- macOS (pour tester le Launch Agent)

### Compiler

```bash
# Avec GUI (nécessite CGO)
CGO_ENABLED=1 go build -o whenidle .

# Sans GUI (headless, plus petit binary)
go build -tags no_gui -o whenidle .
```

### Exécuter les tests

```bash
go test -v ./...
```

### Tests avec couverture

```bash
go test -v -cover ./...
```

### Structure du code

```
.
├── main.go                   # Point d'entrée CLI/GUI
├── gui.go                    # Interface Fyne (system tray)
├── logbuf.go                 # Ring buffer pour logs
├── logbuf_test.go            # Tests du LogBuffer
├── icon.go                   # Génération icônes tray
├── dock_darwin.go            # CGo pour cacher du Dock
├── dock_other.go             # Stub non-Darwin
├── config.go                 # Config JSON (load/save)
├── config_test.go            # Tests config
├── monitor.go                # Surveillance CPU
├── monitor_test.go           # Tests monitor
├── runner.go                 # Gestion processus
├── runner_test.go            # Tests runner
├── config.example.json       # Config par défaut
├── install.sh                # Installation + .app bundle
├── uninstall.sh              # Désinstallation
├── com.whenidle.agent.plist  # Template Launch Agent
└── README.md                 # Cette documentation
```

### Tests

Le projet inclut 24 tests unitaires couvrant :
- Validation de la configuration (save/load round-trip)
- LogBuffer (ring buffer, thread safety, onChange callbacks)
- Machine à états du runner (Stopped → Running → Paused)
- Logique de détection idle/busy du monitor
- Signaling SIGSTOP/SIGCONT/SIGTERM

Couverture : ~65%

### Dépendances

- `github.com/shirou/gopsutil/v4` — Monitoring CPU cross-platform
- `fyne.io/fyne/v2` — Framework GUI et system tray (v2.7.2)

## FAQ

### Q: La tâche ne se lance jamais, pourquoi ?

Vérifiez :
1. Le seuil `cpu_threshold` n'est pas trop bas (votre CPU réel)
2. La durée `idle_duration` n'est pas trop longue
3. Les logs : `tail -f /tmp/whenidle.stdout.log`

### Q: La tâche ne se suspend pas quand je travaille ?

Le monitor vérifie toutes les `check_interval` secondes. Si votre pic CPU est court, il peut ne pas être détecté. Réduisez `check_interval` (ex: 3s).

### Q: Comment vérifier que le daemon fonctionne ?

```bash
ps aux | grep whenidle
launchctl list | grep whenidle
tail -f /tmp/whenidle.stdout.log
```

### Q: La tâche ne s'arrête pas proprement ?

WhenIdle envoie d'abord SIGTERM, puis SIGKILL après 5 secondes. Si votre processus ignore SIGTERM, il sera tué de force.

### Q: Puis-je lancer plusieurs tâches différentes ?

Non, un seul daemon = une seule tâche. Pour plusieurs tâches, créez plusieurs configurations et Launch Agents avec des noms différents.

### Q: L'icône n'apparaît pas dans la barre de menu ?

Vérifiez :
1. Que le Launch Agent pointe bien vers `~/Applications/WhenIdle.app/Contents/MacOS/whenidle`
2. Les logs : `tail -f /tmp/whenidle.stdout.log`
3. Relancez : `launchctl unload ... && launchctl load ...`

### Q: L'app apparaît dans le Dock au lieu de la barre de menu ?

Le bundle `.app` doit avoir `LSUIElement=true` dans `Info.plist`. Ré-exécutez `./install.sh`.

## License

MIT
