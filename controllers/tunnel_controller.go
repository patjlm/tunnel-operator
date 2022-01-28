/*
Copyright 2022.

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

package controllers

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/cloudflare/cloudflare-go"
	tunnelv1alpha1 "github.com/patjlm/tunnel-operator/api/v1alpha1"
)

// TunnelReconciler reconciles a Tunnel object
type TunnelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const tunnelFinalizer = "tunnel.zeeweb.xyz/finalizer"

//+kubebuilder:rbac:groups=tunnel.zeeweb.xyz,resources=tunnels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tunnel.zeeweb.xyz,resources=tunnels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tunnel.zeeweb.xyz,resources=tunnels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *TunnelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("reconciling")

	// Fetch the Tunnel instance
	tunnel := &tunnelv1alpha1.Tunnel{}
	err := r.Get(ctx, req.NamespacedName, tunnel)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Tunnel resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Tunnel")
		return ctrl.Result{}, err
	}

	api, err := r.cloudflareApi()
	if err != nil {
		log.Error(err, "could not initiate cloudflare client")
		return ctrl.Result{}, err
	}

	// Check if the Tunnel instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isToBeDeleted := tunnel.GetDeletionTimestamp() != nil
	if isToBeDeleted {
		if controllerutil.ContainsFinalizer(tunnel, tunnelFinalizer) {
			if tunnel.Status.TunnelID == "" {
				controllerutil.RemoveFinalizer(tunnel, tunnelFinalizer)
				err := r.Update(ctx, tunnel)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
			// Run finalization logic for Tunnel. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			log.Info("deleting tunnel " + tunnel.Status.TunnelID)
			if err := api.DeleteArgoTunnel(ctx, api.AccountID, tunnel.Status.TunnelID); err != nil {
				return ctrl.Result{}, err
			}

			// Remove tunnelFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(tunnel, tunnelFinalizer)
			err := r.Update(ctx, tunnel)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(tunnel, tunnelFinalizer) {
		controllerutil.AddFinalizer(tunnel, tunnelFinalizer)
		err = r.Update(ctx, tunnel)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Tunnel creation
	log.Info("looking up tunnel " + tunnel.Spec.Name)
	cfTunnels, err := api.ArgoTunnels(ctx, api.AccountID)
	if err != nil {
		log.Error(err, "Failed to retrieve the list of tunnels from Cloudflare")
		return reconcile.Result{}, err
	}
	exists := false
	for _, t := range cfTunnels {
		if t.Name == tunnel.Spec.Name && t.DeletedAt == nil {
			exists = true
		}
	}
	if !exists {
		secret := randomSecretB64(32)
		log.Info("creating cloudflare tunnel " + tunnel.Spec.Name)
		cfTunnel, err := api.CreateArgoTunnel(ctx, api.AccountID, tunnel.Spec.Name, secret)
		if err != nil {
			log.Error(err, "Failed to create cloudflare tunnel")
			apimeta.SetStatusCondition(&tunnel.Status.Conditions,
				metav1.Condition{
					Type:    tunnelv1alpha1.TunnelConditionCreatedType,
					Status:  metav1.ConditionFalse,
					Reason:  tunnelv1alpha1.TunnelConditionCreatedFailedReason,
					Message: "Cloudflare tunnel creation failed: " + err.Error(),
				})
			errStatus := r.Status().Update(ctx, tunnel)
			if errStatus != nil {
				log.Error(err, "Failed to update Tunnel status")
				return ctrl.Result{}, errStatus
			}
			return ctrl.Result{}, err
		}
		tunnel.Status.AccountID = api.AccountID
		tunnel.Status.TunnelID = cfTunnel.ID
		apimeta.SetStatusCondition(&tunnel.Status.Conditions,
			metav1.Condition{
				Type:    tunnelv1alpha1.TunnelConditionCreatedType,
				Status:  metav1.ConditionTrue,
				Reason:  tunnelv1alpha1.TunnelConditionCreatedSuccessReason,
				Message: "Cloudflare tunnel created successfully with ID " + cfTunnel.ID,
			})
		s := r.newTunnelSecret(tunnel, secret)
		if err := r.Create(ctx, s); err != nil {
			log.Error(err, "Failed to create tunnel secret")
			log.Info("deleting cloudflare tunnel " + tunnel.Status.TunnelID)
			_ = api.DeleteArgoTunnel(ctx, api.AccountID, tunnel.Status.TunnelID)
			return ctrl.Result{Requeue: true}, err
		}
		if err := r.Status().Update(ctx, tunnel); err != nil {
			log.Error(err, "Failed to update Tunnel status")
			log.Info("deleting cloudflare tunnel " + tunnel.Status.TunnelID)
			_ = api.DeleteArgoTunnel(ctx, api.AccountID, tunnel.Status.TunnelID)
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, err
	}

	tunnelID := tunnel.Status.TunnelID
	// our tunnel already exists but was not created by this resource
	// or the id is not yet filled in the conditions
	if tunnelID == "" {
		apimeta.SetStatusCondition(&tunnel.Status.Conditions,
			metav1.Condition{
				Type:    tunnelv1alpha1.TunnelConditionCreatedType,
				Status:  metav1.ConditionFalse,
				Reason:  tunnelv1alpha1.TunnelConditionCreatedFailedReason,
				Message: "Cloudflare tunnel already exists with name " + tunnel.Spec.Name,
			})
		err := r.Status().Update(ctx, tunnel)
		return ctrl.Result{}, err
	}

	if tunnel.Spec.Ingress != nil {
		proxied := true
		zoneName := os.Getenv("CLOUDFLARE_ZONE_NAME")
		zoneID, _ := api.ZoneIDByName(zoneName)
		// Create missing DNS records
		for _, ingress := range *tunnel.Spec.Ingress {
			tpl := cloudflare.DNSRecord{Name: ingress.HostName, Type: "CNAME"}
			records, err := api.DNSRecords(context.Background(), zoneID, tpl)
			if err != nil {
				log.Error(err, "failed to retrieve CNAME DNS recods from zone "+zoneName)
				return ctrl.Result{}, err
			}
			if len(records) == 0 {
				log.Info("creating cloudflare CNAME record for " + ingress.HostName)
				_, err := api.CreateDNSRecord(ctx, zoneID, cloudflare.DNSRecord{
					Type:    "CNAME",
					Name:    ingress.HostName,
					Proxied: &proxied,
					Content: tunnelID + ".cfargotunnel.com",
				})
				if err != nil {
					log.Error(err, "failed to created DNS record")
					return reconcile.Result{}, err
				}
			}
			// else if not good target or not proxied {
			// 	api.UpdateDNSRecord(ctx)
			// }
			recordedInStatus := inSlice(ingress.HostName, tunnel.Status.IngressHostnames)
			if !recordedInStatus {
				tunnel.Status.IngressHostnames = append(tunnel.Status.IngressHostnames, ingress.HostName)
				err := r.Status().Update(ctx, tunnel)
				return ctrl.Result{}, err
			}
		}
		updatedHostnames := false
		hostnames := []string{}
		for _, statusHostname := range tunnel.Status.IngressHostnames {
			found := false
			for _, ingress := range *tunnel.Spec.Ingress {
				if statusHostname == ingress.HostName {
					found = true
					break
				}
			}
			if !found {
				tpl := cloudflare.DNSRecord{Type: "CNAME", Name: statusHostname}
				records, err := api.DNSRecords(ctx, zoneID, tpl)
				if err != nil {
					log.Error(err, "failed to list DNS records matching "+statusHostname)
					return reconcile.Result{}, err
				}
				for _, record := range records {
					log.Info("deleting DNS CNAME record " + statusHostname)
					api.DeleteDNSRecord(ctx, zoneID, record.ID)
				}
				updatedHostnames = true
			} else {
				hostnames = append(hostnames, statusHostname)
			}
		}
		if updatedHostnames {
			tunnel.Status.IngressHostnames = hostnames
			err := r.Status().Update(ctx, tunnel)
			return reconcile.Result{}, err
		}
		if err := r.updateTunnelSecretConfig(ctx, tunnel); err != nil {
			return reconcile.Result{}, err
		}
	}

	log.Info("nothing to do")
	return ctrl.Result{}, nil
}

func inSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func (r *TunnelReconciler) cloudflareApi() (*cloudflare.API, error) {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	if token == "" {
		return nil, errors.New("missing environment variable CLOUDFLARE_API_TOKEN")
	}
	api, err := cloudflare.NewWithAPIToken(token)
	if api.AccountID == "" {
		api.AccountID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	}
	return api, err
}

func (r *TunnelReconciler) baseTunnelSecret(t *tunnelv1alpha1.Tunnel) *corev1.Secret {
	name := t.Name
	namespace := t.Namespace
	if t.Spec.TunnelSecret != nil {
		if t.Spec.TunnelSecret.Name != "" {
			name = t.Spec.TunnelSecret.Name
		}
		if t.Spec.TunnelSecret.Namespace != "" {
			namespace = t.Spec.TunnelSecret.Namespace
		}
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func (r *TunnelReconciler) newTunnelSecret(t *tunnelv1alpha1.Tunnel, secretB64 string) *corev1.Secret {
	secret := r.baseTunnelSecret(t)
	credentials := map[string]string{
		"AccountTag":   t.Status.AccountID,
		"TunnelID":     t.Status.TunnelID,
		"TunnelName":   t.Spec.Name,
		"TunnelSecret": secretB64,
	}
	credentialsJson, _ := json.Marshal(credentials)
	// 	configYaml := `---
	// ingress:
	// - service: http_status:404
	// tunnel: ` + t.Status.TunnelID
	configYaml, _ := yaml.Marshal(r.tunnelConfig(t))
	secret.StringData = map[string]string{
		"credentials.json": string(credentialsJson),
		"config.yaml":      string(configYaml),
	}
	ctrl.SetControllerReference(t, secret, r.Scheme)
	return secret
}

func (r *TunnelReconciler) updateTunnelSecretConfig(ctx context.Context, t *tunnelv1alpha1.Tunnel) error {
	secret := r.baseTunnelSecret(t)
	objectKey := client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}
	if err := r.Get(ctx, objectKey, secret); err != nil {
		return errors.New("failed to retrieve secret: " + err.Error())
	}
	configYaml, _ := yaml.Marshal(r.tunnelConfig(t))
	secret.Data["config.yaml"] = configYaml
	if err := r.Update(ctx, secret); err != nil {
		return errors.New("failed to update secret: " + err.Error())
	}
	return nil
}

func randomSecretB64(n int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()-_=+")
	b := make([]rune, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return b64.StdEncoding.EncodeToString([]byte(string(b)))
}

// SetupWithManager sets up the controller with the Manager.
func (r *TunnelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tunnelv1alpha1.Tunnel{}).
		Owns(&corev1.Secret{}).
		// WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}

type TunnelConfig struct {
	Tunnel  string                          `yaml:"tunnel"`
	Ingress *[]tunnelv1alpha1.TunnelIngress `yaml:"ingress"`
}

func (r *TunnelReconciler) tunnelConfig(t *tunnelv1alpha1.Tunnel) *TunnelConfig {
	ingresses := []tunnelv1alpha1.TunnelIngress{}
	if t.Spec.Ingress != nil {
		ingresses = append(ingresses, *t.Spec.Ingress...)
	}
	defaultIngress := "http_status:404"
	ingresses = append(ingresses, tunnelv1alpha1.TunnelIngress{Service: &defaultIngress})
	config := &TunnelConfig{
		Tunnel:  t.Status.TunnelID,
		Ingress: &ingresses,
	}
	return config
}
