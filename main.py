from ProfileDownloader import ProfileDownloader
import json

if __name__ == "__main__":
    with open("profile/1_20_1_profile.json", "r") as f:
        profile = json.loads(f.read())
        p = ProfileDownloader(profile)
        p.setup_profile_no_patching()

