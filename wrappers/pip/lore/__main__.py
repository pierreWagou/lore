"""Entry point: downloads the lore binary on first run, then execs it."""

import os
import platform
import stat
import sys
import urllib.request
from importlib.metadata import version
from pathlib import Path

_PLATFORM_MAP = {
    "Darwin": "darwin",
    "Linux": "linux",
    "Windows": "windows",
}

_ARCH_MAP = {
    "x86_64": "amd64",
    "AMD64": "amd64",
    "aarch64": "arm64",
    "arm64": "arm64",
}

_VERSION = version("lore-agent")


def _binary_path() -> Path:
    cache = Path(os.environ.get("XDG_CACHE_HOME", Path.home() / ".cache"))
    ext = ".exe" if platform.system() == "Windows" else ""
    return cache / "lore" / f"lore-{_VERSION}{ext}"


def _download_binary(dest: Path) -> None:
    system = _PLATFORM_MAP.get(platform.system())
    machine = _ARCH_MAP.get(platform.machine())

    if not system or not machine:
        sys.exit(
            f"lore: unsupported platform {platform.system()}/{platform.machine()}"
        )

    ext = ".exe" if system == "windows" else ""
    binary_name = f"lore-{system}-{machine}{ext}"
    url = (
        f"https://github.com/pierreWagou/lore/releases/download"
        f"/v{_VERSION}/{binary_name}"
    )

    dest.parent.mkdir(parents=True, exist_ok=True)
    print(f"lore: downloading {binary_name} v{_VERSION}...", file=sys.stderr)

    try:
        urllib.request.urlretrieve(url, dest)
    except urllib.error.HTTPError as exc:
        sys.exit(
            f"lore: download failed ({exc})\n"
            f"      Manual install: go install github.com/pierreWagou/lore@v{_VERSION}"
        )

    dest.chmod(dest.stat().st_mode | stat.S_IEXEC | stat.S_IXGRP | stat.S_IXOTH)
    print(f"lore: installed to {dest}", file=sys.stderr)


def main() -> None:
    binary = _binary_path()
    if not binary.exists():
        _download_binary(binary)

    os.execv(str(binary), [str(binary)] + sys.argv[1:])


if __name__ == "__main__":
    main()
