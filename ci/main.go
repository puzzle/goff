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
	DefaultImageTag   string
	Release           Releaser
}

type Releaser interface {
	releaseFiles(client *dagger.Client, version string, files []string) error
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

	//get working directory on host, load files, exclude build and ci directory
	source := daggerClient.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"ci/", "build/"},
	})

	//add source to container
	golang = golang.WithDirectory("/src", source)

	//new error group for concurrent execution
	errGroup, _ := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return runTests(daggerClient, golang, ctx)
	})

	errGroup.Go(func() error {
		return buildAndPushDevImage(golang, ctx, g, daggerClient)

	})

	err = errGroup.Wait()
	if err != nil {
		return err
	}

	//If version tag, build binary releases and release them on github
	if g.isReleaseTag() {
		files, err := g.buildAndRelease(ctx, daggerClient, golang)

		if err != nil {
			return err
		}

		err = g.Release.releaseFiles(daggerClient, g.RefName, files)

		if err != nil {
			return err
		}

		//Build patched ArgoCD Repo server
		return g.buildArgoCdRepoServer(ctx, daggerClient)

	}

	return nil
}

func buildAndPushDevImage(golang *dagger.Container, ctx context.Context, g *GoffPipeline, daggerClient *dagger.Client) error {
	goffGitlab, err := golang.
		WithExec([]string{"mkdir", "-p", "/app"}).
		WithEnvVariable("CC", "musl-gcc").
		WithExec([]string{"go", "build", "-o", "/app/goff", "github.com/puzzle/goff"}).
		WithExec([]string{"go", "install", "gitlab.com/gitlab-org/cli/cmd/glab@main"}).
		Sync(ctx)

	if err != nil {
		return err
	}

	if g.RefName == "main" || g.isReleaseTag() {
		err = g.publishImage(ctx, goffGitlab, daggerClient)
	}

	return err
}

func runTests(daggerClient *dagger.Client, golang *dagger.Container, ctx context.Context) error {
	repoServer := daggerClient.Container().From("quay.io/puzzle/argocd-repo-server:latest").
		WithExposedPort(8081).
		WithEntrypoint([]string{"argocd-repo-server"}).WithExec(nil)

	_, err := golang.
		WithServiceBinding("reposerver", repoServer).
		WithExec([]string{"go", "test", "./...", "-v"}).Stdout(ctx)
	return err
}

func (g *GoffPipeline) publishImage(ctx context.Context, golang *dagger.Container, daggerClient *dagger.Client) error {
	goffBin := golang.File("/app/goff")
	glabBin := golang.File("/go/bin/glab")

	goffContainer := daggerClient.Container().From("docker.io/alpine:3.18").
		WithFile("/bin/goff", goffBin).
		WithFile("/bin/glab", glabBin).
		WithExec([]string{"apk", "add", "git", "helm"})

	goffContainer = goffContainer.WithEntrypoint([]string{"/bin/goff"})

	secret := daggerClient.SetSecret("reg-secret", g.RegistrySecret)

	imageName := g.getImageFullUrl("goff")

	_, err := goffContainer.WithRegistryAuth(g.RegistryUrl, g.RegistryUser, secret).Publish(ctx, imageName)

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

func (g *GoffPipeline) buildAndRelease(ctx context.Context, client *dagger.Client, golang *dagger.Container) ([]string, error) {

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

func (g *GitHubReleaser) releaseFiles(client *dagger.Client, version string, files []string) error {

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
