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

numberOfDays = inquirer.select(message="\nHow many days should the user be allowed to use this token? (-1 = test token, 15 seconds)", choices=[5, 7, 14, 21, 30, -1]).execute()

print(f"Authorized for: {numberOfDays} days")

delta = datetime.timedelta(days=max(1, numberOfDays))
if numberOfDays == -1:
    delta = datetime.timedelta(seconds=15)


payload = {
    "name": name,
    "iat": datetime.datetime.now(datetime.timezone.utc),
    "exp": datetime.datetime.now(datetime.timezone.utc) + delta
}

token = jwt.encode(payload, private_key, algorithm="EdDSA")

print(f"Token: {token}")

with open(f"token_{name}.txt", "w") as f:
    f.write(token)