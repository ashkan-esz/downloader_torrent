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
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.DownloadingFile"
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
        "model.DiskInfo": {
            "type": "object",
            "properties": {
                "configs": {
                    "$ref": "#/definitions/model.DiskInfoConfigs"
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
                "totalFilesSizeMb": {
                    "type": "integer"
                }
            }
        },
        "model.DiskInfoConfigs": {
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
                "torrentFilesExpireHour": {
                    "type": "integer"
                },
                "torrentFilesServingConcurrencyLimit": {
                    "type": "integer"
                }
            }
        },
        "model.DownloadingFile": {
            "type": "object",
            "properties": {
                "downloadedSize": {
                    "type": "integer"
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
                }
            }
        },
        "model.LocalFile": {
            "type": "object",
            "properties": {
                "activeDownloads": {
                    "type": "integer"
                },
                "downloadLinks": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "expireTime": {
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
        "model.TorrentStatusRes": {
            "type": "object",
            "properties": {
                "activeDownloadsCounts": {
                    "type": "integer"
                },
                "diskInfo": {
                    "$ref": "#/definitions/model.DiskInfo"
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
