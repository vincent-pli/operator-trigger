package watcher

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	ibmdevv1alpha1 "github.com/vincent-pli/operator-trigger/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"

	knativeHandler "github.com/vincent-pli/operator-trigger/pkg/watcher/handlers/knative"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Watcher struct {
	DiscoveryClient *discovery.DiscoveryClient
	DynamicClient   dynamic.Interface
	K8sClient       client.Client
	SwName          string
	SwNamespace     string
	Instance        *ibmdevv1alpha1.ResourceWatcher
	Log             logr.Logger
	Cache           cache.Cache
}

func (w Watcher) Start(stopCh <-chan struct{}) {
	// get ResourceWatcher cr
	instance := &ibmdevv1alpha1.ResourceWatcher{}
	nameNamespace := types.NamespacedName{
		Name:      w.SwName,
		Namespace: w.SwNamespace,
	}

	err := w.K8sClient.Get(context.TODO(), nameNamespace, instance)
	if err != nil {
		w.Log.Error(err, "get Resourcewatcher %s/%s raise exception", w.SwNamespace, w.SwName)
	}
	w.Instance = instance

	// create cache
	namespaces := []string{}
	for _, namespace := range instance.Spec.Namespaces {
		namespaces = append(namespaces, namespace)
	}
	// cache, err := cache.MultiNamespacedCacheBuilder(namespaces)(ctrl.GetConfigOrDie(), cache.Options{})
	// if err != nil {
	// 	w.Log.Error(err, "create cache raise exception")
	// 	return err
	// }
	// w.Cache = cache
	// w.Log.Info("Create cache watch on: ", "namespaces", namespaces)

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(w.DynamicClient, 0)

	for _, resource := range w.Instance.Spec.Resources {
		kind := resource.Kind
		var resourceStr string

		gv, err := schema.ParseGroupVersion(resource.APIVersion)
		if err != nil {
			w.Log.Error(err, "error parsing APIVersion")
			continue
		}

		preferredResources, err := w.DiscoveryClient.ServerResourcesForGroupVersion(gv.String())
		if err != nil {
			if discovery.IsGroupDiscoveryFailedError(err) {
				// w.Log.Warningf("failed to discover some groups: %v", err.(*discovery.ErrGroupDiscoveryFailed).Groups)
				fmt.Println("failed to discover some groups")
			} else {
				// w.Log.Warningf("failed to discover preferred resources: %v", err)
				fmt.Println("failed to discover preferred resources")
			}
		}

		for _, r := range preferredResources.APIResources {
			if r.Kind == kind && strings.Index(r.Name, "status") == -1 {
				resourceStr = r.Name
			}
		}

		gvr := schema.GroupVersionResource{Resource: resourceStr, Group: gv.Group, Version: gv.Version}

		var handler cache.ResourceEventHandler
		if w.Instance.Spec.Handler == "knative" {
			handler = knativeHandler.NewHandler(w.Instance, w.K8sClient, w.Log)
		} else if w.Instance.Spec.Handler == "keda" {
			w.Log.Info("not implements yet")
		} else {
			w.Log.Info("not implements yet")
		}

		informerFactory.ForResource(gvr).Informer().AddEventHandler(handler)

		fmt.Println(gvr.String())

	}
	informerFactory.Start(stopCh)
	<-stopCh
}
