/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"

	"github.com/vincent-pli/operator-trigger/api/v1alpha1"
	watcher "github.com/vincent-pli/operator-trigger/pkg/watcher"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	name      = "WATCHER_NAME"
	nameSpace = "WATCHER_NAMESPACE"
)

var (
	log    = logf.Log.WithName("cmd")
	scheme = runtime.NewScheme()
)

func main() {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	logf.SetLogger(zap.Logger())

	rwName, defined := os.LookupEnv(name)
	if !defined {
		err := fmt.Errorf("No environment variable found")
		log.Error(err, "required environment variable %q not defined", name)
		os.Exit(1)
	}

	rwNamespace, defined := os.LookupEnv(nameSpace)
	if !defined {
		err := fmt.Errorf("No environment variable found")
		log.Error(err, "required environment variable %q not defined", nameSpace)
		os.Exit(1)
	}
	// Setup Scheme for all resources
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	client, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme, Mapper: nil})
	if err != nil {
		log.Error(err, "exception raised when create client")
		os.Exit(1)
	}

	dynamicClientSet := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
	discoverClientSet := discovery.NewDiscoveryClientForConfigOrDie(ctrl.GetConfigOrDie())

	watchers := watcher.Watcher{
		DiscoveryClient: discoverClientSet,
		DynamicClient:   dynamicClientSet,
		K8sClient:       client,
		SwName:          rwName,
		SwNamespace:     rwNamespace,
		Log:             log,
		Scheme:          runtime.Scheme,
	}

	watchers.Start(stopCh)
}
