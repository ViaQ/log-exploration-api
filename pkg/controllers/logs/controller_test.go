package logscontroller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ViaQ/log-exploration-api/pkg/elastic"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Test_ControllerFilterLogs(t *testing.T) {
	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		TestParams map[string]string
		TestData   []string
		Response   map[string][]string
		Status     int
	}{
		{
			"Filter by no parameters",
			"app",
			false,
			map[string]string{},
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			map[string][]string{"Logs": {"test-log-1", "test-log-2", "test-log-3"}},
			200,
		},
		{
			"Filter by index",
			"app",
			false,
			map[string]string{"index": "app"},
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			map[string][]string{"Logs": {"test-log-1", "test-log-2", "test-log-3"}},
			200,
		},
		{
			"Filter by time",
			"app",
			false,
			map[string]string{"starttime": "2021-03-17T14:22:20+05:30", "finishtime": "2021-03-17T14:23:20+05:30"},
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			map[string][]string{"Logs": {"test-log-1", "test-log-2", "test-log-3"}},
			200,
		},
		{
			"Filter by podname",
			"infra",
			false,
			map[string]string{"podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"}},
			200,
		},
		{
			"Filter by multiple parameters",
			"infra",
			false,
			map[string]string{
				"index":      "infra",
				"podname":    "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal",
				"namespace":  "openshift-kube-scheduler",
				"starttime":  "2021-03-17T14:22:20+05:30",
				"finishtime": "2021-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"}},
			200,
		},
		{
			"Invalid parameters",
			"app",
			false,
			map[string]string{
				"podname":   "hello",
				"namespace": "world",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
		{
			"Invalid timestamp",
			"app",
			false,
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
		{
			"No logs in the given time interval",
			"infra",
			false,
			map[string]string{
				"starttime":  "2022-03-17T14:22:20+05:30",
				"finishtime": "2022-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
	}

	provider := elastic.NewMockedElastisearchProvider()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	NewLogsController(zap.L(), provider, router)

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		logTime, _ := time.Parse(time.RFC3339Nano, "2021-03-17T14:22:40+05:30")
		provider.Cleanup()
		provider.PutDataAtTime(logTime, tt.Index, tt.TestData)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/logs/filter", nil)
		q := req.URL.Query()
		for k, v := range tt.TestParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
		if err != nil {
			t.Errorf("Failed to create HTTP request. E: %v", err)
		}
		router.ServeHTTP(rr, req)
		resp := rr.Body.String()
		status := rr.Code
		expected, err := json.Marshal(tt.Response)
		if err != nil {
			t.Errorf("failed to marshal test data. E: %v", err)
		}
		expectedResp := string(expected)
		if resp != expectedResp {
			t.Errorf("expected response to be %s, got %s", expectedResp, resp)
		}
		if status != tt.Status {
			t.Errorf("expected response to be %v, got %v", tt.Status, status)
		}
	}

}
