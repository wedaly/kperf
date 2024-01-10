package runner

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Azure/kperf/api/types"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// GroupHandler is to deploy job to run several runners with same load profile.
type GroupHandler struct {
	name      string
	namespace string
	uid       string

	count    int
	imageRef string
	profile  types.LoadProfile

	clientset kubernetes.Interface
}

// NewGroupHandler returns new instance of GroupHandler.
//
// The profileUrl input has two formats
//
//  1. file:///absolute_path?count=x
//  2. configmap:///configmap_name?count=x
func NewGroupHandler(clientset kubernetes.Interface, name, ns, profileUrl, imageRef string) (*GroupHandler, error) {
	u, err := url.Parse(profileUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid uri %s: %w", profileUrl, err)
	}

	var count int
	var profile types.LoadProfile

	switch u.Scheme {
	case "file":
		// TODO
	case "configmap":
		// TODO
	default:
		return nil, fmt.Errorf("unsupported scheme %s", u.Scheme)
	}

	return &GroupHandler{
		name:      name,
		namespace: ns,
		count:     count,
		profile:   profile,
		clientset: clientset,
	}, nil
}

// Deploy deploys a group of runners as a job if necessary.
func (h *GroupHandler) Deploy(ctx context.Context, uploadUrl string) error {
	// 1. Use client to check configmap named by h.name
	//   1.1 If not exist
	//	create configmap named by h.name
	//		the configmap has profile.yaml data (marshal h.profile into YAML)
	//   1.2 else
	//	check configmap and verify profile.yaml data is equal to h.profile
	//	if the data is not correct, return error
	// 2. Use client to check job named by h.name
	//    2.1 If not exist
	// 	create job named by h.name
	//    2.2 else
	//	check if the existing job spec is expected
	//	if not, return error
	// 3. Update h.uid = job.Uid
	//
	// NOTE: The job spec should be like
	/*
		apiVersion: batch/v1
		kind: Job
		metadata:
		  name: {{ h.name }}
		  namespace: {{ h.namespace }}
		spec:
		  completions: {{ h.count }}
		  parallelism: {{ h.count }}
		  template:
		    spec:
		      # TODO: affinity support
		      containers:
		      # FIXME:
		      #
		      # We should consider to use `--result` flag to upload data
		      # directly instead of using curl. When `--result=http://xyz:xxx`,
		      # we should upload data into that target url.
		      - args:
		        - kperf
		        - runner
		        - run
		        - --config=/data/config.yaml
		        - --user-agent=$(POD_NAME)
			- --result=/host/$(POD_NS)-$(POD_NAME)-$(POD_UID).json
			- && curl -X POST {{ uploadUrl }} -d @/host/$(POD_NS)-$(POD_NAME)-$(POD_UID).json
		        env:
		        - name: POD_NAME
		          valueFrom:
		            fieldRef:
		              fieldPath: metadata.name
		        - name: POD_UID
		          valueFrom:
		            fieldRef:
		              fieldPath: metadata.uid
		        - name: POD_NS
		          valueFrom:
		            fieldRef:
		              fieldPath: metadata.namespace
		        image: {{ h.image }}
		        imagePullPolicy: Always
		        name: runner
		        volumeMounts:
		        - mountPath: /data/
		          name: config
		        - mountPath: /host
		          name: host-root
		      restartPolicy: Never
		      # TODO: support serviceAccount/serviceAccountName
		      volumes:
		      - configMap:
		          name: {{ h.name }}
		        name: config
		      - hostPath:
		          path: /tmp
		          type: ""
		        name: host-root
	*/
	return fmt.Errorf("not implemented yet")
}

// Status returns the job's status.
func (h *GroupHandler) Status(ctx context.Context) (*batchv1.JobStatus, error) {
	// return the job named by h.name
	return nil, fmt.Errorf("not implemented yet")
}

// Pods returns all the pods controlled by the job.
func (h *GroupHandler) Pods(ctx context.Context) ([]*corev1.Pod, error) {
	// return all the pods controlled by the job.
	return nil, fmt.Errorf("not implemented yet")
}

// Profile returns load profile.
func (h *GroupHandler) Profile() types.LoadProfile {
	return h.profile
}
