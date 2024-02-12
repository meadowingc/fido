#!/usr/bin/env bash

go install github.com/go-task/task/v3/cmd/task@latest
go install github.com/cosmtrek/air@latest

sudo apt update && sudo apt install -y python3 python3-pip
pip3 install linkchecker
