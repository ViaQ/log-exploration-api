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

type testStruct struct {
	TestName   string
	Index      string
	ShouldFail bool
	PathParams map[string]string
	TestParams map[string]string
	TestData   []string
	Response   map[string][]string
	Status     int
	UseToken   bool
}

func initProviderAndRouter() (p *elastic.MockedElasticsearchProvider, r *gin.Engine) {
	provider := elastic.NewMockedElastisearchProvider()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	NewLogsController(zap.L(), provider, router)
	return provider, router
}

func performTests(t *testing.T, tt testStruct, url string, provider *elastic.MockedElasticsearchProvider, g *gin.Engine) {

	t.Log("Running:", tt.TestName)
	logTime, _ := time.Parse(time.RFC3339Nano, "2021-03-17T14:22:40+05:30")
	provider.Cleanup()
	_ = provider.PutDataAtTime(logTime, tt.Index, tt.TestData)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if tt.UseToken {
		req.Header.Set("Authorization", "Bearer abcdefghijklmnopqrstuv")
	}
	q := req.URL.Query()
	for k, v := range tt.TestParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	if err != nil {
		t.Errorf("Failed to create HTTP request. E: %v", err)
	}
	g.ServeHTTP(rr, req)
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

func Test_ControllerFilterLogs(t *testing.T) {
	tests := []testStruct{
		{
			"Filter by no additional parameters",
			"app",
			false,
			map[string]string{},
			map[string]string{},
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			map[string][]string{"Logs": {"test-log-1", "test-log-2", "test-log-3"}},
			200,
			true,
		},
		{
			"Test with no token",
			"app",
			false,
			map[string]string{},
			map[string]string{},
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			map[string][]string{"Unauthorized, Please pass the token": {"authorization token not found"}},
			401,
			false,
		},
		{
			"Filter by index",
			"app",
			false,
			map[string]string{},
			map[string]string{"index": "app"},
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			map[string][]string{"Logs": {"test-log-1", "test-log-2", "test-log-3"}},
			200,
			true,
		},
		{
			"Filter by time",
			"app",
			false,
			map[string]string{},
			map[string]string{"starttime": "2021-03-17T14:22:20+05:30", "finishtime": "2021-03-17T14:23:20+05:30"},
			[]string{"test-log-1", "test-log-2", "test-log-3"},
			map[string][]string{"Logs": {"test-log-1", "test-log-2", "test-log-3"}},
			200,
			true,
		},
		{
			"Filter by podname",
			"infra",
			false,
			map[string]string{},
			map[string]string{"podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"}},
			200,
			true,
		},
		{
			"Filter by multiple parameters",
			"infra",
			false,
			map[string]string{},
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
			true,
		},
		{
			"Invalid parameters",
			"app",
			false,
			map[string]string{},
			map[string]string{
				"podname":   "hello",
				"namespace": "world",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Invalid timestamp",
			"app",
			false,
			map[string]string{},
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"No logs in the given time interval",
			"infra",
			false,
			map[string]string{},
			map[string]string{
				"starttime":  "2022-03-17T14:22:20+05:30",
				"finishtime": "2022-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
	}

	provider, router := initProviderAndRouter()
	for _, tt := range tests {
		url := "/logs/filter"
		performTests(t, tt, url, provider, router)
	}

}

func Test_ControllerFilterContainerLogs(t *testing.T) {

	tests := []testStruct{
		{
			"Filter by container name on no additional query parameters",
			"infra",
			false,
			map[string]string{"containername": "registry", "namespace": "openshift-image-registry", "podname": "image-registry-78b76b488f-9lvnn"},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry"}},
			200,
			true,
		},
		{
			"Test with no token",
			"infra",
			false,
			map[string]string{"containername": "registry", "namespace": "openshift-image-registry", "podname": "image-registry-78b76b488f-9lvnn"},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Unauthorized, Please pass the token": {"authorization token not found"}},
			401,
			false,
		},
		{
			"Filter by logging level on container name",
			"infra",
			false,
			map[string]string{"containername": "openshift-kube-scheduler", "namespace": "openshift-kube-scheduler", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			map[string]string{"level": "info"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"}},
			200,
			true,
		},
		{
			"Filter by time",
			"app",
			false,
			map[string]string{"containername": "openshift", "namespace": "openshift-kube-scheduler", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			map[string]string{
				"starttime":  "2021-03-17T14:22:20+05:30",
				"finishtime": "2021-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift"}},
			200,
			true,
		},
		{
			"Filter by time, and logging level",
			"app",
			false,
			map[string]string{"containername": "openshift", "namespace": "openshift-kube-scheduler", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			map[string]string{
				"starttime":  "2021-03-17T14:22:20+05:30",
				"finishtime": "2021-03-17T14:23:20+05:30",
				"level":      "warn",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"}},
			200,
			true,
		},
		{
			"Invalid podname, valid containername and namespace",
			"audit",
			false,
			map[string]string{"containername": "openshift", "namespace": "openshift-kube-scheduler", "podname": "dummy_podname"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, containername: openshift"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Invalid containername, valid namespace and valid podname",
			"audit",
			false,
			map[string]string{"containername": "dummy_container", "namespace": "openshift-kube-scheduler", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, containername: openshift"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Invalid namespace, valid containername and valid podname",
			"audit",
			false,
			map[string]string{"containername": "openshift", "namespace": "dummy_namespace", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, containername: openshift"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Invalid namespace, invalid containername and valid podname",
			"audit",
			false,
			map[string]string{"containername": "dummy_container", "namespace": "dummy_namespace", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, containername: openshift"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Invalid namespace, valid containername and invalid podname",
			"audit",
			false,
			map[string]string{"containername": "openshift", "namespace": "dummy_namespace", "podname": "dummy_podname"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, containername: openshift"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Valid namespace, invalid containername and invalid podname",
			"audit",
			false,
			map[string]string{"containername": "dummy_container", "namespace": "openshift-kube-scheduler", "podname": "dummy_podname"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, containername: openshift"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Invalid namespace, invalid containername and invalid podname",
			"audit",
			false,
			map[string]string{"containername": "dummy_container", "namespace": "dummy_namespace", "podname": "dummy_podname"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, containername: openshift"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Invalid timestamp",
			"infra",
			false,
			map[string]string{"containername": "openshift-kube", "namespace": "openshift-kube", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"No logs in the given time interval",
			"infra",
			false,
			map[string]string{"containername": "openshift-kube", "namespace": "openshift-kube-scheduler", "podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			map[string]string{
				"starttime":  "2022-03-17T14:22:20+05:30",
				"finishtime": "2022-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
	}

	provider, router := initProviderAndRouter()
	for _, tt := range tests {
		nameSpace := tt.PathParams["namespace"]
		podName := tt.PathParams["podname"]
		containerName := tt.PathParams["containername"]
		url := "/logs/namespace/" + nameSpace + "/pod/" + podName + "/container/" + containerName
		performTests(t, tt, url, provider, router)
	}

}

func Test_ControllerFilterLabelLogs(t *testing.T) {

	tests := []testStruct{
		{
			"Filter label logs",
			"infra",
			false,
			map[string]string{"flat_labels": "app=openshift-kube-scheduler,revision=8,scheduler=true"},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image, flat_labels: app=openshift-kube-scheduler,revision=8",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift, flat_labels: app=cluster-version-operator"},
			map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true"}},
			200,
			true,
		},
		{
			"Test with no token",
			"infra",
			false,
			map[string]string{"flat_labels": "app=openshift-kube-scheduler,revision=8,scheduler=true"},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image, flat_labels: app=openshift-kube-scheduler,revision=8",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift, flat_labels: app=cluster-version-operator"},
			map[string][]string{"Unauthorized, Please pass the token": {"authorization token not found"}},
			401,
			false,
		},
		{
			"Filter label logs and no other parameter",
			"infra",
			false,
			map[string]string{"flat_labels": "app=openshift-kube-scheduler,revision=8,scheduler=true"},
			map[string]string{},
			[]string{"flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true",
				"flat_labels: app=openshift-kube-scheduler,revision=8",
				"flat_labels: app=cluster-version-operator"},
			map[string][]string{"Logs": {"flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true"}},
			200,
			true,
		},
		{
			"Invalid filter labels",
			"infra",
			false,
			map[string]string{"flat_labels": "app=dummy,revision=0,scheduler=false"},
			map[string]string{},
			[]string{"flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true",
				"flat_labels: app=openshift-kube-scheduler,revision=8",
				"flat_labels: app=cluster-version-operator"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Filter by labels and time",
			"infra",
			false,
			map[string]string{"flat_labels": "app=openshift-kube-scheduler,revision=8,scheduler=true"},
			map[string]string{"starttime": "2021-03-17T14:22:20+05:30", "finishtime": "2021-03-17T14:23:20+05:30"},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image, flat_labels: app=openshift-kube-scheduler,revision=8",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift, flat_labels: app=cluster-version-operator"},
			map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true"}},
			200,
			true,
		},
		{
			"Filter by labels and logging level",
			"audit",
			false,
			map[string]string{"flat_labels": "app=openshift-kube-scheduler,revision=8,scheduler=true"},
			map[string]string{
				"level": "info",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, flat_labels: app=openshift-kube-scheduler, revision=8, scheduler=true"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, flat_labels: app=openshift-kube-scheduler, revision=8, scheduler=true"}},
			200,
			true,
		},
		{
			"Invalid timestamp",
			"infra",
			false,
			map[string]string{"flat_labels": "app=openshift-kube-scheduler,revision=8,scheduler=true"},
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, flat_labels: app=openshift-kube-scheduler,revision=8,scheduler=true"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"No logs in the given time interval",
			"infra",
			false,
			map[string]string{"flat_labels": "app=openshift-cluster-version"},
			map[string]string{
				"starttime":  "2022-03-17T14:22:20+05:30",
				"finishtime": "2022-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, flat_labels: app=openshift-cluster-version"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
	}

	provider, router := initProviderAndRouter()

	for _, tt := range tests {
		flatLabels := tt.PathParams["flat_labels"]
		url := "/logs/logs_by_labels/" + flatLabels
		performTests(t, tt, url, provider, router)

	}

}

func Test_ControllerFilterPodLogs(t *testing.T) {

	tests := []testStruct{
		{
			"Filter pod logs",
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
			true,
		},
		{
			"Test with No Token",
			"infra",
			false,
			map[string]string{"namespace": "openshift-image-registry", "podname": "image-registry-78b76b488f-9lvnn"},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Unauthorized, Please pass the token": {"authorization token not found"}},
			401,
			false,
		},
		{
			"Filter by logging level",
			"infra",
			false,
			map[string]string{"podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "namespace": "openshift-kube-scheduler"},
			map[string]string{"level": "info"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"}},
			200,
			true,
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
			true,
		},
		{
			"Filter by time",
			"app",
			false,
			map[string]string{"podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "namespace": "openshift-kube-scheduler"},
			map[string]string{
				"starttime":  "2021-03-17T14:22:20+05:30",
				"finishtime": "2021-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift"}},
			200,
			true,
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
			true,
		},
		{
			"Invalid namespace and valid podname",
			"app",
			false,
			map[string]string{"podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "namespace": "dummy"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Valid namespace and Invalid podname",
			"app",
			false,
			map[string]string{"podname": "dummy", "namespace": "openshift-kube-scheduler"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Invalid namespace and Invalid podname",
			"app",
			false,
			map[string]string{"podname": "dummy", "namespace": "dummy"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
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
			true,
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
			true,
		},
	}

	provider, router := initProviderAndRouter()
	for _, tt := range tests {

		nameSpace := tt.PathParams["namespace"]
		podName := tt.PathParams["podname"]
		url := "/logs/namespace/" + nameSpace + "/pod/" + podName
		performTests(t, tt, url, provider, router)
	}

}

func Test_ControllerFilterNamespaceLogs(t *testing.T) {

	tests := []testStruct{
		{
			"Filter on namespace",
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
			true,
		},
		{
			"Test with no token",
			"infra",
			false,
			map[string]string{"namespace": "openshift-image-registry"},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Unauthorized, Please pass the token": {"authorization token not found"}},
			401,
			false,
		},
		{
			"Filter by namespace and logging level",
			"infra",
			false,
			map[string]string{"namespace": "openshift-kube-scheduler"},
			map[string]string{"level": "info"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"}},
			200,
			true,
		},
		{
			"Filter by namespace and time ",
			"infra",
			false,
			map[string]string{"namespace": "openshift-kube-scheduler"},
			map[string]string{},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube-scheduler"}},
			200,
			true,
		},
		{
			"Filter by namespace, time, and logging level",
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
			true,
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
			true,
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
			true,
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
			true,
		},
	}

	provider, router := initProviderAndRouter()
	for _, tt := range tests {
		nameSpace := tt.PathParams["namespace"]
		url := "/logs/namespace/" + nameSpace
		performTests(t, tt, url, provider, router)
	}

}

func Test_ControllerLogs(t *testing.T) {

	tests := []testStruct{
		{
			"Filter no additional query parameters [Get all logs]",
			"infra",
			false,
			map[string]string{},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Logs": {"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"}},
			200,
			true,
		},
		{
			"Test with no token",
			"infra",
			false,
			map[string]string{},
			map[string]string{},
			[]string{"test-log-1 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: registry",
				"test-log-2 namespace_name: openshift-kube-controller-manager, pod_name: image-registry-78b76b488f-bzgqz, container_name: image",
				"test-log-3 namespace_name: openshift-image-registry, pod_name: image-registry-78b76b488f-9lvnn, container_name: openshift"},
			map[string][]string{"Unauthorized, Please pass the token": {"authorization token not found"}},
			401,
			false,
		},
		{
			"Filter by logging level",
			"infra",
			false,
			map[string]string{},
			map[string]string{"level": "info"},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler",
				"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: warn, container_name: openshift-kube-scheduler"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, level: info, container_name: openshift-kube-scheduler"}},
			200,
			true,
		},
		{
			"Filter by time, and logging level",
			"app",
			false,
			map[string]string{},
			map[string]string{
				"starttime":  "2021-03-17T14:22:20+05:30",
				"finishtime": "2021-03-17T14:23:20+05:30",
				"level":      "warn",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"},
			map[string][]string{"Logs": {"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift, level: warn"}},
			200,
			true,
		},
		{
			"Invalid timestamp",
			"infra",
			false,
			map[string]string{},
			map[string]string{
				"starttime":  "hey",
				"finishtime": "hey",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"Invalid Logging level",
			"infra",
			false,
			map[string]string{},
			map[string]string{
				"level": "invalid-level",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
		{
			"No logs in the given time interval",
			"infra",
			false,
			map[string]string{},
			map[string]string{
				"starttime":  "2022-03-17T14:22:20+05:30",
				"finishtime": "2022-03-17T14:23:20+05:30",
			},
			[]string{"test-log pod_name: openshift-kube-scheduler-ip-10-0-157-165.ec2.internal, namespace_name: openshift-kube-scheduler, container_name: openshift-kube"},
			map[string][]string{"Please check the input parameters": {"Not Found Error"}},
			400,
			true,
		},
	}

	provider, router := initProviderAndRouter()
	for _, tt := range tests {
		url := "/logs"
		performTests(t, tt, url, provider, router)
	}

}
