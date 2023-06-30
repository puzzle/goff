# Examples


## Helm

```bash
#Clone source branch
git clone --branch feature/A git@github.com:schlapzz/goff-github.git ./source/

#Clone target branch
git clone --branch main git@github.com:schlapzz/goff-github.git ./target/

#Render source chart
helm template mychart ./source/helm/mychart --output-dir /tmp/source/out
#Render target Chart
helm template mychart ./target/helm/mychart --output-dir /tmp/target/out

#Diff rendered manifests
goff diff "/tmp/source" "/tmp/target" --output-dir .
```

## Kustomize

Example for Kustomization diff
```bash
#Clone source branch
git clone --branch feature/A git@github.com:schlapzz/goff-github.git ./source/

#Clone target branch
git clone --branch main git@github.com:schlapzz/goff-github.git ./target/

#Build base and all overlays from source branch
goff kustomize build ./source/kustomize --output-dir /tmp/source/out
#Build base and all overlays from target branch
goff kustomize build ./target/kustomize --output-dir /tmp/target/out

#Diff rendered manifests
goff diff "/tmp/source" "/tmp/target" --title=Preview --output-dir .
```

## ArgoCD


```bash
goff argocd -h
Render manifests from ArgoCD resources

Usage:
  goff argocd [flags]
  goff argocd [command]

Available Commands:
  app         Render manifests from ArgoCD Application
  appset      Render ArgoCD Applications manifests from ApplicationSets

Flags:
  -h, --help                help for argocd
  -o, --output-dir string   Output directory (default ".")
  -p, --password string     Repo password
  -i, --ssh-key string      Repo SSH Key
  -u, --username string     Repo username

Global Flags:
  -l, --log-level string   Set loglevel [debug, info, error] (default "error")

Use "goff argocd [command] --help" for more information about a command.
```
### ArgoCD Application
Example for ArgoCD Application diff
```bash
#Clone source branch
git clone --branch feature/A git@github.com:schlapzz/goff-github.git ./source/

#Clone target branch
git clone --branch main git@github.com:schlapzz/goff-github.git ./target/

#Start ArgoCD repo server to render Applications
docker run -d -p 8081:8081 quay.io/argoproj/argocd:latest

#Render all ArgoCD manifests in directory from source branch
goff argocd app "./source/argocd/app" --repo-server="localhost:8081" --output-dir=/tmp/source/
#Render all ArgoCD manifests in directory from target branch
goff argocd app "./target/argocd/app" --repo-server="localhost:8081" --output-dir=/tmp/target/
#Diff rendered Kubernetes manifests
goff diff "/tmp/source" "/tmp/target" --output-dir .
```


### ArgoCD ApplicationSet

Example for ArgoCD ApplicationSet diff
```bash
#Clone source branch
git clone --branch feature/A git@github.com:schlapzz/goff-github.git ./source/

#Clone target branch
git clone --branch main git@github.com:schlapzz/goff-github.git ./target/

#Render all ArgoCD manifests in directory from source branch
goff argocd appset "./source/argocd/app-set" --repo-server="localhost:8081" --output-dir=/tmp/source/
#Render all ArgoCD manifests in directory from target branch
goff argocd appset "./target/argocd/app-set" --repo-server="localhost:8081" --output-dir=/tmp/target/
#Diff rendered Kubernetes manifests
goff diff "/tmp/source" "/tmp/target" --output-dir .
```

