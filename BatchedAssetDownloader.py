import aiofiles
import aiohttp
import asyncio
import os
from tqdm import tqdm

CONCURRENCY = 20

async def download_file(session, url, dest, pbar):
    async with session.get(url) as response:
        response.raise_for_status()
        async with aiofiles.open(dest, "wb") as f:
            async for chunk in response.content.iter_chunked(1024 * 64):
                await f.write(chunk)
    pbar.update(1)

async def bounded_download(semaphore, session, url, dest, pbar):
    async with semaphore:
        os.makedirs(os.path.dirname(dest), exist_ok=True)
        await download_file(session, url, dest, pbar)


async def download_files(urls: list[tuple[str, str]]):

    semaphore = asyncio.Semaphore(CONCURRENCY)
    timeout = aiohttp.ClientTimeout(total=300)
    connector = aiohttp.TCPConnector(limit=CONCURRENCY)

    with tqdm(total=len(urls), desc="Downloading files") as pbar:
        async with aiohttp.ClientSession(
            timeout=timeout,
            connector=connector
        ) as session:
            tasks = [
                bounded_download(
                    semaphore,
                    session,
                    url[0],
                    url[1],
                    pbar
                )
                for url in urls
            ]
            await asyncio.gather(*tasks)
