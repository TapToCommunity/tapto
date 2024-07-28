#!/usr/bin/env python3

import os
import json
import hashlib
import sys
import time
from pathlib import Path
from typing import TypedDict, Union, Optional, List

DB_ID = "mrext/tapto"
DL_URL_PREFIX = "https://github.com/wizzomafizzo/tapto/releases/download/{}"
ZIP_FILENAME = "tapto-mister_arm-{}.zip"
SCRATCH_FOLDER = "_scratch"
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
    zip_filename = ZIP_FILENAME.format(tag[1:])
    
    folders: RepoDbFolders = {
        "Scripts/": RepoDbFoldersItem(tags=None),
    }

    files: RepoDbFiles = {}
    for file in FILES:
        local_path = os.path.join(SCRATCH_FOLDER, file)

        key = "Scripts/{}".format(os.path.basename(local_path))
        size = os.stat(local_path).st_size
        md5 = hashlib.md5(open(local_path, "rb").read()).hexdigest()
        url = "{}/{}/{}".format(DL_URL_PREFIX.format(tag), zip_filename, os.path.basename(local_path))

        file_entry = RepoDbFilesItem(
            hash=md5, size=size, url=url, overwrite=None, reboot=True, tags=["tapto"]
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
    
    # set up release files
    Path(SCRATCH_FOLDER).mkdir(parents=True, exist_ok=True)
    zip_filename = ZIP_FILENAME.format(tag[1:])
    os.system("wget {}/{} -O {}/{}".format(DL_URL_PREFIX.format(tag), zip_filename, SCRATCH_FOLDER, zip_filename))
    os.system("unzip {}/{} -d {}".format(SCRATCH_FOLDER, zip_filename, SCRATCH_FOLDER))

    repo_db = create_tapto_db(tag)
    with open("{}/tapto.json".format(REPO_FOLDER), "w") as f:
        f.write(generate_json(repo_db))

if __name__ == "__main__":
    main()
