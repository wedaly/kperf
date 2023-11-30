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
      kind: pods
      apiVersion: v1
      namespace: default
      name: x1
    shares: 100
  - quorumGet:
      kind: configmap
      apiVersion: v1
      namespace: default
      name: x2
    shares: 150
  - staleList:
      kind: pods
      apiVersion: v1
      namespace: default
      limit: 10000
      seletor: app=x2
    shares: 200
  - quorumList:
      kind: configmap
      apiVersion: v1
      namespace: default
      limit: 10000
      seletor: app=x3
    shares: 400
  - put:
      kind: configmap
      apiVersion: v1
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
	assert.Equal(t, "pods", target.Spec.Requests[0].StaleGet.Kind)
	assert.Equal(t, "v1", target.Spec.Requests[0].StaleGet.APIVersion)
	assert.Equal(t, "default", target.Spec.Requests[0].StaleGet.Namespace)
	assert.Equal(t, "x1", target.Spec.Requests[0].StaleGet.Name)

	assert.NotNil(t, target.Spec.Requests[1].QuorumGet)
	assert.Equal(t, 150, target.Spec.Requests[1].Shares)

	assert.Equal(t, 200, target.Spec.Requests[2].Shares)
	assert.NotNil(t, target.Spec.Requests[2].StaleList)
	assert.Equal(t, "pods", target.Spec.Requests[2].StaleList.Kind)
	assert.Equal(t, "v1", target.Spec.Requests[2].StaleList.APIVersion)
	assert.Equal(t, "default", target.Spec.Requests[2].StaleList.Namespace)
	assert.Equal(t, 10000, target.Spec.Requests[2].StaleList.Limit)
	assert.Equal(t, "app=x2", target.Spec.Requests[2].StaleList.Selector)

	assert.NotNil(t, target.Spec.Requests[3].QuorumList)
	assert.Equal(t, 400, target.Spec.Requests[3].Shares)

	assert.Equal(t, 1000, target.Spec.Requests[4].Shares)
	assert.NotNil(t, target.Spec.Requests[4].Put)
	assert.Equal(t, "configmap", target.Spec.Requests[4].Put.Kind)
	assert.Equal(t, "v1", target.Spec.Requests[4].Put.APIVersion)
	assert.Equal(t, "kperf", target.Spec.Requests[4].Put.Namespace)
	assert.Equal(t, "kperf-", target.Spec.Requests[4].Put.Name)
	assert.Equal(t, 1000, target.Spec.Requests[4].Put.KeySpaceSize)
	assert.Equal(t, 1024, target.Spec.Requests[4].Put.ValueSize)
}
