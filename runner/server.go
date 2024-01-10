package runner

import (
	"fmt"
	"net"
	"net/http"

	"github.com/Azure/kperf/runner/localstore"
)

type Server struct {
	store   *localstore.Store
	lis     net.Listener
	groups  []GroupHandler
	readyCh chan struct{}
}

func NewServer(dataDir string, addr string, groups ...*GroupHandler) (*Server, error) {
	// 1. Listen on addr (ensure addr is available)
	//	lis, err := net.Listen("tcp", addr)
	//
	// 2. Check if summary result exists in local store
	//
	// 	store := localstore.NewStore(dataDir)
	//	r, err := store.OpenReader("summary.json")
	//	if err == nil  {
	//		mark summary report is ready
	//		return &Server{}, nil
	//	}
	//
	// 3. If len(groups) == 0, mark summary report is ready as well.
	//	return &Server, nil
	//
	// 4. Deploy groups
	//
	// 5. If all runner groups are finished, generate summary.
	return nil, fmt.Errorf("not implemented yet")
}

func (s *Server) Run() error {
	// use https://github.com/gorilla/mux to build http routing
	mux := http.NewServeMux()
	return http.Serve(s.lis, mux)
}

// buildListRunnerGroupsHandler is to create handler for /v1/runnergroups.
func (s *Server) buildListRunnerGroupsHandler() (string, http.Handler) {
	return "/v1/runnergroups",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// List all the runner groups to generate response.
			w.WriteHeader(http.StatusNotImplemented)
		})
}

// buildGetRunnerGroupsSummaryHandler is to create handler for /v1/runnergroups/summary.
func (s *Server) buildGetRunnerGroupsSummaryHandler() (string, http.Handler) {
	return "/v1/runnergroups/summary",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// check readyCh channel
			// if it's closed, that means summary report is ready.
			w.WriteHeader(http.StatusNotImplemented)
		})
}

func (s *Server) buildPostRunnerGroupsRunnerResult() (string, http.Handler) {
	return "/v1/runnergroups/{runner}/result",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// write request body into s.store.
			w.WriteHeader(http.StatusNotImplemented)
		})
}
