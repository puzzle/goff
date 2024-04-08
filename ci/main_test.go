package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	mrand "math/rand"
	"os"
	"testing"

	"dagger.io/dagger"
	"github.com/stretchr/testify/assert"
)

type ReleaseStub struct {
}

// releaseFiles implements Releaser
func (*ReleaseStub) releaseFiles(ctx context.Context, version string, files []string, client *dagger.Client) error {
	//Do nothing
	return nil
}

func (*ReleaseStub) releaseDocs(ctx context.Context, version string, daggerClient *dagger.Client) error {
	return nil
}

type ReleaseStubErr struct {
}

// releaseFiles implements Releaser which return error
func (*ReleaseStubErr) releaseFiles(ctx context.Context, version string, files []string, client *dagger.Client) error {
	return errors.New("shoul not be called")
}

func (*ReleaseStubErr) releaseDocs(ctx context.Context, version string, daggerClient *dagger.Client) error {
	return errors.New("shoul not be called")
}

var _ Releaser = &ReleaseStub{}

func randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func TestMain(t *testing.T) {
	os.Chdir("..")
	ctx := context.Background()
	daggerClient, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		t.Fail()
	}
	defer daggerClient.Close()

	testNonMainBranch(ctx, t, daggerClient)
	testMainBranch(ctx, t, daggerClient)
	testNonReleaseTag(ctx, t, daggerClient)
	testReleaseTag(ctx, t, daggerClient)
}

func testNonMainBranch(ctx context.Context, t *testing.T, client *dagger.Client) {
	gp := &GoffPipeline{
		RefType:         "branch",
		RefName:         "dev",
		RegistryUser:    "admin",
		RegistrySecret:  "secret",
		RegistryUrl:     "ttl.sh",
		Release:         &ReleaseStub{},
		DefaultImageTag: randomHex(12),
	}

	err := gp.run()

	assert.Nil(t, err)

	_, err = client.Container().From(gp.getImageFullUrl("goff")).WithExec([]string{"--help"}).Sync(ctx)
	assert.Error(t, err, "container with name '%s' should not exists", gp.getImageFullUrl("goff"))

}

func testMainBranch(ctx context.Context, t *testing.T, daggerClient *dagger.Client) {
	gpMain := &GoffPipeline{
		RefType:         "branch",
		RefName:         "main",
		RegistryUser:    "admin",
		RegistrySecret:  "secret",
		RegistryUrl:     "ttl.sh",
		Release:         &ReleaseStubErr{},
		DefaultImageTag: randomHex(12),
	}

	err := gpMain.run()
	assert.Nil(t, err)

	_, err = daggerClient.Container().From(gpMain.getImageFullUrl("goff")).WithExec([]string{"--help"}).Sync(ctx)
	assert.Nil(t, err, "container '%s' should exists and be functional", gpMain.getImageFullUrl("goff"))
}

func testNonReleaseTag(ctx context.Context, t *testing.T, daggerClient *dagger.Client) {
	gpWrongTag := &GoffPipeline{
		RefType:         "tag",
		RefName:         "birnenbaum",
		RegistryUser:    "admin",
		RegistrySecret:  "secret",
		RegistryUrl:     "ttl.sh",
		Release:         &ReleaseStubErr{},
		DefaultImageTag: randomHex(12),
	}

	err := gpWrongTag.run()
	assert.Nil(t, err)

	_, err = daggerClient.Container().From(gpWrongTag.getImageFullUrl("goff")).WithExec([]string{"--help"}).Sync(ctx)
	assert.Error(t, err, "container with name '%s' should not exists", gpWrongTag.getImageFullUrl("goff"))
}

func testReleaseTag(ctx context.Context, t *testing.T, daggerClient *dagger.Client) {
	gprelease := &GoffPipeline{
		RefType:         "tag",
		RefName:         fmt.Sprintf("v0.%d.%d", mrand.Intn(200), mrand.Intn(200)),
		RegistryUser:    "admin",
		RegistrySecret:  "secret",
		RegistryUrl:     "ttl.sh",
		Release:         &ReleaseStub{},
		DefaultImageTag: randomHex(12),
	}

	err := gprelease.run()
	assert.Nil(t, err)

	_, err = daggerClient.Container().From(gprelease.getImageFullUrl("goff")).WithExec([]string{"--help"}).Sync(ctx)
	assert.Nil(t, err, "container with name '%s' should exists", gprelease.getImageFullUrl("goff"))

	_, err = daggerClient.Container().From(gprelease.getImageFullUrl("goff")).WithExec([]string{"kustomize", "--version"}).Sync(ctx)
	assert.Nil(t, err, "container with name '%s' should exists", gprelease.getImageFullUrl("goff"))
}
