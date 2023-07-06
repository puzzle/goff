![ci](img/goff-logo.png)



[![ci](https://github.com/puzzle/goff/actions/workflows/main.yml/badge.svg)](https://github.com/puzzle/goff/actions/workflows/main.yml)


# GOFF #{{ goff.version }}

Inspired from Kostis Kapelonis (Codefresh.io) talk at the KubeCon about [How to Preview and Diff Your Argo CD Deployments](https://youtu.be/X392bJX0AEs) we relased our own GitOps Diff tool (Goff). This tool helps you to preview your changes in your GitOps Repository.

## How it works


Example for ArgoCD Application diff
```bash
#Render all ArgoCD manifests in directory from source branch
goff argocd app "./source/argocd" --repo-server="repo-server:8081" --output-dir=/tmp/source/
#Render all ArgoCD manifests in directory from target branch
goff argocd app "./target/argocd" --repo-server="repo-server:8081" --output-dir=/tmp/target/
#Diff rendered Kubernetes manifests
goff diff "/tmp/source" "/tmp/target" --output-dir .
```

1. Setup your pipeline in your GitOps repository. You can find examples integrations [for Github, Gitlab and Gitea here](integrations.md)
2. Create a new branch and commit your changes in your ArgoCd Application
 ![GitHub Diff](img/github-argo-diff.png)
3. Run your pipeline, Goff renders the Appication into manifests calculate the diff between the source and target branch.
4. Check the auto generated comment in your Pull request and review the changes
 ![GitHub Diff](img/goff-argo-diff.png)

## Installation

You can download the latest release here
Or you can use the pre built Docker image `docker pull quay.io/puzzle/goff:#{{ goff.version }}`

The image includes following tools:

* goff #{{ goff.version }}
* Gitlab CLI
* Gitea CLI
* Helm
* Git

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
