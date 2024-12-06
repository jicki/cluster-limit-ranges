package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sync"

	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8scloudv1 "github.com/jicki/cluster-limit-ranges/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	DefaultClusterLimitName = "global-limits"
	DefaultLimitRangeName   = "default-limitrange"
	ManagedLabelKey         = "clusterlimit"
	ManagedLabelValue       = "managed"
)

type ClusterLimitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile 核心逻辑
func (r *ClusterLimitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("ClusterLimit", req.NamespacedName)

	var clusterLimit k8scloudv1.ClusterLimit
	if err := r.Get(ctx, req.NamespacedName, &clusterLimit); err != nil {
		if errors.IsNotFound(err) {
			// 如果是资源被删除，则执行清理逻辑
			logger.Info("ClusterLimit resource not found, cleaning up related LimitRanges")
			return r.cleanupLimitRanges(ctx)
		}
		logger.Error(err, "Unable to fetch ClusterLimit")
		return ctrl.Result{}, err
	}

	// 如果资源在删除过程中
	if !clusterLimit.DeletionTimestamp.IsZero() {
		logger.Info("ClusterLimit is being deleted, cleaning up related LimitRanges")
		return r.cleanupLimitRanges(ctx)
	}

	// 扫描并应用 LimitRange
	r.scanAndApplyLimitRanges(ctx, &clusterLimit)
	logger.Info("Reconciling ClusterLimit completed")
	return ctrl.Result{}, nil
}

// scanAndApplyLimitRanges 扫描所有命名空间，应用 LimitRange
func (r *ClusterLimitReconciler) scanAndApplyLimitRanges(ctx context.Context, clusterLimit *k8scloudv1.ClusterLimit) {
	logger := log.FromContext(ctx)

	var namespaces v1core.NamespaceList
	if err := r.List(ctx, &namespaces); err != nil {
		logger.Error(err, "Unable to list namespaces")
		return
	}

	var wg sync.WaitGroup
	for _, ns := range namespaces.Items {
		namespace := ns.Name
		if len(clusterLimit.Spec.IncludeNamespaces) > 0 && !contains(clusterLimit.Spec.IncludeNamespaces, namespace) {
			continue
		}
		if contains(clusterLimit.Spec.ExcludeNamespaces, namespace) {
			continue
		}

		wg.Add(1)
		go func(namespace string) {
			defer wg.Done()
			if err := r.handleNamespaceEvent(ctx, namespace); err != nil {
				logger.Error(err, "Error processing namespace", "namespace", namespace)
			}
		}(namespace)
	}
	wg.Wait()
}

// handleNamespaceEvent 处理命名空间事件
func (r *ClusterLimitReconciler) handleNamespaceEvent(ctx context.Context, namespace string) error {
	logger := log.FromContext(ctx).WithValues("Namespace", namespace)

	// 检查命名空间是否已经存在任意 LimitRange
	var limitRanges v1core.LimitRangeList
	if err := r.List(ctx, &limitRanges, client.InNamespace(namespace)); err != nil {
		logger.Error(err, "Unable to list LimitRanges in namespace", "namespace", namespace)
		return err
	}
	if len(limitRanges.Items) > 0 {
		logger.Info("Namespace already has LimitRange(s), skipping creation", "namespace", namespace)
		return nil
	}

	// 获取所有 ClusterLimit
	var clusterLimits k8scloudv1.ClusterLimitList
	if err := r.List(ctx, &clusterLimits); err != nil {
		logger.Error(err, "Unable to list ClusterLimits")
		return err
	}

	for _, clusterLimit := range clusterLimits.Items {
		if len(clusterLimit.Spec.IncludeNamespaces) > 0 && !contains(clusterLimit.Spec.IncludeNamespaces, namespace) {
			continue
		}
		if contains(clusterLimit.Spec.ExcludeNamespaces, namespace) {
			continue
		}

		logger.Info("Creating LimitRange for namespace", "namespace", namespace)
		return r.createLimitRange(ctx, namespace, clusterLimit.Spec.Limits)
	}

	return nil
}

// createLimitRange 创建新的 LimitRange
func (r *ClusterLimitReconciler) createLimitRange(ctx context.Context, namespace string, limits []k8scloudv1.LimitItem) error {
	logger := log.FromContext(ctx).WithValues("Namespace", namespace)

	limitRange := &v1core.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultLimitRangeName,
			Namespace: namespace,
			Labels: map[string]string{
				ManagedLabelKey: ManagedLabelValue,
			},
		},
		Spec: v1core.LimitRangeSpec{
			// 填充默认配置
		},
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

	err := r.Client.Create(ctx, limitRange)
	if err != nil && !errors.IsAlreadyExists(err) {
		logger.Error(err, "Failed to create LimitRange", "namespace", namespace)
		return err
	}

	logger.Info("Successfully created LimitRange", "namespace", namespace)
	return nil
}

// cleanupLimitRanges 清理由 ClusterLimit 创建的所有 LimitRange
func (r *ClusterLimitReconciler) cleanupLimitRanges(ctx context.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var limitRanges v1core.LimitRangeList
	if err := r.List(ctx, &limitRanges, client.MatchingLabels{ManagedLabelKey: ManagedLabelValue}); err != nil {
		return ctrl.Result{}, err
	}

	for _, lr := range limitRanges.Items {
		if err := r.Delete(ctx, &lr); err != nil {
			logger.Error(err, "Failed to delete LimitRange", "namespace", lr.Namespace)
		} else {
			logger.Info("Deleted LimitRange", "namespace", lr.Namespace)
		}
	}
	return ctrl.Result{}, nil
}

// toResourceList 将 map[string]string 转换为 v1core.ResourceList
func toResourceList(input map[string]string) v1core.ResourceList {
	result := make(v1core.ResourceList)
	for key, value := range input {
		result[v1core.ResourceName(key)] = resource.MustParse(value)
	}
	return result
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// mapNamespaceToClusterLimit 实现 MapFunc
func mapNamespaceToClusterLimit(ctx context.Context, obj client.Object) []reconcile.Request {
	if _, ok := obj.(*v1core.Namespace); !ok {
		return nil
	}
	return []reconcile.Request{
		{NamespacedName: client.ObjectKey{Name: DefaultClusterLimitName}},
	}
}

// SetupWithManager 注册控制器
func (r *ClusterLimitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8scloudv1.ClusterLimit{}).
		Owns(&v1core.LimitRange{}).
		Watches(
			&v1core.Namespace{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				return mapNamespaceToClusterLimit(ctx, obj)
			}),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return true
				},
			}),
		).
		Complete(r)
}
