package knative

import (
	"context"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/apis/duck/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CouldeventHandler struct {
	K8sClient client.Client
	Client    cloudevents.Client
	Sink      string
	Log       logr.Logger
}

var _ cache.ResourceEventHandler = (*CouldeventHandler)(nil)

func (c *CouldeventHandler) NewHandler(instance *ibmdevv1alpha1.ResourceWatcher, client client.Client, logger logr.Logger) (*CouldeventHandler, error) {
	// prepare eventHandler, could be couldevent or k8sevent
	sink, err := c.getSinkURI(instance.Spec.Target, instance.Namespace)
	if err != nil {
		w.Log.Error(err, "get sink raise exception")
	}

	eventsClient, err := cloudevents.NewClientHTTP()
	if err != nil {
		w.Log.Error(err, "failed to create client")
	}

	return &CouldeventHandler{
		K8sClient: client,
		Client:    eventsClient,
		Sink:      sink,
		Log:       logger,
	}
}

func (c *CouldeventHandler) OnAdd(obj interface{}) {
	// Pull metav1.Object out of the object
	if o, err := meta.Accessor(obj); err == nil {
		c.Log.Info("resource added...", "object", o)
	} else {
		c.Log.Error(err, "OnAdd missing Meta",
			"object", obj, "type", fmt.Sprintf("%T", obj))
		return
	}

	// Pull the runtime.Object out of the object
	if o, ok := obj.(runtime.Object); ok {
		c.Log.Info("resource added again...", "object", o)
	} else {
		c.Log.Error(nil, "OnAdd missing runtime.Object",
			"object", obj, "type", fmt.Sprintf("%T", obj))
		return
	}
	event := cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

	ctx := cloudevents.ContextWithTarget(context.Background(), c.Sink)
	// Send that Event.
	if result := c.Client.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}

	c.Log.Info("resource added %v", obj)
}

func (c *CouldeventHandler) OnUpdate(oldObj, newObj interface{}) {
	c.Log.Info("resource update")

}

func (c *CouldeventHandler) OnDelete(obj interface{}) {
	c.Log.Info("resource delete")
}

// GetSinkURI retrieves the sink URI from the object referenced by the given
// ObjectReference.
func (c *CouldeventHandler) getSinkURI(sink *corev1.ObjectReference, namespace string) (string, error) {
	if sink == nil {
		return "", fmt.Errorf("sink ref is nil")
	}

	if sink.Namespace == "" {
		sink.Namespace = namespace
	}

	objIdentifier := fmt.Sprintf("\"%s/%s\" (%s)", sink.Namespace, sink.Name, sink.GroupVersionKind())

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(sink.GroupVersionKind())
	err := c.K8sClient.Get(context.TODO(), client.ObjectKey{Name: sink.Name, Namespace: sink.Namespace}, u)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize sink %s: %v", objIdentifier, err)
	}

	t := v1beta1.AddressableType{}
	err = duck.FromUnstructured(u, &t)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize sink %s: %v", objIdentifier, err)
	}

	if t.Status.Address == nil {
		return "", fmt.Errorf("sink %s does not contain address", objIdentifier)
	}

	if t.Status.Address.URL == nil {
		return "", fmt.Errorf("sink %s contains an empty URL", objIdentifier)
	}

	return t.Status.Address.URL.String(), nil
}
