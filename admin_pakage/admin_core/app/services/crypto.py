import os
import base64
import hashlib
from cryptography.hazmat.primitives.ciphers.aead import AESGCM


def _get_key() -> bytes:
    key = os.getenv("ENCRYPTION_KEY", "changeme_32_byte_encryption_key!!")
    # Derive a 32-byte key from any input using SHA-256
    return hashlib.sha256(key.encode("utf-8")).digest()


def encrypt(plaintext: str) -> str:
    key = _get_key()
    aesgcm = AESGCM(key)
    nonce = os.urandom(12)
    ciphertext = aesgcm.encrypt(nonce, plaintext.encode("utf-8"), None)
    return base64.b64encode(nonce + ciphertext).decode("utf-8")


def decrypt(ciphertext: str) -> str:
    key = _get_key()
    data = base64.b64decode(ciphertext.encode("utf-8"))
    nonce = data[:12]
    ct = data[12:]
    aesgcm = AESGCM(key)
    plaintext = aesgcm.decrypt(nonce, ct, None)
    return plaintext.decode("utf-8")
