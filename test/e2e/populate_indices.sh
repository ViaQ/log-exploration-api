#!/bin/bash

insert_infra() {
  curl -X PUT "localhost:9200/infra-000001?pretty" -H 'Content-Type: application/json' -d'
  {
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0
  },
    "mappings": {
      "properties": {
        "docker.container_id":{"type":"keyword"},
        "kubernetes.namespace_name": { "type": "keyword" },
        "kubernetes.pod_name": { "type": "keyword" },
        "kubernetes.host" : {"type":"keyword"},
        "kubernetes.pod_id" : {"type":"keyword"},
        "kubernetes.master_url" : {"type":"keyword"},
        "kubernetes.namespace_id" : {"type":"keyword"},
        "level":{"type":"keyword"},
        "hostname":{"type":"keyword"},
        "pipeline_metadata.collector.ipaddr4":{"type":"ip"},
        "pipeline_metadata.collector.inputname":{"type":"keyword"},
        "pipeline_metadata.collector.name":{"type":"keyword"},
        "pipeline_metadata.collector.received_at":{"type":"date"},
        "pipeline_metadata.collector.version":{"type":"keyword"},
        "viaq_msg_id":{"type":"keyword"}
    }
  }
}
'
  curl -X POST "localhost:9200/_aliases?pretty" -H 'Content-Type: application/json' -d' { "actions" : [ { "add" : { "index" : "infra-000001", "alias" : "infra" } } ] } '
}

insert_app() {
  curl -X PUT "localhost:9200/app-000001?pretty" -H 'Content-Type: application/json' -d'
  {
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0
  },
    "mappings": {
      "properties": {
        "docker.container_id":{"type":"keyword"},
        "kubernetes.namespace_name": { "type": "keyword" },
        "kubernetes.pod_name": { "type": "keyword" },
        "kubernetes.host" : {"type":"keyword"},
        "kubernetes.pod_id" : {"type":"keyword"},
        "kubernetes.master_url" : {"type":"keyword"},
        "kubernetes.namespace_id" : {"type":"keyword"},
        "level":{"type":"keyword"},
        "hostname":{"type":"keyword"},
        "pipeline_metadata.collector.ipaddr4":{"type":"ip"},
        "pipeline_metadata.collector.inputname":{"type":"keyword"},
        "pipeline_metadata.collector.name":{"type":"keyword"},
        "pipeline_metadata.collector.received_at":{"type":"date"},
        "pipeline_metadata.collector.version":{"type":"keyword"},
        "viaq_msg_id":{"type":"keyword"}
    }
  }
}
'
  curl -X POST "localhost:9200/_aliases?pretty" -H 'Content-Type: application/json' -d' { "actions" : [ { "add" : { "index" : "app-000001", "alias" : "app" } } ] } '
}

insert_audit() {
  curl -X PUT "localhost:9200/audit-000001?pretty" -H 'Content-Type: application/json' -d'
  {
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0
  },
    "mappings": {
      "properties": {
        "docker.container_id":{"type":"keyword"},
        "kubernetes.namespace_name": { "type": "keyword" },
        "kubernetes.pod_name": { "type": "keyword" },
        "kubernetes.host" : {"type":"keyword"},
        "kubernetes.pod_id" : {"type":"keyword"},
        "kubernetes.master_url" : {"type":"keyword"},
        "kubernetes.namespace_id" : {"type":"keyword"},
        "level":{"type":"keyword"},
        "hostname":{"type":"keyword"},
        "pipeline_metadata.collector.ipaddr4":{"type":"ip"},
        "pipeline_metadata.collector.inputname":{"type":"keyword"},
        "pipeline_metadata.collector.name":{"type":"keyword"},
        "pipeline_metadata.collector.received_at":{"type":"date"},
        "pipeline_metadata.collector.version":{"type":"keyword"},
        "viaq_msg_id":{"type":"keyword"}
    }
  }
}
'
  curl -X POST "localhost:9200/_aliases?pretty" -H 'Content-Type: application/json' -d' { "actions" : [ { "add" : { "index" : "audit-000001", "alias" : "audit" } } ] } '
}

insert_indices() {
  ES_STATUS=$(curl --silent 'http://localhost:9200/_cluster/health?pretty=true' | jq .status)
  ES_STATUS_GREEN="\"green\""
  ES_STATUS_YELLOW="\"yellow\""
  BASE_URL="http://localhost:9200/"
  TIME_LIMIT=$((SECONDS+60))

  while [[ $ES_STATUS != $ES_STATUS_GREEN ]] && [[ $ES_STATUS != $ES_STATUS_YELLOW ]]
  do
    if [ $SECONDS -gt $TIME_LIMIT ]; then
      echo "Connection timed out error: Failed to connect to "$BASE_URL
      docker-compose down -v
      exit 1
    fi
    ES_STATUS=$(curl --request GET -sL --url $BASE_URL'_cluster/health?pretty=true' | jq .status)
  done

  insert_infra
  insert_app
  insert_audit
}

check_elasticdump(){

  if ! [ -x "$(command -v elasticdump)" ]; then
    echo "Error: Elasticdump Not Found."
    docker-compose down -v
    exit 1
  fi
  populate_es

}

populate_es(){
  NODE_TLS_REJECT_UNAUTHORIZED=0 ./node_modules/.bin/elasticdump \
    --output=http://localhost:9200 \
    --input=test-logs-mapping.json \
    --type=data
}

insert_indices
check_elasticdump