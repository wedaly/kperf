package runner

var (
	// runnerGroupReleaseLabels is used to mark that helm chart release
	// is managed by kperf.
	runnerGroupReleaseLabels = map[string]string{
		"runnergroups.kperf.io/managed": "true",
	}
)

const (
	// runnerGroupServerChartName should be aligned with ../manifests/runnergroup/server.
	runnerGroupServerChartName = "runnergroup/server"

	// runnerGroupServerReleaseName is the helm releas name for runner groups's server.
	runnerGroupServerReleaseName = "runnergroup-server"

	// runnerGroupServerPort should be aligned with ../manifests/runnergroup/server/templates/pod.yaml.
	runnerGroupServerPort uint16 = 8080

	// runnerGroupReleaseNamespace is used to host runner groups.
	runnerGroupReleaseNamespace = "runnergroups-kperf-io"
)
