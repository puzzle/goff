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
	RefType         string
	RefName         string
	RegistryUser    string
	RegistrySecret  string
	RegistryUrl     string
	DefaultImageTag string
	Release         Releaser
}

type Releaser interface {
	releaseFiles(ctx context.Context, version string, files []string, client *dagger.Client) error
	releaseDocs(ctx context.Context, version string, daggerClient *dagger.Client) error
}

func main() {
	gp := NewFromGithub()
	err := gp.run()
	if err != nil {
		panic(err)
	}
}

func (g *GoffPipeline) run() error {

	// create Dagger client
	ctx := context.Background()
	daggerClient, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer daggerClient.Close()

	golang := createGoBase(daggerClient)

	//new error group for concurrent execution
	errGroup, _ := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return runTests(ctx, golang, daggerClient)
	})

	errGroup.Go(func() error {
		if g.RefName == "main" || g.isReleaseTag() {
			return buildAndPushDevImage(ctx, golang, g, daggerClient)
		}
		return nil

	})

	err = errGroup.Wait()
	if err != nil {
		return err
	}

	//If version tag, build binary releases and release them on github
	if g.isReleaseTag() {
		files, err := g.build(ctx, daggerClient, golang)

		if err != nil {
			return err
		}

		err = g.Release.releaseFiles(ctx, g.RefName, files, daggerClient)

		if err != nil {
			return err
		}

		err = g.Release.releaseDocs(ctx, g.RefName, daggerClient)
		if err != nil {
			return err
		}

		//Build patched ArgoCD Repo server
		return g.buildArgoCdRepoServer(ctx, daggerClient)

	}

	return nil
}

// create golang base
func createGoBase(daggerClient *dagger.Client) *dagger.Container {
	goMod := daggerClient.CacheVolume("go")
	goChache := daggerClient.CacheVolume("go-cache")

	base := daggerClient.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("golang:1.22").
		WithMountedCache("/go/src", goMod).
		WithMountedCache("/root/.cache/go-build", goChache)

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

		//get working directory on host, load files, exclude build and ci directory
	source := daggerClient.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"ci/", "build/"},
	})

	golang = golang.WithDirectory("/src", source)
	return golang
}

func buildAndPushDevImage(ctx context.Context, golang *dagger.Container, g *GoffPipeline, daggerClient *dagger.Client) error {

	goffGitlab, err := golang.
		WithExec([]string{"mkdir", "-p", "/app"}).
		WithEnvVariable("CC", "musl-gcc").
		WithExec([]string{"go", "build", "-o", "/app/goff", "github.com/puzzle/goff"}).
		WithExec([]string{"go", "install", "gitlab.com/gitlab-org/cli/cmd/glab@main"}).
		Sync(ctx)

	if err != nil {
		return err
	}

	goffBin := goffGitlab.File("/app/goff")
	glabBin := goffGitlab.File("/go/bin/glab")

	kustomizeContainer := daggerClient.Container().
		From("registry.k8s.io/kustomize/kustomize:v5.4.1")

	goffContainer := daggerClient.Container().
		From("docker.io/alpine:3.18").
		WithExec([]string{"addgroup", "-g", "1001", "goff"}).
		WithExec([]string{"adduser", "-D", "-u", "1001", "-G", "goff", "goff"}).
		WithFile("/usr/local/bin/kustomize", kustomizeContainer.File("/app/kustomize")).
		WithFile("/bin/goff", goffBin).
		WithFile("/bin/glab", glabBin).
		WithExec([]string{"apk", "add", "git", "helm"})

	goffContainer = goffContainer.
		WithEntrypoint([]string{"/bin/goff"}).
		WithUser("goff")

	secret := daggerClient.SetSecret("reg-secret", g.RegistrySecret)

	imageName := g.getImageFullUrl("goff")

	_, err = goffContainer.WithRegistryAuth(g.RegistryUrl, g.RegistryUser, secret).Publish(ctx, imageName)

	return err

}

func runTests(ctx context.Context, golang *dagger.Container, daggerClient *dagger.Client) error {

	redis, err := daggerClient.Container().From("registry.puzzle.ch/docker.io/redis").
		WithExposedPort(6379).AsService().Start(ctx)

	if err != nil {
		return err
	}

	repoServer, err := daggerClient.Container().From("quay.io/puzzle/argocd-repo-server:latest").
		WithExposedPort(8081).
		WithServiceBinding("argocd-redis", redis).
		AsService().Start(ctx)

	if err != nil {
		return err
	}

	_, err = golang.
		WithServiceBinding("reposerver", repoServer).
		WithExec([]string{"go", "test", "./...", "-v"}).Sync(ctx)

	repoServer.Stop(ctx)
	redis.Stop(ctx)

	return err
}

// Build repo server for GitHub actions becuase they don't yet support overriding the entrypoint
func (g *GoffPipeline) buildArgoCdRepoServer(ctx context.Context, client *dagger.Client) error {

	repoServerContainer := client.Container().From("quay.io/argoproj/argocd:latest").
		WithUser("root").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "netcat", "-y"}).
		WithUser("argocd").
		WithEntrypoint([]string{"argocd-repo-server"})

	secret := client.SetSecret("reg-secret", g.RegistrySecret)

	_, err := repoServerContainer.WithRegistryAuth(g.RegistryUrl, g.RegistryUser, secret).Publish(ctx, g.getImageFullUrl("argocd-repo-server"))

	return err

}

func (g *GoffPipeline) build(ctx context.Context, client *dagger.Client, golang *dagger.Container) ([]string, error) {

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

	return files, err
}

type GitHubReleaser struct {
	GithubAccessToken string
}

func (g *GitHubReleaser) releaseFiles(ctx context.Context, version string, files []string, client *dagger.Client) error {

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

	return err
}

func (g *GitHubReleaser) releaseDocs(ctx context.Context, version string, daggerClient *dagger.Client) error {
	mkdocs := daggerClient.Container().From("python:3-slim")

	token := daggerClient.SetSecret("token", g.GithubAccessToken)

	_, err := mkdocs.
		WithEnvVariable("GOFF_VERSION", version).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "git", "-y"}).
		WithWorkdir("/src").
		WithExec([]string{"pip", "install", "mkdocs", "mkdocs-markdownextradata-plugin"}).
		WithDirectory("/src", daggerClient.Host().Directory(".", dagger.HostDirectoryOpts{
			Include: []string{"mkdocs.yml", "docs/", ".git"},
		})).
		WithSecretVariable("GH_PUSH_TOKEN", token).
		WithExec([]string{"git", "remote", "set-url", "origin", "https://schlapzz:$GH_PUSH_TOKEN@github.com/puzzle/goff.git"}).
		WithExec([]string{"git", "config", "--global", "user.email", "schlatter@puzzle.ch"}).
		WithExec([]string{"git", "config", "--global", "user.name", "schlapzz"}).
		WithExec([]string{"git", "fetch"}).
		WithExec([]string{"git", "checkout", "origin/gh-pages"}).
		WithExec([]string{"mkdocs", "build"}).
		WithExec([]string{"rm", "-rf", "docs"}).
		WithExec([]string{"mv", "site", "docs"}).
		WithExec([]string{"git", "add", "."}).
		WithExec([]string{"git", "commit", "-m", "test"}).
		WithExec([]string{"git", "push", "--force"}).
		Sync(ctx)

	return err
}

func (g *GoffPipeline) isReleaseTag() bool {
	return g.RefType == "tag" && strings.HasPrefix(g.RefName, "v")
}

func (g *GoffPipeline) getImageFullUrl(name string) string {
	tag := g.DefaultImageTag
	if g.isReleaseTag() {
		tag = g.RefName
	}
	return fmt.Sprintf("%s/puzzle/%s:%s", g.RegistryUrl, name, tag)
}

// Load configuration from Github and custom environment variables
func NewFromGithub() GoffPipeline {
	return GoffPipeline{
		RefType:         mustLoadEnv("GITHUB_REF_TYPE"),
		RefName:         mustLoadEnv("GITHUB_REF_NAME"),
		RegistrySecret:  mustLoadEnv("REGISTRY_PASSWORD"),
		RegistryUser:    mustLoadEnv("REGISTRY_USER"),
		RegistryUrl:     mustLoadEnv("REGISTRY_URL"),
		DefaultImageTag: "latest",
		Release: &GitHubReleaser{
			GithubAccessToken: mustLoadEnv("GITHUB_ACCESS_TOKEN"),
		},
	}
}

func mustLoadEnv(env string) string {
	val, found := os.LookupEnv(env)
	if !found {
		panic(fmt.Errorf("env var '%s' not set", env))
	}
	return val
}
