package argocd

import (
	"net"
	"testing"
)

func TestRender(t *testing.T) {
	const repoServerAddr = "reposerver:8081"
	if _, err := net.LookupHost(repoServerAddr); err != nil {
		t.Skipf("no reposerver host found: %v", err)
	}

	err := Render("../testdata/argocd/helm_app.yaml", repoServerAddr, "../out/argocd/app/", RepoCredentails{})
	if err != nil {
		t.Error(err)
	}
}
