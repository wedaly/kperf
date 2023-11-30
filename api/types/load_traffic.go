package types

// LoadProfile defines how to create load traffic from one host to kube-apiserver.
type LoadProfile struct {
	// Version defines the version of this object.
	Version int `json:"version" yaml:"version"`
	// Description is a string value to describe this object.
	Description string `json:"description,omitempty" yaml:"description"`
	// Spec defines behavior of load profile.
	Spec LoadProfileSpec `json:"spec" yaml:"spec"`
}

// LoadProfileSpec defines the load traffic for traget resource.
type LoadProfileSpec struct {
	// Rate defines the maximum requests per second (zero is no limit).
	Rate int `json:"rate" yaml:"rate"`
	// Total defines the total number of requests.
	Total int `json:"total" yaml:"total"`
	// Conns defines total number of long connections used for traffic.
	Conns int `json:"conns" yaml:"conns"`
	// Requests defines the different kinds of requests with weights.
	// The executor should randomly pick by weight.
	Requests []*WeightedRequest
}

// KubeTypeMeta represents metadata of kubernetes object.
type KubeTypeMeta struct {
	// Kind is a string value representing the REST resource the object represents.
	Kind string `json:"kind" yaml:"kind"`
	// APIVersion defines the versioned schema of the representation of an object.
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
}

// WeightedRequest represents request with weight.
// Only one of request types may be specified.
type WeightedRequest struct {
	// Shares defines weight in the same group.
	Shares int `json:"shares" yaml:"shares"`
	// StaleList means this list request with zero resource version.
	StaleList *RequestList `json:"staleList" yaml:"staleList"`
	// QuorumList means this list request without kube-apiserver cache.
	QuorumList *RequestList `json:"quorumList" yaml:"quorumList"`
	// StaleGet means this get request with zero resource version.
	StaleGet *RequestGet `json:"staleGet" yaml:"staleGet"`
	// QuorumGet means this get request without kube-apiserver cache.
	QuorumGet *RequestGet `json:"quorumGet" yaml:"quorumGet"`
	// Put means this is mutating request.
	Put *RequestPut `json:"put" yaml:"put"`
}

// RequestGet defines GET request for target object.
type RequestGet struct {
	// KubeTypeMeta represents object's resource type.
	KubeTypeMeta `yaml:",inline"`
	// Namespace is object's namespace.
	Namespace string `json:"namespace" yaml:"namespace"`
	// Name is object's name.
	Name string `json:"name" yaml:"name"`
}

// RequestList defines LIST request for target objects.
type RequestList struct {
	// KubeTypeMeta represents object's resource type.
	KubeTypeMeta `yaml:",inline"`
	// Namespace is object's namespace.
	Namespace string `json:"namespace" yaml:"namespace"`
	// Limit defines the page size.
	Limit int `json:"limit" yaml:"limit"`
	// Selector defines how to identify a set of objects.
	Selector string `json:"seletor" yaml:"seletor"`
}

// RequestPut defines PUT request for target resource type.
type RequestPut struct {
	// KubeTypeMeta represents object's resource type.
	//
	// NOTE: Currently, it should be configmap or secrets because we can
	// generate random bytes as blob for it. However, for the pod resource,
	// we need to ensure a lot of things are ready, for instance, volumes,
	// resource capacity. It's not easy to generate it randomly. Maybe we
	// can introduce pod template in the future.
	KubeTypeMeta `yaml:",inline"`
	// Namespace is object's namespace.
	Namespace string `json:"namespace" yaml:"namespace"`
	// Name is object's prefix name.
	Name string `json:"name" yaml:"name"`
	// KeySpaceSize is used to generate random number as name's suffix.
	KeySpaceSize int `json:"keySpaceSize" yaml:"keySpaceSize"`
	// ValueSize is the object's size in bytes.
	ValueSize int `json:"valueSize" yaml:"valueSize"`
}
