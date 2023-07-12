# Build

### Build binary from source

```bash
go build -o goff github.com/puzzle/goff 
```

### Build Image with dagger

First set the required env vars.
For releasing a new version, GITHUB_REF_TYPE=tag and a valid version must be set GITHUB_REF_NAME=vX.X.X
```bash
export GITHUB_REF_TYPE=<branch|tag>
export GITHUB_REF_NAME=test
export GITHUB_ACCESS_TOKEN=<Github Access Token>
export REGISTRY_URL=ttl.sh
export REGISTRY_USER=bar
export REGISTRY_PASSWORD=foo
```

```bash
go run ci/main.go 
```

or 

```bash
dagger run go run ci/main.go 
```

