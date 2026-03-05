# WhenIdle

[![CI](https://github.com/frederictriquet/WhenIdle/actions/workflows/ci.yml/badge.svg)](https://github.com/frederictriquet/WhenIdle/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/frederictriquet/WhenIdle/branch/master/graph/badge.svg)](https://codecov.io/gh/frederictriquet/WhenIdle)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

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

WhenIdle supporte **deux modes** de détection d'inactivité :

#### Mode CPU (par défaut)
```
CPU idle (< seuil) pendant N secondes → Lance la tâche
CPU actif (> seuil) → Suspend la tâche (SIGSTOP)
CPU redevient idle → Reprend la tâche (SIGCONT)
Arrêt du daemon → Termine proprement la tâche (SIGTERM/SIGKILL)
```

#### Mode User Idle (nouveau)
```
Aucun événement clavier/souris pendant N secondes → Lance la tâche
Événement clavier/souris détecté → Suspend la tâche (SIGSTOP)
Inactivité retrouvée (N secondes) → Reprend la tâche (SIGCONT)
Arrêt du daemon → Termine proprement la tâche (SIGTERM/SIGKILL)
```

**Note** : En mode User Idle, le seuil CPU est ignoré. La détection utilise `CGEventSourceSecondsSinceLastEventType` (macOS Quartz Event Services) et ne nécessite aucune permission d'accessibilité.

## Fonctionnalités

- ✅ **Interface graphique (GUI)** dans la barre de menu macOS (system tray)
- ✅ **Deux modes de détection d'inactivité** : CPU-based ou User Input-based
- ✅ **Surveillance CPU en temps réel** avec seuil configurable
- ✅ **Détection d'inactivité utilisateur** (clavier/souris) sans permissions additionnelles
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
./scripts/install.sh
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
cp scripts/config.example.json ~/.config/whenidle/config.json

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
| `idle_mode` | string | `"cpu"` | Mode de détection : `"cpu"` (CPU-based) ou `"user_idle"` (clavier/souris) |
| `cpu_threshold` | float | 15.0 | % CPU global en dessous duquel le système est considéré idle (0-100) — **ignoré en mode `user_idle`** |
| `idle_duration` | int | 120 | Nombre de secondes consécutives d'idle avant de lancer la tâche |
| `check_interval` | int | 5 | Intervalle de vérification en secondes |
| `command` | string | **requis** | Chemin complet de la commande à exécuter |
| `args` | []string | [] | Arguments à passer à la commande |
| `working_dir` | string | **requis** | Répertoire de travail pour la commande |
| `log_file` | string | "" | Fichier de log optionnel (vide = stdout) |
| `restart` | bool | false | Redémarre la tâche automatiquement après sa complétion |

### Exemple de configuration

#### Mode CPU (par défaut)
```json
{
    "idle_mode": "cpu",
    "cpu_threshold": 15.0,
    "idle_duration": 120,
    "check_interval": 5,
    "command": "/usr/local/bin/heavy-task",
    "args": ["--option1", "value1"],
    "working_dir": "/Users/fred/projects/heavy-task",
    "log_file": "/tmp/whenidle.log",
    "restart": false
}
```

#### Mode User Idle (clavier/souris)
```json
{
    "idle_mode": "user_idle",
    "idle_duration": 300,
    "check_interval": 5,
    "command": "/usr/local/bin/heavy-task",
    "args": ["--option1", "value1"],
    "working_dir": "/Users/fred/projects/heavy-task",
    "log_file": "/tmp/whenidle.log",
    "restart": false
}
```

**Note** : En mode `user_idle`, le champ `cpu_threshold` est ignoré.

### Calcul de l'idle

#### Mode CPU

Le nombre de vérifications nécessaires est calculé automatiquement :

```
checks_needed = idle_duration / check_interval
```

Exemple : avec `idle_duration=120` et `check_interval=5`, la tâche se lance après **24 vérifications consécutives** à CPU < seuil (soit 2 minutes).

#### Mode User Idle

La détection est immédiate car `CGEventSourceSecondsSinceLastEventType` retourne directement les secondes écoulées depuis le dernier événement clavier/souris. Le monitor vérifie cette valeur toutes les `check_interval` secondes et lance la tâche si elle dépasse `idle_duration`.

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
3. Modifiez les paramètres dans le formulaire :
   - **Idle Mode** : Choisissez "CPU Activity" ou "User Activity"
   - **CPU Threshold** : Seuil CPU (ignoré en mode User Activity)
   - **Idle Duration** : Secondes d'inactivité avant de lancer la tâche
   - **Check Interval** : Fréquence de vérification
   - **Restart when done** : Relancer automatiquement la tâche après complétion
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

### Exemple 4 : Mode User Idle pour tâche nocturne

```json
{
    "idle_mode": "user_idle",
    "idle_duration": 600,
    "check_interval": 10,
    "command": "/usr/local/bin/backup-script",
    "args": ["--full"],
    "working_dir": "/Users/fred/backup",
    "restart": false
}
```

Lance un backup complet après 10 minutes sans activité clavier/souris (ex: la nuit). Ignore l'utilisation CPU, donc la tâche se lance même si d'autres processus consomment du CPU.

## Désinstallation

```bash
./scripts/uninstall.sh
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
| `config.go` | Chargement, sauvegarde et validation JSON (IdleMode) |
| `monitor.go` | Surveillance idle avec stratégie pluggable (CPU ou User Idle) |
| `user_idle_darwin.go` | CGo wrapper pour `CGEventSourceSecondsSinceLastEventType` |
| `user_idle_other.go` | Stub non-Darwin pour `UserIdleSeconds()` |
| `runner.go` | Gestion du processus (exec, SIGSTOP/SIGCONT/SIGTERM) |
| `*_test.go` | Tests unitaires (33 tests) |

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
├── config.go                 # Config JSON (load/save, idle_mode)
├── config_test.go            # Tests config
├── monitor.go                # Surveillance idle (stratégie pluggable)
├── monitor_test.go           # Tests monitor (CPU + User Idle)
├── user_idle_darwin.go       # CGo CoreGraphics (User Idle)
├── user_idle_other.go        # Stub non-Darwin
├── runner.go                 # Gestion processus
├── runner_test.go            # Tests runner
├── scripts/
│   ├── install.sh                # Installation + .app bundle
│   ├── uninstall.sh              # Désinstallation
│   ├── com.whenidle.agent.plist  # Template Launch Agent
│   └── config.example.json       # Config par défaut
└── README.md                 # Cette documentation
```

### Tests

Le projet inclut 33 tests unitaires couvrant :
- Validation de la configuration (save/load round-trip, idle_mode)
- LogBuffer (ring buffer, thread safety, onChange callbacks)
- Machine à états du runner (Stopped → Running → Paused)
- Logique de détection idle/busy du monitor (CPU et User Idle)
- Signaling SIGSTOP/SIGCONT/SIGTERM
- Mode User Idle (détection et pause)

Couverture : ~43%

### Dépendances

- `github.com/shirou/gopsutil/v4` — Monitoring CPU cross-platform
- `fyne.io/fyne/v2` — Framework GUI et system tray (v2.7.2)
- **CGo** — Requis pour Cocoa (Dock), CoreGraphics (User Idle), et Fyne

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

Le bundle `.app` doit avoir `LSUIElement=true` dans `Info.plist`. Ré-exécutez `./scripts/install.sh`.

## License

MIT
