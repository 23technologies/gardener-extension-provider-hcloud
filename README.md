**This Gardener Extension Provider is a work in progress. Code may still break at any time. DO NOT USE IT.**

----

# [Gardener Extension for Hetzner Cloud provider](https://gardener.cloud)

[![CI Build status](https://concourse.ci.gardener.cloud/api/v1/teams/gardener/pipelines/gardener-extension-provider-hcloud-master/jobs/master-head-update-job/badge)](https://concourse.ci.gardener.cloud/teams/gardener/pipelines/gardener-extension-provider-hcloud-master/jobs/master-head-update-job)
[![Go Report Card](https://goreportcard.com/badge/github.com/23technologies/gardener-extension-provider-hcloud)](https://goreportcard.com/report/github.com/23technologies/gardener-extension-provider-hcloud)

Project Gardener implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service.
Its main principle is to leverage Kubernetes concepts for all of its tasks.

Recently, most of the vendor specific logic has been developed [in-tree](https://github.com/gardener/gardener).
However, the project has grown to a size where it is very hard to extend, maintain, and test.
With [GEP-1](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md) we have proposed how the architecture can be changed in a way to support external controllers that contain their very own vendor specifics.
This way, we can keep Gardener core clean and independent.

This controller implements Gardener's extension contract for the Hetzner cloud provider.

All `ControllerRegistration` resources that can be used to register this controller to Gardener can be found under releases.

Please find more information regarding the extensibility concepts and a detailed proposal [here](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md).

## Supported Kubernetes versions

This extension controller supports the following Kubernetes versions:

| Version         | Support     | Conformance test results |
| --------------- | ----------- | ------------------------ |
| Kubernetes 1.20 | untested    | not yet available        |
| Kubernetes 1.19 | untested    | not yet available        |
| Kubernetes 1.18 | untested    | not yet available        |
| Kubernetes 1.17 | 1.17.0+     | not yet available        |
| Kubernetes 1.16 | 1.16.0+     | not yet available        |
| Kubernetes 1.15 | 1.15.0+     | not yet available        |

Please take a look [here](https://github.com/gardener/gardener/blob/master/docs/usage/supported_k8s_versions.md) to see which versions are supported by Gardener in general.

----

## How to start using or developing this extension controller locally

You can run the controller locally on your machine by executing `make start`.

Static code checks and tests can be executed by running `VERIFY=true make all`. We are using Go modules for Golang package dependency management and [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega) for testing.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/23technologies/gardener-extension-provider-hcloud/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

* [Our landing page gardener.cloud](https://gardener.cloud/)
* ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
* ["Gardener Project Update" blog on kubernetes.io](https://kubernetes.io/blog/2019/12/02/gardener-project-update/)
* [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
* [GEP-4 (New `core.gardener.cloud/v1alpha1` API)](https://github.com/gardener/gardener/blob/master/docs/proposals/04-new-core-gardener-cloud-apis.md)
* [Extensibility API documentation](https://github.com/gardener/gardener/tree/master/docs/extensions)
* [Gardener Extensions Golang library](https://godoc.org/github.com/gardener/gardener/extensions/pkg)
* [Gardener API Reference](https://gardener.cloud/api-reference/)
