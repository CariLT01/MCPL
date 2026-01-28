import json
import struct
import lz4.frame
import os
from tqdm import tqdm

# Open asset index
ASSET_INDEX_PATH = "dist/assets/indexes/29.json"

with open(ASSET_INDEX_PATH, "r") as f:
    asset_index = json.load(f)


objects: dict[str, dict[str, str]] = asset_index["objects"]

strings: list[bytes] = []

for key, data in objects.items():
    
    hash = data['hash']
    string_bytes = bytes.fromhex(hash)
    name_bytes = key.encode('utf-8')
    
    packed_string = struct.pack('>B', len(name_bytes)) + name_bytes + string_bytes # String always 20 bytes long
    
    strings.append(packed_string)
    
final_string = b''.join(strings)
compressed = lz4.frame.compress(final_string)

with open("index.mcpack", "wb") as f:
    f.write(compressed)


# Actually form packs

current_pack = []
pack_counter = 0

for data in tqdm(objects.values(), desc="Packing assets"):
    
    hash: str = data["hash"]
    first_2 = hash[:2]
    
    path = f"dist/assets/objects/{first_2}/{hash}"
    
    if os.path.exists(path):
    
        with open(path, "rb") as f:
            content = f.read()
            length = len(content)
            
            packed_string = struct.pack('>I', length) + content
            current_pack.append(packed_string)
    else:
        print(f"Doesn't exist: {path}")
        packed_string = struct.pack(">I", 0)
        current_pack.append(packed_string)
    
    if len(current_pack) >= 500:
        
        final_pack = b''.join(current_pack)
    
        
        with open(f"assetspack{pack_counter}.mcpack", "wb") as f:
            f.write(final_pack)
            
        current_pack = []
        pack_counter += 1
    

# Final

if (pack_counter > 0):

    final_pack = b''.join(current_pack)

    with open(f"assetspack{pack_counter}.mcpack", "wb") as f:
        f.write(final_pack)
        
    current_pack = []
    pack_counter += 1