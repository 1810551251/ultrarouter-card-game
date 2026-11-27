from pathlib import Path
import re

root = Path('/src')
prior = root / 'docs/pr-3-evidence/harness'
adapter = (prior / 'adapter.go.example').read_text()
adapter = adapter.replace('\n\t"gorm.io/driver/sqlite"', '')
adapter = adapter.replace('sqlite.Open(common.SQLitePath)', 'reviewDialector()')
adapter = re.sub(r'^func lockForUpdate.*\n', '', adapter, flags=re.M)
Path('model/adapter.go').write_text(adapter)
common = (prior / 'common.go.example').read_text()
Path('common/common.go').write_text(common + '\ntype RawMessage = json.RawMessage\nconst UserStatusDisabled = 2\n')

source = (root / 'backend/model/gacha_game_test.go').read_text(encoding='utf-8')
helpers = '''package model_test
import (
    "bytes"
    "fmt"
    "net/http"
    "net/http/httptest"
    "sync/atomic"
    "testing"
    "time"
    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/require"
)
var gachaTestUserSequence atomic.Int64
const gachaTestRunID = "consolidated-review"
'''
for name in ['gachaUser', 'gachaState', 'seedGachaState', 'strongGachaState', 'gachaRequest']:
    (function,) = re.findall(rf'^func {name}\(.*?^\}}$', source, re.M | re.S)
    helpers += function + '\n'
Path('model/fixtures_test.go').write_text(helpers, encoding='utf-8')
