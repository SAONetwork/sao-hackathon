// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/collection": {
            "get": {
                "description": "get collection by address",
                "tags": [
                    "Collection"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "The collection id for query",
                        "name": "collectionId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The file id for query",
                        "name": "fileId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The owner for query",
                        "name": "owner",
                        "in": "query"
                    }
                ],
                "responses": {}
            },
            "post": {
                "description": "create collection",
                "tags": [
                    "Collection"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user's ethereum address",
                        "name": "address",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signaturemessage",
                        "name": "signaturemessage",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signature",
                        "name": "signature",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "body for request",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.MockCollection"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/collection/{collectionId}": {
            "delete": {
                "description": "delete collection",
                "tags": [
                    "Collection"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user's ethereum address",
                        "name": "address",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signaturemessage",
                        "name": "signaturemessage",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signature",
                        "name": "signature",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "The collection id for deletion",
                        "name": "collectionId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/collectionFile": {
            "post": {
                "description": "add file to collection",
                "tags": [
                    "Collection"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user's ethereum address",
                        "name": "address",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signaturemessage",
                        "name": "signaturemessage",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signature",
                        "name": "signature",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "body for request",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.MockCollectionRequest"
                        }
                    }
                ],
                "responses": {}
            },
            "delete": {
                "description": "remove file from collection",
                "tags": [
                    "Collection"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user's ethereum address",
                        "name": "address",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signaturemessage",
                        "name": "signaturemessage",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signature",
                        "name": "signature",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "The collection id for deletion",
                        "name": "collectionId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The file id for deletion",
                        "name": "fileId",
                        "in": "query"
                    }
                ],
                "responses": {}
            }
        },
        "/collectionLike": {
            "post": {
                "description": "like collection",
                "tags": [
                    "Collection"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user's ethereum address",
                        "name": "address",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signaturemessage",
                        "name": "signaturemessage",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signature",
                        "name": "signature",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "The collection id for like operation",
                        "name": "collectionId",
                        "in": "query"
                    }
                ],
                "responses": {}
            },
            "delete": {
                "description": "unlike collection",
                "tags": [
                    "Collection"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user's ethereum address",
                        "name": "address",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signaturemessage",
                        "name": "signaturemessage",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signature",
                        "name": "signature",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "The collection id for unlike operation",
                        "name": "collectionId",
                        "in": "query"
                    }
                ],
                "responses": {}
            }
        },
        "/search": {
            "get": {
                "description": "search files, collections and users etc.",
                "tags": [
                    "Search"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user's ethereum address",
                        "name": "address",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signaturemessage",
                        "name": "signaturemessage",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "user's ethereum signature",
                        "name": "signature",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "The key you want to search",
                        "name": "key",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Set search scope, file/collection/user",
                        "name": "scope",
                        "in": "query"
                    }
                ],
                "responses": {}
            }
        }
    },
    "definitions": {
        "main.MockCollection": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "ethAddr": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "labels": {
                    "type": "string"
                },
                "preview": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                },
                "type": {
                    "type": "integer"
                }
            }
        },
        "main.MockCollectionRequest": {
            "type": "object",
            "properties": {
                "collectionIds": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "fileId": {
                    "type": "integer"
                },
                "status": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
