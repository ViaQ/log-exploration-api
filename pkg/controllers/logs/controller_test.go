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

func Test_ControllerFilterContainerLogs(t *testing.T) {

	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		TestParams map[string]string
		PathParams map[string]string
		TestData   []string
		Response   map[string][]string
		Status     int
	}{
		{
			"Filter by container name on no additional query parameters",
			"infra",
			false,
			map[string]string{},
			map[string]string{"containername": "registry", "namespace": "openshift-image-registry", "podname": "image-registry-78b76b488f-9lvnn"},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry"}},
			200,
		},
		{
			"Filter by logging level on container name",
			"infra",
			false,
			map[string]string{"level": "info"},
			map[string]string{"containername": "openshift-kube-scheduler", "namespace": "openshift-kube-scheduler", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"}},
			200,
		},
		{
			"Filter by time, and logging level",
			"app",
			false,
			map[string]string{
				"starttime":  "2021-03-17T14:22:20+05:30",
				"finishtime": "2021-03-17T14:23:20+05:30",
				"level":      "warn",
			},
			map[string]string{"containername": "openshift", "namespace": "openshift-kube-scheduler", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"}},
			200,
		},
		{
			"Invalid parameters",
			"audit",
			false,
			map[string]string{},
			map[string]string{"containername": "dummy_container", "namespace": "dummy_namespace", "podname": "dummy_podname"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
		{
			"Invalid timestamp",
			"infra",
			false,
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			map[string]string{"containername": "openshift-kube", "namespace": "openshift-kube", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
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
			map[string]string{"containername": "openshift-kube", "namespace": "openshift-kube-scheduler", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
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
		req, err := http.NewRequest(http.MethodGet, "/logs/namespace/"+tt.PathParams["namespace"]+"/pod/"+tt.PathParams["podname"]+"/container/"+tt.PathParams["containername"], nil)
		q := req.URL.Query()
		for k, v := range tt.TestParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
		if err != nil {
			t.Errorf("Failed to create HTTP request. E: %v", err)
		}
		fmt.Println(req.URL.String())
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

func Test_ControllerFilterLabelLogs(t *testing.T) {

	tests := []struct {
		TestName      string
		Index         string
		ShouldFail    bool
		LabelsList    []string
		UrlPathLabels string
		TestParams    map[string]string
		TestData      []string
		Response      map[string][]string
		Status        int
	}{{
		"Filter by container name on no additional query parameters",
		"infra",
		false,
		[]string{"app=openshift-kube-scheduler", "revision=8", "scheduler=true"},
		"app=openshift-kube-scheduler,revision=8,scheduler=true",
		map[string]string{},
		[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true",
			"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image, flat_labels: app=openshift-kube-scheduler,revision=8",
			"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift, flat_labels: app=cluster-version-operator"},
		map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true"}},
		200,
	},
		{
			"Filter by time",
			"infra",
			false,
			[]string{"app=openshift-kube-scheduler", "revision=8", "scheduler=true"},
			"app=openshift-kube-scheduler,revision=8,scheduler=true",
			map[string]string{"starttime": "2021-03-17T14:22:20+05:30", "finishtime": "2021-03-17T14:23:20+05:30"},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image, flat_labels: app=openshift-kube-scheduler,revision=8",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift, flat_labels: app=cluster-version-operator"},
			map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true"}},
			200,
		},
		{
			"Filter by labels and logging level",
			"audit",
			false,
			[]string{"app=openshift-kube-scheduler", "revision=8", "scheduler=true"},
			"app=openshift-kube-scheduler,revision=8,scheduler=true",
			map[string]string{
				"level": "info",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, flat_labels: app=openshift-kube-scheduler, revision=8, scheduler=true"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, flat_labels: app=openshift-kube-scheduler, revision=8, scheduler=true"}},
			200,
		},
		{
			"Invalid timestamp",
			"infra",
			false,
			[]string{"app=openshift-kube-scheduler", "revision=8", "scheduler=true"},
			"app=openshift-kube-scheduler,revision=8,scheduler=true",
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
		{
			"No logs in the given time interval",
			"infra",
			false,
			[]string{"app=openshift-cluster-version"},
			"app=openshift-cluster-version",
			map[string]string{
				"starttime":  "2022-03-17T14:22:20+05:30",
				"finishtime": "2022-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, flat_labels: app=openshift-cluster-version"},
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
		Url := "/logs/logs_by_labels/" + tt.UrlPathLabels

		req, err := http.NewRequest(http.MethodGet, Url, nil)
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

func Test_ControllerFilterPodLogs(t *testing.T) {

	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		PathParams map[string]string
		TestParams map[string]string
		TestData   []string
		Response   map[string][]string
		Status     int
	}{
		{
			"Filter by container name on no additional query parameters",
			"infra",
			false,
			map[string]string{"namespace": "openshift-image-registry", "podname": "image-registry-78b76b488f-9lvnn"},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"}},
			200,
		},
		{
			"Filter by logging level on container name",
			"infra",
			false,
			map[string]string{"podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "namespace": "openshift-kube-scheduler"},
			map[string]string{"level": "info"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"}},
			200,
		},
		{
			"Invalid Path Parameters",
			"infra",
			false,
			map[string]string{"podname": "cluster-kube-scheduler-ip-10-0-157-165.ec2.internal", "namespace": "openshift-kube-scheduler"},
			map[string]string{"level": "info"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
		{
			"Filter by time, and logging level",
			"app",
			false,
			map[string]string{"podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "namespace": "openshift-kube-scheduler"},
			map[string]string{
				"starttime":  "2021-03-17T14:22:20+05:30",
				"finishtime": "2021-03-17T14:23:20+05:30",
				"level":      "warn",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"}},
			200,
		},
		{
			"Invalid timestamp",
			"infra",
			false,
			map[string]string{"podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "namespace": "openshift-kube-scheduler"},
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
		{
			"No logs in the given time interval",
			"infra",
			false,
			map[string]string{"podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "namespace": "openshift-kube-scheduler"},
			map[string]string{
				"starttime":  "2022-03-17T14:22:20+05:30",
				"finishtime": "2022-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
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
		req, err := http.NewRequest(http.MethodGet, "/logs/namespace/"+tt.PathParams["namespace"]+"/pod/"+tt.PathParams["podname"], nil)
		q := req.URL.Query()
		for k, v := range tt.TestParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
		if err != nil {
			t.Errorf("Failed to create HTTP request. E: %v", err)
		}
		fmt.Println(req.URL.String())
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

func Test_ControllerFilterNamespaceLogs(t *testing.T) {

	tests := []struct {
		TestName   string
		Index      string
		ShouldFail bool
		PathParams map[string]string
		TestParams map[string]string
		TestData   []string
		Response   map[string][]string
		Status     int
	}{
		{
			"Filter by container name on no additional query parameters",
			"infra",
			false,
			map[string]string{"namespace": "openshift-image-registry"},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"}},
			200,
		},
		{
			"Filter by logging level on container name",
			"infra",
			false,
			map[string]string{"namespace": "openshift-kube-scheduler"},
			map[string]string{"level": "info"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"}},
			200,
		},
		{
			"Filter by time, and logging level",
			"app",
			false,
			map[string]string{"namespace": "openshift-kube-scheduler"},
			map[string]string{
				"starttime":  "2021-03-17T14:22:20+05:30",
				"finishtime": "2021-03-17T14:23:20+05:30",
				"level":      "warn",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"}},
			200,
		},
		{
			"Invalid Namespace",
			"audit",
			false,
			map[string]string{"namespace": "cluster-scheduler"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
		{
			"Invalid timestamp",
			"infra",
			false,
			map[string]string{"namespace": "openshift-kube-scheduler"},
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
		{
			"No logs in the given time interval",
			"infra",
			false,
			map[string]string{"namespace": "openshift-kube-scheduler"},
			map[string]string{
				"starttime":  "2022-03-17T14:22:20+05:30",
				"finishtime": "2022-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
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
		req, err := http.NewRequest(http.MethodGet, "/logs/namespace/"+tt.PathParams["namespace"], nil)
		q := req.URL.Query()
		for k, v := range tt.TestParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
		if err != nil {
			t.Errorf("Failed to create HTTP request. E: %v", err)
		}
		fmt.Println(req.URL.String())
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

func Test_ControllerLogs(t *testing.T) {

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
			"Filter no additional query parameters [Get all logs]",
			"infra",
			false,
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"}},
			200,
		},
		{
			"Filter by logging level",
			"infra",
			false,
			map[string]string{"level": "info"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler",
				"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: warn, container_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"}},
			200,
		},
		{
			"Filter by time, and logging level",
			"app",
			false,
			map[string]string{
				"starttime":  "2021-03-17T14:22:20+05:30",
				"finishtime": "2021-03-17T14:23:20+05:30",
				"level":      "warn",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"}},
			200,
		},
		{
			"Invalid timestamp",
			"infra",
			false,
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
		},
		{
			"Invalid Logging level",
			"infra",
			false,
			map[string]string{
				"level": "invalid-level",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
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
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
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
		req, err := http.NewRequest(http.MethodGet, "/logs", nil)
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
