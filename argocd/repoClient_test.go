package argocd

import "testing"

func TestRender(t *testing.T) {
	err := Render("../testdata/argocd/helm_app.yaml", "localhost:8081", "../out/argocd/app/", RepoCredentails{})
	if err != nil {
		t.Error(err)
	}
}
