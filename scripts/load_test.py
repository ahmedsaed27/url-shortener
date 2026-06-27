import asyncio
import os
import random
import string
import time
from statistics import mean

import aiohttp

BASE_URL = os.getenv("BASE_URL", "http://localhost:8080")

DURATION_SECONDS = int(os.getenv("DURATION_SECONDS", "60"))
CONCURRENCY = int(os.getenv("CONCURRENCY", "200"))
READ_RATIO = 0.9
REQUEST_TIMEOUT_SECONDS = 10
SCENARIO = os.getenv("SCENARIO", "mixed")
BURST_REQUESTS = int(os.getenv("BURST_REQUESTS", "200"))

created_codes = []
latencies_ms = []
status_counts = {}
error_counts = {}
errors = 0
rate_limit_headers_seen = 0


def random_url() -> str:
    token = "".join(random.choices(string.ascii_letters + string.digits, k=16))
    return f"https://example.com/products/{token}"


def record_status(status: int):
    status_counts[status] = status_counts.get(status, 0) + 1


def record_error(exc: Exception):
    global errors

    errors += 1
    name = type(exc).__name__
    error_counts[name] = error_counts.get(name, 0) + 1


def record_rate_limit_headers(resp: aiohttp.ClientResponse):
    global rate_limit_headers_seen

    if resp.status != 429:
        return

    required = ("X-RateLimit-Limit", "X-RateLimit-Remaining", "Retry-After")
    if all(resp.headers.get(header) for header in required):
        rate_limit_headers_seen += 1


async def create_url(session: aiohttp.ClientSession):
    start = time.perf_counter()

    try:
        async with session.post(
            f"{BASE_URL}/api/urls",
            json={"url": random_url()},
        ) as resp:
            record_status(resp.status)
            record_rate_limit_headers(resp)

            if resp.status == 201:
                try:
                    data = await resp.json()
                    if "code" in data:
                        created_codes.append(data["code"])
                except Exception as exc:
                    record_error(exc)
            else:
                await resp.read()

    except Exception as exc:
        record_error(exc)
    finally:
        latencies_ms.append((time.perf_counter() - start) * 1000)


async def redirect_url(session: aiohttp.ClientSession):
    if not created_codes:
        await create_url(session)
        return

    code = random.choice(created_codes)
    start = time.perf_counter()

    try:
        async with session.get(
            f"{BASE_URL}/{code}",
            allow_redirects=False,
        ) as resp:
            record_status(resp.status)
            record_rate_limit_headers(resp)
            await resp.read()

    except Exception as exc:
        record_error(exc)
    finally:
        latencies_ms.append((time.perf_counter() - start) * 1000)


async def worker(session: aiohttp.ClientSession, end_time: float):
    while time.time() < end_time:
        if random.random() < READ_RATIO:
            await redirect_url(session)
        else:
            await create_url(session)


async def run_scenario(session: aiohttp.ClientSession):
    if SCENARIO == "create-limit":
        await asyncio.gather(*(create_url(session) for _ in range(BURST_REQUESTS)))
        return 0

    if SCENARIO == "resolve-limit":
        await create_url(session)
        if not created_codes:
            raise RuntimeError("could not create a URL for the resolve-limit scenario")
        await asyncio.gather(*(redirect_url(session) for _ in range(BURST_REQUESTS)))
        return 0

    if SCENARIO == "normal":
        await create_url(session)
        for _ in range(5):
            await redirect_url(session)
        return 0

    if SCENARIO != "mixed":
        raise ValueError(
            "SCENARIO must be mixed, normal, create-limit, or resolve-limit"
        )

    print("Warming up with 20 URLs...")
    for _ in range(20):
        await create_url(session)
    print(f"Created warmup codes: {len(created_codes)}")

    start_time = time.time()
    end_time = start_time + DURATION_SECONDS
    tasks = [
        asyncio.create_task(worker(session, end_time))
        for _ in range(CONCURRENCY)
    ]
    await asyncio.gather(*tasks)
    return time.time() - start_time


def percentile(values, p):
    if not values:
        return 0

    values = sorted(values)
    index = int((len(values) - 1) * p)
    return values[index]


async def main():
    timeout = aiohttp.ClientTimeout(total=REQUEST_TIMEOUT_SECONDS)
    connector = aiohttp.TCPConnector(
        limit=CONCURRENCY * 2,
        limit_per_host=CONCURRENCY * 2,
        ttl_dns_cache=300,
    )

    async with aiohttp.ClientSession(timeout=timeout, connector=connector) as session:
        print(f"Starting {SCENARIO} scenario...")
        start_time = time.time()
        measured_time = await run_scenario(session)
        total_time = measured_time or (time.time() - start_time)
        total_requests = sum(status_counts.values()) + errors

        print("\n===== Load Test Results =====")
        print(f"Duration: {total_time:.2f}s")
        print(f"Concurrency: {CONCURRENCY}")
        print(f"Total requests: {total_requests}")
        print(f"Requests/sec: {total_requests / total_time:.2f}")
        print(f"Created codes: {len(created_codes)}")
        print(f"Errors: {errors}")
        print(f"Error counts: {error_counts}")
        print(f"Status counts: {status_counts}")
        print(f"429 responses with expected headers: {rate_limit_headers_seen}")

        if latencies_ms:
            print(f"Avg latency: {mean(latencies_ms):.2f} ms")
            print(f"P50 latency: {percentile(latencies_ms, 0.50):.2f} ms")
            print(f"P95 latency: {percentile(latencies_ms, 0.95):.2f} ms")
            print(f"P99 latency: {percentile(latencies_ms, 0.99):.2f} ms")


if __name__ == "__main__":
    asyncio.run(main())
