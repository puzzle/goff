package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
	"golang.org/x/sync/errgroup"
)

type GoffPipeline struct {
	GithubAccessToken string
	RefType           string
	RefName           string
	RegistryUser      string
	RegistrySecret    string
	RegistryUrl       string
	Release           Releaser
}

type Releaser interface {
	releaseFiles(client *dagger.Client, version string, files []string)
}

func main() {
	gp := NewFromGithub()
	gp.run()
}

func (g *GoffPipeline) run() {

	// create Dagger client
	ctx := context.Background()
	daggerClient, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer daggerClient.Close()

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
		if g.RefName == "main" || g.isReleaseTag() {

			err = g.publishImage(ctx, goffGitlab, daggerClient)
		}

		return err

	})

	err = errGroup.Wait()
	if err != nil {
		panic(err)
	}

	//If version tag, build binary releases and release them on github
	if g.isReleaseTag() {
		files := g.buildAndRelease(ctx, daggerClient, golang)
		var r Releaser = g
		r.releaseFiles(daggerClient, g.RefName, files)
		//Build patched ArgoCD Repo server
		g.buildArgoCdRepoServer(ctx, daggerClient)
	}
}

func NewFromGithub() GoffPipeline {
	return GoffPipeline{
		RefType:        mustLoadEnv("GITHUB_REF_TYPE"),
		RefName:        mustLoadEnv("GITHUB_REF_NAME"),
		RegistrySecret: mustLoadEnv("REGISTRY_PASSWORD"),
		RegistryUser:   mustLoadEnv("REGISTRY_USER"),
		RegistryUrl:    mustLoadEnv("REGISTRY_URL"),
	}
}

func (g *GoffPipeline) publishImage(ctx context.Context, golang *dagger.Container, daggerClient *dagger.Client) error {
	goffBin := golang.File("/app/goff")
	glabBin := golang.File("/go/bin/glab")

	goffContainer := daggerClient.Container().From("docker.io/alpine:3.18").
		WithFile("/bin/goff", goffBin).
		WithFile("/bin/glab", glabBin).
		WithExec([]string{"apk", "add", "git", "helm"})

	goffContainer = goffContainer.WithEntrypoint([]string{"/bin/goff"})

	tag := "latest"
	if g.isReleaseTag() {
		tag = g.RefName
	}
	imageName := fmt.Sprintf("quay.io/puzzle/goff:%s", tag)
	secret := daggerClient.SetSecret("reg-secret", g.RegistrySecret)

	_, err := goffContainer.WithRegistryAuth("quay.io", g.RegistryUser, secret).Publish(ctx, imageName)

	return err
}

//Build repo server for GitHub actions becuase they don't yet support overriding the entrypoint
func (g *GoffPipeline) buildArgoCdRepoServer(ctx context.Context, client *dagger.Client) {

	repoServerContainer := client.Container().From("quay.io/argoproj/argocd:latest").
		WithUser("root").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "netcat", "-y"}).
		WithUser("argocd").
		WithEntrypoint([]string{"argocd-repo-server"})

	secret := client.SetSecret("reg-secret", g.RegistrySecret)

	_, err := repoServerContainer.WithRegistryAuth(g.RegistryUrl, g.RegistryUser, secret).Publish(ctx, g.getImageFullUrl("argocd-repo-server"))
	if err != nil {
		panic(err)
	}

}

func (g *GoffPipeline) buildAndRelease(ctx context.Context, client *dagger.Client, golang *dagger.Container) []string {

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
				outFile := fmt.Sprintf("./build/goff-%s-%s-%s", oss, arch, g.RefName)
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

	return files
}

func (g *GoffPipeline) releaseFiles(client *dagger.Client, version string, files []string) {

	buildDir := client.Host().Directory("build/")
	ghContainer := client.Container().From("ghcr.io/supportpal/github-gh-cli").
		WithEnvVariable("GITHUB_TOKEN", g.GithubAccessToken).
		WithDirectory("/build", buildDir).
		WithExec([]string{"gh", "-R", "puzzle/goff", "release", "create", version})

	for _, f := range files {
		ghContainer = ghContainer.
			WithExec([]string{"gh", "-R", "puzzle/goff", "release", "upload", version, f})
	}

	_, err := ghContainer.Sync(context.Background())
	if err != nil {
		panic(err)
	}
}

func (g *GoffPipeline) isReleaseTag() bool {
	return g.RefType == "tag" && strings.HasPrefix(g.RefName, "v")
}

func (g *GoffPipeline) getImageFullUrl(name string) string {
	tag := "latest"
	if g.isReleaseTag() {
		tag = g.RefName
	}
	return fmt.Sprintf("%s/puzzle/%s:%s", g.RegistryUrl, name, tag)
}

func mustLoadEnv(env string) string {
	val, found := os.LookupEnv(env)
	if !found {
		panic(fmt.Errorf("Env var '%s' not set", env))
	}
	return val
}
