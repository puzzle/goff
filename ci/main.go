package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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

	//Get Registry password as secret
	secret := daggerClient.SetSecret("reg-secret", os.Getenv("REGISTRY_PASSWORD"))

	regUser, ok := os.LookupEnv("REGISTRY_USER")
	if !ok {
		panic(fmt.Errorf("Env var REGISTRY_USER not set"))
	}

	// create golang base
	goMod := daggerClient.CacheVolume("go")
	base := daggerClient.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("golang:1.20").
		WithMountedCache("/go/src", goMod)

	//install musl for alpine builds on base image
	base = base.
		WithWorkdir("/src").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "musl-tools", "-y"})

	// get working directory on host
	modDir := daggerClient.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{"go.mod", "go.sum"},
	})

	//download go modules
	golang := base.WithDirectory("/src", modDir).
		WithExec([]string{"go", "mod", "download"})

	// get working directory on host
	source := daggerClient.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"ci/", "build/"},
	})

	//test and build
	golang = golang.
		WithDirectory("/src", source).
		WithExec([]string{"mkdir", "-p", "/app"}).
		WithEnvVariable("CC", "musl-gcc").
		WithExec([]string{"go", "test", "./...", "-v"}).
		WithExec([]string{"go", "build", "-o", "/app/goff", "goff"}).
		WithExec([]string{"go", "install", "gitlab.com/gitlab-org/cli/cmd/glab@main"}) //download gitlab cli

	goffBin := golang.File("/app/goff")
	glabBin := golang.File("/go/bin/glab")

	//Add GOFF and Gitlab CLI to our standard build container
	goffContainer := daggerClient.Container().From("docker.io/alpine:3.18").
		WithFile("/bin/goff", goffBin).
		WithFile("/bin/glab", glabBin).
		WithEntrypoint([]string{"/bin/goff"})

		//Push into registry
	_, err = goffContainer.WithRegistryAuth("quay.io", regUser, secret).Publish(ctx, "quay.io/puzzle/goff")
	if err != nil {
		panic(err)
	}

	refType := os.Getenv("GITHUB_REF_TYPE")
	refName := os.Getenv("GITHUB_REF_NAME")

	//If version tag, build binary releases and release them on github
	if refType == "tag" && strings.HasPrefix(refName, "v") {
		buildAndRelease(daggerClient, golang, refName)
	}

	if refName == "main" {
		//Build patched ArgoCD Repo server
		buildArgoCdRepoServer(ctx, regUser, secret, daggerClient)
	}

}

//Build repo server for GitHub actions becuase they don't yet support overriding the entrypoint
func buildArgoCdRepoServer(ctx context.Context, regUser string, regSecret *dagger.Secret, client *dagger.Client) {

	repoServerContainer := client.Container().From("quay.io/argoproj/argocd:latest").
		WithUser("root").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "netcat", "-y"}).
		WithUser("argocd").
		WithEntrypoint([]string{"argocd-repo-server"})

	_, err := repoServerContainer.WithRegistryAuth("quay.io", regUser, regSecret).Publish(ctx, "quay.io/puzzle/argocd-repo-server")
	if err != nil {
		panic(err)
	}

}

func buildAndRelease(client *dagger.Client, golang *dagger.Container, version string) {

	targets := make(map[string][]string)
	targets["linux"] = []string{"amd64", "386", "arm"}
	targets["windows"] = []string{"amd64", "386"}
	targets["darwin"] = []string{"amd64"}

	files := make([]string, 0)

	for os, target := range targets {
		for i := range target {
			arch := target[i]
			outFile := fmt.Sprintf("./build/goff-%s-%s-%s", os, arch, version)
			golang = golang.
				WithEnvVariable("GOOS", os).
				WithEnvVariable("GOARCH", arch).
				WithEnvVariable("CC", "").
				WithExec([]string{"go", "build", "-o", outFile, "goff"})
			files = append(files, outFile)
		}
	}

	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if accessToken == "" {
		panic("GITHUB_ACCESS_TOKEN env var is missing")
	}

	_, err := golang.Directory("build/").Export(context.Background(), "./build")
	if err != nil {
		panic(err)
	}

	ghContainer := client.Container().From("ghcr.io/supportpal/github-gh-cli").
		WithEnvVariable("GITHUB_TOKEN", accessToken).
		WithDirectory("/build", golang.Directory("build/")).
		WithExec([]string{"gh", "-R", "schlapzz/goff", "release", "create", version})

	for _, f := range files {
		ghContainer = ghContainer.
			WithExec([]string{"gh", "-R", "schlapzz/goff", "release", "upload", version, f})
	}

	//Evaluate
	_, err = ghContainer.Sync(context.Background())
	if err != nil {
		panic(err)
	}
}
