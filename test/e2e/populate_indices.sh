#!/bin/bash

populate_es() {
    curl -XPUT http://localhost:9200/infra-000001
    curl -XPUT http://localhost:9200/audit-000001
    curl -XPUT http://localhost:9200/app-000001
}

populate_es

# To insert data 
NODE_TLS_REJECT_UNAUTHORIZED=0 ./node_modules/.bin/elasticdump \
    --output=http://localhost:9200 \
    --input=test-logs-mapping.json \
    --type=data