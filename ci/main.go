package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
	"golang.org/x/sync/errgroup"
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

	refType := os.Getenv("GITHUB_REF_TYPE")
	refName := os.Getenv("GITHUB_REF_NAME")

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
	golang = golang.WithDirectory("/src", source)

	errGroup, _ := errgroup.WithContext(ctx)

	errGroup.Go(func() error {

		repoServer := daggerClient.Container().From("quay.io/puzzle/argocd-repo-server:latest").
			WithExposedPort(8081).
			WithEntrypoint([]string{"argocd-repo-server"}).WithExec(nil)

		_, err := golang.
			WithServiceBinding("reposerver", repoServer).
			WithExec([]string{"go", "test", "./...", "-v"}).Stdout(ctx)
		return err
	})

	errGroup.Go(func() error {
		goffGitlab, err := golang.
			WithExec([]string{"mkdir", "-p", "/app"}).
			WithEnvVariable("CC", "musl-gcc").
			WithExec([]string{"go", "build", "-o", "/app/goff", "github.com/puzzle/goff"}).
			WithExec([]string{"go", "install", "gitlab.com/gitlab-org/cli/cmd/glab@main"}).
			Sync(ctx)

		if err != nil {
			return err
		}

		//Add GOFF and Gitlab CLI to our standard build container
		//Push into registry
		if refName == "main" || isReleaseTag(refName, refType) {
			tag := "latest"
			if isReleaseTag(refName, refType) {
				tag = refName
			}
			err = publishImage(ctx, goffGitlab, daggerClient, regUser, secret, tag)
		}

		return err

	})

	err = errGroup.Wait()
	if err != nil {
		panic(err)
	}

	//If version tag, build binary releases and release them on github
	if isReleaseTag(refName, refType) {
		buildAndRelease(ctx, daggerClient, golang, refName)
		//Build patched ArgoCD Repo server
		buildArgoCdRepoServer(ctx, regUser, secret, daggerClient)
	}

}

func publishImage(ctx context.Context, golang *dagger.Container, daggerClient *dagger.Client, regUser string, secret *dagger.Secret, tag string) error {
	goffBin := golang.File("/app/goff")
	glabBin := golang.File("/go/bin/glab")

	goffContainer := daggerClient.Container().From("docker.io/alpine:3.18").
		WithFile("/bin/goff", goffBin).
		WithFile("/bin/glab", glabBin).
		WithExec([]string{"apk", "add", "git", "helm"})

	goffContainer = goffContainer.WithEntrypoint([]string{"/bin/goff"})

	imageName := fmt.Sprintf("quay.io/puzzle/goff:%s", tag)
	_, err := goffContainer.WithRegistryAuth("quay.io", regUser, secret).Publish(ctx, imageName)

	return err
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

func buildAndRelease(ctx context.Context, client *dagger.Client, golang *dagger.Container, version string) {

	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if accessToken == "" {
		panic("GITHUB_ACCESS_TOKEN env var is missing")
	}

	targets := make(map[string][]string)
	targets["linux"] = []string{"amd64", "386", "arm"}
	targets["windows"] = []string{"amd64", "386"}
	targets["darwin"] = []string{"amd64"}

	files := make([]string, 0)

	errGroup, ctx := errgroup.WithContext(ctx)

	for os, target := range targets {
		for i := range target {
			arch := target[i]
			oss := os
			errGroup.Go(func() error {
				var buildErr error
				outFile := fmt.Sprintf("./build/goff-%s-%s-%s", oss, arch, version)
				_, buildErr = golang.
					WithEnvVariable("GOOS", oss).
					WithEnvVariable("GOARCH", arch).
					WithEnvVariable("CC", "").
					WithExec([]string{"go", "build", "-o", outFile, "github.com/puzzle/goff"}).
					File(outFile).Export(ctx, outFile)

				if buildErr != nil {
					return buildErr
				}

				files = append(files, outFile)

				return nil
			})
		}
	}

	err := errGroup.Wait()
	if err != nil {
		panic(err)
	}

	buildDir := client.Host().Directory("build/")
	ghContainer := client.Container().From("ghcr.io/supportpal/github-gh-cli").
		WithEnvVariable("GITHUB_TOKEN", accessToken).
		WithDirectory("/build", buildDir).
		WithExec([]string{"gh", "-R", "puzzle/goff", "release", "create", version})

	for _, f := range files {
		ghContainer = ghContainer.
			WithExec([]string{"gh", "-R", "puzzle/goff", "release", "upload", version, f})
	}

	//Evaluate
	_, err = ghContainer.Sync(context.Background())
	if err != nil {
		panic(err)
	}
}

func isReleaseTag(refName, refType string) bool {
	return refType == "tag" && strings.HasPrefix(refName, "v")
}
