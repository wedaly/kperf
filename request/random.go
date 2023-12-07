package request

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"github.com/Azure/kperf/api/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// WeightedRandomRequests is used to generate requests based on LoadProfileSpec.
type WeightedRandomRequests struct {
	once         sync.Once
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	reqBuilderCh chan RequestBuilder

	shares      []int
	reqBuilders []RequestBuilder
}

// NewWeightedRandomRequests creates new instance of WeightedRandomRequests.
func NewWeightedRandomRequests(spec *types.LoadProfileSpec) (*WeightedRandomRequests, error) {
	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid load profile spec: %v", err)
	}

	shares := make([]int, 0, len(spec.Requests))
	reqBuilders := make([]RequestBuilder, 0, len(spec.Requests))
	for _, r := range spec.Requests {
		shares = append(shares, r.Shares)

		var builder RequestBuilder
		switch {
		case r.StaleList != nil:
			builder = newRequestListBuilder(r.StaleList, "0")
		case r.QuorumList != nil:
			builder = newRequestListBuilder(r.QuorumList, "")
		case r.StaleGet != nil:
			builder = newRequestGetBuilder(r.StaleGet, "0")
		case r.QuorumGet != nil:
			builder = newRequestGetBuilder(r.QuorumGet, "")
		default:
			return nil, fmt.Errorf("only support get/list")
		}
		reqBuilders = append(reqBuilders, builder)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &WeightedRandomRequests{
		ctx:          ctx,
		cancel:       cancel,
		reqBuilderCh: make(chan RequestBuilder),
		shares:       shares,
		reqBuilders:  reqBuilders,
	}, nil
}

// Run starts to random pick request.
func (r *WeightedRandomRequests) Run(ctx context.Context, total int) {
	defer r.wg.Done()
	r.wg.Add(1)

	sum := 0
	for sum < total {
		builder := r.randomPick()
		select {
		case r.reqBuilderCh <- builder:
			sum += 1
		case <-r.ctx.Done():
			return
		case <-ctx.Done():
			return
		}
	}
}

// Chan returns channel to get random request.
func (r *WeightedRandomRequests) Chan() chan RequestBuilder {
	return r.reqBuilderCh
}

func (r *WeightedRandomRequests) randomPick() RequestBuilder {
	sum := 0
	for _, s := range r.shares {
		sum += s
	}

	rndInt, err := rand.Int(rand.Reader, big.NewInt(int64(sum)))
	if err != nil {
		panic(err)
	}

	rnd := rndInt.Int64()
	for i := range r.shares {
		s := int64(r.shares[i])
		if rnd < s {
			return r.reqBuilders[i]
		}
		rnd -= s
	}
	panic("unreachable")
}

// Stop stops request generator.
func (r *WeightedRandomRequests) Stop() {
	r.once.Do(func() {
		r.cancel()
		r.wg.Wait()
		close(r.reqBuilderCh)
	})
}

// RequestBuilder is used to build rest.Request.
type RequestBuilder interface {
	Build(cli rest.Interface) (method string, _ *rest.Request)
}

type requestGetBuilder struct {
	version         schema.GroupVersion
	resource        string
	namespace       string
	name            string
	resourceVersion string
}

func newRequestGetBuilder(src *types.RequestGet, resourceVersion string) *requestGetBuilder {
	return &requestGetBuilder{
		version: schema.GroupVersion{
			Group:   src.Group,
			Version: src.Version,
		},
		resource:        src.Resource,
		namespace:       src.Namespace,
		name:            src.Name,
		resourceVersion: resourceVersion,
	}
}

// Build implements RequestBuilder.Build.
func (b *requestGetBuilder) Build(cli rest.Interface) (string, *rest.Request) {
	// https://kubernetes.io/docs/reference/using-api/#api-groups
	apiPath := "apis"
	if b.version.Group == "" {
		apiPath = "api"
	}

	comps := make([]string, 2, 5)
	comps[0], comps[1] = apiPath, b.version.Version
	if b.namespace != "" {
		comps = append(comps, "namespaces", b.namespace)
	}
	comps = append(comps, b.resource, b.name)

	return "GET", cli.Get().AbsPath(comps...).
		SpecificallyVersionedParams(
			&metav1.GetOptions{ResourceVersion: b.resourceVersion},
			scheme.ParameterCodec,
			schema.GroupVersion{Version: "v1"},
		)
}

type requestListBuilder struct {
	version         schema.GroupVersion
	resource        string
	namespace       string
	limit           int64
	labelSelector   string
	resourceVersion string
}

func newRequestListBuilder(src *types.RequestList, resourceVersion string) *requestListBuilder {
	return &requestListBuilder{
		version: schema.GroupVersion{
			Group:   src.Group,
			Version: src.Version,
		},
		resource:        src.Resource,
		namespace:       src.Namespace,
		limit:           int64(src.Limit),
		labelSelector:   src.Selector,
		resourceVersion: resourceVersion,
	}
}

// Build implements RequestBuilder.Build.
func (b *requestListBuilder) Build(cli rest.Interface) (string, *rest.Request) {
	// https://kubernetes.io/docs/reference/using-api/#api-groups
	apiPath := "apis"
	if b.version.Group == "" {
		apiPath = "api"
	}

	comps := make([]string, 2, 5)
	comps[0], comps[1] = apiPath, b.version.Version
	if b.namespace != "" {
		comps = append(comps, "namespaces", b.namespace)
	}
	comps = append(comps, b.resource)

	return "LIST", cli.Get().AbsPath(comps...).
		SpecificallyVersionedParams(
			&metav1.ListOptions{
				LabelSelector:   b.labelSelector,
				ResourceVersion: b.resourceVersion,
				Limit:           b.limit,
			},
			scheme.ParameterCodec,
			schema.GroupVersion{Version: "v1"},
		)
}
