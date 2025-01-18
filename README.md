# provider-dummymessageservice

`provider-dummymessageservice` is a minimal [Crossplane](https://crossplane.io/) Provider
that is meant to be used as a dummymessageservice for implementing new Providers. It comes
with the following features that are meant to be refactored:

- A `ProviderConfig` type that only points to a credentials `Secret`.
- A `MyType` resource type that serves as an example managed resource.
- A managed resource controller that reconciles `MyType` objects and simply
  prints their configuration in its `Observe` method.

## Developing

1. Use this repository as a dummymessageservice to create a new one.
1. Run `make submodules` to initialize the "build" Make submodule we use for CI/CD.
1. Rename the provider by running the following command:
```shell
  export provider_name=MyProvider # Camel case, e.g. GitHub
  make provider.prepare provider=${provider_name}
```
4. Add your new type by running the following command:
```shell
  export group=sample # lower case e.g. core, cache, database, storage, etc.
  export type=MyType # Camel casee.g. Bucket, Database, CacheCluster, etc.
  make provider.addtype provider=${provider_name} group=${group} kind=${type}
```
5. Replace the *sample* group with your new group in apis/{provider}.go
5. Replace the *mytype* type with your new type in internal/controller/{provider}.go
5. Replace the default controller and ProviderConfig implementations with your own
5. Run `make reviewable` to run code generation, linters, and tests. Ignore linters and tests hehe
5. Run `make build` to build the provider. Notice the `xpkg saved to ...` line
6. Run `make push-xpkg UPBOUND_USER=dummy XPKG_VER=v0.0.1 XPKG_PATH=/path/to/xpkg`

Refer to Crossplane's [CONTRIBUTING.md] file for more information on how the
Crossplane community prefers to work. The [Provider Development][provider-dev]
guide may also be of use.

[CONTRIBUTING.md]: https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md
[provider-dev]: https://github.com/crossplane/crossplane/blob/master/contributing/guide-provider-development.md

## Running

`make run-example`

## Re-using package version

Crossplane creates a ReplicaSet for the provider with `imagePullPolicy: IfNotPresent` so overwriting an already pulled version of image won't update it.

Workaround:
1. Uninstall the provider `kubectl delete -f examples/provider/provider.yaml`
2. Override finalizers of CRs managed by it `make override-finalizers`
3. Find the node on which the pod is running: `kubectl get pod <pod-name> -n <namespace> -o wide`
3. `ssh <node_ip>`
4. Delete the image, cri-o example: `sudo crictl images; sudo crictl rmi <image_id>`

## Sources
- [Creating Custom Providers in Crossplane](https://medium.com/@dan.morita/creating-custom-crossplane-providers-ade76dcc571a)
- [Creating and Pushing Packages](https://docs.upbound.io/upbound-marketplace/packages/)
