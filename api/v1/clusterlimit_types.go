package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ClusterLimit 定义 ClusterLimit 的 CRD Schema
type ClusterLimit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterLimitSpec   `json:"spec,omitempty"`
	Status            ClusterLimitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterLimitList 是 ClusterLimit 的集合
type ClusterLimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLimit `json:"items"`
}

// ClusterLimitSpec 定义 LimitRange 配置
type ClusterLimitSpec struct {
	Limits            []LimitItem `json:"limits,omitempty"`
	IncludeNamespaces []string    `json:"includeNamespaces,omitempty"`
	ExcludeNamespaces []string    `json:"excludeNamespaces,omitempty"`
}

// ClusterLimitStatus 表示 CR 的状态
type ClusterLimitStatus struct {
	AppliedNamespaces []string `json:"appliedNamespaces,omitempty"`
}

// LimitItem 定义每个限制条目
type LimitItem struct {
	Type           string            `json:"type,omitempty"`
	Default        map[string]string `json:"default,omitempty"`
	DefaultRequest map[string]string `json:"defaultRequest,omitempty"`
	Max            map[string]string `json:"max,omitempty"`
	Min            map[string]string `json:"min,omitempty"`
}

// 手动实现 ClusterLimitSpec 的 DeepCopyInto 方法
func (in *ClusterLimitSpec) DeepCopyInto(out *ClusterLimitSpec) {
	*out = *in
	if in.Limits != nil {
		in, out := &in.Limits, &out.Limits
		*out = make([]LimitItem, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.IncludeNamespaces != nil {
		in, out := &in.IncludeNamespaces, &out.IncludeNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExcludeNamespaces != nil {
		in, out := &in.ExcludeNamespaces, &out.ExcludeNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// 手动实现 ClusterLimitStatus 的 DeepCopyInto 方法
func (in *ClusterLimitStatus) DeepCopyInto(out *ClusterLimitStatus) {
	*out = *in
	if in.AppliedNamespaces != nil {
		in, out := &in.AppliedNamespaces, &out.AppliedNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// 手动实现 LimitItem 的 DeepCopyInto 方法
func (in *LimitItem) DeepCopyInto(out *LimitItem) {
	*out = *in
	if in.Default != nil {
		in, out := &in.Default, &out.Default
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.DefaultRequest != nil {
		in, out := &in.DefaultRequest, &out.DefaultRequest
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Max != nil {
		in, out := &in.Max, &out.Max
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Min != nil {
		in, out := &in.Min, &out.Min
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// 手动实现 ClusterLimit 的 DeepCopyInto 方法
func (in *ClusterLimit) DeepCopyInto(out *ClusterLimit) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta) // 调用 ObjectMeta 的 DeepCopyInto 方法
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy 是 ClusterLimit 的深拷贝方法
func (in *ClusterLimit) DeepCopy() *ClusterLimit {
	if in == nil {
		return nil
	}
	out := new(ClusterLimit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject 手动实现 ClusterLimit 的 DeepCopyObject 方法
func (in *ClusterLimit) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// 手动实现 ClusterLimitList 的 DeepCopyInto 方法
func (in *ClusterLimitList) DeepCopyInto(out *ClusterLimitList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta) // 调用 ListMeta 的 DeepCopyInto 方法
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterLimit, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy 是 ClusterLimitList 的深拷贝方法
func (in *ClusterLimitList) DeepCopy() *ClusterLimitList {
	if in == nil {
		return nil
	}
	out := new(ClusterLimitList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject 手动实现 ClusterLimitList 的 DeepCopyObject 方法
func (in *ClusterLimitList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
