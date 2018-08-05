#!/bin/bash
set -u -e -o pipefail -o xtrace

# for testing
# ./capturer -addr 127.0.0.1:5332 -pcap foo.pcap

rm -rf ./capturer
go build -o ./capturer github.com/ravendb/ravendb-go-client/cmd/capture
# mark it as owend by root so that it has root priviledges even when
# not invoked by root.
# it needs root priviledges to capture packets
sudo chown root:wheel ./capturer
# set "follow user id on execution" bit so that it inherits root priviledges
# from file ownership
sudo chmod +s ./capturer

rm -rf ./pcap_convert
go build -o ./pcap_convert  github.com/ravendb/ravendb-go-client/cmd/pcap_to_txt
