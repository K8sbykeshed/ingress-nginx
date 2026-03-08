/*
Copyright 2026 The Kubernetes Authors.

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

package framework

import (
	"github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Global shared variables, not great but this is e2e anyway
var (
	kubeClientSet *kubernetes.Clientset
	kubeConfig    *rest.Config
	ingressClass  string
)

const (
	sharedNamespace = "e2e-shared"
)

// NewSharedFramework makes a new framework and sets up a BeforeEach/AfterEach for
// you (you can write additional before/after each functions).
// It does not create a new environment and instead relies on a shared ingress-nginx instance
func NewSharedFramework(baseName string, opts ...func(*Framework)) *Framework {
	defer ginkgo.GinkgoRecover()

	f := &Framework{
		BaseName: baseName,
		shared:   true,
	}
	// set framework options
	for _, o := range opts {
		o(f)
	}

	ginkgo.BeforeEach(f.BeforeEachShared)
	ginkgo.AfterEach(f.AfterEachShared)

	return f
}

// InitializeSharedClients initializes the kube client and config for shared environment
// This should be called on all Ginkgo parallel processes
func InitializeSharedClients(t require.TestingT) {
	var err error

	kubeConfig, err = loadConfig()
	require.NoError(t, err, "error loading config")

	// TODO: remove after k8s v1.22
	kubeConfig.WarningHandler = rest.NoWarnings{}

	kubeClientSet, err = kubernetes.NewForConfig(kubeConfig)
	require.NoError(t, err, "error creating client")

	// Set ingressClass to the expected name based on namespace
	ingressClass = "ic-" + sharedNamespace
}

// BootstrapSharedEnvironment creates the shared namespace, IngressClass, and ingress controller
// This should only be called once (on Ginkgo process 1)
func BootstrapSharedEnvironment(t require.TestingT) {
	var err error

	_, err = createKubeSharedNamespace(sharedNamespace, kubeClientSet)
	require.NoError(t, err, "error creating the namespace for deployment")

	ingressClass, err = CreateIngressClass(sharedNamespace, kubeClientSet)
	require.NoError(t, err, "creating IngressClass")

	err = newIngressController(sharedNamespace, "")
	require.NoError(t, err, "deploying the ingress controller")
}

// createKubeSharedNamespace creates a new shared namespace in the cluster
func createKubeSharedNamespace(baseName string, c kubernetes.Interface) (string, error) {
	return createNamespace(baseName, nil, c, true)
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEachShared() {
	var err error

	// Creates the namespace and sets the configuration for it
	f.CreateEnvironment()

	err = f.updateIngressNGINXPod()
	assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller pod information")

	f.WaitForNginxListening(80)

	// If HTTPBun is enabled deploy an instance to the namespace
	if f.HTTPBunEnabled {
		f.HTTPBunIP = f.NewHttpbunDeployment()
	}
}

// AfterEachShared deletes the namespace, after reading its events and dumps the log
// in case of an error
func (f *Framework) AfterEachShared() {

	// Remove the test namespace
	defer f.DestroyEnvironment()

	if !ginkgo.CurrentSpecReport().Failed() || ginkgo.CurrentSpecReport().State.Is(ginkgotypes.SpecStateInterrupted) {
		return
	}

	f.DumpLogs()
}
