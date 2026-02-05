import base64
from cryptography.hazmat.primitives.asymmetric import ed25519
from cryptography.hazmat.primitives import serialization

# Generate Ed25519 keypair
private_key = ed25519.Ed25519PrivateKey.generate()
public_key = private_key.public_key()

# Extract raw bytes
private_bytes = private_key.private_bytes(
    encoding=serialization.Encoding.Raw,
    format=serialization.PrivateFormat.Raw,
    encryption_algorithm=serialization.NoEncryption()
)

public_bytes = public_key.public_bytes(
    encoding=serialization.Encoding.Raw,
    format=serialization.PublicFormat.Raw
)

# Encode in Base64 for copy-paste
b64_private = base64.b64encode(private_bytes).decode()
b64_public = base64.b64encode(public_bytes).decode()

with open("private.key", "w") as f:
    f.write(b64_private)
with open("public.key", "w") as f:
    f.write(b64_public)