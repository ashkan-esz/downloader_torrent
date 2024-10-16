# Go Torrent

the torrent downloader and management service of the [downloader_api](https://github.com/ashkan-esz/downloader_api) project.

## Motivation

To Download from torrent and serve local files and online streaming 

## How to use


## Environment Variables

To run this project, you will need to add the following environment variables to your .env file

| Prop                            | Description                                                                              | Required | Default Value |
|---------------------------------|------------------------------------------------------------------------------------------|----------|---------------|
| **`PORT`**                      | server port                                                                              | `false`  | 3000          |
| **`POSTGRES_DATABASE_URL`**     |                                                                                          | `true`   |               |
| **`MONGODB_DATABASE_NAME`**     |                                                                                          | `true`   |               |
| **`MONGODB_DATABASE_URL`**      |                                                                                          | `true`   |               |
| **`REDIS_URL`**                 |                                                                                          | `true`   |               |
| **`REDIS_PASSWORD`**            |                                                                                          | `true`   |               |
| **`SENTRY_DNS`**                | see [sentry.io](https://sentry.io)                                                       | `false`  |               |
| **`SENTRY_RELEASE`**            | see [sentry release](https://docs.sentry.io/product/releases/.)                          | `false`  |               |
| **`ACCESS_TOKEN_SECRET`**       |                                                                                          | `true`   |               |
| **`REFRESH_TOKEN_SECRET`**      |                                                                                          | `true`   |               |
| **`WAIT_REDIS_CONNECTION_SEC`** |                                                                                          | `true`   |               |
| **`CORS_ALLOWED_ORIGINS`**      | address joined by `---` example: https://download-admin.com---https:download-website.com | `false`  |               |
| **`PRINT_ERRORS`**              |                                                                                          | `false`  | false         |
| **`DONT_CONVERT_MKV`**          | dont convert mkv to mp4 on stream request                                                | `false`  | false         |
| **`SERVER_ADDRESS`**            | the url of the server                                                                    | `true`   |               |
| **`MAIN_SERVER_ADDRESS`**       | the url of the downloader_api (main server)                                              | `true`   |               |
| **`DOWNLOAD_ADDRESS`**          | download address prefix like `http://d5.exmaple.com`                                     | `true`   |               |
| **`DOMAIN`**                    | base domain, used for cookies domain and subdomain                                       | `true`   |               |

>**NOTE: check [configs schema](https://github.com/ashkan-esz/downloader_api/blob/master/docs/CONFIGS.README.md) for other configs that read from db.**

## Future updates

- [x]  Fast and light.
- [ ]  Documentation.
- [ ]  Write test.

## Contributing

Contributions are always welcome!

See `contributing.md` for ways to get started.

Please adhere to this project's `code of conduct`.

## Support

Contributions, issues, and feature requests are welcome!
Give a ⭐️ if you like this project!

## Related

- [downloader_api](https://github.com/ashkan-esz/downloader_api)
- [downloader_goChat](https://github.com/ashkan-esz/downloader_gochat)
- [downloader_adminPanel](https://github.com/ashkan-esz/downloader_adminpanel)
- [downloader_app](https://github.com/ashkan-esz/downloader_app)

## Author

**Ashkan Esz**

- [Profile](https://github.com/ashkan-esz "Ashkan esz")
- [Email](mailto:ashkanaz2828@gmail.com?subject=Hi "Hi!")
