package operator2

import (
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1 "github.com/openshift/api/config/v1"
)

// FIXME: we need to handle Authentication config object properly, namely:
// - honor Type field being set to none and don't create the OSIN
//   deployment in that case
// - the OAuthMetadata settings should be better respected in the code,
//   currently there is no special handling around it (see configmap.go).
// - the WebhookTokenAuthenticators field is currently not being handled
//   anywhere
//
// Note that the configMap from the reference in the OAuthMetadata field is
// used to fill the data in the /.well-known/oauth-authorization-server
// endpoint, but since that endpoint belongs to the apiserver, its syncing is
// handled in cluster-kube-apiserver-operator
func (c *authOperator) handleAuthConfigInner() (*configv1.Authentication, error) {
	// always make sure this function does not rely on defaulting from defaultAuthConfig

	authConfigNoDefaults, err := c.authentication.Get(globalConfigName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		authConfigNoDefaults, err = c.authentication.Create(&configv1.Authentication{
			ObjectMeta: defaultGlobalConfigMeta(),
		})
	}
	if err != nil {
		return nil, err
	}

	expectedReference := configv1.ConfigMapNameReference{
		Name: targetName,
	}

	if authConfigNoDefaults.Status.IntegratedOAuthMetadata == expectedReference {
		return authConfigNoDefaults, nil
	}

	authConfigNoDefaults.Status.IntegratedOAuthMetadata = expectedReference
	return c.authentication.UpdateStatus(authConfigNoDefaults)
}

func (c *authOperator) handleAuthConfig() (*configv1.Authentication, error) {
	auth, err := c.handleAuthConfigInner()
	if err != nil {
		return nil, err
	}
	return defaultAuthConfig(auth), nil
}

func defaultAuthConfig(authConfig *configv1.Authentication) *configv1.Authentication {
	out := authConfig.DeepCopy() // do not mutate informer cache

	if len(out.Spec.Type) == 0 {
		out.Spec.Type = configv1.AuthenticationTypeIntegratedOAuth
	}

	return out
}
