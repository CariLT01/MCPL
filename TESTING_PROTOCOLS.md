# TESTING PROTOCOLS

**GENERIC CHECKS**
- Launch check: perform a launch, if it crashes: fail
- SP check: two clients, connect, #2 disconnect, #2 reconnect, #1 disconnect. Verify state after with F3.
- WS check: upload world, download world via code, upload world w/ code, download world via code & without code. Delete world from dashboard after.

**TEST #1 -- NEW INSTALLATION**
1. Delete folder if exists
2. Install using new version
3. Perform LAUNCH check and SP check and WS check (3)

**TEST #2 -- UPDATE INSTALLATION**
1. Download old instance from OneDrive
2. Install
3. Perform LAUNCH check and SP check and WS check (3)

**TEST #3 -- RELAUNCH**
1. After performing #2, relaunch
2. Perform LAUNCH check and SP check and WS check (2)

