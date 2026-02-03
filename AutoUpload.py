import time, os, base64
from playwright.sync_api import sync_playwright
from playwright_stealth import Stealth

PROFILE_DIR = "SECRET/browserProfile/"
UPLOAD_URL = "https://csmv-my.sharepoint.com/my?id=%2Fpersonal%2F2473858%5Fcsmv%5Fqc%5Fca%2FDocuments%2FSecondaire%2FSecondaire%203%2Fdev&viewid=9c0d0ee0%2Dca45%2D4c6b%2D9d5d%2Dea5db53331aa"

class UploadWorker:
    
    def __init__(self, filePath: str):
        
        self.file = filePath
    
    def initialize_and_run(self):
        
        stealth = Stealth()
        
        with stealth.use_sync(sync_playwright()) as p:
            
            self.context = p.chromium.launch_persistent_context(
                user_data_dir=PROFILE_DIR,
                headless=False,
                args=["--disable-blink-features=AutomationControlled"]
            )
            
            stealth.apply_stealth_sync(self.context)
            
            self._goto_upload_page()
            
            self._drop_file()
            
            time.sleep(9999999)
            
    def _goto_upload_page(self):
        
        self.page = self.context.new_page()
        self.page.goto(UPLOAD_URL)
        
    def _drop_file_execute(self, file_path: str):
        
        dropzone_selector = '[data-automationid="main"]'
        
        file_name = os.path.basename(file_path)

        with open(file_path, "rb") as f:
            file_bytes = f.read()

        file_base64 = base64.b64encode(file_bytes).decode()

        self.page.evaluate(
            """
            async ({ selector, fileName, fileBase64 }) => {
                const dropZone = document.querySelector(selector);
                if (!dropZone) {
                    throw new Error("Dropzone not found");
                }

                const binary = atob(fileBase64);
                const array = new Uint8Array(binary.length);
                for (let i = 0; i < binary.length; i++) {
                    array[i] = binary.charCodeAt(i);
                }

                const file = new File([array], fileName);
                const dataTransfer = new DataTransfer();
                dataTransfer.items.add(file);

                const events = ["dragenter", "dragover", "drop"];
                for (const eventType of events) {
                    const event = new DragEvent(eventType, {
                        dataTransfer,
                        bubbles: true,
                        cancelable: true,
                    });
                    dropZone.dispatchEvent(event);
                }
            }
            """,
            {
                "selector": dropzone_selector,
                "fileName": file_name,
                "fileBase64": file_base64,
            },
        )
        
    def _drop_file(self):
        
        dropzone_selector = '[data-automationid="main"]'
        
        self.page.wait_for_selector(dropzone_selector, state="visible")
        
        print("Dropzone is ready")
        
        self._drop_file_execute(self.file)
        
        print("File dropped")


TEST_FILE = "build/MC_1.21.11_L1_EVALUATION.exe"

if __name__ == "__main__":
    
    worker = UploadWorker(TEST_FILE)
    worker.initialize_and_run()
    
    time.sleep(999999999999999999)