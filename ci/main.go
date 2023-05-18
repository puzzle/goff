package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
)

func main() {

	ReleaseOnGitHub("v0.1.0")
	return
	// create Dagger client
	ctx := context.Background()
	daggerClient, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer daggerClient.Close()

	// get working directory on host
	modDir := daggerClient.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{"go.mod", "go.sum"},
	})

	// get working directory on host
	source := daggerClient.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"ci"},
	})

	// build application
	goMod := daggerClient.CacheVolume("go")
	base := daggerClient.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("golang:1.20").
		WithMountedCache("/go/src", goMod)

	base = base.
		WithWorkdir("/src").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "musl-tools", "-y"})

	golang := base.WithDirectory("/src", modDir).
		WithExec([]string{"go", "mod", "download"})

	golang = golang.
		WithDirectory("/src", source).
		WithExec([]string{"mkdir", "-p", "/app"}).
		WithEnvVariable("CC", "musl-gcc").
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

	refType := os.Getenv("GITHUB_REF_TYPE")
	refName := os.Getenv("GITHUB_REF_NAME")

	if refType == "tag" && strings.HasPrefix(refName, "v") {
		buildAndRelease(golang, refName)
	}
}

func buildAndRelease(golang *dagger.Container, version string) {

	targets := make(map[string][]string)
	targets["linux"] = []string{"amd64", "386", "arm"}
	targets["windows"] = []string{"amd64", "386"}
	targets["darwin"] = []string{"amd64"}

	for os, target := range targets {
		for i := range target {
			arch := target[i]
			outFile := fmt.Sprintf("./build/goff-%s-%s-%s", os, arch, version)
			golang = golang.
				WithEnvVariable("GOOS", os).
				WithEnvVariable("GOARCH", arch).
				WithExec([]string{"go", "build", "-o", outFile, "goff"})
		}
	}

	_, err := golang.Directory("build/").Export(context.Background(), "./build")
	if err != nil {
		panic(err)
	}
}

func ReleaseOnGitHub(tag string) {

	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if accessToken == "" {
		panic("GITHUB_ACCESS_TOKEN env var is missing")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	releaseName := fmt.Sprintf("GOFF %s", tag)

	release := &github.RepositoryRelease{
		TagName: &tag,
		Name:    &releaseName,
	}

	_, _, err := client.Repositories.CreateRelease(ctx, "schlapzz", "goff", release)

	files, err := ioutil.ReadDir("build/")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		fmt.Println("upload file: " + f.Name())
		file, err := os.Open(filepath.Join("build/", f.Name()))
		if err != nil {
			panic(err)
		}
		_, _, err = client.Repositories.UploadReleaseAsset(ctx, "schlapzz", "goff", 0, &github.UploadOptions{
			Name:  f.Name(),
			Label: "release",
		}, file)
		if err != nil {
			panic(err)
		}
	}

}
