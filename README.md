# Go Torrent

the torrent downloader and management service of the [downloader_api](https://github.com/ashkan-esz/downloader_api) project.

## Motivation

To Download from torrent and serve local files and online streaming 

## How to use


## Environment Variables

To run this project, you will need to add the following environment variables to your .env file

| Prop                                   | Description                                                                              | Required | Default Value |
|----------------------------------------|------------------------------------------------------------------------------------------|----------|---------------|
| **`PORT`**                             | server port                                                                              | `false`  | 3000          |
| **`MONGODB_DATABASE_NAME`**            |                                                                                          | `true`   |               |
| **`MONGODB_DATABASE_URL`**             |                                                                                          | `true`   |               |
| **`REDIS_URL`**                        |                                                                                          | `true`   |               |
| **`REDIS_PASSWORD`**                   |                                                                                          | `true`   |               |
| **`SENTRY_DNS`**                       | see [sentry.io](https://sentry.io)                                                       | `false`  |               |
| **`SENTRY_RELEASE`**                   | see [sentry release](https://docs.sentry.io/product/releases/.)                          | `false`  |               |
| **`ACCESS_TOKEN_SECRET`**              |                                                                                          | `true`   |               |
| **`REFRESH_TOKEN_SECRET`**             |                                                                                          | `true`   |               |
| **`WAIT_REDIS_CONNECTION_SEC`**        |                                                                                          | `true`   |               |
| **`CORS_ALLOWED_ORIGINS`**             | address joined by `---` example: https://download-admin.com---https:download-website.com | `false`  |               |
| **`MAIN_SERVER_ADDRESS`**              | the url of the downloader_api (main server)                                              | `true`   |               |
| **`PRINT_ERRORS`**                     |                                                                                          | `false`  | false         |

>**NOTE: check [configs schema](https://github.com/ashkan-esz/downloader_api/blob/master/readme/CONFIGS.README.md) for other configs that read from db.**

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
- [downloader_adminPanel](https://github.com/ashkan-esz/downloader_adminpanel)
- [downloader_app](https://github.com/ashkan-esz/downloader_app)

## Author

**Ashkan Esz**

- [Profile](https://github.com/ashkan-esz "Ashkan esz")
- [Email](mailto:ashkanaz2828@gmail.com?subject=Hi "Hi!")
