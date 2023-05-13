Integrations
===

We provide and support following CI Tools

# Github

//TODO

# Gitlab

//TODO

## OpenShift Considerations

If you are using Gitlab Runners on OpenShift, yu have to tweak your runner configurations to run the ArgoCD Repo Server as service.

```toml
      [runners.kubernetes]
        [[runners.kubernetes.volumes.empty_dir]]
          name = "argo-repo-gpg"
          mount_path = "/app/config/gpg/keys"
```