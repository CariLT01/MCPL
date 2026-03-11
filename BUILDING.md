# Creating Your Environment (Very Detailed Guide)

**If you plan to fork this project, please see LICENSE**

> [!WARNING]
> The process described below involves several technical steps. If you have never worked with developer tools before, follow every instruction carefully and do not skip steps.

This guide explains how to create your own portable `.EXE` launcher for personal use.

The overall process is:

1. Install required software (Java, Python, 7-Zip).
2. Prepare the project files.
3. Generate the Minecraft distribution files.
4. Compress the files properly.
5. Generate signing keys and tokens.
6. Build the final `.EXE` launcher.

This project is designed **only for Windows**. Running it on Linux or macOS may cause errors.

---

# 1. Install Java (Required for Minecraft)

Minecraft requires Java to run. You must download a **JRE (Java Runtime Environment)** as a **ZIP file**, not an installer.

A JRE is recommended instead of a JDK because it is smaller.

### Step-by-step

1. Open your web browser.

2. Go to this website:

   https://adoptium.net/temurin/releases

3. On the page:

   - Locate the **Java Version selector** (top bar).
   - Choose the required Java version for your Minecraft version.

   Example:
   - Minecraft **1.21.11** requires **Java 25**.

4. In the **Operating System** section:
   - Click **Windows**

5. In the **Package Type** section:
   - Select **JRE**

6. In the **Archive Type** section:
   - Click **ZIP**

7. Click the **Download** button.

8. After the file downloads:
   - Open your **Downloads** folder.
   - Right-click the downloaded `.zip` file.
   - Click **Extract All…**
   - Click **Extract**.

9. Navigate to your project folder.

10. Open the following folder inside your project:

```
dist
```

11. Inside `dist`, create a new folder:

```
java
```

12. Move the extracted Java folder into:

```
dist/java/
```

13. Rename the folder so it follows this format:

```
java/java[VERSION]
```

Example for Java 25:

```
java/java25
```

14. Confirm that the following file exists:

```
java/java25/bin/java.exe
```

If `java.exe` is not inside the `bin` folder, the structure is incorrect.

---

# 2. Install Python

Python is required to generate the distribution files and build the launcher.

### Step-by-step

1. Open your web browser.

2. Go to:

https://www.python.org

3. Click **Downloads**.

4. Click **Download Python 3.13.7** (or the version listed in the project).

5. Open the downloaded installer.

6. IMPORTANT: On the first screen:

   Check the box:

```
Add Python to PATH
```

7. Click:

```
Install Now
```

8. Wait for the installation to finish.

9. Close the installer.

10. Restart your computer.

> [!WARNING]
> If the `py` command does not work later, restarting the computer usually fixes it.

---

# 3. Create a Python Virtual Environment

This isolates the project’s dependencies.

### Step-by-step

1. Open the project folder.

2. In the **address bar** of File Explorer, type:

```
cmd
```

3. Press **Enter**.

A **Command Prompt** window will open in the project directory.

4. Type this command and press **Enter**:

```
py -m venv .venv
```

This creates a virtual environment folder named:

```
.venv
```

---

# 4. Install Project Dependencies

Now install the required Python packages.

In the same command window, type:

```
py -m pip install -r requirements.txt
```

Press **Enter**.

Wait until all packages finish installing.

---

# 5. Configure the Minecraft Profile

The default profile is located at:

```
profiles/1_21_11_profile.json
```

This file controls:

- Minecraft version
- Loader type
- Fabric version
- Java version

### If you want to use the default setup

Skip this section.

### If you want to modify it

1. Open the file:

```
profiles/1_21_11_profile.json
```

2. Open it with **Notepad** or **VS Code**.

3. Edit the following fields.

**Minecraft version**

```
"version"
```

Example:

```
"version": "1.21.11"
```

**Loader type**

```
"loader"
```

Options:

```
fabric
vanilla
```

Example:

```
"loader": "fabric"
```

**Fabric loader version**

```
"loaderVersion"
```

Example:

```
"loaderVersion": "0.18.4"
```

Leave it empty for vanilla:

```
"loaderVersion": ""
```

**Java folder**

```
"java"
```

Example for Java 25:

```
"java": "java25"
```

> [!WARNING]
> Some Minecraft or Fabric versions may not work with this launcher.

Save the file when finished.

---

# 6. Generate the Distribution Files

Now you will generate the necessary launcher files.

1. Open **Command Prompt** in the project root again.

2. Run the command:

```
py main.py
```

This process will download and prepare:

- Minecraft assets
- Libraries
- Game files

Wait until the script finishes.

> [!NOTE]
> Forge and NeoForge are **not supported** yet.

---

# 7. Install 7-Zip

You need 7-Zip to compress the files.

### Step-by-step

1. Open your browser.

2. Go to:

https://www.7-zip.org/download.html

3. Download the **64-bit Windows version**.

4. Open the installer.

5. Click:

```
Install
```

6. Wait for installation to finish.

7. Click:

```
Close
```

---

# 8. Create the Distribution Archives

1. Open the folder:

```
dist
```

2. Hold **CTRL** and click these folders:

```
assets
libraries
java
```

3. Right-click the selected folders.

4. Click:

```
7-Zip → Add to archive
```

5. In the window:

Set:

```
Archive format: 7z
```

6. Name the archive:

```
static.7z
```

7. Click **OK**.

---

### Create dynamic archive

1. Still inside `dist/`.

2. Select **all remaining folders** EXCEPT:

```
assets
libraries
java
```

3. Right-click the selection.

4. Click:

```
7-Zip → Add to archive
```

5. Set:

```
Archive format: 7z
```

6. Name it:

```
dynamic.7z
```

7. Click **OK**.

---

# 9. Move Archives to Launcher Data Folder

Move both files:

```
static.7z
dynamic.7z
```

Into:

```
launcher/src/data/bin
```

---

# 10. Generate Delta Files

Open Command Prompt in the project root.

Run:

```
py DeltaUpload.py
```

Two files will be created:

```
skippable.bin
delta.bin
```

Move them into:

```
launcher/src/data/bin
```

Then rename:

```
skippable.bin
```

to

```
unskippable.bin
```

---

# 11. Add Launcher Background

Create two PNG images:

```
launcher_background.png
launcher_background_blurred.png
```

Requirements:

```
Format: PNG
Resolution: 640x400
```

Place them in:

```
launcher/src/data/bin
```

---

# 12. Generate Signing Keys and Tokens

This system limits who can run the launcher.

1. Open Command Prompt.

2. Navigate to the signing directory:

```
cd signing
```

3. Run:

```
py keyGen.py
```

This creates:

```
public key
private key
```

> [!WARNING]
> Never share the private key with anyone.

---

### Create a token

Run:

```
py signing.py
```

The script will ask for:

**Username**

Example format:

```
Bob (friend)
Alice (tester)
```

Then choose a **token duration**.

A token will be generated.

---

# 13. Add the Public Key to the Launcher

Open:

```
launcher/src/config/Config.go
```

Find the public key section.

Replace the existing key with your newly generated public key.

Save the file.

---

# 14. Create token.go

Navigate to:

```
launcher/src/data
```

Create a new file:

```
token.go
```

Paste the following code inside:

```go
package main

var ISSUED_TOKEN = "YOUR_TOKEN_HERE"
```

If you use **AutoBuild**, you do not need to manually paste the token.

---

# 15. Build the Launcher

The easiest method is the AutoBuild script.

Open Command Prompt in the **project root**.

Run:

```
py AutoBuild.py
```

This script will:

- Compile the Go launcher
- Insert tokens
- Build an `.EXE` for each token

The finished executables will appear in:

```
build/
```

---

# 16. Testing the Launcher

1. Open the `build` folder.

2. Double-click the `.EXE` file.

3. Verify that:

- The launcher opens
- The background loads
- Minecraft launches correctly

---

> [!WARNING]
> Distributing this launcher widely is **your responsibility**. The project authors are not liable for misuse.

See the `LICENSE` file for details.