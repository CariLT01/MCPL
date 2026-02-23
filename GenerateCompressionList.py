from pathlib import Path

def loop_and_add_to_set(s: set, path: Path, root: Path):
    
    for item_path in path.rglob("*"):
        if item_path.is_file():
            s.add(str(item_path.relative_to(root)))


assetsFolder = Path("dist/assets")
librariesFolder = Path("dist/libraries")
javaFolder = Path("dist/java")
root = Path("dist")

assetsFolderItems: set = set()
librariesFolderItems: set = set()
javaFolderItems: set = set()

entries = []


loop_and_add_to_set(assetsFolderItems, assetsFolder, root)
loop_and_add_to_set(librariesFolderItems, librariesFolder, root)
loop_and_add_to_set(javaFolderItems, javaFolder, root)

for item in root.rglob("*"):
    if item.is_file():
        if str(item.absolute()) not in assetsFolderItems and str(item.absolute()) not in librariesFolderItems and str(item.absolute()) not in javaFolderItems:
            entries.append(str(item.relative_to(root)))

entries.extend(librariesFolderItems)
entries.extend(javaFolderItems)
entries.extend(assetsFolderItems)

# Deduplicate while preserving order
seen = set()
deduped_entries = []
for e in entries:
    if e not in seen:
        seen.add(e)
        deduped_entries.append(e)

with open("compressionList.txt", "w") as f:
    f.write("\n".join(deduped_entries))
