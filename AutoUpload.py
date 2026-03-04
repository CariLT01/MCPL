"""
Docstring for AutoUpload

AutoUpload script for automatically uploading builds to OneDrive.
Only works on Windows.
"""

import time, os
from playwright.sync_api import sync_playwright
from playwright_stealth import Stealth
from playwright.sync_api import Page
import ctypes
import re
from tqdm import tqdm
import threading
import shutil
from UploadUrl import UPLOAD_URL


def print(*args):
    """
    Docstring for print

    :param args: Description
    """
    thread_context = threading.current_thread()
    tqdm.write(f"[ {thread_context.name} ]: " + " ".join(map(str, args)))


PROFILE_DIR = "SECRET/browserProfile/"


try:
    ctypes.windll.user32.SetProcessDPIAware()
except Exception:
    pass


class UploadWorker:

    def __init__(self, filePath: str, loginLaunch: bool = False):

        self.file = filePath
        self.loginLaunch = loginLaunch

    def initialize_and_run(self, index: int):

        self.index = index

        stealth = Stealth()

        with stealth.use_sync(sync_playwright()) as p:

            profile_directory = ""
            if self.loginLaunch:
                profile_directory = PROFILE_DIR
            else:
                profile_directory = f"SECRET/browserProfileThreaded_{index}"

            if not self.loginLaunch:
                if os.path.exists(profile_directory) == False:
                    if os.path.exists(PROFILE_DIR) == False:
                        raise RuntimeError(
                            "Profile directory cannot be found, can't copy"
                        )
                    print(
                        f"Copying profile to {profile_directory} for multi-threaded execution"
                    )
                    shutil.copytree(PROFILE_DIR, profile_directory)

            self.context = p.chromium.launch_persistent_context(
                user_data_dir=profile_directory,
                headless=not self.loginLaunch,
                args=[
                    "--disable-blink-features=AutomationControlled",
                    "--start-maximized",
                ],
                viewport=None,
            )

            stealth.apply_stealth_sync(self.context)

            self._goto_upload_page()

            if self.loginLaunch:
                self.page.wait_for_selector(
                    'button[data-automationid="AddNew"]',
                    state="visible",
                    timeout=240_000,
                )
                time.sleep(2)
                self.context.close()
                print("Login complete. Please relaunch")
                exit(0)
                return

            self._drop_file()

    def _goto_upload_page(self):

        self.page = self.context.new_page()
        self.page.goto(UPLOAD_URL, timeout=240_000)
        print(f"Goto upload page")

    def _drop_file_execute(self, page: Page, file_path: str):
        print(f"Waiting 5 seconds before executing drop...")
        print(f"Executing drop... Path: {file_path}")

        print(f"> Clicking on add new file")

        page.wait_for_selector(
            'button[data-automationid="AddNew"]', state="visible", timeout=240_000
        )

        page.click('button[data-automationid="AddNew"]')
        time.sleep(2)
        print(f"> Clicking on Upload File")
        # Use the file chooser context manager
        with page.expect_file_chooser(timeout=240_000) as fc_info:
            page.click(
                'button[data-automationid="uploadFile"]'
            )  # triggers file chooser
        file_chooser = fc_info.value

        # Set the file programmatically
        print(f"> Adding file")
        file_chooser.set_files(os.path.abspath(file_path))
        print(f"> Wait for splitbuttonprimary")
        print(f"> Click on replace")
        try:

            found = False

            try:
                page.wait_for_selector(
                    'span[data-automationid="splitbuttonprimary"]:has-text("Replace")',
                    state="visible",
                    timeout=30_000,
                )
            except Exception as e:
                print(">> Failed to find selector english")
            else:
                print(">> Found selector eng")
                page.click(
                    'span[data-automationid="splitbuttonprimary"]:has-text("Replace")'
                )
                found = True

            if not found:
                try:
                    page.wait_for_selector(
                        'span[data-automationid="splitbuttonprimary"]:has-text("Remplacer")',
                        state="visible",
                        timeout=30_000,
                    )

                except Exception as e:
                    print(">> Failed to find selector french")
                else:
                    print(">> Found selector french")
                    page.click(
                        'span[data-automationid="splitbuttonprimary"]:has-text("Remplacer")'
                    )

        except Exception as e:
            print(f"Failed to find replace button: {e}. Skipping replace button step.")
        time.sleep(5)
        print("Drop complete")

    def _track_progress(self):

        chose_specific_selector = False

        while True:
            try:
                print(">> Waiting for progress bar selector")
                try:
                    self.page.wait_for_selector(
                        'div.ms-ProgressIndicator-progressBar[role="progressbar"]',
                        state="visible",
                        timeout=20_000,
                    )

                except Exception as e:
                    print(f"timed out: {e}, trying next (#2)")
                else:
                    chose_specific_selector = True
                    break
                try:
                    self.page.wait_for_selector(
                        "div.ms-ProgressIndicator-progressBar",
                        state="visible",
                        timeout=20_000,
                    )
                except Exception as e:
                    print(f"timed out: {e}, trying next (#3)")
                else:
                    chose_specific_selector = False
                    break
                try:
                    self.page.wait_for_selector(
                        'div.ms-ProgressIndicator-progressBar[role="progressbar"]',
                        state="visible",
                        timeout=20_000,
                    )
                except Exception as e:
                    raise RuntimeError(
                        "cannot wait for pg bar: not found after 3 methods"
                    )
                else:
                    chose_specific_selector = True
                    break
            except Exception as e:
                print(f"> Failed to wait for progress bar: {e}")
            else:
                break
            time.sleep(2)
        bars = self.page.locator(
            'div.ms-ProgressIndicator-progressBar[role="progressbar"]'
            if chose_specific_selector
            else "div.ms-ProgressIndicator-progressBar"
        )
        count = bars.count()

        # wait for any bar to be visible

        for barIndex in range(count):
            bar = bars.nth(barIndex)
            print(f">> Wait for bar {barIndex} to be visible")
            bar.wait_for(state="visible", timeout=240_000)
            print(f">> Bar {barIndex} is visible")

        bar = tqdm(
            total=100,
            desc=f"Uploading {os.path.basename(self.file)}",
            position=self.index,
        )
        last = 0
        while True:
            try:

                bar_counts = bars.count()

                if bar_counts <= 0:  # 0 means it's gone
                    print("Progress bar no longer in DOM. Exiting loop.")
                    break
                width_accumulated = 0

                for barIndex in range(bar_counts):
                    progress_bar = bars.nth(barIndex)
                    style = progress_bar.get_attribute("style")

                    match = re.search(r"width:\s*([\d.]+%)", style or "")
                    if match:
                        width = match.group(1)
                        width_accumulated += float(width.replace("%", ""))
                        # print(f"[DBG]: Accumulate: {width} at {width_accumulated}")
                    else:
                        print("Width not found in style.")

                width_avg = width_accumulated / bar_counts

                bar.update(width_avg - last)
                last = width_avg

            except Exception as e:
                print(f"> Cannot update progress bar: {e}")
            time.sleep(1)
        bar.close()

    def _drop_file(self):

        dropzone_selector = '[data-automationid="main"]'

        self.page.wait_for_selector(dropzone_selector, state="visible", timeout=240_000)

        print("Dropzone is ready")

        self._drop_file_execute(self.page, self.file)
        self._track_progress()

        print("File dropped")
        time.sleep(2)
        print("Closing browser")
        self.context.close()


class TasksExecutor:

    def __init__(self, build_directory: str = "build/"):
        self.build_directory = build_directory
        pass

    def _get_files(self, build_directory: str):

        files = []

        for item in os.listdir(os.path.abspath(build_directory)):
            if os.path.isfile(os.path.join(os.path.abspath(build_directory), item)):
                if item.endswith(".exe") and item.startswith("MC_"):
                    files.append(os.path.join(os.path.abspath(build_directory), item))

        return files

    def _execute_task(self, index: int, filename: str):

        w = UploadWorker(filename)
        w.initialize_and_run(index)

    def execute_tasks(self):
        tasks = self._get_files(self.build_directory)
        tasks.reverse()
        threads: list[threading.Thread] = []

        for index, task in enumerate(tasks):
            print(f"Starting thread {index} for {task}")
            path: str = task

            filename = path.split("\\")[len(path.split("\\")) - 1]
            t = threading.Thread(
                target=self._execute_task, args=(index, task), name=f"Worker/{filename}"
            )
            t.start()
            threads.append(t)
            time.sleep(0.1)

        for t in threads:
            t.join()


TEST_FILE = "build/MC_1.21.11_L1.1_EVALUATION3.exe"

if __name__ == "__main__":

    if os.path.exists(PROFILE_DIR) == False:
        loginWorker = UploadWorker("LOGIN_WORKER", True)
        loginWorker.initialize_and_run(0)
        exit(0)

    executor = TasksExecutor()
    executor.execute_tasks()
