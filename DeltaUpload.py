import os
import xxhash
import lz4.frame
import struct
from pathlib import Path
from typing import cast

def get_hash_from_file(path: str):
    h = xxhash.xxh64()
    
    with open(path, "rb") as f:
        h = xxhash.xxh64()
        with open(path, "rb") as f:
            for chunk in iter(lambda: f.read(8192), b""):
                h.update(chunk)
        return h.hexdigest()

def loop_through_directory(path_hashes_ptr: dict[str, str], root: Path):
    for item_path in root.rglob("*"):
        if item_path.is_file():
            hash = get_hash_from_file(str(item_path.absolute()))
            
            # print(f"Path: {item_path.absolute()}")
            # print(f"Hash: {hash}")
            
            path_hashes_ptr[str(item_path.absolute())] = hash

pathHashes: dict[str, str] = {}

assetsRoot = Path("dist/assets/")
javaRoot = Path("dist/java")
librariesRoot = Path("dist/libraries")

loop_through_directory(pathHashes, assetsRoot)
loop_through_directory(pathHashes, javaRoot)
loop_through_directory(pathHashes, librariesRoot)

print(pathHashes)

entries: list[bytes] = []

for path, hash in pathHashes.items():
    relative = os.path.relpath(path, "dist")
    
    hashBytes = bytes.fromhex(hash)
    
    relativePath = b"F" + relative.encode("utf-8")
    if relative.startswith("assets\\objects"):
        # take last element
        relativePath = relative.split("\\")[-1]
        
        relativePathBin = b"A" + bytes.fromhex(relativePath)
        
        relativePath = relativePathBin
        # print(relativePathBin)
    
    strSize = struct.pack("B", len(relativePath))
    print(f"path: {relativePath} original: {relative}, hash: {hash} size: {len(relativePath)}")
    
    entry = strSize + relativePath + hashBytes
    
    entries.append(entry)

final_file = b"".join(entries)

with open("delta.bin", "wb") as f:
    f.write(final_file)


# write non-skippable files

root = Path("dist")
skippableFilesEntries: list[bytes] = []

for item_path in root.rglob("*"):
    if item_path.is_file():
        if pathHashes.get(str(item_path.absolute())) == None:
            if os.path.basename(item_path.absolute()) == "compressionList.txt": continue
            relative = os.path.relpath(item_path.absolute(), "dist")
            
            relativePath = relative.encode("utf-8")
            
            strSize = struct.pack("B", len(relativePath))
            
            entry = strSize + relativePath
            
            skippableFilesEntries.append(entry)

with open("skippable.bin", "wb") as f:
    f.write(b"".join(skippableFilesEntries))