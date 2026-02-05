import re
import jwt
import base64
from cryptography.hazmat.primitives.asymmetric import ed25519

CODE_PATH = input("Enter a path >")
with open(CODE_PATH, "rb") as f:
    data = f.read()

with open("public.key", "rb") as f:
    public_key = ed25519.Ed25519PublicKey.from_public_bytes(base64.b64decode(f.read()))
    
jwt_regex = re.compile(rb"eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+(?:={0,2})\.[A-Za-z0-9_-]+(?:={0,2})")

matches = jwt_regex.findall(data)
for match in matches:
    
    token = match.decode().strip()
    
    print(f"Token: {token}")
    payload: dict = {}
    try:
        payload = jwt.decode(token, public_key, algorithms=["EdDSA"])
    except jwt.InvalidSignatureError:
        print(f"Failed to decode: signature is not valid")
    
    name = payload.get("name")
    iat = payload.get("iat")
    exp = payload.get("exp")
    
    print(f"Name (original attributed owner): {name}")
    print(f"Issued At: {iat}")
    print(f"Expires At: {exp}")
    print("\n")