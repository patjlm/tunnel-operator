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
	"encoding/json"
	"errors"
	"time"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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

	CF := Cloudflare{ctx: ctx, log: log}
	api, err := CF.Api()
	// api, err := r.cloudflareApi()
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
			for _, hostname := range tunnel.Status.IngressHostnames {
				if err := CF.DeleteDNSRecords("CNAME", hostname); err != nil {
					return reconcile.Result{}, err
				}
			}
			if tunnel.Spec.Run {
				dep := r.deploymentForTunnelRun(tunnel)
				err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, dep)
				if err != nil && !apierrors.IsNotFound(err) {
					return reconcile.Result{}, err
				} else if err == nil {
					var zero int32 = 0
					dep.Spec.Replicas = &zero
					if err := r.Update(ctx, dep); err != nil {
						return reconcile.Result{}, err
					}
				}
			}
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
		secretB64 := CF.NewTunnelSecretB64()
		log.Info("creating cloudflare tunnel " + tunnel.Spec.Name)
		cfTunnel, err := api.CreateArgoTunnel(ctx, api.AccountID, tunnel.Spec.Name, secretB64)
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
		s := r.newTunnelSecret(tunnel, secretB64)
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
		// Create missing DNS records
		for _, ingress := range *tunnel.Spec.Ingress {
			if err := CF.CreateTunnelDNSRecord(ingress.HostName, tunnel); err != nil {
				return reconcile.Result{}, err
			}
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
				if err := CF.DeleteDNSRecords("CNAME", statusHostname); err != nil {
					return reconcile.Result{}, err
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

	if tunnel.Spec.Run {
		// Set the deploymentSpec in the Tunnel resource so it gets easy to be updated
		if tunnel.Spec.DeploymentSpec == nil {
			spec := tunnel.DefaultDeploymentSpec()
			tunnel.Spec.DeploymentSpec = &spec
			err := r.Update(ctx, tunnel)
			return reconcile.Result{}, err
		}
		found := &appsv1.Deployment{}
		err = r.Get(ctx, req.NamespacedName, found)
		if err != nil && apierrors.IsNotFound(err) {
			// Define a new deployment
			dep := r.deploymentForTunnelRun(tunnel)
			log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err = r.Create(ctx, dep)
			if err != nil {
				log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}

		if found.Labels["tunnel-id"] != tunnel.Status.TunnelID {
			dep := r.deploymentForTunnelRun(tunnel)
			err := r.Update(ctx, dep)
			if err != nil {
				log.Error(err, "failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
				return ctrl.Result{}, err
			}
			// Ask to requeue after 1 minute in order to give enough time for the
			// pods be created on the cluster side and the operand be able
			// to do the next update step accurately.
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
	} else {
		found := &appsv1.Deployment{}
		err = r.Get(ctx, req.NamespacedName, found)
		if err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "failed to get Deployment")
			return reconcile.Result{}, err
		}
		if err == nil {
			log.Info("deleting deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			if err = r.Delete(ctx, found); err != nil {
				return reconcile.Result{}, err
			}
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

func (r *TunnelReconciler) newTunnelSecret(t *tunnelv1alpha1.Tunnel, secretB64 string) *corev1.Secret {
	secret := t.BaseTunnelSecret()
	credentials := map[string]string{
		"AccountTag":   t.Status.AccountID,
		"TunnelID":     t.Status.TunnelID,
		"TunnelName":   t.Spec.Name,
		"TunnelSecret": secretB64,
	}
	credentialsJson, _ := json.Marshal(credentials)
	configYaml, _ := yaml.Marshal(tunnelConfig(t))
	secret.StringData = map[string]string{
		"credentials.json": string(credentialsJson),
		"config.yaml":      string(configYaml),
	}
	ctrl.SetControllerReference(t, secret, r.Scheme)
	return secret
}

func (r *TunnelReconciler) updateTunnelSecretConfig(ctx context.Context, t *tunnelv1alpha1.Tunnel) error {
	secret := t.BaseTunnelSecret()
	objectKey := client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}
	if err := r.Get(ctx, objectKey, secret); err != nil {
		return errors.New("failed to retrieve secret: " + err.Error())
	}
	configYaml, _ := yaml.Marshal(tunnelConfig(t))
	secret.Data["config.yaml"] = configYaml
	if err := r.Update(ctx, secret); err != nil {
		return errors.New("failed to update secret: " + err.Error())
	}
	return nil
}

func (r *TunnelReconciler) deploymentForTunnelRun(t *tunnelv1alpha1.Tunnel) *appsv1.Deployment {
	dep := t.DeploymentForTunnelRun()
	ctrl.SetControllerReference(t, dep, r.Scheme)
	return dep
}

// SetupWithManager sets up the controller with the Manager.
func (r *TunnelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tunnelv1alpha1.Tunnel{}).
		Owns(&corev1.Secret{}).
		Owns(&appsv1.Deployment{}).
		// WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
