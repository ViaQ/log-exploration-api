package logscontroller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ViaQ/log-exploration-api/pkg/elastic"
	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Test_ControllerFilterLogs(t *testing.T) {
	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		TestData   []string
	}{
		{
			"Filter by no parameters",
			"app",
			false,
			[]string{"test-log-1", "test-log-2", "test-log-3"},
		},
		{
			"Filter by index",
			"app",
			false,
			[]string{"test-log-1", "test-log-2", "test-log-3"},
		},
		{
			"Filter by time",
			"app",
			false,
			[]string{"test-log-1", "test-log-2", "test-log-3"},
		},
		{
			"Filter by podname",
			"infra",
			false,
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
		},
		{
			"Filter by multiple parameters",
			"infra",
			false,
			[]string{"test-log-1 pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
		},
		{
			"Invalid parameters",
			"app",
			false,
			[]string{"test-log-1 pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
		},
		{
			"Invalid timestamp",
			"app",
			false,
			[]string{"test-log-1 pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
		},
		{
			"No logs in the given time interval",
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
		logTime, _ := time.Parse(time.RFC3339Nano, "2021-03-17T14:22:40+05:30")
		provider.PutDataAtTime(logTime, tt.Index, tt.TestData)

		var resp string
		var expected []byte
		var status, expectedStatus int
		parameters := logs.Parameters{
			Index:      "",
			Level:      "",
			FinishTime: "",
			StartTime:  "",
			Podname:    "",
			Namespace:  "",
		}
		switch tt.TestName {
		case "Filter by no parameters":
			rr := httptest.NewRecorder()
			gctx, _ := gin.CreateTestContext(rr)
			body := new(bytes.Buffer)
			json.NewEncoder(body).Encode(parameters)
			req, err := http.NewRequestWithContext(gctx, http.MethodGet, "/logs/filter", body)
			if err != nil {
				t.Errorf("Failed to create HTTP request. E: %v", err)
			}
			router.ServeHTTP(rr, req)
			resp = rr.Body.String()
			status = rr.Code
			expected, err = json.Marshal(map[string]interface{}{"logs": tt.TestData})
			if err != nil {
				t.Errorf("failed to marshal test data. E: %v", err)
			}
			expectedStatus = http.StatusOK
		case "Filter by index":
			rr := httptest.NewRecorder()
			gctx, _ := gin.CreateTestContext(rr)
			parameters.Index = tt.Index
			body := new(bytes.Buffer)
			json.NewEncoder(body).Encode(parameters)
			req, err := http.NewRequestWithContext(gctx, http.MethodGet, "/logs/filter", body)
			if err != nil {
				t.Errorf("Failed to create HTTP request. E: %v", err)
			}
			router.ServeHTTP(rr, req)
			resp = rr.Body.String()
			status = rr.Code
			expected, err = json.Marshal(map[string]interface{}{"logs": tt.TestData})
			if err != nil {
				t.Errorf("failed to marshal test data. E: %v", err)
			}
			expectedStatus = http.StatusOK
		case "Filter by time":
			rr := httptest.NewRecorder()
			gctx, _ := gin.CreateTestContext(rr)
			parameters.FinishTime = "2021-03-17T14:23:20+05:30"
			parameters.StartTime = "2021-03-17T14:22:20+05:30"
			body := new(bytes.Buffer)
			json.NewEncoder(body).Encode(parameters)
			req, err := http.NewRequestWithContext(gctx, http.MethodGet, "/logs/filter", body)
			if err != nil {
				t.Errorf("Failed to create HTTP request. E: %v", err)
			}
			router.ServeHTTP(rr, req)
			resp = rr.Body.String()
			status = rr.Code
			expected, err = json.Marshal(map[string]interface{}{"logs": tt.TestData})
			if err != nil {
				t.Errorf("failed to marshal test data. E: %v", err)
			}
			expectedStatus = http.StatusOK
		case "Filter by podname":
			rr := httptest.NewRecorder()
			gctx, _ := gin.CreateTestContext(rr)
			parameters.Podname = "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"
			body := new(bytes.Buffer)
			json.NewEncoder(body).Encode(parameters)
			req, err := http.NewRequestWithContext(gctx, http.MethodGet, "/logs/filter", body)
			if err != nil {
				t.Errorf("Failed to create HTTP request. E: %v", err)
			}
			router.ServeHTTP(rr, req)
			resp = rr.Body.String()
			status = rr.Code
			expected, err = json.Marshal(map[string]interface{}{"logs": tt.TestData})
			if err != nil {
				t.Errorf("failed to marshal test data. E: %v", err)
			}
			expectedStatus = http.StatusOK
		case "Filter by multiple parameters":
			rr := httptest.NewRecorder()
			gctx, _ := gin.CreateTestContext(rr)
			parameters.Index = tt.Index
			parameters.FinishTime = "2021-03-17T14:23:20+05:30"
			parameters.StartTime = "2021-03-17T14:22:20+05:30"
			parameters.Podname = "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"
			parameters.Namespace = "openshift-kube-scheduler"
			body := new(bytes.Buffer)
			json.NewEncoder(body).Encode(parameters)
			req, err := http.NewRequestWithContext(gctx, http.MethodGet, "/logs/filter", body)
			if err != nil {
				t.Errorf("Failed to create HTTP request. E: %v", err)
			}
			router.ServeHTTP(rr, req)
			resp = rr.Body.String()
			status = rr.Code
			expected, err = json.Marshal(map[string]interface{}{"logs": tt.TestData})
			if err != nil {
				t.Errorf("failed to marshal test data. E: %v", err)
			}
			expectedStatus = http.StatusOK
		case "Invalid parameters":
			rr := httptest.NewRecorder()
			gctx, _ := gin.CreateTestContext(rr)
			parameters.Index = tt.Index
			parameters.FinishTime = "2021-03-17T14:23:20+05:30"
			parameters.StartTime = "2021-03-17T14:22:20+05:30"
			parameters.Podname = "hello"
			parameters.Namespace = "world"
			body := new(bytes.Buffer)
			json.NewEncoder(body).Encode(parameters)
			req, err := http.NewRequestWithContext(gctx, http.MethodGet, "/logs/filter", body)
			if err != nil {
				t.Errorf("Failed to create HTTP request. E: %v", err)
			}
			router.ServeHTTP(rr, req)
			resp = rr.Body.String()
			status = rr.Code
			expected, err = json.Marshal(map[string]interface{}{"Please Check input parameters": fmt.Errorf("No logs found!")})
			if err != nil {
				t.Errorf("failed to marshal test data. E: %v", err)
			}
			expectedStatus = http.StatusBadRequest
		case "Invalid timestamp":
			rr := httptest.NewRecorder()
			gctx, _ := gin.CreateTestContext(rr)
			parameters.FinishTime = "hey"
			parameters.StartTime = "hey"
			body := new(bytes.Buffer)
			json.NewEncoder(body).Encode(parameters)
			req, err := http.NewRequestWithContext(gctx, http.MethodGet, "/logs/filter", body)
			if err != nil {
				t.Errorf("Failed to create HTTP request. E: %v", err)
			}
			router.ServeHTTP(rr, req)
			resp = rr.Body.String()
			status = rr.Code
			expected, err = json.Marshal(map[string]interface{}{"Please Check input parameters": fmt.Errorf("No logs found!")})
			if err != nil {
				t.Errorf("failed to marshal test data. E: %v", err)
			}
			expectedStatus = http.StatusBadRequest
		case "No logs in given time interval":
			rr := httptest.NewRecorder()
			gctx, _ := gin.CreateTestContext(rr)
			parameters.FinishTime = "2022-03-17T14:23:20+05:30"
			parameters.StartTime = "2022-03-17T14:22:20+05:30"
			body := new(bytes.Buffer)
			json.NewEncoder(body).Encode(parameters)
			req, err := http.NewRequestWithContext(gctx, http.MethodGet, "/logs/filter", nil)
			if err != nil {
				t.Errorf("Failed to create HTTP request. E: %v", err)
			}
			router.ServeHTTP(rr, req)
			resp = rr.Body.String()
			status = rr.Code
			expected, err = json.Marshal(map[string]interface{}{"Please Check input parameters": fmt.Errorf("No logs found!")})
			if err != nil {
				t.Errorf("failed to marshal test data. E: %v", err)
			}
			expectedStatus = http.StatusBadRequest
		}

		expectedResp := string(expected)
		if resp != expectedResp {
			t.Errorf("expected response to be %s, got %s", expectedResp, resp)
		}
		if status != expectedStatus {
			t.Errorf("expected response to be %v, got %v", expectedStatus, status)
		}
	}

}
