package health

import (
	"encoding/json"
	"github.com/ViaQ/log-exploration-api/pkg/elastic"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testStruct struct {
	TestName       string
	ShouldFail     bool
	ReadinessState bool
	Response       map[string]string
	Status         int
}

func initProviderAndRouter() (p *elastic.MockedElasticsearchProvider, r *gin.Engine) {
	provider := elastic.NewMockedElastisearchProvider()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	NewHealthController(router, provider)
	return provider, router
}

func performTests(t *testing.T, tt testStruct, url string, provider *elastic.MockedElasticsearchProvider, g *gin.Engine) {

	t.Log("Running:", tt.TestName)
	provider.Cleanup()
	provider.UpdateReadinessState(tt.ReadinessState)
	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, url, nil)

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

func TestHealthController_ReadinessHandler(t *testing.T) {
	tests := []testStruct{
		{
			"ES is ready",
			false,
			true,
			map[string]string{"Message": "Success"},
			200,
		},
		{
			"ES is not ready",
			false,
			false,
			map[string]string{"Message": "failed to connect to esClient"},
			400,
		},
	}
	provider, router := initProviderAndRouter()
	for _, tt := range tests {
		url := "/ready"
		performTests(t, tt, url, provider, router)
	}
}
