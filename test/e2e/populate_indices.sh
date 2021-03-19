#!/bin/bash

populate_es() {
    curl -XPUT http://localhost:9200/infra-000001
    curl -XPUT http://localhost:9200/audit-000001
    curl -XPUT http://localhost:9200/app-000001
}


populate_es