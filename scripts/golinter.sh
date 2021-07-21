#!/bin/bash
# shellcheck disable=SC2207
set +e

OP="${1}"
path="${2}"

function filter() {
    res=$(
        golangci-lint run --no-config --issues-exit-code=1 --deadline=2m --disable-all \
            --enable=gofmt \
            --enable=gosimple \
            --enable=deadcode \
            --enable=unconvert \
            --enable=interfacer \
            --enable=varcheck \
            --enable=structcheck \
            --enable=goimports \
            --enable=misspell \
            --enable=golint \
            --exclude=underscores \
            --exclude-use-default=false\
            --enable=staticcheck \
    	      --enable=gocyclo \
    	      --enable=staticcheck \
    	      --enable=golint \
    	      --enable=unused \
    	      --enable=gotype \
    	      --enable=gotypex
    )


    if [[ ${#res} -gt "0" ]]; then
        echo -e "${res}"
        exit 1
    fi
}

function test() {
    cd "${path}" >/dev/null || exit
    golangci-lint run --no-config --issues-exit-code=1 --deadline=2m --disable-all \
        --enable=gofmt \
        --enable=gosimple \
        --enable=deadcode \
        --enable=unconvert \
        --enable=interfacer \
        --enable=varcheck \
        --enable=structcheck \
        --enable=goimports \
        --enable=misspell \
        --enable=golint \
        --exclude=underscores \
        --exclude-use-default=false

    cd - >/dev/null || exit
}

function main() {
    if [ "${OP}" == "filter" ]; then
        filter
    elif [ "${OP}" == "test" ]; then
        test
    fi
}

# run script
main