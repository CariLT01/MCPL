"""
strip_sounds.py: Strips all sounds in the assets folder


"""
import json
import os
from typing import Any

INDEX = 30

with open(f"dist/assets/indexes/{INDEX}.json", "r") as f:
    indexData = json.load(f)

objects: dict[str, Any] = indexData["objects"]

for name, assetObject in objects.items():
    if name.endswith(".ogg"):
        
    
        assetHash = assetObject["hash"]
        assetPath = f"dist/assets/objects/{assetHash[:2]}/{assetHash}"
        if not os.path.exists(assetPath):
            continue
        print(f"Deleting audio: {name} at {assetPath}")
       
        os.remove(assetPath)

# loop through empty folders

dist_path = "dist/assets/objects"

for folder in os.listdir(dist_path):
    files = os.listdir(os.path.join(dist_path, folder))
    if len(files) == 0:
        print(f"Deleting empty folder: {folder}")
        os.rmdir(os.path.join(os.path.normpath(dist_path), folder))