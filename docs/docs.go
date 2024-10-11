// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "description": "get the status of server.",
                "tags": [
                    "System"
                ],
                "summary": "Show the status of server.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/v1/admin/fetch_configs": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Reload db configs and dynamic configs.",
                "tags": [
                    "Admin"
                ],
                "summary": "Fetch Configs",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseOKModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/partial_download/:filename": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "serve files downloaded from torrent",
                "tags": [
                    "Serve-Files"
                ],
                "summary": "serve file",
                "parameters": [
                    {
                        "type": "string",
                        "description": "filename",
                        "name": "filename",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "download/stream range",
                        "name": "Range",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseOKModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/stream/play/:filename": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "stream downloaded file.",
                "tags": [
                    "Stream-Media"
                ],
                "summary": "Stream Media",
                "parameters": [
                    {
                        "type": "boolean",
                        "description": "doesn't convert mkv to mp4",
                        "name": "noConversion",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "crf value for mkv to mp4 conversion",
                        "name": "crf",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseOKModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/stream/status": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "get streaming status and converting files.",
                "tags": [
                    "Stream-Media"
                ],
                "summary": "Stream Status",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.StreamStatusRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/torrent/cancel/:filename": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "cancel downloading torrent file.",
                "tags": [
                    "Torrent-Download"
                ],
                "summary": "Cancel Download",
                "parameters": [
                    {
                        "type": "string",
                        "description": "filename",
                        "name": "filename",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseOKModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/torrent/download/:movieId": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "download from torrent into local storage",
                "tags": [
                    "Torrent-Download"
                ],
                "summary": "Download Torrent",
                "parameters": [
                    {
                        "type": "string",
                        "description": "movieId",
                        "name": "movieId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "link to torrent file or magnet link",
                        "name": "link",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "boolean",
                        "description": "starts downloading now. need permission 'admin_manage_torrent' to work",
                        "name": "downloadNow",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "size of file in megabyte",
                        "name": "size",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.DownloadRequestRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/torrent/extend_expire_time/:filename": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "add time to expiration of local files",
                "tags": [
                    "Torrent-Download"
                ],
                "summary": "Extend Expire Time",
                "parameters": [
                    {
                        "type": "string",
                        "description": "filename",
                        "name": "filename",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseOKWithDataModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/torrent/limits": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Return configs and limits",
                "tags": [
                    "Torrent-Download"
                ],
                "summary": "Link State",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/service.TorrentUsageRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/torrent/my_downloads": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Return my download requests status",
                "tags": [
                    "Torrent-Download"
                ],
                "summary": "My Downloads",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/service.TorrentUsageRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/torrent/my_usage": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Return status of usage and limits",
                "tags": [
                    "Torrent-Download"
                ],
                "summary": "My Usage",
                "parameters": [
                    {
                        "type": "boolean",
                        "description": "send queued downloads",
                        "name": "embedQueuedDownloads",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/service.TorrentUsageRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/torrent/queue_link_state": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Return status of link, its downloading or in the queue. fastest response, use to show live progress",
                "tags": [
                    "Torrent-Download"
                ],
                "summary": "Link State",
                "parameters": [
                    {
                        "type": "string",
                        "description": "link of the queued download",
                        "name": "link",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "name of file",
                        "name": "filename",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/service.TorrentUsageRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/torrent/remove/:filename": {
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "remove downloaded torrent file.",
                "tags": [
                    "Torrent-Download"
                ],
                "summary": "Remove Download",
                "parameters": [
                    {
                        "type": "string",
                        "description": "filename",
                        "name": "filename",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseOKModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/torrent/status": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "get downloading files and storage usage etc.",
                "tags": [
                    "Torrent-Download"
                ],
                "summary": "Torrent Status",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.TorrentStatusRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.ConvertingFile": {
            "type": "object",
            "properties": {
                "duration": {
                    "type": "number"
                },
                "name": {
                    "type": "string"
                },
                "progress": {
                    "type": "string"
                },
                "size": {
                    "type": "integer"
                }
            }
        },
        "model.DownloadQueueStats": {
            "type": "object",
            "properties": {
                "capacity": {
                    "type": "integer"
                },
                "dequeueWorkersSleep": {
                    "type": "boolean"
                },
                "enqueueCounter": {
                    "type": "integer"
                },
                "size": {
                    "type": "integer"
                },
                "workers": {
                    "type": "integer"
                }
            }
        },
        "model.DownloadRequestRes": {
            "type": "object",
            "properties": {
                "downloadingFile": {
                    "$ref": "#/definitions/model.DownloadingFile"
                },
                "message": {
                    "type": "string"
                },
                "queueIndex": {
                    "type": "integer"
                }
            }
        },
        "model.DownloadingFile": {
            "type": "object",
            "properties": {
                "botId": {
                    "type": "string"
                },
                "chatId": {
                    "type": "string"
                },
                "done": {
                    "type": "boolean"
                },
                "downloadedSize": {
                    "type": "integer"
                },
                "enqueueSource": {
                    "type": "string"
                },
                "error": {},
                "metaFileName": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "size": {
                    "type": "integer"
                },
                "startTime": {
                    "type": "string"
                },
                "state": {
                    "type": "string"
                },
                "titleId": {
                    "type": "string"
                },
                "titleName": {
                    "type": "string"
                },
                "titleType": {
                    "type": "string"
                },
                "torrentUrl": {
                    "type": "string"
                },
                "userId": {
                    "type": "integer"
                }
            }
        },
        "model.LocalFile": {
            "type": "object",
            "properties": {
                "activeDownloads": {
                    "type": "integer"
                },
                "downloadLink": {
                    "type": "string"
                },
                "expireTime": {
                    "type": "string"
                },
                "lastDownloadTime": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "size": {
                    "type": "integer"
                },
                "streamLink": {
                    "type": "string"
                },
                "totalDownloads": {
                    "type": "integer"
                }
            }
        },
        "model.Stats": {
            "type": "object",
            "properties": {
                "configs": {
                    "$ref": "#/definitions/model.StatsConfigs"
                },
                "downloadingFilesCurrentSizeMb": {
                    "type": "integer"
                },
                "downloadingFilesFinalSizeMb": {
                    "type": "integer"
                },
                "localFilesSizeMb": {
                    "type": "integer"
                },
                "remainingSpaceMb": {
                    "type": "integer"
                },
                "torrentDownloadTimeoutMin": {
                    "type": "integer"
                },
                "totalFilesSizeMb": {
                    "type": "integer"
                }
            }
        },
        "model.StatsConfigs": {
            "type": "object",
            "properties": {
                "downloadFileSizeLimitMb": {
                    "type": "integer"
                },
                "downloadSpaceLimitMb": {
                    "type": "integer"
                },
                "downloadSpaceThresholdMb": {
                    "type": "integer"
                },
                "torrentDownloadConcurrencyLimit": {
                    "type": "integer"
                },
                "torrentDownloadDisabled": {
                    "type": "boolean"
                },
                "torrentFileExpireDelayFactor": {
                    "type": "number"
                },
                "torrentFileExpireExtendHour": {
                    "type": "integer"
                },
                "torrentFilesExpireHour": {
                    "type": "integer"
                },
                "torrentFilesServingConcurrencyLimit": {
                    "type": "integer"
                },
                "torrentFilesServingDisabled": {
                    "type": "boolean"
                },
                "torrentSendResultToBot": {
                    "type": "boolean"
                },
                "torrentUserEnqueueLimit": {
                    "type": "integer"
                }
            }
        },
        "model.StreamStatusRes": {
            "type": "object",
            "properties": {
                "convertingFiles": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.ConvertingFile"
                    }
                }
            }
        },
        "model.Tasks": {
            "type": "object",
            "properties": {
                "autoDownloader": {
                    "type": "string"
                },
                "dbsInvalidLocalLinksRemover": {
                    "type": "string"
                },
                "diskSpaceCleaner": {
                    "type": "string"
                },
                "expiredFilesRemover": {
                    "type": "string"
                },
                "inCompleteDownloadsRemover": {
                    "type": "string"
                },
                "invalidLocalFilesRemover": {
                    "type": "string"
                },
                "orphanMetaFilesRemover": {
                    "type": "string"
                }
            }
        },
        "model.TorrentStatusRes": {
            "type": "object",
            "properties": {
                "activeDownloadsCounts": {
                    "type": "integer"
                },
                "downloadQueueStats": {
                    "$ref": "#/definitions/model.DownloadQueueStats"
                },
                "downloadingFiles": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.DownloadingFile"
                    }
                },
                "localFiles": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.LocalFile"
                    }
                },
                "stats": {
                    "$ref": "#/definitions/model.Stats"
                },
                "tasks": {
                    "$ref": "#/definitions/model.Tasks"
                }
            }
        },
        "response.ResponseErrorModel": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "errorMessage": {}
            }
        },
        "response.ResponseOKModel": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "errorMessage": {
                    "type": "string"
                }
            }
        },
        "response.ResponseOKWithDataModel": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {},
                "errorMessage": {
                    "type": "string"
                }
            }
        },
        "service.EnqueueSource": {
            "type": "string",
            "enum": [
                "admin",
                "user",
                "user-bot",
                "auto-downloader"
            ],
            "x-enum-varnames": [
                "Admin",
                "User",
                "UserBot",
                "AutoDownloader"
            ]
        },
        "service.QueueItem": {
            "type": "object",
            "properties": {
                "botId": {
                    "type": "string"
                },
                "botUsername": {
                    "type": "string"
                },
                "chatId": {
                    "type": "string"
                },
                "enqueueSource": {
                    "$ref": "#/definitions/service.EnqueueSource"
                },
                "enqueueTime": {
                    "type": "string"
                },
                "titleId": {
                    "type": "string"
                },
                "titleType": {
                    "type": "string"
                },
                "torrentLink": {
                    "type": "string"
                },
                "userId": {
                    "type": "integer"
                }
            }
        },
        "service.TorrentUsageRes": {
            "type": "object",
            "properties": {
                "downloading": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.DownloadingFile"
                    }
                },
                "firstUseAt": {
                    "type": "string"
                },
                "leachLimit": {
                    "type": "integer"
                },
                "localFiles": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.LocalFile"
                    }
                },
                "queueItems": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/service.QueueItem"
                    }
                },
                "queueItemsIndex": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "searchLimit": {
                    "type": "integer"
                },
                "torrentLeachGb": {
                    "type": "number"
                },
                "torrentSearch": {
                    "type": "integer"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "description": "Type \"Bearer\" followed by a space and JWT token.",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "2.0",
	Host:             "download.movieTracker.site",
	BasePath:         "/",
	Schemes:          []string{"https"},
	Title:            "Go Torrent",
	Description:      "Torrent service of the downloader_api project.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
