package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestLoadProfileUnmarshalFromYAML(t *testing.T) {
	in := `
version: 1
description: test
spec:
  rate: 100
  total: 10000
  conns: 2
  requests:
  - staleGet:
      group: core
      version: v1
      resource: pods
      namespace: default
      name: x1
    shares: 100
  - quorumGet:
      group: core
      version: v1
      resource: configmaps
      namespace: default
      name: x2
    shares: 150
  - staleList:
      group: core
      version: v1
      resource: pods
      namespace: default
      limit: 10000
      seletor: app=x2
    shares: 200
  - quorumList:
      group: core
      version: v1
      resource: configmaps
      namespace: default
      limit: 10000
      seletor: app=x3
    shares: 400
  - put:
      group: core
      version: v1
      resource: configmaps
      namespace: kperf
      name: kperf-
      keySpaceSize: 1000
      valueSize: 1024
    shares: 1000
`

	target := LoadProfile{}
	require.NoError(t, yaml.Unmarshal([]byte(in), &target))
	assert.Equal(t, 1, target.Version)
	assert.Equal(t, "test", target.Description)
	assert.Equal(t, 100, target.Spec.Rate)
	assert.Equal(t, 10000, target.Spec.Total)
	assert.Equal(t, 2, target.Spec.Conns)
	assert.Len(t, target.Spec.Requests, 5)

	assert.Equal(t, 100, target.Spec.Requests[0].Shares)
	assert.NotNil(t, target.Spec.Requests[0].StaleGet)
	assert.Equal(t, "pods", target.Spec.Requests[0].StaleGet.Resource)
	assert.Equal(t, "v1", target.Spec.Requests[0].StaleGet.Version)
	assert.Equal(t, "core", target.Spec.Requests[0].StaleGet.Group)
	assert.Equal(t, "default", target.Spec.Requests[0].StaleGet.Namespace)
	assert.Equal(t, "x1", target.Spec.Requests[0].StaleGet.Name)

	assert.NotNil(t, target.Spec.Requests[1].QuorumGet)
	assert.Equal(t, 150, target.Spec.Requests[1].Shares)

	assert.Equal(t, 200, target.Spec.Requests[2].Shares)
	assert.NotNil(t, target.Spec.Requests[2].StaleList)
	assert.Equal(t, "pods", target.Spec.Requests[2].StaleList.Resource)
	assert.Equal(t, "v1", target.Spec.Requests[2].StaleList.Version)
	assert.Equal(t, "core", target.Spec.Requests[0].StaleGet.Group)
	assert.Equal(t, "default", target.Spec.Requests[2].StaleList.Namespace)
	assert.Equal(t, 10000, target.Spec.Requests[2].StaleList.Limit)
	assert.Equal(t, "app=x2", target.Spec.Requests[2].StaleList.Selector)

	assert.NotNil(t, target.Spec.Requests[3].QuorumList)
	assert.Equal(t, 400, target.Spec.Requests[3].Shares)

	assert.Equal(t, 1000, target.Spec.Requests[4].Shares)
	assert.NotNil(t, target.Spec.Requests[4].Put)
	assert.Equal(t, "configmaps", target.Spec.Requests[4].Put.Resource)
	assert.Equal(t, "v1", target.Spec.Requests[4].Put.Version)
	assert.Equal(t, "core", target.Spec.Requests[0].StaleGet.Group)
	assert.Equal(t, "kperf", target.Spec.Requests[4].Put.Namespace)
	assert.Equal(t, "kperf-", target.Spec.Requests[4].Put.Name)
	assert.Equal(t, 1000, target.Spec.Requests[4].Put.KeySpaceSize)
	assert.Equal(t, 1024, target.Spec.Requests[4].Put.ValueSize)
}

func TestWeightedRequest(t *testing.T) {
	for _, r := range []struct {
		name   string
		req    *WeightedRequest
		hasErr bool
	}{
		{
			name:   "shares < 0",
			req:    &WeightedRequest{Shares: -1},
			hasErr: true,
		},
		{
			name:   "no request setting",
			req:    &WeightedRequest{Shares: 10},
			hasErr: true,
		},
		{
			name: "empty version",
			req: &WeightedRequest{
				Shares: 10,
				StaleGet: &RequestGet{
					KubeGroupVersionResource: KubeGroupVersionResource{
						Resource: "pods",
					},
				},
			},
			hasErr: true,
		},
		{
			name: "empty resource",
			req: &WeightedRequest{
				Shares: 10,
				StaleGet: &RequestGet{
					KubeGroupVersionResource: KubeGroupVersionResource{
						Group:   "core",
						Version: "v1",
					},
				},
			},
			hasErr: true,
		},
		{
			name: "wrong limit",
			req: &WeightedRequest{
				Shares: 10,
				StaleList: &RequestList{
					KubeGroupVersionResource: KubeGroupVersionResource{
						Group:    "core",
						Version:  "v1",
						Resource: "pods",
					},
					Limit: -1,
				},
			},
			hasErr: true,
		},
		{
			name: "no error",
			req: &WeightedRequest{
				Shares: 10,
				StaleGet: &RequestGet{
					KubeGroupVersionResource: KubeGroupVersionResource{
						Group:    "core",
						Version:  "v1",
						Resource: "pods",
					},
					Namespace: "default",
					Name:      "testing",
				},
			},
		},
	} {
		r := r
		t.Run(r.name, func(t *testing.T) {
			err := r.req.Validate()
			if r.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
