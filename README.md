# Log exploration API server
This is the backend for the log exploration application for the Openshift Logging Stack

### Supported features
 * Fetch logs from specific log group
 * Fetch logs for specific time window
 * Fetch logs for the pod for the specific time window

### Build
`make build` - to build the application <br/>
`make test` - to run unit tests

### Metrics
This application uses Prometheus for monitoring metrics.
See [Official Prometheus Docs](https://prometheus.io/docs/guides/go-application/)
and this [Prometheus User Guide](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/user-guides/getting-started.md)
for more information.

#### Dependencies
- [oc](https://docs.openshift.com/container-platform/4.8/cli_reference/openshift_cli/getting-started-cli.html)
   installed
- an existing Openshift cluster you can connect to
- the following log-exploration-api objects should have already been created : deployment, service, and route
    - if not use the commands : <br/>
    `oc apply -f log-exploration-api-deployment.yaml`<br/>
      `oc apply -f log-exploration-service.yaml`<br/>
      `oc apply -f log-exploration-api-route.yaml`<br/>

#### To run metrics from log-exploration-api :
`oc apply -f log-exploration-api-service-monitor`<br/>
ServiceMonitor describes the set of targets to be monitored by Prometheus.<br/>
`oc apply -f log-exploration-api-namespace` <br/>
This enables cluster monitoring of the namespace.<br/>


#### See Metrics in a Browser
Use the command : 
`oc get route`<br/>
Copy and paste the log-exploration-api HOST/PORT into your browser followed by the endpoint `/metrics`. 

#### See Metrics in the Prometheus UI via Openshift Console 
Go to the Openshift Console > Monitoring > Metrics > Prometheus UI. 
You will be prompted to 'Login with Openshift'. Sign in with your cluster username and password. This will log you into 
the Prometheus homepage. At the top of the screen you should see a dialogue box to enter expressions. 

Enter the following expressions to query for log-exploration-api's custom metrics: 
<br/>
`custom_metric_http_requests_total` <br/>
This expression obtains the total number of requests at each endpoint. <br/>
`custom_metric_http_response_time_seconds_bucket`<br/>
`custom_metric_http_response_time_seconds_count`<br/>
`custom_metric_http_response_time_seconds_sum`<br/>
These expressions return information about response time when endpoints are called. <br/>
`custom_metric_response_status`<br/>
This expression returns the counts for response statuses (e.g., 200, 404). <br/>

#### Metrics Troubleshooting: Verify Prometheus Discovers Our Target
On the Prometheus UI find the navigation bar and click Status > Targets. If Prometheus is able to monitor 
our application you should see the target named 
`openshift-logging/log-exploration-api-service-monitor/0`. 


