# MCPL

> [!NOTE]
> MCPL is experimental and may have issues in certain environments. It may fail to install or get stuck on a specific task/file.

> [!NOTE]
> MCPL is only available for Windows at the moment.

Custom tooling to get Minecraft Java edition to run portably in a single self-extracting launcher/installer. Developed originally as an experiment to see if it was possible to run Minecraft: Java Edition on a school computer (the answer was yes). These tools will help you create a portable EXE that will be able to run Minecraft on any Windows computer. This repository does not include any assets or code from Mojang, there are scripts inside that are dedicated to downloading them and setting up the environment.

# Features
- **Fully self-contained**: Not dependent on DLLs, system files, or any other system library. All libraries are statically linked.
- **Secure Distribution**: Allows copies to expire to limit distribution, with secure tokens signed with the modern digital signing algorithm: EdDSA.
- **Fully offline**: Compiled clients do not need to download files from Mojang, allowing it to run even on restricted networks and decreasing initial startup time.
- **Resilient Open to LAN multiplayer**: StealthPipe (included by default, can be removed) allows Open to LAN to work on hostile networks. See the other repository for more info.
- **Resilient to Reverse Engineering**: Symbols are stripped automatically from the final production build

> [!NOTE]
> Due to the fully offline nature of MCPL, a prebuilt EXE cannot be distributed, as that would violate copyright laws.

# Creating your environment

It should be noted that these steps are not trivial to follow if you aren't a developer or software engineer.

There are many steps required as to creating your own portable EXE for personal use. For this, you need first, create the zip contents, zip it up, and compile with the launcher to create a runnable Go binary. This project was designed to work with Windows only, so you may have many issues doing this for any other operating system other than Windows, such as Linux.

### Java
You will need to download a Java runtime as a ZIP and etract it under the Java folder in dist. This will allow Minecraft to run in the first place. A JRE is recommended over a JDK, since JREs are smaller and take up less space than a JDK. Any Java runtime should work, like Hotspot or Adoptium.

### Creating the distribution contents
To get started, you need to have some knowledge about Python and its packages. Since no requirements.txt exists yet in the project, you will have to create one manually and figure out the package names to install.

1. **Install Python**: Download and install Python here: [https://www.python.org](https://www.python.org)
   If needed, refresh the terminal session to register the new environment variables.
2. **Create a virtual environment**: Run (Windows) to create a virtual environment:
   ```
   py -m venv .venv
   ```
3. **Install dependencies**: Run at the root of the project:
   ```
   py -m pip install -r requirements.txt
   ```
4. **Modify the profile**: Have a look at `profiles/1_20_1_profile.json`. This is an example for 1.21.11 fabric 0.18.3 (weird, I know).
5. **Setting up a Java installation**: For the Java version, it will simply copy the contents of the java installation from the `java/` folder under the root of the project. Install the version used by your profile under the `java` folder. By default, `java/java25` should contain the JRE 25 installation (just like the previous step).
6. **Run the file**: Run `main.py`. This will download all required assets from Mojang and the Fabric loader.

> [!NOTE]
> Forge and Neoforge are not supported yet!

After you have confirmed everything works, it is time to package it up, and that is what the next step will be about.

### Packing up the distribution contents

For this, you will need to install the 7-Zip file manager, as it provides an extremely efficient LZMA2 compression algorithm. Once you have that installed, compress the whole folder in a .7z file. Then, move the dist.7z (or whatever the name) created by 7-Zip to the launcher/ folder, and rename it to data.7z. Now that you have done that, the file is ready to be packed within the executable.

### Signing, and tokens

This launcher is designed for limited distribution in mind. To achieve that, it uses offline tokens, which are designed to be validatable offline with a hardcoded public key.

Make sure you CD into the signing directory before running these scripts.
After you have made sure you have all the required packages installed for all three of these scripts, you may run keyGen first. This will create a public & private keypair. Keep the private one private, never share or leak it.
Then, run signing.py. This will ask you for a username, which allows you to track identity if it ever gets leaked, but that's hard to say for sure, since the token string is sometimes harder to find. Enter one, and press enter. Then, select a duration. Finally, a token will be created. **The username schema must be SOMETHING (SOMETHING ELSE)**.

> [!CAUTION]
> **Never** share or leak the private key! This will make the security provided by tokens obsolete and lead to uncontrolled distribution.

You may use the AutoBuild.py script located at the root of the project to automatically build an EXE for each token found in the signing directory.

After you have a public key, go into the main.go file located under launcher/ and replace the public key with your public key.

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
