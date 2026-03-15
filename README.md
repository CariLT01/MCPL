# MCPL

> [!NOTE]
> MCPL is experimental and may have issues in certain environments. It may fail to install or get stuck on a specific task/file.

> [!NOTE]
> MCPL is only available for Windows at the moment.

Custom tooling to get Minecraft Java edition to run portably in a single self-extracting launcher/installer. Developed originally as an experiment to see if it was possible to run Minecraft: Java Edition on a school computer (the answer was yes). These tools will help you create a portable EXE that will be able to run Minecraft on any Windows computer. This repository does not include any assets or code from Mojang, there are scripts inside that are dedicated to downloading them and setting up the environment.

# Features
- **Fully self-contained**: Not dependent on DLLs, system files, or any other system library. All libraries are statically linked.
- **Offline Signing**: Uses secure tokens signed with the modern digital signing algorithm: EdDSA.
- **Fully offline**: Not dependent on any server or even the Internet.

> [!NOTE]
> Due to the fully offline nature of MCPL, a prebuilt EXE cannot be distributed, as that would violate copyright laws.

# Building

Please see BUILDING.md to create your own launcher!

# Legal Notice
This tool was made to be used with lawful purposes, such as for entertainment on devices that may potentially collect sensitive information such as credentials. We do not support or encourage piracy of any kind.

# License
These tools are licensed under the MIT license. You may fork, redistribute, and modify it for personal or commercial purposes.

IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
