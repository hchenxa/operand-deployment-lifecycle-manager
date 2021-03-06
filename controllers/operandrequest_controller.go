//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package controllers

import (
	"context"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	olmv1 "github.com/operator-framework/api/pkg/operators/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	operatorv1alpha1 "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
	"github.com/IBM/operand-deployment-lifecycle-manager/controllers/constant"
)

// OperandRequestReconciler reconciles a OperandRequest object
type OperandRequestReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}
type clusterObjects struct {
	namespace     *corev1.Namespace
	operatorGroup *olmv1.OperatorGroup
	subscription  *olmv1alpha1.Subscription
}

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *OperandRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// Fetch the OperandRequest instance
	requestInstance := &operatorv1alpha1.OperandRequest{}
	if err := r.Get(context.TODO(), req.NamespacedName, requestInstance); err != nil {
		// Error reading the object - requeue the request.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	klog.V(1).Infof("Reconciling OperandRequest: %s", req.NamespacedName)

	// Update labels for the request
	if requestInstance.UpdateLabels() {
		if err := r.Update(context.TODO(), requestInstance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Set the init status for OperandRequest instance
	if !requestInstance.InitRequestStatus() {
		if err := r.Status().Update(context.TODO(), requestInstance); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.addFinalizer(requestInstance); err != nil {
		return ctrl.Result{}, err
	}

	// Remove finalizer when DeletionTimestamp none zero
	if !requestInstance.ObjectMeta.DeletionTimestamp.IsZero() {

		// Check and clean up the subscriptions
		err := r.checkFinalizer(requestInstance)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Update finalizer to allow delete CR
		removed := requestInstance.RemoveFinalizer()
		if removed {
			err = r.Update(context.TODO(), requestInstance)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.reconcileOperator(req.NamespacedName); err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile the Operand
	merr := r.reconcileOperand(req.NamespacedName)

	if len(merr.Errors) != 0 {
		return ctrl.Result{}, merr
	}

	// Check if all csv deploy succeed
	if requestInstance.Status.Phase != operatorv1alpha1.ClusterPhaseRunning {
		klog.V(2).Info("Waiting for all operators and operands to be deployed successfully ...")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	klog.V(1).Infof("Finished reconciling OperandRequest: %s", req.NamespacedName)
	return ctrl.Result{}, nil
}

func (r *OperandRequestReconciler) addFinalizer(cr *operatorv1alpha1.OperandRequest) error {
	if cr.GetDeletionTimestamp() == nil {
		added := cr.EnsureFinalizer()
		if added {
			// Update CR
			err := r.Update(context.TODO(), cr)
			if err != nil {
				klog.Errorf("Failed to update the OperandRequest %s in the namespace %s: %s", cr.Name, cr.Namespace, err)
				return err
			}
		}
	}
	return nil
}

func (r *OperandRequestReconciler) checkFinalizer(requestInstance *operatorv1alpha1.OperandRequest) error {
	klog.V(2).Infof("Deleting OperandRequest %s in the namespace %s", requestInstance.Name, requestInstance.Namespace)
	existingSub := &olmv1alpha1.SubscriptionList{}

	opts := []client.ListOption{
		client.MatchingLabels(map[string]string{constant.OpreqLabel: "true"}),
	}

	if err := r.List(context.TODO(), existingSub, opts...); err != nil {
		return err
	}
	if len(existingSub.Items) == 0 {
		return nil
	}
	// Delete all the subscriptions that created by current request
	if err := r.absentOperatorsAndOperands(requestInstance); err != nil {
		return err
	}
	return nil
}

func getRegistryToRequestMapper(mgr manager.Manager) handler.ToRequestsFunc {
	return func(object handler.MapObject) []ctrl.Request {
		mgrClient := mgr.GetClient()
		requestList := &operatorv1alpha1.OperandRequestList{}
		opts := []client.ListOption{
			client.MatchingLabels(map[string]string{object.Meta.GetNamespace() + "." + object.Meta.GetName() + "/registry": "true"}),
		}

		_ = mgrClient.List(context.TODO(), requestList, opts...)

		requests := []ctrl.Request{}
		for _, request := range requestList.Items {
			namespaceName := types.NamespacedName{Name: request.Name, Namespace: request.Namespace}
			req := ctrl.Request{NamespacedName: namespaceName}
			requests = append(requests, req)
		}
		return requests
	}
}

func getConfigToRequestMapper(mgr manager.Manager) handler.ToRequestsFunc {
	return func(object handler.MapObject) []ctrl.Request {
		mgrClient := mgr.GetClient()
		requestList := &operatorv1alpha1.OperandRequestList{}
		opts := []client.ListOption{
			client.MatchingLabels(map[string]string{object.Meta.GetNamespace() + "." + object.Meta.GetName() + "/config": "true"}),
		}

		_ = mgrClient.List(context.TODO(), requestList, opts...)

		requests := []ctrl.Request{}
		for _, request := range requestList.Items {
			namespaceName := types.NamespacedName{Name: request.Name, Namespace: request.Namespace}
			req := ctrl.Request{NamespacedName: namespaceName}
			requests = append(requests, req)
		}
		return requests
	}
}

func (r *OperandRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.OperandRequest{}).
		Watches(&source.Kind{Type: &operatorv1alpha1.OperandRegistry{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: getRegistryToRequestMapper(mgr),
		}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldObject := e.ObjectOld.(*operatorv1alpha1.OperandRegistry)
				newObject := e.ObjectNew.(*operatorv1alpha1.OperandRegistry)
				return !reflect.DeepEqual(oldObject.Spec, newObject.Spec)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				// Evaluates to false if the object has been confirmed deleted.
				return !e.DeleteStateUnknown
			},
		})).
		Watches(&source.Kind{Type: &operatorv1alpha1.OperandConfig{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: getConfigToRequestMapper(mgr),
		}, builder.WithPredicates(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				// Evaluates to false if the object has been confirmed deleted.
				return !e.DeleteStateUnknown
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldObject := e.ObjectOld.(*operatorv1alpha1.OperandConfig)
				newObject := e.ObjectNew.(*operatorv1alpha1.OperandConfig)
				return !reflect.DeepEqual(oldObject.Spec, newObject.Spec)
			},
		})).Complete(r)
}
