#!/usr/bin/env python3
"""Tiny OpenAI-compatible fake server for local integration tests.

It is NOT a model and must never be used for the competition demo. It exists so
the full Go ingestion + SSE + citation path can be tested without an API key.

Usage:
    python scripts/fake-openai-server.py --port 18080
    AI_BASE_URL=http://localhost:18080/v1 AI_API_KEY=test AI_MOCK_MODE=false \
      go run ./cmd/api -ingest
"""
import argparse
import json
import re
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


def embedding(text, dim=64):
    vec = [0.0] * dim
    for ch in text:
        vec[ord(ch) % dim] += 1.0
    norm = sum(v * v for v in vec) ** 0.5 or 1.0
    return [round(v / norm, 6) for v in vec]


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        print("[fake-openai] " + fmt % args, flush=True)

    def _json(self, payload, status=200):
        raw = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def do_POST(self):
        length = int(self.headers.get("Content-Length", "0"))
        try:
            body = json.loads(self.rfile.read(length).decode("utf-8"))
        except Exception:
            self._json({"error": {"message": "bad json"}}, 400)
            return

        if self.path.endswith("/embeddings"):
            inputs = body.get("input") or []
            if isinstance(inputs, str):
                inputs = [inputs]
            data = [{"index": i, "embedding": embedding(text)} for i, text in enumerate(inputs)]
            self._json({"data": data, "model": body.get("model", "fake-embedding")})
            return

        if self.path.endswith("/chat/completions"):
            messages = body.get("messages") or []
            prompt = messages[-1].get("content", "") if messages else ""
            ids = []
            for match in re.findall(r"\[S(\d+)\]", prompt):
                if match not in ids:
                    ids.append(match)
            first = "S" + ids[0] if ids else "S1"
            second = "S" + ids[1] if len(ids) > 1 else first
            answer = (
                "根据馆藏档案与 3D 讲解，这件服饰可以从形制、工艺和纹样三个层面理解。"
                f"馆藏记录提供了名称、民族与地域信息[{first}]；3D 讲解进一步说明了对应部位的结构与工艺[{second}]。"
                "这些内容适合用于导览，但不能作为文物鉴定结论。"
            )
            if not body.get("stream"):
                self._json({
                    "choices": [{"message": {"content": answer}, "finish_reason": "stop"}],
                    "usage": {"prompt_tokens": 80, "completion_tokens": 60, "total_tokens": 140},
                })
                return

            self.send_response(200)
            self.send_header("Content-Type", "text/event-stream; charset=utf-8")
            self.send_header("Cache-Control", "no-cache")
            self.end_headers()
            for i in range(0, len(answer), 18):
                chunk = answer[i:i + 18]
                payload = {"choices": [{"delta": {"content": chunk}}]}
                self.wfile.write(("data: " + json.dumps(payload, ensure_ascii=False) + "\n\n").encode("utf-8"))
                self.wfile.flush()
                time.sleep(0.005)
            usage = {"prompt_tokens": 80, "completion_tokens": 60, "total_tokens": 140}
            self.wfile.write(("data: " + json.dumps({"choices": [], "usage": usage}) + "\n\n").encode("utf-8"))
            self.wfile.write(b"data: [DONE]\n\n")
            self.wfile.flush()
            return

        self._json({"error": {"message": "not found"}}, 404)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int, default=18080)
    args = parser.parse_args()
    print(f"[fake-openai] listening on http://127.0.0.1:{args.port}/v1", flush=True)
    ThreadingHTTPServer(("127.0.0.1", args.port), Handler).serve_forever()
