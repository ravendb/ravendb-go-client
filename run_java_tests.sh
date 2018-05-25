#!/bin/bash

export RAVENDB_JAVA_TEST_SERVER_PATH=${HOME}/Documents/RavenDB/Server/Raven.Server
cd ../ravendb-jvm-client/

# Running just one maven test: https://stackoverflow.com/a/18136440/2898
mvn -Dtest=HiLoTest test
#mvn test
