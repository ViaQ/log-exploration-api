package logscontroller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ViaQ/log-exploration-api/pkg/elastic"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Test_ControllerFilterByIndex(t *testing.T) {
	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		TestData   []string
		Error	   error
	}{
		{
			"Filter by App index",
			"app",
			false,
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			nil,
		},
		{
			"Filter by Infra index",
			"infra",
			false,
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			nil,
		},
		{
			"Filter by Audit index",
			"audit",
			false,
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			nil,
		},
		{
			"Filter by Invalid index",
			"api",
			false,
			[]string{},
			fmt.Errorf("Not Found Error"),
		},
	}

	provider := elastic.NewMockedElastisearchProvider()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	NewLogsController(zap.L(), provider, router)

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)

		provider.PutDataIntoIndex(tt.Index, tt.TestData)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/logs/indexfilter/"+tt.Index, nil)
		if err != nil {
			t.Errorf("Failed to create HTTP request. E: %v", err)
		}
		router.ServeHTTP(rr, req)
		resp := rr.Body.String()
		status := rr.Code

		expected, err := json.Marshal(map[string]interface{}{"Logs": tt.TestData})
		expectedStatus := http.StatusOK
		if err != nil {
			t.Errorf("failed to marshal test data. E: %v", err)
		}

		if !(tt.Index == "app" || tt.Index == "infra" || tt.Index == "audit") {
			expected, err = json.Marshal(map[string]interface{}{"Invalid Index Entered ": tt.Error})
			expectedStatus = http.StatusBadRequest
		}
		
		expectedResp := string(expected)
		if resp != expectedResp {
			t.Errorf("expected response to be %s, got %s", expectedResp, resp)
		}
		if status != expectedStatus {
			t.Errorf("expected status code: %v found: %v", expectedStatus, status)
		}
	}

}

func Test_ControllerGetAllLogs(t *testing.T) {
	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		TestData   []string
	}{
		{
			"Get all logs",
			"infra",
			false,
			[]string{"test-log-1", "test-log-2", "test-log-3"},
		},
	}

	provider := elastic.NewMockedElastisearchProvider()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	NewLogsController(zap.L(), provider, router)

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)

		provider.PutDataIntoIndex(tt.Index, tt.TestData)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/logs/", nil)
		if err != nil {
			t.Errorf("Failed to create HTTP request. E: %v", err)
		}
		router.ServeHTTP(rr, req)
		resp := rr.Body.String()

		expected, err := json.Marshal(map[string]interface{}{"Logs": tt.TestData})
		if err != nil {
			t.Errorf("failed to marshal test data. E: %v", err)
		}
		expectedResp := string(expected)

		if resp != expectedResp {
			t.Errorf("expected response to be %s, got %s", expectedResp, resp)
		}
	}

}


func Test_ControllerFilterByTime(t *testing.T) {
	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		TestData   []string
	}{
		{
			"Filter by time",
			"infra",
			false,
			[]string{"test-log-1", "test-log-2", "test-log-3"},
		},
	}

	provider := elastic.NewMockedElastisearchProvider()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	NewLogsController(zap.L(), provider, router)

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)

		logTime,_ := time.Parse(time.RFC3339Nano, "2021-03-17T14:22:40+05:30")
		provider.PutDataAtTime(logTime, tt.TestData)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/logs/timefilter/2021-03-17T14:22:20+05:30/2021-03-17T14:23:20+05:30", nil)
		if err != nil {
			t.Errorf("Failed to create HTTP request. E: %v", err)
		}
		router.ServeHTTP(rr, req)
		resp := rr.Body.String()

		expected, err := json.Marshal(map[string]interface{}{"Logs": tt.TestData})
		if err != nil {
			t.Errorf("failed to marshal test data. E: %v", err)
		}
		expectedResp := string(expected)

		if resp != expectedResp {
			t.Errorf("expected response to be %s, got %s", expectedResp, resp)
		}
	}

}

func Test_FilterLogsByPodName(t *testing.T) {
	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		TestData   []string
	}{
		{
			"Filter by Podname",
			"infra",
			false,
			[]string{"test-log-1 test-log-1 pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
		},
	}

	provider := elastic.NewMockedElastisearchProvider()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	NewLogsController(zap.L(), provider, router)

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)

		provider.PutDataIntoIndex(tt.Index, tt.TestData)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/logs/podnamefilter/openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", nil)
		if err != nil {
			t.Errorf("Failed to create HTTP request. E: %v", err)
		}
		router.ServeHTTP(rr, req)
		resp := rr.Body.String()

		expected, err := json.Marshal(map[string]interface{}{"Logs": tt.TestData})
		if err != nil {
			t.Errorf("failed to marshal test data. E: %v", err)
		}
		expectedResp := string(expected)

		if resp != expectedResp {
			t.Errorf("expected response to be %s, got %s", expectedResp, resp)
		}
	}

}

func Test_FilterLogsMultipleParameters(t *testing.T) {
	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		TestData   []string
	}{
		{
			"Filter by multiple parameters",
			"infra",
			false,
			[]string{"test-log-1 pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
		},
	}

	provider := elastic.NewMockedElastisearchProvider()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	NewLogsController(zap.L(), provider, router)

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		logTime,_ := time.Parse(time.RFC3339Nano, "2021-03-17T14:22:40+05:30")
		provider.PutDataAtTime(logTime, tt.TestData)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/logs/multifilter/openshift-kube-scheduler/openshift-kube-scheduler-ip-10-0-157-165.ec2.internal/2021-03-17T14:22:20+05:30/2021-03-17T14:23:20+05:30", nil)
		if err != nil {
			t.Errorf("Failed to create HTTP request. E: %v", err)
		}
		router.ServeHTTP(rr, req)
		resp := rr.Body.String()

		expected, err := json.Marshal(map[string]interface{}{"Logs": tt.TestData})
		if err != nil {
			t.Errorf("failed to marshal test data. E: %v", err)
		}
		expectedResp := string(expected)

		if resp != expectedResp {
			t.Errorf("expected response to be %s, got %s", expectedResp, resp)
		}
	}

}