# kubectl-tekton-imagebuild

Ein kubectl-Plugin zum Auflisten von Tekton PipelineRuns mit dem Label `imagebuild.ba.de/imagebuildschedule`.

## Installation

### Voraussetzungen

- Go 1.21 oder höher
- kubectl installiert und konfiguriert
- Zugriff auf einen Kubernetes-Cluster mit Tekton Pipelines

### Build und Installation

1. Repository klonen:
```bash
git clone https://github.com/kraema/kubectl-plugins.git
cd kubectl-plugins/kubectl-tekton-imagebuild
```

2. Dependencies installieren:
```bash
go mod download
```

3. Plugin bauen:
```bash
go build -o kubectl-tekton-imagebuild .
```

4. Plugin in PATH installieren:
```bash
# Option 1: In /usr/local/bin installieren (systemweit)
sudo cp kubectl-tekton-imagebuild /usr/local/bin/

# Option 2: In ~/bin installieren (Benutzer-spezifisch)
mkdir -p ~/bin
cp kubectl-tekton-imagebuild ~/bin/
# Sicherstellen, dass ~/bin im PATH ist
export PATH=$PATH:~/bin
```

## Verwendung

### Grundlegende Befehle

```bash
# Alle PipelineRuns mit dem imagebuild.ba.de/imagebuildschedule Label im aktuellen Namespace
kubectl tekton-imagebuild

# PipelineRuns mit spezifischem Schedule-Wert
kubectl tekton-imagebuild --schedule=daily
kubectl tekton-imagebuild -s weekly

# PipelineRuns aus allen Namespaces
kubectl tekton-imagebuild --all-namespaces
kubectl tekton-imagebuild -A

# PipelineRuns in spezifischem Namespace
kubectl tekton-imagebuild --namespace=production
kubectl tekton-imagebuild -n production

# Kombinierte Filter
kubectl tekton-imagebuild -n production --schedule=daily
```

### Ausgabeformate

```bash
# Standard Tabellenformat
kubectl tekton-imagebuild

# Erweiterte Tabellenansicht mit mehr Details
kubectl tekton-imagebuild -o wide

# JSON-Ausgabe
kubectl tekton-imagebuild -o json

# YAML-Ausgabe
kubectl tekton-imagebuild -o yaml

# Tabelle ohne Header (für Scripting)
kubectl tekton-imagebuild --no-headers
```

### Hilfe

```bash
# Hilfe anzeigen
kubectl tekton-imagebuild --help
kubectl tekton-imagebuild -h
```

## Beispielausgabe

### Standard Tabellenformat
```
NAME                          NAMESPACE    STATUS      SCHEDULE    AGE
imagebuild-app1-xyz123       production   Succeeded   daily       2h
imagebuild-app2-abc456       production   Running     weekly      1d
imagebuild-app3-def789       staging      Failed      hourly      30m
```

### Erweiterte Ansicht (-o wide)
```
NAME                    NAMESPACE    PIPELINE           STATUS      SCHEDULE    START TIME                   COMPLETION TIME              AGE
imagebuild-app1-xyz123  production   build-pipeline     Succeeded   daily       2024-01-15T10:00:00Z        2024-01-15T10:15:00Z        2h
imagebuild-app2-abc456  production   deploy-pipeline    Running     weekly      2024-01-14T08:00:00Z        -                           1d
```

## Label-Schema

Das Plugin filtert PipelineRuns basierend auf dem Label:
- **Label Key**: `imagebuild.ba.de/imagebuildschedule`
- **Label Value**: Beliebiger String (z.B. `daily`, `weekly`, `hourly`, `manual`)

### Beispiel PipelineRun mit Label

```yaml
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  name: imagebuild-myapp-123
  labels:
    imagebuild.ba.de/imagebuildschedule: "daily"
spec:
  pipelineRef:
    name: build-pipeline
  # ... weitere Konfiguration
```

## Entwicklung

### Tests ausführen

```bash
# Alle Tests ausführen
go test ./...

# Tests mit Coverage
go test -cover ./...

# Verbose Output
go test -v ./...
```

### Code formatieren

```bash
go fmt ./...
```

### Code-Qualität prüfen

```bash
go vet ./...
```

## Fehlerbehebung

### Plugin wird nicht gefunden

Stellen Sie sicher, dass:
1. Das Plugin ausführbar ist: `chmod +x kubectl-tekton-imagebuild`
2. Das Plugin im PATH ist: `which kubectl-tekton-imagebuild`
3. Der Name mit `kubectl-` beginnt

### Keine PipelineRuns werden angezeigt

Überprüfen Sie:
1. Ob PipelineRuns mit dem entsprechenden Label existieren:
   ```bash
   kubectl get pipelineruns -A -l imagebuild.ba.de/imagebuildschedule
   ```
2. Ob Sie den richtigen Namespace verwenden
3. Ob Sie die nötigen Berechtigungen haben:
   ```bash
   kubectl auth can-i list pipelineruns.tekton.dev
   ```

### Verbindungsprobleme

Wenn das Plugin sich nicht mit dem Cluster verbinden kann:
1. Überprüfen Sie Ihre kubeconfig: `kubectl config current-context`
2. Testen Sie die Verbindung: `kubectl get nodes`
3. Nutzen Sie explizit eine kubeconfig: `kubectl tekton-imagebuild --kubeconfig=/path/to/config`

## Lizenz

MIT