package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
	// create Dagger client
	ctx := context.Background()
	daggerClient, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer daggerClient.Close()

	// get working directory on host
	source := daggerClient.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"ci"},
	})

	// build application
	goMod := daggerClient.CacheVolume("go")
	golang := daggerClient.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("golang:1.20").
		WithMountedCache("/go/src", goMod)

	golang = golang.WithDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "musl-tools", "-y"}).
		WithEnvVariable("CC", "musl-gcc").
		WithExec([]string{"go", "install"}).
		WithExec([]string{"mkdir", "-p", "/app"}).
		WithExec([]string{"go", "test", "./...", "-v"}).
		WithExec([]string{"go", "build", "-o", "/app/goff", "goff"}).
		WithExec([]string{"go", "install", "gitlab.com/gitlab-org/cli/cmd/glab@main"})

	goffBin := golang.File("/app/goff")
	glabBin := golang.File("/go/bin/glab")

	goffContainer := daggerClient.Container().From("registry.puzzle.ch/cicd/alpine-base").
		WithFile("/bin/goff", goffBin).
		WithFile("/bin/glab", glabBin).
		WithEntrypoint([]string{"/bin/goff"})

	secret := daggerClient.SetSecret("reg-secret", os.Getenv("REGISTRY_PASSWORD"))

	regUser, ok := os.LookupEnv("REGISTRY_USER")
	if !ok {
		panic(fmt.Errorf("Env var REGISTRY_USER not set"))
	}

	_, err = goffContainer.WithRegistryAuth("registry.puzzle.ch", regUser, secret).Publish(ctx, "registry.puzzle.ch/cicd/goff")
	if err != nil {
		panic(err)
	}

	//Build repo server for GitHub actions becuase they don't yet support overriding the entrypoint
	repoServerContainer := daggerClient.Container().From("quay.io/argoproj/argocd:latest").
		WithUser("root").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "netcat", "-y"}).
		WithUser("argocd").
		WithEntrypoint([]string{"argocd-repo-server"})

	_, err = repoServerContainer.WithRegistryAuth("registry.puzzle.ch", regUser, secret).Publish(ctx, "registry.puzzle.ch/cicd/argocd-repo-server")
	if err != nil {
		panic(err)
	}

}
