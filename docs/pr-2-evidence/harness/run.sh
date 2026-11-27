#!/bin/sh
set -eu
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -qy --no-install-recommends ca-certificates curl gcc libc6-dev git mysql-server postgresql python3
service mysql start
service postgresql start
mysql --protocol=socket -e 'CREATE DATABASE gacha_authority_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci'
runuser -u postgres -- createuser root
runuser -u postgres -- createdb -O root gacha_authority_test
curl -fsSL https://go.dev/dl/go1.25.5.linux-amd64.tar.gz | tar -xz -C /tmp
export PATH=/tmp/go/bin:$PATH
export REVIEW_REVISION=$(git -c safe.directory=/src -C /src rev-parse HEAD)
root=/src/docs/pr-2-evidence/harness
cd "$(mktemp -d)"
mkdir model common controller middleware
for file in /src/backend/model/*.go /src/backend/model/*.json; do
    case "$file" in *_test.go) continue ;; esac
    cp "$file" model/
done
cp /src/backend/model/gacha_ladder_test.go /src/backend/model/gacha_contract_database_test.go model/
cp /src/backend/controller/gacha.go controller/
cp /src/backend/integration/gacha-database.go.example model/bootstrap.go
cp /src/docs/pr-3-evidence/harness/go.mod.example go.mod
cp "$root/engine.go.example" model/engine.go
cp "$root/session.go.example" middleware/session.go
cp "$root/matrix_test.go.example" model/review_test.go
python3 "$root/setup.py"
go mod edit -require=github.com/gin-gonic/gin@v1.10.1 -require=gorm.io/driver/mysql@v1.6.0 -require=gorm.io/driver/postgres@v1.6.0
go test -mod=mod -p 1 -parallel 1 -count=1 -v ./model -run '^TestConsolidatedReview$' > /out/database.txt || { cat /out/database.txt; exit 1; }
cat /out/database.txt
