/*
Copyright 2021.

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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
	"time"

	backwoodsv1 "github.com/backwoods-devops/archimedes/api/v1"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	conditionTypeConfigmapCreated = "ConfigmapCreated"
	conditionReasonCreated        = "Created"
	conditionReasonCreateFailed   = "CreateFailed"
	conditionReasonUpdated        = "Updated"
	conditionReasonUpdateFailed   = "UpdateFailed"
	conditionReasonMergeFailed    = "MergeFailed"
)

// ArchimedesPropertyReconciler reconciles a ArchimedesProperty object
type ArchimedesPropertyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=archimedes.backwoods-devops.io,resources=archimedesproperties,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=archimedes.backwoods-devops.io,resources=archimedesproperties/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=archimedes.backwoods-devops.io,resources=archimedesproperties/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core;coordination.k8s.io,resources=configmaps;leases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

//Reconcile is part of the main kubernetes reconciliation loop
func (r *ArchimedesPropertyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := r.Log.WithValues("archimedesproperty", req.NamespacedName)

	instance := &backwoodsv1.ArchimedesProperty{}

	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	commit, propTemplate, err := gitConfig(instance)
	if err != nil {
		log.Error(err, "Problem reading property template repo")
	}

	t := template.Must(template.New("properties").Parse(string(propTemplate)))
	sourceConfig := instance.Spec.SourceConfig
	cg := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(sourceConfig), &cg)
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	t.Execute(&tpl, cg)

	var data = make(map[string]string)
	data["commit"] = commit
	data["repoUrl"] = instance.Spec.RepoUrl
	data["revision"] = instance.Spec.Revision
	data["path"] = instance.Spec.PropertiesPath

	switch pt := instance.Spec.PropertyType; pt {
	case "kvp":
		scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(tpl.String())))
		for scanner.Scan() {
			s := strings.Split(scanner.Text(), "=")
			data[s[0]] = s[1]
		}
	case "key":
		if instance.Spec.KeyName != "" {
			data[instance.Spec.KeyName] = strings.TrimSpace(tpl.String())
		} else {
			err := errors.NewBadRequest("Missing keyName")
			log.Error(err, "Could not create Kubernetes configmap")
		}
	default:
		err := errors.NewBadRequest("Invalid PropertyType")
		log.Error(err, "Valid types (kvp, key).")
	}

	configmap, err := newConfigMap(instance, data)
	if err != nil {
		log.Error(err, "Could not create Kubernetes configmap")
		r.updateConditions(ctx, log, instance, conditionReasonCreateFailed, err.Error(), metav1.ConditionFalse)
		return ctrl.Result{}, err
	}
	// Set Archimedes Property instance as the owner and controller
	err = ctrl.SetControllerReference(instance, configmap, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Check if this ConfigMap already exists
	found := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: configmap.Name, Namespace: configmap.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Configmap", "Configmap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
		err = r.Create(ctx, configmap)
		if err != nil {
			log.Error(err, "Could not create configmap")
			r.updateConditions(ctx, log, instance, conditionReasonCreateFailed, err.Error(), metav1.ConditionFalse)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Could not create configmap")
		r.updateConditions(ctx, log, instance, conditionReasonCreateFailed, err.Error(), metav1.ConditionFalse)
		return ctrl.Result{}, err
	}
	log.Info("Updating a configmap", "Configmap.Namespace", configmap.Namespace, "Configmap.Name", configmap.Name)
	err = r.Update(ctx, configmap)
	if err != nil {
		log.Error(err, "Could not update configmap")
		r.updateConditions(ctx, log, instance, conditionReasonUpdateFailed, err.Error(), metav1.ConditionFalse)
		return ctrl.Result{}, err
	}
	r.updateConditions(ctx, log, instance, conditionReasonUpdated, "Configmap was updated", metav1.ConditionTrue)
	return ctrl.Result{}, nil
}

func newConfigMap(r *backwoodsv1.ArchimedesProperty, data map[string]string) (*corev1.ConfigMap, error) {
	labels := map[string]string{
		"created-by": "archimedes-property-operator",
	}
	for k, v := range r.ObjectMeta.Labels {
		labels[k] = v
	}
	annotations := map[string]string{}
	for k, v := range r.ObjectMeta.Annotations {
		annotations[k] = v
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.Spec.ConfigMapName,
			Namespace:   r.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Data: data,
	}, nil
}

func (r *ArchimedesPropertyReconciler) updateConditions(ctx context.Context, log logr.Logger, instance *backwoodsv1.ArchimedesProperty, reason, message string, status metav1.ConditionStatus) {
	instance.Status.Conditions = []metav1.Condition{{
		Type:               conditionTypeConfigmapCreated,
		Status:             status,
		ObservedGeneration: instance.GetGeneration(),
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             reason,
		Message:            message,
	}}
	err := r.Status().Update(ctx, instance)
	if err != nil {
		log.Error(err, "Could not update status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArchimedesPropertyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&backwoodsv1.ArchimedesProperty{}).
		Complete(r)
}

func gitConfig(r *backwoodsv1.ArchimedesProperty) (string, []byte, error) {

	dir, err := ioutil.TempDir("/tmp", "archimedes_")
	if err != nil {
		return "", nil, err
	}
	defer os.RemoveAll(dir)

	user := os.Getenv("USER")
	pass := os.Getenv("PASS")
	var certs []byte

	if _, err := os.Stat(r.Spec.CAPath); err == nil {
		certs, err = ioutil.ReadFile(r.Spec.CAPath)
		if err != nil {
			return "", nil, err
		}
	}

	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: r.Spec.RepoUrl,
		Auth: &http.BasicAuth{
			Username: user,
			Password: pass,
		},
		ReferenceName:     plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", r.Spec.Revision)),
		SingleBranch:      true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		CABundle:          certs,
	})
	if err != nil {
		return "", nil, err
	}
	ref, err := repo.Head()
	if err != nil {
		return "", nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", nil, err
	}

	propTemplate, err := ioutil.ReadFile(dir + "/" + r.Spec.PropertiesPath)
	if err != nil {
		return commit.Hash.String(), propTemplate, err
	}

	return commit.Hash.String(), propTemplate, nil

}
