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

- ✅ **Surveillance CPU en temps réel** avec seuil configurable
- ✅ **Suspension/reprise automatique** via SIGSTOP/SIGCONT
- ✅ **Process group signaling** pour gérer toute l'arborescence de processus
- ✅ **Launch Agent macOS** pour démarrage automatique au login
- ✅ **Configuration JSON** simple et flexible
- ✅ **Logging configurable** (stdout ou fichier)
- ✅ **Graceful shutdown** avec timeout SIGKILL
- ✅ **Binary standalone** sans dépendances runtime (3.4 MB)

## Installation

### Prérequis

- macOS 10.14+ (Darwin)
- Go 1.18+ (pour compiler depuis les sources)

### Installation automatique

```bash
cd /path/to/WhenIdle
./install.sh
```

Cette commande :
1. Compile le binary `whenidle`
2. L'installe dans `~/.local/bin/`
3. Crée une configuration par défaut dans `~/.config/whenidle/config.json`
4. Configure le Launch Agent dans `~/Library/LaunchAgents/`

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

### Démarrer le daemon

```bash
launchctl load ~/Library/LaunchAgents/com.whenidle.agent.plist
```

### Arrêter le daemon

```bash
launchctl unload ~/Library/LaunchAgents/com.whenidle.agent.plist
```

### Redémarrer le daemon

```bash
launchctl unload ~/Library/LaunchAgents/com.whenidle.agent.plist
launchctl load ~/Library/LaunchAgents/com.whenidle.agent.plist
```

### Vérifier le statut

```bash
launchctl list | grep whenidle
```

### Consulter les logs

Les logs sont écrits dans :
- **Stdout du daemon** : `/tmp/whenidle.stdout.log`
- **Stderr du daemon** : `/tmp/whenidle.stderr.log`
- **Log de la tâche** : Configuré via `log_file` (ou mélangé avec stdout)

```bash
# Logs du daemon
tail -f /tmp/whenidle.stdout.log

# Logs en temps réel
log stream --predicate 'processImagePath contains "whenidle"'
```

### Exécution manuelle (sans Launch Agent)

```bash
~/.local/bin/whenidle --config ~/.config/whenidle/config.json
```

Arrêt : `Ctrl+C`

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

```
┌─────────────────────────────────────────────┐
│              WhenIdle                        │
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
| `main.go` | Point d'entrée, orchestration, signal handling |
| `config.go` | Chargement et validation de la configuration |
| `monitor.go` | Surveillance CPU avec boucle de polling (gopsutil) |
| `runner.go` | Gestion du processus (exec, SIGSTOP/SIGCONT/SIGTERM) |
| `*_test.go` | Tests unitaires |

## Développement

### Prérequis

- Go 1.18+
- macOS (pour tester le Launch Agent)

### Compiler

```bash
go build -o whenidle .
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
├── main.go              # Orchestration principale
├── config.go            # Configuration JSON
├── monitor.go           # Surveillance CPU
├── runner.go            # Gestion du processus
├── config_test.go       # Tests de configuration
├── monitor_test.go      # Tests du monitor
├── runner_test.go       # Tests du runner
├── config.example.json  # Exemple de configuration
├── install.sh           # Script d'installation
├── uninstall.sh         # Script de désinstallation
├── com.whenidle.agent.plist  # Template Launch Agent
└── README.md            # Cette documentation
```

### Tests

Le projet inclut 13 tests unitaires couvrant :
- Validation de la configuration
- Machine à états du runner (Stopped → Running → Paused)
- Logique de détection idle/busy du monitor
- Signaling SIGSTOP/SIGCONT/SIGTERM

Couverture : ~60%

### Dépendances

Une seule dépendance externe :
- `github.com/shirou/gopsutil/v4` — Monitoring CPU cross-platform

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

## License

MIT
