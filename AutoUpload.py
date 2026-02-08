import time, os
from playwright.sync_api import sync_playwright
from playwright_stealth import Stealth
from playwright.sync_api import Page
import ctypes
import re
from tqdm import tqdm
import threading
import shutil

print = tqdm.write

PROFILE_DIR = "SECRET/browserProfile/"
UPLOAD_URL = "https://csmv-my.sharepoint.com/my?id=%2Fpersonal%2F2473858%5Fcsmv%5Fqc%5Fca%2FDocuments%2FSecondaire%2FSecondaire%203%2Fdev&viewid=9c0d0ee0%2Dca45%2D4c6b%2D9d5d%2Dea5db53331aa"

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
                        raise RuntimeError("Profile directory cannot be found, can't copy")
                    print(f"Copying profile to {profile_directory} for multi-threaded execution")
                    shutil.copytree(PROFILE_DIR, profile_directory)
            
                
                        
            self.context = p.chromium.launch_persistent_context(
                user_data_dir=profile_directory,
                headless=True,
                args=["--disable-blink-features=AutomationControlled","--start-maximized"],
                viewport=None
            )
            
            stealth.apply_stealth_sync(self.context)
            
            self._goto_upload_page()
            
            if self.loginLaunch:
                self.page.wait_for_selector('button[data-automationid="AddNew"]', state='visible', timeout=240_000)
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
        
        page.wait_for_selector('button[data-automationid="AddNew"]', state='visible', timeout=240_000)
        
        page.click('button[data-automationid="AddNew"]')
        time.sleep(2)
        print(f"> Clicking on Upload File")
        # Use the file chooser context manager
        with page.expect_file_chooser(timeout=240_000) as fc_info:
            page.click('button[data-automationid="uploadFile"]')  # triggers file chooser
        file_chooser = fc_info.value

        # Set the file programmatically
        print(f"> Adding file")
        file_chooser.set_files(os.path.abspath(file_path))
        print(f"> Wait for splitbuttonprimary")
        print(f"> Click on replace")
        try:
            page.wait_for_selector('span[data-automationid="splitbuttonprimary"]', state='visible', timeout=240_000)
            page.click('span[data-automationid="splitbuttonprimary"]')
        except Exception as e:
            print(f"Failed to find replace button: {e}. Skipping replace button step.")
        time.sleep(5)
        print("Drop complete")
        
    def _track_progress(self):
        self.page.wait_for_selector('div.ms-ProgressIndicator-progressBar', state='visible', timeout=240_000)
        progress_bar = self.page.locator('div.ms-ProgressIndicator-progressBar')
        progress_bar.wait_for(state='visible', timeout=240_000)
        bar = tqdm(total=100, desc=f"Uploading {os.path.basename(self.file)}", position=self.index)
        last = 0
        while True:
            
            if not progress_bar.count():  # 0 means it's gone
                print("Progress bar no longer in DOM. Exiting loop.")
                break
            
            style = progress_bar.get_attribute('style')
            
            match = re.search(r'width:\s*([\d.]+%)', style or '')
            if match:
                width = match.group(1)
                # print(f"Progress bar width: {width}")
                w_perc = float(width.replace("%", ""))
                bar.update(w_perc - last)
                last = w_perc
                
                
            else:
                print("Width not found in style.")
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
        
        threads: list[threading.Thread] = []
        
        for index, task in enumerate(tasks):
            print(f"Starting thread {index} for {task}")
            t = threading.Thread(target=self._execute_task, args=(index, task))
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