#!/bin/sh
set -eu
root=/src/docs/pr-3-evidence
legacy=7fbc923711ebe7d30c8f737c13b2687124ac1e83
current=d5271c30f34faee8ea88866b3e6548623e11e503
mkdir /pr3-state
for phase in legacy current; do
    case "$phase" in
        legacy) revision=$legacy; testname=ProduceLegacy ;;
        current) revision=$current; testname=UpgradeAndClaim ;;
    esac
    echo "Testing $phase $revision"
    workspace=$(mktemp -d)
    cd "$workspace"
    git -c safe.directory=/src -C /src archive "$revision" backend/model backend/integration/gacha-database.go.example > source.tar
    tar xf source.tar
    mkdir model common
    for file in backend/model/*.go backend/model/*.json; do
        case "$file" in *_test.go) continue ;; esac
        cp "$file" model/
    done
    cp backend/integration/gacha-database.go.example model/bootstrap.go
    cp "$root/harness/common.go.example" common/common.go
    cp "$root/harness/adapter.go.example" model/adapter.go
    cp "$root/harness/go.mod.example" go.mod
    cp "$root/shared_test.go.example" model/review_shared_test.go
    cp "$root/${phase}_test.go.example" model/review_phase_test.go
    python3 "$root/harness/helpers.py" "$phase"
    go mod tidy
    go test -p 1 -parallel 1 -count=1 -v ./model -run "^TestGachaPR3${testname}$"
done
