"""Print the local Codex ChatGPT email using only app-server account/read."""

import asyncio
import json
import os
import signal


async def read_account():
    proc = await asyncio.create_subprocess_exec(
        "codex", "app-server", "--listen", "stdio://",
        stdin=asyncio.subprocess.PIPE,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.DEVNULL,
        start_new_session=True,
    )

    async def send(message):
        proc.stdin.write((json.dumps(message) + "\n").encode())
        await proc.stdin.drain()

    async def response(request_id):
        while True:
            line = await proc.stdout.readline()
            if not line:
                raise RuntimeError("Transport closed")
            message = json.loads(line)
            if not isinstance(message, dict):
                raise ValueError("Invalid response")
            if message.get("id") == request_id:
                if "error" in message:
                    raise RuntimeError("Account lookup failed")
                return message.get("result")

    try:
        await send({"method": "initialize", "id": 0, "params": {
            "clientInfo": {"name": "whoami", "version": "1.0.0"},
        }})
        await response(0)
        await send({"method": "initialized", "params": {}})
        await send({"method": "account/read", "id": 1,
                    "params": {"refreshToken": False}})
        return await response(1)
    finally:
        if proc.returncode is None:
            try:
                os.killpg(proc.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass
        try:
            await asyncio.wait_for(proc.wait(), timeout=2)
        except asyncio.TimeoutError:
            try:
                os.killpg(proc.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            await proc.wait()


async def main():
    email = "unknown"
    try:
        result = await asyncio.wait_for(read_account(), timeout=15)
        account = result.get("account") if isinstance(result, dict) else None
        if isinstance(account, dict) and account.get("type") == "chatgpt":
            value = account.get("email")
            if isinstance(value, str) and value.strip() and len(value.splitlines()) == 1:
                email = value.strip()
    except (OSError, RuntimeError, ValueError, asyncio.TimeoutError):
        pass
    print(email)


if __name__ == "__main__":
    asyncio.run(main())
