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