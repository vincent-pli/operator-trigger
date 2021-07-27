module github.com/vincent-pli/operator-trigger

go 1.16

require (
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/go-logr/logr v0.4.0
	github.com/kedacore/keda/v2 v2.3.0
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.12.0
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.26.0
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/pkg v0.0.0-20210712150822-e8973c6acbf7
	sigs.k8s.io/controller-runtime v0.8.3
)

replace k8s.io/client-go => k8s.io/client-go v0.20.0
