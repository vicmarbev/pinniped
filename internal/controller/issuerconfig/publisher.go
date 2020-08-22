/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package issuerconfig

import (
	"context"
	"encoding/base64"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/suzerain-io/controller-go"
	pinnipedcontroller "github.com/suzerain-io/pinniped/internal/controller"
	crdpinnipedv1alpha1 "github.com/suzerain-io/pinniped/kubernetes/1.19/api/apis/crdpinniped/v1alpha1"
	pinnipedclientset "github.com/suzerain-io/pinniped/kubernetes/1.19/client-go/clientset/versioned"
	crdpinnipedv1alpha1informers "github.com/suzerain-io/pinniped/kubernetes/1.19/client-go/informers/externalversions/crdpinniped/v1alpha1"
)

const (
	ClusterInfoNamespace = "kube-public"

	clusterInfoName         = "cluster-info"
	clusterInfoConfigMapKey = "kubeconfig"

	configName = "pinniped-config"
)

type publisherController struct {
	namespace                      string
	serverOverride                 *string
	pinnipedClient                 pinnipedclientset.Interface
	configMapInformer              corev1informers.ConfigMapInformer
	credentialIssuerConfigInformer crdpinnipedv1alpha1informers.CredentialIssuerConfigInformer
}

func NewPublisherController(
	namespace string,
	serverOverride *string,
	pinnipedClient pinnipedclientset.Interface,
	configMapInformer corev1informers.ConfigMapInformer,
	credentialIssuerConfigInformer crdpinnipedv1alpha1informers.CredentialIssuerConfigInformer,
	withInformer pinnipedcontroller.WithInformerOptionFunc,
) controller.Controller {
	return controller.New(
		controller.Config{
			Name: "publisher-controller",
			Syncer: &publisherController{
				namespace:                      namespace,
				serverOverride:                 serverOverride,
				pinnipedClient:                 pinnipedClient,
				configMapInformer:              configMapInformer,
				credentialIssuerConfigInformer: credentialIssuerConfigInformer,
			},
		},
		withInformer(
			configMapInformer,
			pinnipedcontroller.NameAndNamespaceExactMatchFilterFactory(clusterInfoName, ClusterInfoNamespace),
			controller.InformerOption{},
		),
		withInformer(
			credentialIssuerConfigInformer,
			pinnipedcontroller.NameAndNamespaceExactMatchFilterFactory(configName, namespace),
			controller.InformerOption{},
		),
	)
}

func (c *publisherController) Sync(ctx controller.Context) error {
	configMap, err := c.configMapInformer.
		Lister().
		ConfigMaps(ClusterInfoNamespace).
		Get(clusterInfoName)
	notFound := k8serrors.IsNotFound(err)
	if err != nil && !notFound {
		return fmt.Errorf("failed to get %s configmap: %w", clusterInfoName, err)
	}
	if notFound {
		klog.InfoS(
			"could not find config map",
			"configmap",
			klog.KRef(ClusterInfoNamespace, clusterInfoName),
		)
		return nil
	}

	kubeConfig, kubeConfigPresent := configMap.Data[clusterInfoConfigMapKey]
	if !kubeConfigPresent {
		klog.InfoS("could not find kubeconfig configmap key")
		return nil
	}

	config, _ := clientcmd.Load([]byte(kubeConfig))

	var certificateAuthorityData, server string
	for _, v := range config.Clusters {
		certificateAuthorityData = base64.StdEncoding.EncodeToString(v.CertificateAuthorityData)
		server = v.Server
		break
	}

	if c.serverOverride != nil {
		server = *c.serverOverride
	}

	credentialIssuerConfig := crdpinnipedv1alpha1.CredentialIssuerConfig{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: c.namespace,
		},
		Status: crdpinnipedv1alpha1.CredentialIssuerConfigStatus{
			Strategies: []crdpinnipedv1alpha1.CredentialIssuerConfigStrategy{},
			KubeConfigInfo: &crdpinnipedv1alpha1.CredentialIssuerConfigKubeConfigInfo{
				Server:                   server,
				CertificateAuthorityData: certificateAuthorityData,
			},
		},
	}
	if err := c.createOrUpdateCredentialIssuerConfig(ctx.Context, &credentialIssuerConfig); err != nil {
		return err
	}

	return nil
}

func (c *publisherController) createOrUpdateCredentialIssuerConfig(
	ctx context.Context,
	newCredentialIssuerConfig *crdpinnipedv1alpha1.CredentialIssuerConfig,
) error {
	existingCredentialIssuerConfig, err := c.credentialIssuerConfigInformer.
		Lister().
		CredentialIssuerConfigs(c.namespace).
		Get(newCredentialIssuerConfig.Name)
	notFound := k8serrors.IsNotFound(err)
	if err != nil && !notFound {
		return fmt.Errorf("could not get credentialissuerconfig: %w", err)
	}

	credentialIssuerConfigsClient := c.pinnipedClient.CrdV1alpha1().CredentialIssuerConfigs(c.namespace)
	if notFound {
		if _, err := credentialIssuerConfigsClient.Create(
			ctx,
			newCredentialIssuerConfig,
			metav1.CreateOptions{},
		); err != nil {
			return fmt.Errorf("could not create credentialissuerconfig: %w", err)
		}
	} else if !equal(existingCredentialIssuerConfig, newCredentialIssuerConfig) {
		// Update just the fields we care about.
		newServer := newCredentialIssuerConfig.Status.KubeConfigInfo.Server
		newCA := newCredentialIssuerConfig.Status.KubeConfigInfo.CertificateAuthorityData
		existingCredentialIssuerConfig.Status.KubeConfigInfo.Server = newServer
		existingCredentialIssuerConfig.Status.KubeConfigInfo.CertificateAuthorityData = newCA

		if _, err := credentialIssuerConfigsClient.Update(
			ctx,
			existingCredentialIssuerConfig,
			metav1.UpdateOptions{},
		); err != nil {
			return fmt.Errorf("could not update credentialissuerconfig: %w", err)
		}
	}

	return nil
}

func equal(a, b *crdpinnipedv1alpha1.CredentialIssuerConfig) bool {
	return a.Status.KubeConfigInfo.Server == b.Status.KubeConfigInfo.Server &&
		a.Status.KubeConfigInfo.CertificateAuthorityData == b.Status.KubeConfigInfo.CertificateAuthorityData
}
