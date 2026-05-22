"""
Docstring for AutoBuild

Batched Go application builder for each user, each with its own token.
"""

import os
import re

# Collect token files

TOKEN_PATH = "signing/"
LAUNCHER_CODE_PATH = "launcher/src/data/token.go"
VERSION_NAME = "MC_26.1_L1.2"

token_files: list[str] = []
for file in os.listdir(TOKEN_PATH):
    if file.endswith(".txt") and file.startswith("token_"):
        token_files.append(os.path.join(TOKEN_PATH, file))

# Replace & Build
PATTERN = r'var\s+ISSUED_TOKEN\s*=\s*".*?"'

current_cwd = os.getcwd()

for file in token_files:

    # Replace and build

    with open(file, "r") as f:
        token = f.read()

    with open(LAUNCHER_CODE_PATH, "r") as f:
        code = f.read()

    with open(LAUNCHER_CODE_PATH + "_copy.txt", "w") as f:
        f.write(code)

    new_code = re.sub(PATTERN, f'var ISSUED_TOKEN = "{token}"', code)

    if new_code == code:
        raise RuntimeError("Token file not found in launcher code")

    with open(LAUNCHER_CODE_PATH, "w") as f:
        f.write(new_code)
    # Build
    print(f"Building for: {file}")

    filename = file.split("_")[1].split("(")[0].strip()

    print(f"Building to: MC_Launcher_{filename}.exe")

    main_file_abs = os.path.abspath("./launcher/")

    v_name = VERSION_NAME
    if "EXP" in filename:
        v_name = "EXP"

    output_file_abs = os.path.abspath(f"./build/{v_name}_{filename}.exe")

    print(f"Main file: {main_file_abs}")
    print(f"Output file: {output_file_abs}")

    os.chdir("launcher")

    print(f"Building MC_Launcher_{filename}.exe ...")

    os.system(
        f'go build -ldflags="-s -w -H=windowsgui" -trimpath -o {output_file_abs} -v {main_file_abs}'
    )

    os.chdir(current_cwd)
