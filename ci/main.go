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
	golang := daggerClient.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("golang:1.20")

	golang = golang.WithDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"mkdir", "-p", "/app"}).
		WithExec([]string{"go", "build", "-o", "/app/goff", "goff"}).
		WithExec([]string{"go", "install", "gitlab.com/gitlab-org/cli/cmd/glab@main"})

	goffBin := golang.File("/app/goff")
	glabBin := golang.File("/go/bin/glab")

	goofContainer := daggerClient.Container().From("registry.puzzle.ch/cicd/alpine-base").
		WithFile("/bin/goff", goffBin).
		WithFile("/bin/glab", glabBin).
		WithEntrypoint([]string{"/bin/goff"})

	secret := daggerClient.SetSecret("reg-secret", os.Getenv("REGISTRY_PASSWORD"))

	addr, err := goofContainer.WithRegistryAuth("registry.puzzle.ch", "cschlatter", secret).Publish(ctx, "registry.puzzle.ch/cicd/goff")
	if err != nil {
		panic(err)
	}

	// print ref
	fmt.Println("Published at:", addr)
}
