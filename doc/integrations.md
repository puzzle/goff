Integrations
===

We provide and support following CI Tools

# Github

//TODO

## ArgoCD Applications Considerations

Due the limitation of Github we can not override the entrypoint of a service image.
Therefore we recommend to use our argo-cd-repo server image which just overrides the entrypoint and install netcat
for the health probe.

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