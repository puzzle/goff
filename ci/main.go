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
		WithExec([]string{"go", "build", "-o", "/app/goff", "goff"})

	goffBin := golang.File("/app/goff")

	templates := daggerClient.Host().Directory("templates")

	goofContainer := daggerClient.Container().From("registry.puzzle.ch/cicd/ubi9-base").
		WithFile("/app/goff", goffBin).
		WithDirectory("/app/templates", templates).
		WithEntrypoint([]string{"/app/goff"})

	secret := daggerClient.SetSecret("gh-secret", os.Getenv("REGISTRY_PASSWORD"))

	addr, err := goofContainer.WithRegistryAuth("registry.puzzle.ch", "cschlatter", secret).Publish(ctx, "registry.puzzle.ch/cicd/goff")
	if err != nil {
		panic(err)
	}

	// print ref
	fmt.Println("Published at:", addr)
}
