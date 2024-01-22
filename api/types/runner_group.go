package types

// RunnerGroup defines a set of runners with same load profile.
type RunnerGroup struct {
	// Name is the name of runner group.
	Name string `json:"name" yaml:"name"`
	// Spec is specification of the desired behavior of the runner group.
	Spec *RunnerGroupSpec `json:"spec" yaml:"spec"`
}

// RunnerGroupSpec is to descibe how the runner group works.
type RunnerGroupSpec struct {
	// Count is the number of runners.
	Count int32 `json:"count" yaml:"count"`
	// Profile defines what the load traffic looks like.
	Profile *LoadProfile `json:"loadProfile,omitempty" yaml:"loadProfile"`
	// NodeAffinity defines how to deploy runners into dedicated nodes
	// which have specific labels.
	NodeAffinity map[string][]string `json:"nodeAffinity,omitempty" yaml:"nodeAffinity,omitempty"`
	// ServiceAccount is the name of the ServiceAccount to use to run runners.
	ServiceAccount *string `json:"serviceAccount,omitempty" yaml:"serviceAccount,omitempty"`
	// OwnerReference is to mark the runner group depending on this object.
	//
	// FORMAT: APIVersion:Kind:Name:UID
	OwnerReference *string `json:"ownerReference,omitempty" yaml:"ownerReference,omitempty"`
}
