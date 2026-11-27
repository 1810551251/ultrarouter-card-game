from pathlib import Path
import json,sys

records=[]
for path in sorted(Path('/out').glob('*-tests.jsonl')):
    tests={};failures=[];skipped=[];packages={};output={};build=[]
    for line in path.read_text().splitlines():
        try:event=json.loads(line)
        except ValueError:continue
        action=event.get('Action');test=event.get('Test');package=event.get('Package')
        if event.get('Output'):
            output.setdefault(package+'/'+test if test else str(package),[]).append(event['Output'])
            if action=='build-output':build.append(event['Output'])
        if test and action in ('pass','fail','skip'):
            tests[package+'/'+test]=action
            if action=='fail':failures.append(package+'/'+test)
            if action=='skip':skipped.append(package+'/'+test)
        elif not test and action in ('pass','fail'):packages[package]=action
    row=dict(revision=path.name.removesuffix('-tests.jsonl'),passed=sum(x=='pass' for x in tests.values()),failed=len(failures),skipped=len(skipped),failures=failures,skip_names=skipped,packages=packages)
    records.append(row)
    for name in failures:
        print('FAILURE '+name+'\n'+''.join(output.get(name,[]))[-10000:])
    for name,result in packages.items():
        if result=='fail':print('PACKAGE '+name+'\n'+''.join(output.get(name,[]))[-6000:])
    if build:print('BUILD '+''.join(build)[-6000:])
Path('/out/summary.json').write_text(json.dumps(records,indent=2)+'\n')
print('SUMMARY '+json.dumps(records))
for path in Path('/out').rglob('*'):
    if path.is_file():path.chmod(0o644)
