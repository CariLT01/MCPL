# Minecraft Java Portable

Custom tooling to get Minecraft Java edition to run portably in a single self-extracting launcher/installer. Developed originally as an experiment to see if it was possible to run Minecraft: Java Edition on a school computer (the answer was yes). These tools will help you create a portable EXE that will be able to run Minecraft on any Windows computer. This repository does not include any assets or code from Mojang, there are scripts inside that are dedicated to downloading them and setting up the environment.

# Features
These tools allow you to create a fully self-contained executable that will aid you in launching Minecraft: Java Edition portably without prior installation or any system dependencies like DLLs that may cause certain problems. It ensures that it will run predictably on any Windows device. Extra features are included:
- StealthPipe allows Minecraft Open to LAN to work on networks that traditionally block it by rerouting it through WebSockets and a relay. You can see its own technical report under another repository.
- Optimization mods like Sodium & Lithium ensure higher performance on lower-end hardware and fixes many issues with different GPU drivers.

## Creating your environment

For creating your own environment, you must have **extensive prior knowledge in Python, and Go, and general knowledge about Minecraft, operating systems, and build pipelines**. These instructions will not provide troubleshooting steps, so you are on your own if you do encounter issues (although artificial assistant will help).

There are many steps required as to creating your own portable EXE for personal use. For this, you need first, create the zip contents, zip it up, and compile with the launcher to create a runnable Go binary. This project was designed to work with Windows only, so you may have many issues doing this for any other operating system other than Windows, such as Linux.

### Creating the distribution contents
To get started, you need to have some knowledge about Python and its packages. Since no requirements.txt exists yet in the project, you will have to create one manually and figure out the package names to install.

So first, install Python via [https://www.python.org](https://www.python.org). Pick the latest version. Then after, install the packages needed. To figure out what packages are needed, simply run it over and over again, see what is missing, and install it. Sometimes, the import name will differ from the package name, and in this case, you may use an AI assistant to help you out. This isn't ideal, but this will be fixed in a future release.

After, modify the profile used in `main.py`. Have a look at the profiles folder, which will tell Python what version and modloader should be used. By default, Minecraft version 1.21.11 and Fabric 0.18.3 are used. Currently, only the vanilla game and the fabric loaders are supported. Forge and NeoForge are not supported yet, and support for them will come soon.

After, run `main.py`. This will download all required assets from Mojang & Fabric (if it applies to your profile) and create the assets/ folder along with the batch file for launching.
For testing, you can use the batch file for launching the game, but in production, it is recommended to remove it.

After you have confirmed everything works, it is time to package it up, and that is what the next step will be about.

### Packing up the distribution contents

For this, you will need to install the 7-Zip file manager, as it provides an extremely efficient LZMA2 compression algorithm. Once you have that installed, compress the whole folder in a .7z file. Then, move the dist.7z (or whatever the name) created by 7-Zip to the launcher/ folder, and rename it to data.7z. Now that you have done that, the file is ready to be packed within the executable.

### Signing, and tokens

This launcher is designed for limited distribution in mind. To achieve that, it uses offline tokens, which are designed to be validatable offline with a hardcoded public key. There are many steps to generating a valid token. First, notice how the signing directory is empty. To resolve this, you will need to create three scripts: `keygen.py`, `signing.py`, and `verifyTokens.py`. The names can be modified if needed.

Paste this into `keygen.py`. This script will create a public & private keypair for token validation. Make sure you keep this secure, and, especially, keep your private key private.

```python
import base64
from cryptography.hazmat.primitives.asymmetric import ed25519
from cryptography.hazmat.primitives import serialization

# Generate Ed25519 keypair
private_key = ed25519.Ed25519PrivateKey.generate()
public_key = private_key.public_key()

# Extract raw bytes
private_bytes = private_key.private_bytes(
    encoding=serialization.Encoding.Raw,
    format=serialization.PrivateFormat.Raw,
    encryption_algorithm=serialization.NoEncryption()
)

public_bytes = public_key.public_bytes(
    encoding=serialization.Encoding.Raw,
    format=serialization.PublicFormat.Raw
)

# Encode in Base64 for copy-paste
b64_private = base64.b64encode(private_bytes).decode()
b64_public = base64.b64encode(public_bytes).decode()

with open("private.key", "w") as f:
    f.write(b64_private)
with open("public.key", "w") as f:
    f.write(b64_public)
```

Again, like the previous Python code, run it and install all the required packages.

Then, for `signing.py`, paste this code. This will aid you in creating a valid token and selecting a duration, for different levels of trust between your users.

```python
import base64
import jwt
import datetime
from InquirerPy import inquirer

from cryptography.hazmat.primitives.asymmetric import ed25519

with open("public.key") as f:
    b64_public = f.read()
    
with open("private.key") as f:
    b64_private = f.read()

private_bytes = base64.b64decode(b64_private)
public_bytes = base64.b64decode(b64_public)

# Private key for signing
private_key = ed25519.Ed25519PrivateKey.from_private_bytes(private_bytes)

# Public key for verification
public_key = ed25519.Ed25519PublicKey.from_public_bytes(public_bytes)

name = input("Enter Attribution Name >")

numberOfDays = inquirer.select(message="\nHow many days should the user be allowed to use this token?", choices=[5, 7, 14, 21, 30]).execute()

print(f"Authorized for: {numberOfDays} days")

payload = {
    "name": name,
    "iat": datetime.datetime.now(datetime.timezone.utc),
    "exp": datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(days=numberOfDays)
}

token = jwt.encode(payload, private_key, algorithm="EdDSA")

print(f"Token: {token}")

with open(f"token_{name}.txt", "w") as f:
    f.write(token)
```

And finally, you may also paste the code for `verifyTokens.py`. However, it is entirely optional. This script aids you into managing the different tokens you have issued, track expiration dates, and helps you renew tokens when they eventually do expire. If you choose to not include this script, you will have to manage expiration dates yourself, which may not be as straightforward.

```python
import jwt
from datetime import datetime, timezone, timedelta
from InquirerPy import inquirer
from rich.console import Console
from rich.panel import Panel
import os

console = Console()

def get_expiration_status(exp_timestamp):
    if not exp_timestamp: return "N/A"
    now = datetime.now(timezone.utc)
    exp_date = datetime.fromtimestamp(exp_timestamp, tz=timezone.utc)
    diff = exp_date - now
    
    if diff.total_seconds() <= 0:
        return "[bold red]Already expired[/bold red]"
    return f"Expires in {diff.days}d {diff.seconds // 3600}h"

def load_tokens(files):
    token_data = {}
    for file in files:
        try:
            with open(file, "r") as f:
                token = f.read().strip()
                payload = jwt.decode(token, options={"verify_signature": False})
                
                # We use the name (or filename) as the key for the search
                display_name = payload.get("name") or payload.get("sub") or file
                display_name = f"{display_name} - {get_expiration_status(payload.get('exp'))}"
                
                
                token_data[display_name] = {
                    "payload": payload,
                    "filename": file
                }
        except Exception as e:
            console.print(f"[red]Error reading {file}: {e}[/red]")
    return token_data

def check_token_expiration(files):
    for file in files:
        with open(file, "r") as f:
            token = f.read().strip()
            payload = jwt.decode(token, options={"verify_signature": False})
            
            exp = datetime.fromtimestamp(payload.get("exp"))
            name = payload.get("name")
            expDelta = exp - datetime.now()
            if exp - datetime.now() < timedelta(days=5):
                console.print(f"[bold red]Token near expiration date or already expired: '{name}'. Inspect & Renew. Expires in: {str(expDelta.days)} days [/bold red]")
    

def main():
    
    
    console.print(f"[green]TOKEN MANAGEMENT TOOL[/green]")
    console.print(f"Time now: {datetime.now()}")
    
    # 1. Load your files
    files = []
    
    for file in os.listdir("."):
        if file.startswith("token_"):
            files.append(file)
    
    check_token_expiration(files)
    data = load_tokens(files)

    if not data:
        console.print("[red]No valid tokens found.[/red]")
        return

    # 2. The Interactive Selection with Fuzzy Search
    selected_name = inquirer.fuzzy(
        message="Search/Select a token to inspect:",
        choices=list(data.keys()),
        match_exact=False,
    ).execute()

    # 3. Pull the specific data and display it
    token_info = data[selected_name]
    payload = token_info["payload"]
    
    # Beautiful Rich Output
    console.print("\n")
    
    delta = datetime.fromtimestamp(payload.get("exp")) - datetime.now()
    
    
    console.print(Panel(
        f"[bold cyan]Name:[/bold cyan] {selected_name}\n"
        f"[bold cyan]File:[/bold cyan] {token_info['filename']}\n"
        f"[bold cyan]Issued At:[/bold cyan] {datetime.fromtimestamp(payload.get('iat', 0))}\n"
        f"[bold cyan]Expires At:[/bold cyan] {datetime.fromtimestamp(payload.get('exp', 0))}\n"
        f"[bold yellow]Status:[/bold yellow] {get_expiration_status(payload.get('exp'))}\n\n",
        
        title="Token Inspection",
        expand=False,
        border_style="green"
    ))
    print(f"""
This copy will stop working in {delta.days} days at {datetime.fromtimestamp(payload.get('exp'))}.
If you would like to keep using it, please ask me for a new copy.""")

if __name__ == "__main__":
    main()
```

Make sure you CD into the signing directory before running these scripts.
After you have made sure you have all the required packages installed for all three of these scripts, you may run keyGen first. This will create a public & private keypair. Keep the private one private, never share or leak it.
Then, run signing.py. This will ask you for a username, which allows you to track identity if it ever gets leaked, but that's hard to say for sure, since the token string is sometimes harder to find. Enter one, and press enter. Then, select a duration. Finally, a token will be created. **The username schema must be SOMETHING (SOMETHING ELSE)**.

You may use the AutoBuild.py script located at the root of the project to automatically build an EXE for each token found in the signing directory.

After you have a public key, go into the main.go file located under launcher/ and replace the public key with your public key.

### Java
You will need to download a Java runtime as a ZIP and etract it under the Java folder in dist. This will allow Minecraft to run in the first place. A JRE is recommended over a JDK, since JREs are smaller and take up less space than a JDK. Any Java runtime should work, like Hotspot or Adoptium.

### Building Preparations

It's recommended to use the AutoBuild.py script to build. But first, under the launcher directory, you must create a token.go file with the following contents:
```go
package main

var ISSUED_TOKEN = "YOUR_TOKEN_HERE"

```
You don't actually have to paste your token in ISSUED_TOKEN if you are using AutoBuild. It will automatically replace it with the current targetted token.
It is also recommended to change the launcher version, and game version variables in main.go under the launcher directory. This will make it easier to know when to clean and reinstall when the launcher needs updating.
The built executables will be under the build/ folder. You can confirm if it works by opening it up and trying out the game.
Finally, before building, you may also customize the look of the launcher with elements such as a custom background and a cooler name.

# Building
Once you have all the setup done, building should be easy. You can simply run the AutoBuild script located under the root of the project. If you skipped the signing and tokens step, this script won't work and it won't build anything. Make sure you have something named `token_something (something else)` under the 'signing' directory even if the contents are empty. This will build an executable for each user (token). The token is directly baked into the code, so it is hard to reverse-engineer.

# Distribution
It is strongly recommended to not widely distribute it, as it may get you in trouble from violating local rules and enforcement or law enforcement depending on your region, or you may receive DMCAs from Mojang/Microsoft. If you wish to use this, please use it for yourself only and avoid uploading it or sharing it on the Internet.

# Legal Notice
This tool was made to be used with lawful purposes, such as for entertainment on devices that may potentially collect sensitive information such as credentials. We do not support or encourage piracy of any kind.

# License
These tools are licensed under the MIT license. You may fork, redistribute, and modify it for personal or commercial purposes.
