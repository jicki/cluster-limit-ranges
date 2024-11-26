package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion 定义组和版本
	GroupVersion = schema.GroupVersion{Group: "jicki.cn", Version: "v1"}

	// SchemeBuilder 是 API 的构建器
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme 添加类型到 Scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

// 添加类型到 Scheme
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, &ClusterLimit{}, &ClusterLimitList{})
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
