package env

import (
	"reflect"
	"testing"
)

// Test_podNamesToEnvVars ensures the following requirements.
//
//     - pod names are properly transformed to upper case
//     - the order of pod names aligns to the order of their corresponding env vars
//     - the given prefix is properly prepended
//
func Test_podNamesToEnvVars(t *testing.T) {
	testCases := []struct {
		PodNames []string
		Prefix   string
		Expected []string
	}{
		{
			PodNames: []string{
				"master-438601543-dxvjb",
				"worker-2528079433-dpl9w",
				"worker-2528079433-fh4hs",
			},
			Prefix: "K8S_ENDPOINT_UPDATER_POD_",
			Expected: []string{
				"K8S_ENDPOINT_UPDATER_POD_MASTER_438601543_DXVJB",
				"K8S_ENDPOINT_UPDATER_POD_WORKER_2528079433_DPL9W",
				"K8S_ENDPOINT_UPDATER_POD_WORKER_2528079433_FH4HS",
			},
		},
		{
			PodNames: []string{
				"master_438601543_dxvjb",
				"worker_2528079433_dpl9w",
				"worker_2528079433_fh4hs",
			},
			Prefix: "Prefix",
			Expected: []string{
				"PrefixMASTER_438601543_DXVJB",
				"PrefixWORKER_2528079433_DPL9W",
				"PrefixWORKER_2528079433_FH4HS",
			},
		},
	}

	for _, testCase := range testCases {
		result := podNamesToEnvVars(testCase.PodNames, testCase.Prefix)
		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Fatal("expected", testCase.Expected, "got", result)
		}
	}
}
