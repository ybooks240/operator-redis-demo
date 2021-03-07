/*
Copyright 2021 james.liu.

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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devv1 "github.com/ybooks240/api/v1"
)

// JamesRedisReconciler reconciles a JamesRedis object
type JamesRedisReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dev.ybooks240.github.com,resources=jamesredis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dev.ybooks240.github.com,resources=jamesredis/status,verbs=get;update;patch

func (r *JamesRedisReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("jamesredis", req.NamespacedName)

	log.Info("welcome to use jamesRedis")
	// your logic here

	// 1.首先获取实例
	var jamesRedis devv1.JamesRedis

	if err := r.Get(ctx, req.NamespacedName, &jamesRedis); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. 创建svc
	var svc corev1.Service
	svc.Name = jamesRedis.Name
	svc.Namespace = jamesRedis.Namespace

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r, &svc, func() error {
			log.Info("MutateService")
			MutateService(&jamesRedis, &svc)
			return controllerutil.SetControllerReference(&jamesRedis, &svc, r.Scheme)
		})
		log.Info("createOrUpdate", "service", or)
		return err
	}); err != nil {
		//log.Error(err, "认真查看", "ERROR")
		log.Info("认真查看", err)
		return ctrl.Result{}, nil

	}

	// 3. 创建sts
	var sts appsv1.StatefulSet
	sts.Name = jamesRedis.Name
	sts.Namespace = jamesRedis.Namespace

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r, &sts, func() error {
			log.Info("MutateStatefulSet")
			MutateStatefulSet(&jamesRedis, &sts)
			return controllerutil.SetControllerReference(&jamesRedis, &sts, r.Scheme)
		})
		log.Info("createOrUpdate", "statefulSet", or)
		return err
	}); err != nil {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *JamesRedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devv1.JamesRedis{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
