package main

import (
	"dagger/ci/internal/dagger"
	"fmt"
)

type Build struct {
	// +private
	Source *Directory
}

func (ci *CI) Build() *Build {
	return &Build{
		Source: ci.Source,
	}
}

func golangBaseImage(source *Directory) *Container {
	return dag.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From(fmt.Sprintf("golang:%s-alpine", goVersion)).
		WithMountedCache("/go/src", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithWorkdir("/src").
		WithFile("go.mod", source.File("go.mod")).
		WithFile("go.sum", source.File("go.sum")).
		WithExec([]string{"go", "mod", "download"}).
		WithDirectory("/src", source, ContainerWithDirectoryOpts{
			Exclude: []string{
				"ci/",
				"build/",
				"goff",
			},
		})
}

func (b *Build) binary() *File {
	return golangBaseImage(b.Source).
		WithExec([]string{"go", "build", "-o", "goff", "-ldflags", "-s -w"}).
		File("/src/goff")
}

// build `goff` container image
func (b *Build) Image() *Container {
	containerOpts := dagger.ContainerOpts{Platform: "linux/amd64"}
	return dag.Container(containerOpts).
		From(alpineBaseImage).
		WithExec([]string{"addgroup", "-g", "1001", "goff"}).
		WithExec([]string{"adduser", "-D", "-u", "1001", "-G", "goff", "goff"}).
		WithFile("/usr/local/bin/kustomize", dag.Container(containerOpts).From(kustomizeImage).File("/app/kustomize")).
		WithFile("/usr/local/bin/goff", b.binary()).
		WithExec([]string{"apk", "add", "git", "helm"}).
		WithUser("goff")
}
