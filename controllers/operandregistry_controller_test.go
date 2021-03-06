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

// import (
// 	"context"
// 	"testing"

// 	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/types"
// 	"k8s.io/client-go/kubernetes/scheme"
// 	"k8s.io/client-go/tools/record"
// 	"sigs.k8s.io/controller-runtime/pkg/client/fake"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"

// 	v1alpha1 "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
// )

// // TestRegistryController runs OperandRegistryReconciler.Reconcile() against a
// // fake client that tracks a OperandRegistry object.
// func TestRegistryController(t *testing.T) {
// 	var (
// 		name              = "common-service"
// 		namespace         = "ibm-common-service"
// 		operatorNamespace = "ibm-operators"
// 	)

// 	req := getReconcileRequest(name, namespace)
// 	r := getReconciler(name, namespace, operatorNamespace)

// 	initReconcile(t, r, req)

// }

// func initReconcile(t *testing.T, r OperandRegistryReconciler, req reconcile.Request) {
// 	assert := assert.New(t)

// 	_, err := r.Reconcile(req)
// 	assert.NoError(err)

// 	registry := &v1alpha1.OperandRegistry{}
// 	err = r.Get(context.TODO(), req.NamespacedName, registry)
// 	assert.NoError(err)
// 	assert.Equalf(v1alpha1.RegistryReady, registry.Status.Phase, "OperandRegistry(%s) phase should be %s", req.NamespacedName, v1alpha1.RegistryReady)

// }

// func getReconciler(name, namespace, operatorNamespace string) OperandRegistryReconciler {
// 	s := scheme.Scheme
// 	v1alpha1.SchemeBuilder.AddToScheme(s)
// 	corev1.SchemeBuilder.AddToScheme(s)
// 	olmv1alpha1.SchemeBuilder.AddToScheme(s)

// 	initData := initClientData(name, namespace, operatorNamespace)

// 	// Create a fake client to mock API calls.
// 	client := fake.NewFakeClient(initData.objs...)

// 	// Return a OperandRegistryReconciler object with the scheme and fake client.
// 	return OperandRegistryReconciler{
// 		Scheme:   s,
// 		Client:   client,
// 		Recorder: record.NewFakeRecorder(10),
// 	}
// }

// // Mock request to simulate Reconcile() being called on an event for a watched resource
// func getReconcileRequest(name, namespace string) reconcile.Request {
// 	return reconcile.Request{
// 		NamespacedName: types.NamespacedName{
// 			Name:      name,
// 			Namespace: namespace,
// 		},
// 	}
// }

// type DataObj struct {
// 	objs []runtime.Object
// }

// func initClientData(name, namespace, operatorNamespace string) *DataObj {
// 	return &DataObj{
// 		objs: []runtime.Object{
// 			operandRegistry(name, namespace, operatorNamespace),
// 			catalogSource(),
// 		},
// 	}
// }

// // Return OperandRegistry obj
// func operandRegistry(name, namespace, operatorNamespace string) *v1alpha1.OperandRegistry {
// 	return &v1alpha1.OperandRegistry{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      name,
// 			Namespace: namespace,
// 		},
// 		Spec: v1alpha1.OperandRegistrySpec{
// 			Operators: []v1alpha1.Operator{
// 				{
// 					Name:            "etcd",
// 					Namespace:       operatorNamespace,
// 					SourceName:      "community-operators",
// 					SourceNamespace: "openshift-marketplace",
// 					PackageName:     "etcd",
// 					Channel:         "singlenamespace-alpha",
// 				},
// 				{
// 					Name:            "jenkins",
// 					Namespace:       operatorNamespace,
// 					SourceName:      "community-operators",
// 					SourceNamespace: "openshift-marketplace",
// 					PackageName:     "jenkins-operator",
// 					Channel:         "alpha",
// 				},
// 			},
// 		},
// 	}
// }

// func catalogSource() *olmv1alpha1.CatalogSource {
// 	return &olmv1alpha1.CatalogSource{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "community-operators",
// 			Namespace: "openshift-marketplace",
// 		},
// 		Spec: olmv1alpha1.CatalogSourceSpec{
// 			Address:     "community-operators.openshift-marketplace.svc:50051",
// 			DisplayName: "Community Operators",
// 			Icon: olmv1alpha1.Icon{
// 				Data:      "",
// 				MediaType: "",
// 			},
// 			Publisher:  "Red Hat",
// 			SourceType: "grpc",
// 		},
// 		Status: olmv1alpha1.CatalogSourceStatus{
// 			GRPCConnectionState: &olmv1alpha1.GRPCConnectionState{
// 				Address:           "community-operators.openshift-marketplace.svc:50051",
// 				LastObservedState: "READY",
// 			},
// 			RegistryServiceStatus: &olmv1alpha1.RegistryServiceStatus{
// 				Protocol: "grpc",
// 			},
// 		},
// 	}
// }
