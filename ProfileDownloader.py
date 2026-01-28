import requests
import asyncio
from BatchedAssetDownloader import download_files
from AssetDownloader import AssetDownloader
import os
import shutil
from typing import Literal

TARGET_OPERATING_SYSTEM = "windows"

class ProfileDownloader:
    
    def __init__(self, profile):
        
        self.profile_version = profile["version"]
        self.libraries = []
        self.asset_index_URL: str | None = None
        self.asset_index_ID = -1
        self.client_URL: str | None = None
        self.class_path = ""
        
        self.loader: Literal["fabric", "forge", "vanilla"] = profile["loader"]["loaderName"]
        self.loader_version = profile["loader"]["loaderVersion"]
        self.main_class = "net.minecraft.client.main.Main"
        self.java_version = profile["java"]
    
    def setup_profile_no_patching(self):
        
        self.get_version_information()
        self.download_client()
        self.download_libraries()
        self.setup_fabric_loader()
        self.download_assets()
        
        self.copy_java()
        self.generate_batch_file()
    
    def setup_profile_forge(self, libraries_list: str):
        
        self.get_version_information()
        self.download_assets()
        
        libs = []
        classpaths = []
        
        with open(libraries_list, "r") as f:
            libs = f.readlines()
        f.close()
        
        for libr in libs:
            lib = libr.strip()
            package = lib.split("libraries/")[1]
            
            os.makedirs(f"dist/libraries/{os.path.dirname(package)}", exist_ok=True)
            
            shutil.copyfile(lib, f"dist/libraries/{package}")
            classpaths.append(f"libraries/{package}")

        os.makedirs(f"dist/libraries/net/minecraftforge/forge/{self.loader_version}/", exist_ok=True)
        shutil.copy(f"extra/forge-{self.profile_version}-{self.loader_version}-client.jar", f"dist/libraries/net/minecraftforge/forge/{self.loader_version}/forge-{self.profile_version}-{self.loader_version}-client.jar")

        classpaths.append(f"libraries/net/minecraftforge/forge/{self.loader_version}/forge-{self.profile_version}-{self.loader_version}-client.jar")

        self.class_path = ";".join(classpaths)
        
        self.copy_java()
        self.main_class = "cpw.mods.modlauncher.Launcher"
        
        self.generate_batch_file()
        
        
        
    
    def copy_java(self):
        
        if os.path.exists("dist/java"): return
        
        print(f"Setting up Java")
        shutil.copytree(f"java/{self.java_version}", "dist/java")
    
    def generate_batch_file(self):
        
        TEMPLATE =  f"""
@echo off

set CP="{self.class_path}"
set ASSET_INDEX={self.asset_index_ID}
set USERNAME=%1

echo Launching Minecraft... Please wait...
echo > New: You can now close this window without closing the game!

.\\java\\bin\\java.exe -Xmx4G -Xms1G ^
--enable-native-access=ALL-UNNAMED ^
-cp "%CP%" ^
{self.main_class} ^
{"--launchTarget forgeclient ^" if self.loader == "forge" else ""}
--accessToken 0 ^
--version {f"{self.profile_version}-forge-{self.loader_version}" if self.loader == "forge" else self.profile_version} ^
--assetsDir %CD%\\assets ^
--assetIndex %ASSET_INDEX% ^
--username %USERNAME%
        """
        
        with open("dist/launch.bat", "w") as f:
            f.write(TEMPLATE)
        
        
    
    def download_client(self):
        
        client_path = f"dist/libraries/minecraft-{self.profile_version}-client.jar"
        self.class_path = ";".join([f"libraries/minecraft-{self.profile_version}-client.jar", self.class_path])
        if (os.path.exists(client_path)):
            return
        
        print("Downloading client...")
        
        if self.client_URL == None:
            raise ValueError("Client URL not found")
        
        asyncio.run(download_files([(self.client_URL, f"dist/libraries/minecraft-{self.profile_version}-client.jar")]))
        
        
    
    def download_libraries(self):
        
        print("Downloading libraries...")
        
        asyncio.run(download_files(self.libraries))
    
    def setup_fabric_loader(self):
        
        if self.loader == "vanilla": return
        
        if self.loader == "forge": return
        
        version_data = requests.get(f"https://meta.fabricmc.net/v2/versions/loader/{self.profile_version}/{self.loader_version}/profile/json")
        if version_data.status_code != 200:
            raise ValueError("Version not found")
        
        version_data = version_data.json()
        
        self.main_class = version_data["mainClass"]
        
        libraries = version_data["libraries"]
        
        additional_libraries = []
        class_paths = []
        
        for library in libraries:
            
            name = library["name"]
            directory_path = name.replace(".", "/").split(":")[0]
            splitted = name.split(":")
            directory_path = f"{directory_path}/{splitted[1]}/{splitted[2]}"
            file_name = splitted[1] + "-" + splitted[2] + ".jar"
            
            url = f"https://maven.fabricmc.net/{directory_path}/{file_name}"
            fs_path = f"dist/libraries/{file_name}"
            
            class_paths.append(f"libraries/{file_name}")
            
            if os.path.exists(fs_path): continue
            
            additional_libraries.append(
                (url, fs_path)
            )
        
        asyncio.run(download_files(additional_libraries))
        
        self.class_path = ";".join([self.class_path, *class_paths])
        
    
    def download_assets(self):
        
        downloader = AssetDownloader()
        
        if self.asset_index_URL == None:
            raise ValueError("Asset index URL not found")
        if self.asset_index_ID == -1:
            raise ValueError("Asset index ID not found")
        
        downloader.downloadAssets(self.asset_index_URL, self.asset_index_ID)
        
    def get_version_information(self):
        
        versions_list = requests.get(f"https://launchermeta.mojang.com/mc/game/version_manifest.json").json()
        
        version_manifest_URL = None
        
        for version in versions_list["versions"]:
            
            id = version["id"]
            
            if id == self.profile_version:
                
                version_manifest_URL = version["url"]
                
                print(f"Version manifest: {version_manifest_URL}")
                
                break
            
        if version_manifest_URL == None:
            raise ValueError("Version not found")
        
        
        # Get version manifest
        
        version_manifest_data = requests.get(version_manifest_URL).json()
        
        self.asset_index_URL = version_manifest_data["assetIndex"]["url"]
        self.asset_index_ID = int(version_manifest_data["assetIndex"]["id"])
        self.client_URL = version_manifest_data["downloads"]["client"]["url"]
        
        # Get libraries
        
        class_paths = []
        
        for library in version_manifest_data["libraries"]:
            
            available_for_os = False
            
            if library.get("rules"):
                for rule in library["rules"]:
                    if rule["action"] == "allow" and rule["os"]["name"] == TARGET_OPERATING_SYSTEM:
                        available_for_os = True
                        break
            else:
                available_for_os = True
            
            artifact_url = library["downloads"]["artifact"]["url"]
            
            if available_for_os == False:
                print(f"Library {artifact_url} not available for {TARGET_OPERATING_SYSTEM}")
                continue
            
            
            artifact_path = library["downloads"]["artifact"]["path"]
            
            splitted = artifact_path.split("/")
            artifact_jar_name = splitted[len(splitted) - 1]
            
            path = f"dist/libraries/{artifact_jar_name}"
            class_paths.append(f"libraries/{artifact_jar_name}")
            
            if os.path.exists(path) == True:
                print(f"Library already downloaded: {path}")
                continue            
            self.libraries.append((artifact_url, path))
            
            
        
        
        self.class_path = ";".join([self.class_path, *class_paths])
        
        
        
        
        
    