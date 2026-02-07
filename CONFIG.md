# Configuration Guide

## Setup

1. Copy the example configuration:
   ```bash
   cp config.json.example config.json
   ```

2. **IMPORTANT:** Change the default password immediately!

## Changing the Password

The example `config.json` includes a default password hash. **You must change this before using the application**, especially in production.

### Generate a new password hash:

**On macOS/Linux:**
```bash
echo -n "your_secure_password" | shasum -a 256
```

**On Windows (PowerShell):**
```powershell
$password = "your_secure_password"
$hash = [System.Security.Cryptography.SHA256]::Create().ComputeHash([System.Text.Encoding]::UTF8.GetBytes($password))
[System.BitConverter]::ToString($hash).Replace("-", "").ToLower()
```

### Update config.json:

Replace the `password_hash` value with your generated hash:

```json
{
  "api_key": "add-token",
  "port": "8080",
  "password_hash": "YOUR_GENERATED_HASH_HERE",
  "token_hashes": []
}
```

## Configuration Fields

- **api_key**: Legacy field (kept for backward compatibility)
- **port**: Server port (default: 8080)
- **password_hash**: SHA-256 hash of your master password
- **token_hashes**: Array of generated token hashes (managed automatically by the app)

## Security Notes

- Never commit `config.json` to version control (it's in `.gitignore`)
- The `password_hash` is used to authenticate token generation requests
- Generated tokens are stored as hashes in `token_hashes`
- For production deployments, use environment variables instead of config files
