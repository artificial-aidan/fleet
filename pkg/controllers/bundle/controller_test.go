package bundle

import (
	"context"
	"testing"

	fleetv1alpha1 "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/fleet/pkg/generated/controllers/fleet.cattle.io"
	"github.com/rancher/fleet/pkg/manifest"
	"github.com/rancher/fleet/pkg/target"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/pkg/generated/controllers/core"
	"github.com/rancher/wrangler/pkg/ratelimit"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
)

func TestOnChanged(t *testing.T) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// if you want to change the loading rules (which files in which order), you can do so here

	configOverrides := &clientcmd.ConfigOverrides{}
	// if you want to change override values or bind them to flags, there are methods to help you

	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	client, err := cfg.ClientConfig()
	if err != nil {
		panic(err)
	}
	client.RateLimiter = ratelimit.None

	scheme := runtime.NewScheme()
	fleetv1alpha1.AddToScheme(scheme)

	scf, err := controller.NewSharedControllerFactoryFromConfig(client, scheme)
	if err != nil {
		panic(err)
	}

	err = scf.Start(context.Background(), 10)
	if err != nil {
		panic(err)
	}

	core, err := core.NewFactoryFromConfigWithOptions(client, &core.FactoryOptions{
		SharedControllerFactory: scf,
	})
	if err != nil {
		panic(err)
	}
	corev := core.Core().V1()

	fleet, err := fleet.NewFactoryFromConfigWithOptions(client, &fleet.FactoryOptions{
		SharedControllerFactory: scf,
	})
	if err != nil {
		panic(err)
	}
	fleetv := fleet.Fleet().V1alpha1()
	targetManager := target.New(
		fleetv.Cluster().Cache(),
		fleetv.ClusterGroup().Cache(),
		fleetv.Bundle().Cache(),
		fleetv.BundleNamespaceMapping().Cache(),
		corev.Namespace().Cache(),
		manifest.NewStore(fleetv.Content()),
		fleetv.BundleDeployment().Cache())

	err = scf.Start(context.Background(), 10)
	if err != nil {
		panic(err)
	}
	scf.SharedCacheFactory().WaitForCacheSync(context.Background())

	h := &handler{
		targets: targetManager,
	}

	h.OnClusterChange("", &fleetv1alpha1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name:      "virtual-dev-aidan",
			Namespace: "fleet-default",
			Labels: map[string]string{
				"name": "virtual-dev-aidan",
			},
		},
		Spec: fleetv1alpha1.ClusterSpec{},
	},
	)
}
