from pathlib import Path
import argparse
import hashlib
import importlib.util
import json
import sys
import time
import zipfile


ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description='Create a reviewed source archive without local secrets or database files.')
    parser.add_argument('--name', default='starfrost-game', help='Short archive name using letters, digits and hyphens.')
    parser.add_argument('--private-markers', help='Private JSON marker list, kept outside the published files.')
    arguments = parser.parse_args()
    if not arguments.name or any(character not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-' for character in arguments.name):
        raise RuntimeError('Archive name must contain only letters, digits and hyphens.')
    module_path = ROOT/'scripts/audit-release.py'
    spec = importlib.util.spec_from_file_location('release_audit', module_path)
    audit = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(audit)
    files = audit.release_files()
    scanned_hashes = {path:hashlib.sha256(path.read_bytes()).hexdigest() for path in files}
    markers = audit.load_private_markers(arguments.private_markers, files)
    findings, reviewed = audit.scan(files, markers)
    if findings:
        raise RuntimeError('Run python scripts/audit-release.py and resolve its findings before packaging.')
    directory = ROOT/'.local/releases'
    directory.mkdir(parents=True, exist_ok=True)
    target = directory/f'{arguments.name}-{time.strftime("%Y%m%d-%H%M%S")}.zip'
    entries = []
    with zipfile.ZipFile(target, 'x', compression=zipfile.ZIP_DEFLATED) as archive:
        for path in files:
            content = path.read_bytes()
            relative = path.relative_to(ROOT).as_posix()
            if hashlib.sha256(content).hexdigest() != scanned_hashes[path]:
                raise RuntimeError('Source changed while packaging; preserve the partial archive and retry after review.')
            entries.append({'path': relative, 'sha256': hashlib.sha256(content).hexdigest(), 'bytes': len(content)})
            archive.writestr(arguments.name+'/'+relative, content)
        archive.writestr(arguments.name+'/FILE-MANIFEST.json', json.dumps(entries, indent=2)+'\n')
    print('Created source archive: '+str(target))
    print('SHA-256: '+hashlib.sha256(target.read_bytes()).hexdigest())
    print('Original game artwork is included; confirm the independent artwork license in ASSETS.md before public distribution.')


if __name__ == '__main__':
    try:
        main()
    except (RuntimeError, OSError, ValueError) as error:
        print(str(error), file=sys.stderr)
        raise SystemExit(1)
