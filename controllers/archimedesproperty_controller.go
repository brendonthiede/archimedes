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

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	backwoodsv1 "github.com/backwoodsautomation/archimedes/api/v1"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-logr/logr"
)

const (
	conditionTypeConfigmapCreated = "ConfigmapCreated"
	//conditionReasonFetchFailed  = "FetchFailed"
	conditionReasonCreated      = "Created"
	conditionReasonCreateFailed = "CreateFailed"
	conditionReasonUpdated      = "Updated"
	conditionReasonUpdateFailed = "UpdateFailed"
	conditionReasonMergeFailed  = "MergeFailed"
)

// ArchimedesPropertyReconciler reconciles a ArchimedesProperty object
type ArchimedesPropertyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=backwoods.backwoodsautomation.com,resources=archimedesproperties,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=backwoods.backwoodsautomation.com,resources=archimedesproperties/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=backwoods.backwoodsautomation.com,resources=archimedesproperties/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ArchimedesProperty object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ArchimedesPropertyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("archimedesproperty", req.NamespacedName)
	log.Info("Starting Reconcile func")

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
	commit := gitConfig(instance.Spec.RepoUrl, instance.Spec.Revision)
	props, err := ioutil.ReadFile("/tmp/application/" + instance.Spec.PropertiesPath)
	if err != nil {
		log.Error(err, "Could not find property template in application repo")
	}
	t := template.Must(template.New("properties").Parse(string(props)))
	//mock data for testing should already be passed with instance from helm chart creating roos property k8s resource
	// yamlFile, err := ioutil.ReadFile("./config/samples/cluster-global.yaml")
	// if err != nil {
	//  log.Error(err, "Could not load sample cluster-global data")
	// }
	clusterGlobalInput := instance.Spec.SourceConfig
	log.Info(string(props))
	cg := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(clusterGlobalInput), &cg)
	if err != nil {
		panic(err)
	}
	fmt.Println("Source-Config: ")
	fmt.Println(cg)
	var tpl bytes.Buffer
	t.Execute(&tpl, cg)
	var data = make(map[string]string)
	data["commit"] = commit
	data["repoUrl"] = instance.Spec.RepoUrl
	data["revision"] = instance.Spec.Revision
	data["path"] = instance.Spec.PropertiesPath
	scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(tpl.String())))
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), "=")
		data[s[0]] = s[1]
	}
	configmap, err := newConfigMap(instance, data)
	if err != nil {
		// Error while creating the Kubernetes configmap - requeue the request.
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
			Name:        r.Name,
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

func gitConfig(url, revision string) string {
	dir := "/tmp/application"
	err := os.RemoveAll(dir)
	if err != nil {
		fmt.Println(err)
	}
	user := os.Getenv("USER")
	pass := os.Getenv("PASS")
	certs, err := ioutil.ReadFile("/etc/archimedes-property-operator/ca.crt")
	if err != nil {
		fmt.Println(err)
	}
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
		Auth: &http.BasicAuth{
			Username: user,
			Password: pass,
		},
		ReferenceName:     plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", revision)),
		SingleBranch:      true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		CABundle:          certs,
	})
	if err != nil {
		fmt.Println(err)
	}
	ref, err := r.Head()
	if err != nil {
		fmt.Println(err)
	}
	commit, err := r.CommitObject(ref.Hash())
	return commit.Hash.String()
}
