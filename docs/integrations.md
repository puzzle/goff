# Integrations

## Github

```yaml
on:
  pull_request:
    types: [opened, synchronize]
    paths: 
     - "argocd/**"

permissions:
  contents: read
  pull-requests: write

name: Diff GitOps Environments

jobs:
  diff-env-argo:
    # The type of runner that the job will run on
    runs-on: ubuntu-latest
    services:
      reposerver:
        image: quay.io/puzzle/argocd-repo-server:latest
        ports:
          - "8081:8081"
        options: >-
          --health-cmd "nc -z localhost 8081"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    container:
      image: quay.io/puzzle/goff:latest
    steps:
      - name: Checkout PR
        uses: actions/checkout@v3
        with:
         path: source
      - name: Checkout Target of PR
        uses: actions/checkout@v3
        with:
          path: target
          ref: ${{ github.event.pull_request.base.ref }}
      - run: |
         goff argocd app "./source/argocd" --repo-server="reposerver:8081" --output-dir=/tmp/source/
         goff argocd app "./target/argocd" --repo-server="reposerver:8081" --output-dir=/tmp/target/
         goff diff "/tmp/source" "/tmp/target" --output-dir .
      - name: comment PR
        uses: thollander/actions-comment-pull-request@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          filePath: diff.md
```

## Gitlab

```yaml
stages:
 - merge

create_comment:
    stage: merge
    image: quay.io/puzzle/goff:latest
    script:
     - echo "Target branch $CI_MERGE_REQUEST_SOURCE_BRANCH_NAME"
     - git clone -b $CI_MERGE_REQUEST_TARGET_BRANCH_NAME --single-branch $CI_REPOSITORY_URL /tmp/checkout/target
     - mkdir -p /tmp/source
     - mkdir -p /tmp/target
     - cd kustomize
     - goff kustomize build . --output-dir=/tmp/out/source
     - cd /tmp/checkout/target/kustomize
     - goff kustomize build . --output-dir=/tmp/out/target
     - goff diff "/tmp/out/source" "/tmp/out/target" --output-dir /tmp/
     - glab auth login --hostname gitlab.puzzle.ch -t $GITLAB_ACCESS_TOKEN
     - glab mr note -m "$(cat /tmp/diff.md)" $CI_MERGE_REQUEST_SOURCE_BRANCH_NAME
    rules:
    - if: $CI_PIPELINE_SOURCE == 'merge_request_event'
      changes: 
        - kustomize/**/*
```

## Gitea