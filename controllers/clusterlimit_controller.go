package controllers

import (
	"context"
	"fmt"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8scloudv1 "github.com/jicki/cluster-limit-ranges/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

type ClusterLimitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile 核心逻辑
func (r *ClusterLimitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var clusterLimit k8scloudv1.ClusterLimit
	if err := r.Get(ctx, req.NamespacedName, &clusterLimit); err != nil {
		if errors.IsNotFound(err) {
			// ClusterLimit 资源不存在，进行相关的清理操作
			logger.Info("ClusterLimit resource not found, cleaning up related LimitRanges")
			return r.cleanupLimitRanges(ctx)
		}
		// 其他错误，返回错误信息
		logger.Error(err, "unable to fetch ClusterLimit")
		return ctrl.Result{}, err
	}

	// 如果 ClusterLimit 在删除中，删除相关 LimitRange
	if !clusterLimit.DeletionTimestamp.IsZero() {
		logger.Info("ClusterLimit is being deleted, cleaning up related LimitRanges")
		return r.cleanupLimitRanges(ctx)
	}

	// 定时扫描所有命名空间，处理未添加的 LimitRange
	r.scanAndApplyLimitRanges(ctx, &clusterLimit)

	// 每30分钟扫描一次
	return ctrl.Result{RequeueAfter: 30 * time.Minute}, nil
}

// scanAndApplyLimitRanges 扫描所有命名空间，应用 LimitRange
func (r *ClusterLimitReconciler) scanAndApplyLimitRanges(ctx context.Context, clusterLimit *k8scloudv1.ClusterLimit) {
	logger := log.FromContext(ctx)

	// 获取所有命名空间
	var namespaces v1core.NamespaceList
	if err := r.List(ctx, &namespaces); err != nil {
		logger.Error(err, "unable to list namespaces")
		return
	}

	for _, ns := range namespaces.Items {
		namespace := ns.Name

		// 处理 includeNamespaces 和 excludeNamespaces
		if len(clusterLimit.Spec.IncludeNamespaces) > 0 && !contains(clusterLimit.Spec.IncludeNamespaces, namespace) {
			continue
		}
		if contains(clusterLimit.Spec.ExcludeNamespaces, namespace) {
			continue
		}

		// 检查是否已经存在 LimitRange
		var existingLimitRange v1core.LimitRange
		err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: "cluster-limit"}, &existingLimitRange)
		if errors.IsNotFound(err) {
			// 如果不存在，则创建新的 LimitRange
			logger.Info("Creating LimitRange for new namespace", "namespace", namespace)
			err = r.applyLimitRange(ctx, namespace, clusterLimit.Spec.Limits)
			if err != nil {
				logger.Error(err, "failed to create LimitRange", "namespace", namespace)
			}
		} else if err != nil {
			// 获取 LimitRange 时出现错误
			logger.Error(err, "failed to get existing LimitRange", "namespace", namespace)
		}
	}
}

// applyLimitRange 应用 LimitRange 到指定命名空间
func (r *ClusterLimitReconciler) applyLimitRange(ctx context.Context, namespace string, limits []k8scloudv1.LimitItem) error {
	logger := log.FromContext(ctx)

	limitRange := &v1core.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-limit",
			Namespace: namespace,
		},
		Spec: v1core.LimitRangeSpec{},
	}

	for _, limit := range limits {
		item := v1core.LimitRangeItem{
			Type:           v1core.LimitType(limit.Type),
			Default:        toResourceList(limit.Default),
			DefaultRequest: toResourceList(limit.DefaultRequest),
			Max:            toResourceList(limit.Max),
			Min:            toResourceList(limit.Min),
		}
		limitRange.Spec.Limits = append(limitRange.Spec.Limits, item)
	}

	// 创建或更新 LimitRange
	err := r.Client.Create(ctx, limitRange)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Info("LimitRange already exists, updating it", "namespace", namespace)
			return r.Client.Update(ctx, limitRange)
		}
		logger.Error(err, "failed to create LimitRange", "namespace", namespace)
		return err
	}
	logger.Info("Successfully created LimitRange", "namespace", namespace)
	return nil
}

// cleanupLimitRanges 清理由 ClusterLimit 创建的所有 LimitRange
func (r *ClusterLimitReconciler) cleanupLimitRanges(ctx context.Context) (ctrl.Result, error) {
	var limitRanges v1core.LimitRangeList
	if err := r.List(ctx, &limitRanges); err != nil {
		return ctrl.Result{}, err
	}

	for _, lr := range limitRanges.Items {
		// 检查是否是由该 ClusterLimit 创建的 LimitRange
		if lr.Name == "cluster-limit" {
			if err := r.Delete(ctx, &lr); err != nil {
				return ctrl.Result{}, err
			}
			fmt.Printf("Deleted LimitRange %s in namespace %s\n", lr.Name, lr.Namespace)
		}
	}

	return ctrl.Result{}, nil
}

// contains 判断字符串是否在数组中
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// toResourceList 将 map[string]string 转换为 v1core.ResourceList
func toResourceList(input map[string]string) v1core.ResourceList {
	result := make(v1core.ResourceList)
	for key, value := range input {
		result[v1core.ResourceName(key)] = resource.MustParse(value)
	}
	return result
}

// SetupWithManager 注册控制器
func (r *ClusterLimitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8scloudv1.ClusterLimit{}).
		Complete(r)
}
