import requests
import os
from tqdm import tqdm
import json
import asyncio
import aiohttp
import aiofiles
import subprocess
from concurrent.futures import ProcessPoolExecutor, as_completed

from BatchedAssetDownloader import download_files



def crunch_file(data):
    """
    data is a tuple: (file_path, mode)
    modes: 'aggressive' (<1s), 'music' (>30s), 'standard' (everything else)
    """
    file_path, mode = data
    temp_path = file_path + ".tmp.ogg"
    
    # Define profiles based on your goals
    if mode == 'aggressive':
        # Extreme savings for short sfx
        audio_opts = ['-ac', '1', '-ar', '32000', '-q:a', '-1']
    elif mode == 'music':
        # Music is non-directional in MC, so Mono is a HUGE win (50% reduction)
        # We keep 44.1kHz and slightly higher quality (-q 0) so it doesn't sound "tinny"
        audio_opts = ['-ac', '2', '-ar', '44100', '-q:a', '3']
    else:
        # Standard: slight VBR reduction
        audio_opts = ['-ac', '1', '-q:a', '2']

    cmd = ['ffmpeg', '-y', '-i', file_path] + audio_opts + ['-loglevel', 'error', temp_path]
    
    try:
        subprocess.run(cmd, check=True)
        if os.path.exists(temp_path):
            if os.path.getsize(temp_path) < os.path.getsize(file_path):
                os.replace(temp_path, file_path)
                return True
            else:
                os.remove(temp_path)
    except Exception:
        if os.path.exists(temp_path): os.remove(temp_path)
    return False

def start_smart_compression(ogg_data_dict):
    """
    ogg_data_dict should be: { file_path: duration_float }
    """
    workers = os.cpu_count() or 1
    
    # Prepare the task list with modes
    tasks = []
    for path, duration in ogg_data_dict.items():
        if duration < 1.0:
            mode = 'aggressive'
        elif duration > 30.0:
            mode = 'music'
        else:
            mode = 'standard'
        tasks.append((path, mode))

    with ProcessPoolExecutor(max_workers=workers) as executor:
        futures = {executor.submit(crunch_file, t): t for t in tasks}
        
        with tqdm(total=len(tasks), desc="Optimizing Assets", unit="file", dynamic_ncols=True) as pbar:
            saved_count = 0
            for future in as_completed(futures):
                if future.result():
                    saved_count += 1
                pbar.update(1)
    print(f"Finished executing")

class AssetDownloader:
    
    def __init__(self):
        ...
    
    def downloadAssets(self, assetIndexUrl: str, versionIndex: int):
        
        assetIndexData = requests.get(assetIndexUrl).json()
        
        # Save it
        
        os.makedirs("dist/assets/indexes", exist_ok=True)
        
        with open(f"dist/assets/indexes/{versionIndex}.json", "w", encoding='utf-8') as f:
            f.write(json.dumps(assetIndexData, separators=(',', ':'), ensure_ascii=False))
        f.close()
        
        # Download assets
        
        
        objects = assetIndexData["objects"]
        
        urls = []
        toDownloadPaths: set = set([])
        
        for name, obj in tqdm(objects.items(), desc="Updating asset index"):
            
            # Remove some bloat
            sname: str = name
            if (sname.startswith("realms/")): continue
            if (sname.startswith("minecraft/lang/")): continue
            
            
            sha1_hash = obj["hash"]
            first_2 = sha1_hash[:2]
            

            
            asset_path = f"dist/assets/objects/{first_2}/{sha1_hash}"
            
            if os.path.exists(asset_path): continue
            
            os.makedirs(f"dist/assets/objects/{first_2}", exist_ok=True)
            
            downloadUrl = f"https://resources.download.minecraft.net/{first_2}/{sha1_hash}"
            
            urls.append((downloadUrl, f"dist/assets/objects/{first_2}/{sha1_hash}"))
            toDownloadPaths.add(f"dist/assets/objects/{first_2}/{sha1_hash}")
        
       
        asyncio.run(download_files(urls))
        
        # Rewrite every JSON file
        
        for name, obj in tqdm(objects.items(), desc="Minifying JSON"):
            sname: str = name
            if sname.endswith(".json") == False: continue
            sha1_hash = obj["hash"]
            first_2 = sha1_hash[:2]
            asset_path = f"dist/assets/objects/{first_2}/{sha1_hash}"
            try:
                jsonObject = json.load(open(asset_path, 'r'))
                with open(asset_path, 'w', encoding='utf-8') as f:
                    json.dump(jsonObject, f, separators=(',', ':'), ensure_ascii=False)
            except Exception as e:
                print(f"Minify failed for: {asset_path}: {e}")
                
                
        return
    
        # Don't compress
        ogg_lengths: dict[str, float] = {}
        
        for name, obj in tqdm(objects.items(), desc="Gathering audio files"):
            
            sname: str = name
            if sname.endswith(".ogg") == False: continue
            sha1_hash = obj["hash"]
            first_2 = sha1_hash[:2]
            asset_path = f"dist/assets/objects/{first_2}/{sha1_hash}"
            if (asset_path in toDownloadPaths) == False: continue
            length = get_ogg_duration(asset_path)
            

            ogg_lengths[asset_path] = length
                
        print(f"Found {len(ogg_lengths)} eligible audio files for compression")
        
        start_smart_compression(ogg_lengths)
            