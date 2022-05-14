# notion-front

A very-very-very simple frontend to serve Notion export as a standalone website.

## How to run

1. Build docker image:

   ```shell
   docker build -t notion-front .
   ```

2. Export Notion content with the following settings:

   * Export format: **HTML**
   * Include content: **Everything**
   * Include subpages: **Yes**
   * Create folders for subpages: **Yes**

3. Run docker image:

   * mount extracted Notion export archive as `/content` volume (this volume path can be overriden using `SOURCE_DIR` variable)
   * mount a cache directory as `/cache` volume (this volume path can be overriden using `CACHE_DIR` variable)
   * bind port `80` (this port path can be overriden using `LISTEN_ADDR` variable)

   Example:

   ```shell
   docker run -d -v $(pwd):/content -p 8000:80 notion-front:latest
   ```

## Parameters

| Env variable  | Default value | Description               |
| ------------- | ------------- | ------------------------- |
| `SOURCE_DIR`  | `/content`    | Path to content directory |
| `CACHE_DIR`   | `/cache`      | Path to cache directory   |
| `LISTEN_ADDR` | `0.0.0.0:80`  | HTTP endpoint to listen   |

## License

[MIT](LICENSE)
