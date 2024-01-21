#!/usr/bin/env python3

import os
import json
import hashlib
import sys
import time
from pathlib import Path
from zipfile import ZipFile
from typing import TypedDict, Union, Optional, List

DB_ID = "mrext/tapto"
DL_URL = "https://github.com/wizzomafizzo/tapto/releases/download/{}"
RELEASES_FOLDER = "_bin/releases"
REPO_FOLDER = "scripts/mister/repo"
FILES = [
    "tapto.sh",
    "taptui.sh",
]


class RepoDbFilesItem(TypedDict):
    hash: str
    size: int
    url: Optional[str]
    overwrite: Optional[bool]
    reboot: Optional[bool]
    tags: List[str]


RepoDbFiles = dict[str, RepoDbFilesItem]


class RepoDbFoldersItem(TypedDict):
    tags: Optional[list[Union[str, int]]]


RepoDbFolders = dict[str, RepoDbFoldersItem]


class RepoDb(TypedDict):
    db_id: str
    timestamp: int
    files: RepoDbFiles
    folders: RepoDbFolders
    base_files_url: Optional[str]


def create_tapto_db(tag: str) -> RepoDb:
    folders: RepoDbFolders = {
        "Scripts/": RepoDbFoldersItem(tags=None),
    }

    files: RepoDbFiles = {}
    for file in FILES:
        local_path = os.path.join(RELEASES_FOLDER, file)

        key = "Scripts/{}".format(os.path.basename(local_path))
        size = os.stat(local_path).st_size
        md5 = hashlib.md5(open(local_path, "rb").read()).hexdigest()
        url = "{}/{}".format(DL_URL.format(tag), os.path.basename(local_path))

        file_entry = RepoDbFilesItem(
            hash=md5, size=size, url=url, overwrite=None, reboot=None, tags=[Path(local_path).stem]
        )

        files[key] = file_entry

    return RepoDb(
        db_id=DB_ID,
        timestamp=int(time.time()),
        files=files,
        folders=folders,
        base_files_url=None,
    )


def remove_nulls(v: any) -> any:
    if isinstance(v, dict):
        return {key: remove_nulls(val) for key, val in v.items() if val is not None}
    else:
        return v


def generate_json(repo_db: RepoDb) -> str:
    return json.dumps(remove_nulls(repo_db), indent=4)


def main():
    tag = sys.argv[1]

    repo_db = create_tapto_db(tag)
    with open("{}/tapto.json".format(REPO_FOLDER), "w") as f:
        f.write(generate_json(repo_db))

if __name__ == "__main__":
    main()
