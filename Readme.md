[![ci](https://github.com/puzzle/goff/actions/workflows/main.yml/badge.svg)](https://github.com/puzzle/goff/actions/workflows/main.yml)

# GOFF

Inspired from Kostis Kapelonis (Codefresh.io) talk at the KubeCon about [How to Preview and Diff Your Argo CD Deployments](https://youtu.be/X392bJX0AEs) we relased our own GitOps Diff tool (Goff). This tool helps you to preview your changes in your GitOps Repository.

## How it works

[Checkout the examples](doc/)

### Kustomize example

Example for Kustomization diff
```bash
#Build base and all overlays from source branch
goff kustomize build ./source/kustomize --output-dir /tmp/source/out
#Build base and all overlays from target branch
goff kustomize build ./target/kustomize --output-dir /tmp/target/out
#Diff rendered manifests
goff diff "/tmp/source" "/tmp/target" --title=Preview --output-dir .
```

1. Create a new branch and commit your changes in your Kustomize deployment
 ![GitHub Diff](docs/img/github-diff.png)
2. Run your pipeline, Goff renders the Base and the Overlays and calculate the diff between the source and target branch.
3. Check the auto generated comment in your Pull request and review the changes
 ![GitHub Diff](docs/img/goff-diff.png)

### ArgoCD Application

## Usage

```bash
GitOps Diff Tool

Usage:
  goff [command]

Available Commands:
  argocd      Render manifests from ArgoCD resources
  completion  Generate the autocompletion script for the specified shell
  diff        Diff files [sourceDir] [targetDir]
  help        Help about any command
  kustomize   Generate a DOT file to visualize the dependencies between your kustomize components
  split       Split manifests [manifestFile]

Flags:
  -h, --help              help for goff
  -l, --logLevel string   Set loglevel [debug, info, error] (default "error")

Use "goff [command] --help" for more information about a command.
```

## Supported Tools

| Tooling               | Support                                       |
|-----------------------|----------------------------------------------|
| Plain manifests       | ✅                                          |
| Helm                  | ✅ Supported through plain manifests        |
| Kustomize             | ✅                                          |
| ArgoCD Application    | ✅ Needs a local ArgoCD Repo server instance             |
| ArgoCD ApplicationSet |  🚧 Not yet fully supported (List generators only)                |

## Documentation

here
