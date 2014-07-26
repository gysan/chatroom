#!/bin/sh
protoc --go_out=. --plugin=/opt/workspace-go/bin/protoc-gen-go MessageRouting.proto