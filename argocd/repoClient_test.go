package argocd

import "testing"

func TestRender(t *testing.T) {
	//deactivate until https://github.com/dagger/dagger/issues/5382 is fixed
	err := Render("../testdata/argocd/helm_app.yaml", "reposerver:8081", "../out/argocd/app/", RepoCredentails{})
	if err != nil {
		t.Error(err)
	}
}
