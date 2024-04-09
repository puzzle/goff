package main

import "context"

const (
	goVersion       = "1.22"
	alpineBaseImage = "docker.io/alpine:3.19.1"
	kustomizeImage  = "registry.k8s.io/kustomize/kustomize:v5.4.1"
	redisImage      = "redis"
	argocdImage     = "quay.io/argoproj/argocd:latest"
)

type CI struct {
	// Project source directory.
	// +private
	Source *Directory
}

func New(
	// Project source directory.
	source *Directory,
) *CI {
	return &CI{
		Source: source,
	}
}

func (ci *CI) Test(ctx context.Context) error {
	// Sadly integration tests do not work at the moment :(
	// https://github.com/dagger/dagger/issues/6951
	_, err := golangBaseImage(ci.Source).
		WithExec([]string{"go", "test", "-v", "./..."}).
		WithServiceBinding("reposerver", ci.RepoServer()).
		Sync(ctx)
	return err
}

func (ci *CI) RepoServer() *Service {
	redis := dag.Container().
		From(redisImage).
		WithExposedPort(6379).
		AsService()

	return dag.Container().
		From(argocdImage).
		WithServiceBinding("redis", redis).
		WithDefaultArgs([]string{"--redis", "redis:6379"}).
		WithExposedPort(8081).
		WithEntrypoint([]string{"argocd-repo-server"}).
		AsService()
}
