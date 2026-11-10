#!/usr/bin/env python3
"""Assemble an offline Windows QA kit from explicitly supplied local artifacts."""
import argparse
import hashlib
from pathlib import Path, PurePosixPath
import shutil
from zipfile import BadZipFile, ZIP_DEFLATED, ZipFile


ROOT = Path(__file__).resolve().parents[2]
RUNTIMES = (
    ("pwsh_74", "pwsh-7.4.20", "pwsh.exe"),
    ("pwsh_76", "pwsh-7.6.6", "pwsh.exe"),
    ("python", "python", "python.exe"),
    ("pester", "pester", "Pester.psd1"),
)
SOURCE_SUFFIXES = {".cmd", ".cs", ".json", ".ps1", ".psd1", ".psm1", ".py", ".template"}


def add_runtime(archive, source, directory, executable):
    with ZipFile(source) as runtime:
        if executable not in runtime.namelist():
            raise ValueError(f"Runtime archive has no {executable} at its root: {source.name}")
        for entry in runtime.infolist():
            path = PurePosixPath(entry.filename)
            if path.is_absolute() or ".." in path.parts or any(c in entry.filename for c in ("\\", ":")):
                raise ValueError(f"Unsafe path in runtime archive: {source.name}")
            if entry.is_dir():
                continue
            with runtime.open(entry) as content, archive.open(f"runtime/{directory}/{path}", "w") as target:
                shutil.copyfileobj(content, target)


def build(args):
    output = args.output.resolve()
    checksum = output.with_suffix(output.suffix + ".sha256")
    if output.exists() or checksum.exists():
        raise FileExistsError(f"QA kit or checksum already exists: {output}")
    output.parent.mkdir(parents=True, exist_ok=True)
    with output.open("xb") as stream:
        try:
            with ZipFile(stream, "w", ZIP_DEFLATED, compresslevel=3) as archive:
                launcher = (ROOT / "powershell/tests/Run-Windows-QA.cmd").read_text()
                archive.writestr("Run-Windows-QA.cmd", launcher.replace("\n", "\r\n"))
                archive.writestr("README.txt", "Windows x64. Extract this ZIP and run Run-Windows-QA.cmd. Pass 7.6.6 to use that runtime. Return contract-report.json and the console output. The test uses a local mock API and needs no cloud credentials.\r\n")
                archive.write(args.cli, "zstack-cli.exe")
                archive.write(ROOT / "LICENSE", "LICENSE")
                for path in sorted((ROOT / "powershell").rglob("*")):
                    relative = path.relative_to(ROOT)
                    if (path.is_file() and path.suffix in SOURCE_SUFFIXES and "packer" not in relative.parts
                            and not any(part.startswith(".") for part in relative.parts)):
                        archive.write(path, relative.as_posix())
                for attr, directory, executable in RUNTIMES:
                    add_runtime(archive, getattr(args, attr), directory, executable)
        except (BadZipFile, OSError, ValueError):
            stream.close()
            output.unlink()
            raise
    digest = hashlib.sha256()
    with output.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    with checksum.open("x") as stream:
        stream.write(f"{digest.hexdigest()}  {output.name}\n")
    print(output)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cli", type=Path, required=True, help="Windows x64 CLI binary")
    parser.add_argument("--output", type=Path, required=True)
    for attr, directory, executable in RUNTIMES:
        parser.add_argument(f"--{attr.replace('_', '-')}", type=Path, required=True, help=f"ZIP for {directory}, with {executable} at its root")
    build(parser.parse_args())


if __name__ == "__main__":
    main()
