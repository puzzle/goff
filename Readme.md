GOFF
===

The GitOps Diff tool. Review your changes in deep.

# Usage

```bash
Helper tool to show changes between .....

Usage:
  goff [command]

Available Commands:
  argocd      Render manifests from ArgoCD Application
  completion  Generate the autocompletion script for the specified shell
  diff        Diff files [sourceDir] [targetDir]
  help        Help about any command
  kustomize   Generate a DOT file to visualize the dependencies betw

Flags:
  -h, --help     help for goff
  
  -t, --toggle   Help message for toggle

Use "goff [command] --help" for more information about a command.
```

# Build binary from source
```bash
go build -o goff goff 
```


# Build Image with dagger
```bash
export REGISTRY_PASSWORD=....
export REGISTRY_USER=....
go run ci/main.go 
```

If you wanna try the new fancy Dagger TUI
```bash
export _EXPERIMENTAL_DAGGER_TUI=1
export REGISTRY_PASSWORD=....
export REGISTRY_USER=....
dagger run go run ci/main.go
```
