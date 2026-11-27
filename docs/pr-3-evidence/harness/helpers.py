"""Copy only named fixture functions verbatim from the tested revision."""
from pathlib import Path
import re
import sys

legacy = sys.argv[1] == 'legacy'
imports = '''package model_test
import (
    "testing"
    "time"
    "github.com/QuantumNous/new-api/model"
    "github.com/stretchr/testify/require"
'''
if legacy:
    imports += '''    "fmt"
    "sync/atomic"
    "github.com/QuantumNous/new-api/common"
'''
body = imports + ')\n'
if legacy:
    body += 'var gachaTestUserSequence atomic.Int64\nconst gachaTestRunID = "pr3-fixture"\n'
helpers = {'gacha_game_test.go': ['gachaState']}
if legacy:
    helpers['gacha_game_test.go'] += ['seedGachaState', 'strongGachaState', 'gachaUser']
else:
    helpers['gacha_voyage_sweep_test.go'] = ['requireGachaTestMail']
for filename, names in helpers.items():
    source = (Path('backend/model') / filename).read_text(encoding='utf-8')
    for name in names:
        (function,) = re.findall(rf'^func {name}\(.*?^\}}$', source, re.M | re.S)
        body += function + '\n'
Path('model/review_helpers_test.go').write_text(body, encoding='utf-8')
