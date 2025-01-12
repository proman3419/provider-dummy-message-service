/*
Copyright 2022 The Crossplane Authors.

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

package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-dummymessageservice/apis/core/v1alpha1"
	apisv1alpha1 "github.com/crossplane/provider-dummymessageservice/apis/v1alpha1"
	"github.com/crossplane/provider-dummymessageservice/internal/features"
)

const (
	errNotMessage   = "managed resource is not a Message custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

// NoOpService -> DummyMessageService
// A NoOpService does nothing.
type DummyMessageService struct {
}

var (
	newDummyMessageService = func(_ []byte) (*DummyMessageService, error) {
		return &DummyMessageService{}, nil
	}
)

// Setup adds a controller that reconciles Message managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.MessageGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.MessageGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: newDummyMessageService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Message{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(creds []byte) (*DummyMessageService, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Message)
	if !ok {
		return nil, errors.New(errNotMessage)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newServiceFn(data)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service *DummyMessageService
}

type Message struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
}

type Messages struct {
	Messages []Message `json:"messages"`
}

func getApiUrl() string {
	namespace := "dummy-message-service"
	serviceName := "dummy-message-service-svc"
	apiPort := 8000
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, namespace, apiPort)
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Message)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMessage)
	}

	messageToObserve := Message{Content: cr.Spec.ForProvider.Content}

	resp, err := http.Get(fmt.Sprintf("%s/messages", getApiUrl()))
	if err != nil {
		log.Printf("Failed to send GET /messages: %v\n", err)
		return managed.ExternalObservation{ResourceExists: false}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to send GET /messages: %v\n", err)
		return managed.ExternalObservation{ResourceExists: false}, err
	}

	var messages Messages
	err = json.Unmarshal(body, &messages)
	if err != nil {
		log.Printf("Failed to parse JSON: %v\n", err)
		return managed.ExternalObservation{ResourceExists: false}, err
	}

	for _, message := range messages.Messages {
		if message.Content == messageToObserve.Content {
			log.Printf("Observed a message with content: '%s'\n", messageToObserve.Content)
			return managed.ExternalObservation{
				// Return false when the external resource does not exist. This lets
				// the managed resource reconciler know that it needs to call Create to
				// (re)create the resource, or that it has successfully been deleted.
				ResourceExists: true,

				// Return false when the external resource exists, but it not up to date
				// with the desired managed resource state. This lets the managed
				// resource reconciler know that it needs to call Update.
				ResourceUpToDate: true,

				// Return any details that may be required to connect to the external
				// resource. These will be stored as the connection secret.
				ConnectionDetails: managed.ConnectionDetails{},
			}, nil
		}
	}
	log.Printf("Didn't observe a message with content: '%s'\n", messageToObserve.Content)
	return managed.ExternalObservation{ResourceExists: false}, err
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Message)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMessage)
	}

	messageToCreate := Message{Content: cr.Spec.ForProvider.Content}

	resp, err := http.Post(fmt.Sprintf("%s/message?content=%s", getApiUrl(), url.QueryEscape(messageToCreate.Content)),
		"application/json", nil)
	if err != nil {
		log.Printf("Failed to send POST /message: %v\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to send POST /message: %v\n", err)
	}
	log.Printf("Response: %s\n", body)

	var createdMessage Message
	err = json.Unmarshal([]byte(body), &createdMessage)
	if err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
	}
	log.Printf("Created message: %+v\n", createdMessage)

	cr.Status.AtProvider.Id = createdMessage.Id
	cr.Status.AtProvider.Content = createdMessage.Content

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	_, ok := mg.(*v1alpha1.Message)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotMessage)
	}

	// Not covered by this demo

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Message)
	if !ok {
		return errors.New(errNotMessage)
	}

	messageToDelete := Message{Id: cr.Spec.ForProvider.Id}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/message?id_=%d", getApiUrl(), messageToDelete.Id),
		bytes.NewBuffer([]byte(nil)))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to send DELETE /message: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to send DELETE /message: %v\n", err)
	}
	log.Printf("Response: %s\n", body)

	return nil
}
