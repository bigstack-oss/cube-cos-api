package api

import "github.com/swaggo/swag"

const docTemplate = `{
    "openapi": "3.0.0",
    "info": {
        "description": "",
        "title": "Cube COS API",
        "contact": {},
        "version": "1.0.0"
    },
    "paths": {
        "/api/v1/logout": {
            "post": {
                "operationId": "logout",
                "tags": [
                    "Logout"
                ],
                "summary": "Logout from the system and redirect to login page",
                "responses": {
                    "302": {
                        "description": "Found and redirect to login page"
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/me": {
            "get": {
                "operationId": "getMe",
                "tags": [
                    "User Info"
                ],
                "summary": "Retrieve the own user info",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the own user info successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetMeResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Own user info",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "name": "admin"
                                            },
                                            "msg": "fetch own user info successfully",
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch own user info: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters": {
            "get": {
                "operationId": "getDataCenters",
                "description": "Retrieve the list of data centers",
                "tags": [
                    "Data Centers"
                ],
                "summary": "Retrieve the list of data centers",
                "responses": {
                    "200": {
                        "description": "Retrieve the list of data centers successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetDataCentersResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch data centers: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/events": {
            "get": {
                "operationId": "getEvents",
                "tags": [
                    "Events"
                ],
                "summary": "Retrieve the list of events",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "type",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of event to query, the value can be only 'system', 'host', and 'instance'.",
                        "example": "system"
                    },
                    {
                        "in": "query",
                        "name": "start",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The start time of the event to query, the value should be in RFC3339 format (default is 24 hours ago).",
                        "example": "2025-01-01T01:00:00Z"
                    },
                    {
                        "in": "query",
                        "name": "stop",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The end time of the event to query, the value should be in RFC3339 format (default is now).",
                        "example": "2025-01-01T01:00:00Z"
                    },
                    {
                        "in": "query",
                        "name": "pageNum",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The page number of the event chunking to fetch (default is 1).",
                        "example": 1
                    },
                    {
                        "in": "query",
                        "name": "pageSize",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The size per page of the events to return (default is unlimit).",
                        "example": 10
                    },
                    {
                        "in": "query",
                        "name": "watch",
                        "required": false,
                        "schema": {
                            "type": "boolean"
                        },
                        "description": "The toggle to enable http chunked transfer for continues server push.",
                        "example": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of events successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetEventsResponse"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 400
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "invalid 'start' time: 2021-09-01T111:00:00Z"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "bad request"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch events: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/events/abstract": {
            "get": {
                "operationId": "getAbstractedEvents",
                "tags": [
                    "Events"
                ],
                "summary": "Retrieve the abstracted events",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "type",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of event to query, the value can be only 'system', 'host', and 'instance'.",
                        "example": "system"
                    },
                    {
                        "in": "query",
                        "name": "limit",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The limit of the abstracted events to return (default is 10).",
                        "example": 10
                    },
                    {
                        "in": "query",
                        "name": "watch",
                        "required": false,
                        "schema": {
                            "type": "boolean"
                        },
                        "description": "The toggle to enable http chunked transfer for continues server push.",
                        "example": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the abstracted events successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetAbstractedEventsResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Abstracted recent events",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "events": [
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "NET00003I",
                                                        "description": "instance \"ccc449e4-a26c-47ac-afc1-c792ab1ed20a\" at 192.168.0.10 is reachable",
                                                        "host": "",
                                                        "category": "net",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "net",
                                                            "id": "ccc449e4-a26c-47ac-afc1-c792ab1ed20a",
                                                            "ip": "192.168.0.10"
                                                        },
                                                        "time": "2025-02-04T06:05:08Z"
                                                    },
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "NET00003I",
                                                        "description": "instance \"441ddbcb-c6a3-48cd-933c-c416d52032b8\" at 192.168.0.127 is reachable",
                                                        "host": "",
                                                        "category": "net",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "net",
                                                            "id": "441ddbcb-c6a3-48cd-933c-c416d52032b8",
                                                            "ip": "192.168.0.127"
                                                        },
                                                        "time": "2025-02-03T19:30:15Z"
                                                    },
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "NET00003I",
                                                        "description": "instance \"17fa8b83-3f9c-4541-a4a8-10b972f19bd2\" at 10.254.1.149 is reachable",
                                                        "host": "",
                                                        "category": "net",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "net",
                                                            "id": "17fa8b83-3f9c-4541-a4a8-10b972f19bd2",
                                                            "ip": "10.254.1.149"
                                                        },
                                                        "time": "2025-02-03T19:15:15Z"
                                                    },
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "SDN00002I",
                                                        "description": "PROJ001 deleted virtual port 192.168.1.87 (fa:16:3e:b1:5c:89)",
                                                        "host": "",
                                                        "category": "sdn",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "sdn",
                                                            "device_id": "32f4993f-c692-49c9-9b5e-e1bd1014699e",
                                                            "network_id": "cae0e230-9713-4717-a94e-7e2a50cb2d86",
                                                            "port_id": "5f20ef48-a1ea-4ef9-8463-6a790b2026c2"
                                                        },
                                                        "time": "2025-02-03T17:32:09Z"
                                                    },
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "SDN00002I",
                                                        "description": "PROJ001 deleted virtual port 192.168.1.5 (fa:16:3e:ce:ab:6b)",
                                                        "host": "",
                                                        "category": "sdn",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "sdn",
                                                            "device_id": "4222d5e8-ccb2-4069-af14-d8f0fe3fbfea",
                                                            "network_id": "cae0e230-9713-4717-a94e-7e2a50cb2d86",
                                                            "port_id": "1caa0da6-8c58-4a80-9bac-2eda77c23c92"
                                                        },
                                                        "time": "2025-02-03T17:32:09Z"
                                                    },
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "NET00003I",
                                                        "description": "instance \"e5a85381-023c-4087-a607-c74687dc3b52\" at 10.254.1.235 is reachable",
                                                        "host": "",
                                                        "category": "net",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "net",
                                                            "id": "e5a85381-023c-4087-a607-c74687dc3b52",
                                                            "ip": "10.254.1.235"
                                                        },
                                                        "time": "2025-02-03T17:30:15Z"
                                                    },
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "NET00003I",
                                                        "description": "instance \"66cdfddb-efe1-4b67-b864-76c5d106524c\" at 10.254.131.183 is reachable",
                                                        "host": "",
                                                        "category": "net",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "net",
                                                            "id": "66cdfddb-efe1-4b67-b864-76c5d106524c",
                                                            "ip": "10.254.131.183"
                                                        },
                                                        "time": "2025-02-03T17:30:15Z"
                                                    },
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "NET00003I",
                                                        "description": "instance \"f682fc8d-2c8d-4557-9b84-049e5a72b713\" at 192.168.0.157 is reachable",
                                                        "host": "",
                                                        "category": "net",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "net",
                                                            "id": "f682fc8d-2c8d-4557-9b84-049e5a72b713",
                                                            "ip": "192.168.0.157"
                                                        },
                                                        "time": "2025-02-03T17:15:15Z"
                                                    },
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "NET00003I",
                                                        "description": "instance \"6942ec52-f089-434d-ba5c-6137015b91b1\" at 192.168.0.108 is reachable",
                                                        "host": "",
                                                        "category": "net",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "net",
                                                            "id": "6942ec52-f089-434d-ba5c-6137015b91b1",
                                                            "ip": "192.168.0.108"
                                                        },
                                                        "time": "2025-02-03T16:50:15Z"
                                                    },
                                                    {
                                                        "type": "system",
                                                        "severity": "Info",
                                                        "id": "NET00003I",
                                                        "description": "instance \"2d9ce9a4-5fe9-4011-84ae-8ecb20f64594\" at 192.168.0.111 is reachable",
                                                        "host": "",
                                                        "category": "net",
                                                        "service": "",
                                                        "metadata": {
                                                            "category": "net",
                                                            "id": "2d9ce9a4-5fe9-4011-84ae-8ecb20f64594",
                                                            "ip": "192.168.0.111"
                                                        },
                                                        "time": "2025-02-03T16:35:15Z"
                                                    }
                                                ],
                                                "limit": {
                                                    "number": 10,
                                                    "description": "the top 10 recent events"
                                                }
                                            },
                                            "msg": "fetch event abstract successfully",
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 400
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "invalid 'start' time: 2021-09-01T111:00:00Z"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "bad request"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch events: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/healths": {
            "get": {
                "operationId": "GetHealths",
                "tags": [
                    "Health"
                ],
                "summary": "Retrieve the overall health status of all modules",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "watch",
                        "required": false,
                        "schema": {
                            "type": "boolean"
                        },
                        "description": "The toggle to enable http chunked transfer for continues server push.",
                        "example": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of health successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetHealthsResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Health check ok",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "overall": {
                                                    "status": {
                                                        "current": "ok"
                                                    }
                                                },
                                                "services": [
                                                    {
                                                        "name": "clusterLink",
                                                        "category": "core",
                                                        "status": {
                                                            "current": "ok"
                                                        },
                                                        "modules": [
                                                            {
                                                                "name": "link",
                                                                "isAutoRepairable": false,
                                                                "status": {
                                                                    "current": "ok"
                                                                }
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch health successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example2": {
                                        "summary": "Health check ng",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "overall": {
                                                    "status": {
                                                        "current": "ng",
                                                        "description": "ceph has 2 ceph_osd down"
                                                    }
                                                },
                                                "services": [
                                                    {
                                                        "name": "storage",
                                                        "category": "storage",
                                                        "status": {
                                                            "current": "ng"
                                                        },
                                                        "modules": [
                                                            {
                                                                "name": "ceph_osd",
                                                                "isAutoRepairable": false,
                                                                "status": {
                                                                    "current": "ng",
                                                                    "description": "ceph has 2 ceph_osd down"
                                                                }
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch health successfully",
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch health checks: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            },
            "patch": {
                "tags": [
                    "Health"
                ],
                "summary": "Repair the health for all modules",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    }
                ],
                "responses": {
                    "202": {
                        "description": "The Request of the all modules repair is accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PatchRepairResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Accept the repair request",
                                        "value": {
                                            "code": 202,
                                            "msg": "the request of all modules repair is accepted and repairing",
                                            "status": "accepted"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "409": {
                        "description": "Conflict",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 409
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "the repair process is already running"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "conflict"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to request repair"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/healths/{module}": {
            "patch": {
                "operationId": "repairModuleHealth",
                "tags": [
                    "Health"
                ],
                "summary": "Repair the unhealthy module",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "path",
                        "name": "module",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the module to repair. The module name is like 'ceph_osd', 'nova', and so on.",
                        "example": "all, ceph_osd, or other module names"
                    }
                ],
                "responses": {
                    "202": {
                        "description": "The Request of the unhealthy module repair is accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PatchModuleRepairResponse"
                                }
                            }
                        }
                    },
                    "409": {
                        "description": "Conflict",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 409
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "the repair process is already running"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "conflict"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch health checks: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/healths/{module}/history": {
            "get": {
                "operationId": "getHealthHistory",
                "tags": [
                    "Health"
                ],
                "summary": "Retrieve the health history of module",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "path",
                        "name": "module",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the module to retrieve health history, it could be 'ceph_osd', 'nova', and so on.",
                        "example": "'ceph_osd', 'nova', and so on."
                    },
                    {
                        "in": "query",
                        "name": "start",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The start time of the health history to query, the value should be in RFC3339 format (default is 24 hours ago).",
                        "example": "2025-01-01T01:00:00Z"
                    },
                    {
                        "in": "query",
                        "name": "stop",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The end time of the health history to query, the value should be in RFC3339 format (default is now).",
                        "example": "2025-01-01T01:00:00Z"
                    },
                    {
                        "in": "query",
                        "name": "pageNum",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The page number of the health history chunking to fetch (default is 1).",
                        "example": 1
                    },
                    {
                        "in": "query",
                        "name": "pageSize",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The size per page of the health history to return (default is unlimit).",
                        "example": 10
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the health history of module successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetModuleHealthHistoryResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Historical health check records",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "health": {
                                                    "category": "cloud computing",
                                                    "service": "compute",
                                                    "module": "nova",
                                                    "history": [
                                                        {
                                                            "time": "2025-02-01T03:00:00Z",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-01T02:55:00Z",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-01T02:50:00Z",
                                                            "status": "ng",
                                                            "error": {
                                                                "type": "service down",
                                                                "reason": "1 node down",
                                                                "nodes": [
                                                                    "example-node-0"
                                                                ],
                                                                "description": "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
                                                                "details": "{ ... best effort error summary / direction ...}",
                                                                "log": "http://datacenter1:8888/log/nova/dell180-20250205113459-b3gc.log"
                                                            }
                                                        },
                                                        {
                                                            "time": "2025-02-01T02:45:00Z",
                                                            "status": "ok"
                                                        }
                                                    ]
                                                },
                                                "page": {
                                                    "total": 10,
                                                    "number": 1,
                                                    "size": 2
                                                }
                                            },
                                            "msg": "retrieve health history of module successfully",
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to retrieve health history"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/integrations": {
            "get": {
                "operationId": "getIntegrations",
                "tags": [
                    "Integrations"
                ],
                "summary": "Retrieve the list of integrated applications",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of integrated applications successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetIntegrationsResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Integrated applications",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "name": "keycloak",
                                                    "isHeaderShortcutEnabled": true,
                                                    "description": "Keycloak Dashboard",
                                                    "isBuiltIn": true,
                                                    "url": "https://example-datat-center.host:10443/auth/admin"
                                                },
                                                {
                                                    "name": "openstack",
                                                    "isHeaderShortcutEnabled": true,
                                                    "description": "OpenStack Dashboard",
                                                    "isBuiltIn": true,
                                                    "url": "https://example-datat-center.host/horizon"
                                                },
                                                {
                                                    "name": "rancher",
                                                    "isHeaderShortcutEnabled": true,
                                                    "description": "Rancher Dashboard",
                                                    "isBuiltIn": true,
                                                    "url": "https://example-datat-center.host:10443"
                                                },
                                                {
                                                    "name": "ceph",
                                                    "isHeaderShortcutEnabled": true,
                                                    "description": "Ceph Dashboard",
                                                    "isBuiltIn": true,
                                                    "url": "https://example-datat-center.host:7443/ceph/#/dashboard"
                                                }
                                            ],
                                            "msg": "fetch integrations successfully",
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch integrations: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/licenses": {
            "get": {
                "operationId": "getLicenses",
                "tags": [
                    "Licenses"
                ],
                "summary": "Retrieve the list of licenses",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "pageNum",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The page number of the event chunking to fetch (default is 1).",
                        "example": 1
                    },
                    {
                        "in": "query",
                        "name": "pageSize",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The size per page of the events to return (default is unlimit).",
                        "example": 10
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of licenses successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetLicensesResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Licenses",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "licenses": [
                                                    {
                                                        "type": "trial",
                                                        "hostname": "dell200",
                                                        "serial": "  ",
                                                        "product": {
                                                            "name": "",
                                                            "features": null
                                                        },
                                                        "issue": {
                                                            "by": "Bigstack co., ltd.",
                                                            "to": "*",
                                                            "hardware": "*",
                                                            "date": "2025-02-04 09:54:45 UTC"
                                                        },
                                                        "quantity": {
                                                            "type": "",
                                                            "vcpu": 0
                                                        },
                                                        "serviceLevelAgreement": {
                                                            "uptime": 0,
                                                            "period": "",
                                                            "meanTimeBetweenFailure": "",
                                                            "meanTimeToRecovery": ""
                                                        },
                                                        "expiry": {
                                                            "date": "2025-04-05 09:54:45 UTC",
                                                            "days": 58
                                                        }
                                                    }
                                                ],
                                                "page": {
                                                    "total": 1,
                                                    "number": 1,
                                                    "size": 1
                                                }
                                            },
                                            "msg": "fetch licenses successfully",
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch licenses: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the metrics of data center",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "watch",
                        "required": false,
                        "schema": {
                            "type": "boolean"
                        },
                        "description": "The toggle to enable http chunked transfer for continues server push.",
                        "example": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the summary of data center successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "object",
                                            "properties": {
                                                "host": {
                                                    "type": "object",
                                                    "properties": {
                                                        "role": {
                                                            "type": "object",
                                                            "properties": {
                                                                "controlConverged": {
                                                                    "type": "integer",
                                                                    "example": 10
                                                                },
                                                                "control": {
                                                                    "type": "integer",
                                                                    "example": 3
                                                                },
                                                                "compute": {
                                                                    "type": "integer",
                                                                    "example": 5
                                                                },
                                                                "storage": {
                                                                    "type": "integer",
                                                                    "example": 2
                                                                },
                                                                "others": {
                                                                    "type": "integer",
                                                                    "example": 0
                                                                }
                                                            }
                                                        },
                                                        "usage": {
                                                            "type": "object",
                                                            "properties": {
                                                                "vcpu": {
                                                                    "type": "object",
                                                                    "properties": {
                                                                        "totalCores": {
                                                                            "type": "integer",
                                                                            "example": 80
                                                                        },
                                                                        "usedCores": {
                                                                            "type": "integer",
                                                                            "example": 31
                                                                        },
                                                                        "usedPercent": {
                                                                            "type": "number",
                                                                            "example": 38.75
                                                                        },
                                                                        "freeCores": {
                                                                            "type": "integer",
                                                                            "example": 49
                                                                        },
                                                                        "freePercent": {
                                                                            "type": "number",
                                                                            "example": 61.25
                                                                        }
                                                                    }
                                                                },
                                                                "memory": {
                                                                    "type": "object",
                                                                    "properties": {
                                                                        "totalMiB": {
                                                                            "type": "integer",
                                                                            "example": 257371
                                                                        },
                                                                        "usedMiB": {
                                                                            "type": "integer",
                                                                            "example": 98255
                                                                        },
                                                                        "usedPercent": {
                                                                            "type": "number",
                                                                            "example": 38.2
                                                                        },
                                                                        "freeMiB": {
                                                                            "type": "integer",
                                                                            "example": 159116
                                                                        },
                                                                        "freePercent": {
                                                                            "type": "number",
                                                                            "example": 61.8
                                                                        }
                                                                    }
                                                                },
                                                                "storage": {
                                                                    "type": "object",
                                                                    "properties": {
                                                                        "totalMiB": {
                                                                            "type": "integer",
                                                                            "example": 102400
                                                                        },
                                                                        "usedMiB": {
                                                                            "type": "integer",
                                                                            "example": 51200
                                                                        },
                                                                        "usedPercent": {
                                                                            "type": "number",
                                                                            "example": 50
                                                                        },
                                                                        "freeMiB": {
                                                                            "type": "integer",
                                                                            "example": 51200
                                                                        },
                                                                        "freePercent": {
                                                                            "type": "number",
                                                                            "example": 50
                                                                        }
                                                                    }
                                                                }
                                                            }
                                                        }
                                                    }
                                                },
                                                "vm": {
                                                    "type": "object",
                                                    "properties": {
                                                        "status": {
                                                            "type": "object",
                                                            "properties": {
                                                                "total": {
                                                                    "type": "integer",
                                                                    "example": 10
                                                                },
                                                                "running": {
                                                                    "type": "integer",
                                                                    "example": 8
                                                                },
                                                                "stopped": {
                                                                    "type": "integer",
                                                                    "example": 2
                                                                },
                                                                "paused": {
                                                                    "type": "integer",
                                                                    "example": 0
                                                                },
                                                                "suspend": {
                                                                    "type": "integer",
                                                                    "example": 0
                                                                },
                                                                "error": {
                                                                    "type": "integer",
                                                                    "example": 0
                                                                },
                                                                "unknown": {
                                                                    "type": "integer",
                                                                    "example": 0
                                                                }
                                                            }
                                                        },
                                                        "usage": {
                                                            "type": "object",
                                                            "properties": {
                                                                "vcpu": {
                                                                    "type": "object",
                                                                    "properties": {
                                                                        "totalCores": {
                                                                            "type": "integer",
                                                                            "example": 80
                                                                        },
                                                                        "usedCores": {
                                                                            "type": "integer",
                                                                            "example": 31
                                                                        },
                                                                        "freeCores": {
                                                                            "type": "integer",
                                                                            "example": 49
                                                                        },
                                                                        "usedPercent": {
                                                                            "type": "number",
                                                                            "example": 38.75
                                                                        },
                                                                        "freePercent": {
                                                                            "type": "number",
                                                                            "example": 61.25
                                                                        }
                                                                    }
                                                                },
                                                                "memory": {
                                                                    "type": "object",
                                                                    "properties": {
                                                                        "totalMiB": {
                                                                            "type": "integer",
                                                                            "example": 257371
                                                                        },
                                                                        "usedMiB": {
                                                                            "type": "integer",
                                                                            "example": 98255
                                                                        },
                                                                        "freeMiB": {
                                                                            "type": "integer",
                                                                            "example": 159116
                                                                        },
                                                                        "usedPercent": {
                                                                            "type": "number",
                                                                            "example": 38.2
                                                                        },
                                                                        "freePercent": {
                                                                            "type": "number",
                                                                            "example": 61.8
                                                                        }
                                                                    }
                                                                },
                                                                "storage": {
                                                                    "type": "object",
                                                                    "properties": {
                                                                        "totalMiB": {
                                                                            "type": "integer",
                                                                            "example": 102400
                                                                        },
                                                                        "usedMiB": {
                                                                            "type": "integer",
                                                                            "example": 51200
                                                                        },
                                                                        "freeMiB": {
                                                                            "type": "integer",
                                                                            "example": 51200
                                                                        },
                                                                        "usedPercent": {
                                                                            "type": "number",
                                                                            "example": 50
                                                                        },
                                                                        "freePercent": {
                                                                            "type": "number",
                                                                            "example": 50
                                                                        }
                                                                    }
                                                                }
                                                            }
                                                        }
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch summary successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch nodes: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/{metricGroup}/{resourceType}": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the metrics of data center, host, and vm",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "path",
                        "name": "metricGroup",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the metric group to query, the value can be only 'cpu', 'memory', 'storage', and 'network'",
                        "example": "cpu"
                    },
                    {
                        "in": "path",
                        "name": "resourceType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of resource to query, the value can be only 'hosts' and 'vms'",
                        "example": "hosts"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'usage', 'iops', 'bandwidth', 'latency', 'ingress', and 'egress'.",
                        "example": 1
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'summary', 'timeSeries', and 'rank'.",
                        "example": 1
                    },
                    {
                        "in": "query",
                        "name": "start",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The start time of the event to query, the value should be in RFC3339 format (default is 24 hours ago).",
                        "example": "2025-01-01T01:00:00Z"
                    },
                    {
                        "in": "query",
                        "name": "stop",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The end time of the event to query, the value should be in RFC3339 format (default is now).",
                        "example": "2025-01-01T01:00:00Z"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the summary of data center successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "object",
                                            "properties": {
                                                "vm": {
                                                    "type": "object",
                                                    "properties": {
                                                        "total": {
                                                            "type": "integer",
                                                            "example": 10
                                                        },
                                                        "running": {
                                                            "type": "integer",
                                                            "example": 8
                                                        },
                                                        "stopped": {
                                                            "type": "integer",
                                                            "example": 2
                                                        },
                                                        "paused": {
                                                            "type": "integer",
                                                            "example": 0
                                                        },
                                                        "suspend": {
                                                            "type": "integer",
                                                            "example": 0
                                                        },
                                                        "error": {
                                                            "type": "integer",
                                                            "example": 0
                                                        },
                                                        "unknown": {
                                                            "type": "integer",
                                                            "example": 0
                                                        }
                                                    }
                                                },
                                                "role": {
                                                    "type": "object",
                                                    "properties": {
                                                        "controlConverged": {
                                                            "type": "integer",
                                                            "example": 10
                                                        },
                                                        "control": {
                                                            "type": "integer",
                                                            "example": 3
                                                        },
                                                        "compute": {
                                                            "type": "integer",
                                                            "example": 5
                                                        },
                                                        "storage": {
                                                            "type": "integer",
                                                            "example": 2
                                                        },
                                                        "others": {
                                                            "type": "integer",
                                                            "example": 0
                                                        }
                                                    }
                                                },
                                                "metrics": {
                                                    "type": "object",
                                                    "properties": {
                                                        "vcpu": {
                                                            "type": "object",
                                                            "properties": {
                                                                "totalCores": {
                                                                    "type": "integer",
                                                                    "example": 80
                                                                },
                                                                "usedCores": {
                                                                    "type": "integer",
                                                                    "example": 31
                                                                },
                                                                "freeCores": {
                                                                    "type": "integer",
                                                                    "example": 49
                                                                }
                                                            }
                                                        },
                                                        "memory": {
                                                            "type": "object",
                                                            "properties": {
                                                                "totalMiB": {
                                                                    "type": "integer",
                                                                    "example": 257371
                                                                },
                                                                "usedMiB": {
                                                                    "type": "integer",
                                                                    "example": 98255
                                                                },
                                                                "freeMiB": {
                                                                    "type": "integer",
                                                                    "example": 159116
                                                                }
                                                            }
                                                        },
                                                        "storage": {
                                                            "type": "object",
                                                            "properties": {
                                                                "totalMiB": {
                                                                    "type": "integer",
                                                                    "example": 102400
                                                                },
                                                                "usedMiB": {
                                                                    "type": "integer",
                                                                    "example": 51200
                                                                },
                                                                "freeMiB": {
                                                                    "type": "integer",
                                                                    "example": 51200
                                                                }
                                                            }
                                                        }
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch summary successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch nodes: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/cpu/hosts": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the cpu metrics of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'usage'.",
                        "example": 1
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'timeSeries'.",
                        "example": 1
                    },
                    {
                        "in": "query",
                        "name": "start",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The start time of the event to query, the value should be in RFC3339 format (default is 24 hours ago).",
                        "example": "2025-01-01T01:00:00Z"
                    },
                    {
                        "in": "query",
                        "name": "stop",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The end time of the event to query, the value should be in RFC3339 format (default is now).",
                        "example": "2025-01-01T01:00:00Z"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the cpu usage of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "object",
                                            "properties": {
                                                "totalCores": {
                                                    "type": "integer",
                                                    "example": 80
                                                },
                                                "usedCores": {
                                                    "type": "integer",
                                                    "example": 31
                                                },
                                                "freeCores": {
                                                    "type": "integer",
                                                    "example": 49
                                                },
                                                "usedPercent": {
                                                    "type": "number",
                                                    "example": 38.75
                                                },
                                                "freePercent": {
                                                    "type": "number",
                                                    "example": 61.25
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch summary successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/storage/hosts?metricType=bandwidth": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the storage bandwidth time series of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'bandwidth'.",
                        "example": "bandwidth"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'timeSeries'.",
                        "example": "timeSeries"
                    },
                    {
                        "in": "query",
                        "name": "start",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The start time of the metric to query, the value should be in RFC3339 format (default is 24 hours ago).",
                        "example": "2025-01-01T01:00:00Z"
                    },
                    {
                        "in": "query",
                        "name": "stop",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The end time of the metric to query, the value should be in RFC3339 format (default is now).",
                        "example": "2025-01-01T01:00:00Z"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the storage bandwidth time series of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "oneOf": [
                                        {
                                            "type": "object",
                                            "properties": {
                                                "code": {
                                                    "type": "integer",
                                                    "example": 200
                                                },
                                                "data": {
                                                    "type": "object",
                                                    "properties": {
                                                        "read": {
                                                            "type": "array",
                                                            "items": {
                                                                "type": "object",
                                                                "properties": {
                                                                    "time": {
                                                                        "type": "string",
                                                                        "example": "2025-01-01T01:00:00Z"
                                                                    },
                                                                    "bytes": {
                                                                        "type": "number",
                                                                        "example": 682.6666
                                                                    }
                                                                }
                                                            }
                                                        },
                                                        "write": {
                                                            "type": "array",
                                                            "items": {
                                                                "type": "object",
                                                                "properties": {
                                                                    "time": {
                                                                        "type": "string",
                                                                        "example": "2025-01-01T01:00:00Z"
                                                                    },
                                                                    "bytes": {
                                                                        "type": "number",
                                                                        "example": 4.5
                                                                    }
                                                                }
                                                            }
                                                        }
                                                    }
                                                },
                                                "msg": {
                                                    "type": "string",
                                                    "example": "fetch metrics successfully"
                                                },
                                                "status": {
                                                    "type": "string",
                                                    "example": "ok"
                                                }
                                            }
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "status": {
                                                    "type": "string",
                                                    "example": "ok"
                                                }
                                            }
                                        }
                                    ]
                                },
                                "examples": {
                                    "example1": {
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "read": [
                                                    {
                                                        "time": "2025-01-01T01:00:00Z",
                                                        "bytes": 682.6666
                                                    }
                                                ],
                                                "write": [
                                                    {
                                                        "time": "2025-01-01T01:00:00Z",
                                                        "bytes": 4.5
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example2": {
                                        "value": {
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/storage/hosts?metricType=iops": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the storage iops time series of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'iops'.",
                        "example": "iops"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'timeSeries'.",
                        "example": "timeSeries"
                    },
                    {
                        "in": "query",
                        "name": "start",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The start time of the metric to query, the value should be in RFC3339 format (default is 24 hours ago).",
                        "example": "2025-01-01T01:00:00Z"
                    },
                    {
                        "in": "query",
                        "name": "stop",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The end time of the metric to query, the value should be in RFC3339 format (default is now).",
                        "example": "2025-01-01T01:00:00Z"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the storage iops time series of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "object",
                                            "properties": {
                                                "read": {
                                                    "type": "array",
                                                    "items": {
                                                        "type": "object",
                                                        "properties": {
                                                            "time": {
                                                                "type": "string",
                                                                "example": "2025-01-01T01:00:00Z"
                                                            },
                                                            "ops": {
                                                                "type": "number",
                                                                "example": 1.0833
                                                            }
                                                        }
                                                    }
                                                },
                                                "write": {
                                                    "type": "array",
                                                    "items": {
                                                        "type": "object",
                                                        "properties": {
                                                            "time": {
                                                                "type": "string",
                                                                "example": "2025-01-01T01:00:00Z"
                                                            },
                                                            "ops": {
                                                                "type": "number",
                                                                "example": 5.8166
                                                            }
                                                        }
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/storage/hosts?metricType=latency": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the storage latency time series of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'latency'.",
                        "example": "latency"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'timeSeries'.",
                        "example": "timeSeries"
                    },
                    {
                        "in": "query",
                        "name": "start",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The start time of the metric to query, the value should be in RFC3339 format (default is 24 hours ago).",
                        "example": "2025-01-01T01:00:00Z"
                    },
                    {
                        "in": "query",
                        "name": "stop",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The end time of the metric to query, the value should be in RFC3339 format (default is now).",
                        "example": "2025-01-01T01:00:00Z"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the storage latency time series of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "object",
                                            "properties": {
                                                "read": {
                                                    "type": "array",
                                                    "items": {
                                                        "type": "object",
                                                        "properties": {
                                                            "time": {
                                                                "type": "string",
                                                                "example": "2025-01-01T01:00:00Z"
                                                            },
                                                            "ms": {
                                                                "type": "number",
                                                                "example": 10736.9833
                                                            }
                                                        }
                                                    }
                                                },
                                                "write": {
                                                    "type": "array",
                                                    "items": {
                                                        "type": "object",
                                                        "properties": {
                                                            "time": {
                                                                "type": "string",
                                                                "example": "2025-01-01T01:00:00Z"
                                                            },
                                                            "ms": {
                                                                "type": "number",
                                                                "example": 1615349.9
                                                            }
                                                        }
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/storage/hosts?metricType=usage": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the storage usage time series of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'usage'.",
                        "example": "latency"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'rank'.",
                        "example": "rank"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the storage usage rank of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "properties": {
                                                    "id": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "name": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 24.6699
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 75.3301
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/cpu/hosts?metricType=usage&reportType=rank": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the cpu usage rank of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'usage'.",
                        "example": "latency"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'rank'.",
                        "example": "rank"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the cpu usage rank of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "properties": {
                                                    "id": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "name": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 10.5263
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 89.4737
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/memory/hosts?metricType=usage&reportType=rank": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the memory usage rank of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'usage'.",
                        "example": "latency"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'rank'.",
                        "example": "rank"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the memory usage rank of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "properties": {
                                                    "id": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "name": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 52.6702
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 47.3298
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/network/hosts?metricType=ingress&reportType=rank": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the network ingress rank of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'ingress'.",
                        "example": "ingress"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'rank'.",
                        "example": "rank"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the network ingress rank of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "properties": {
                                                    "id": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "name": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "packets": {
                                                        "type": "number",
                                                        "example": 5392.2666
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/network/hosts?metricType=egress&reportType=rank": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the network egress rank of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'ingress'.",
                        "example": "ingress"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'rank'.",
                        "example": "rank"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the network ingress rank of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "properties": {
                                                    "id": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "name": {
                                                        "type": "string",
                                                        "example": "example-data-center"
                                                    },
                                                    "packets": {
                                                        "type": "number",
                                                        "example": 1724.5302
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/cpu/vms?metricType=usage&reportType=rank": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the cpu usage rank of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'usage'.",
                        "example": "ingress"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'rank'.",
                        "example": "rank"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the cpu usage rank of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "properties": {
                                                    "id": {
                                                        "type": "string",
                                                        "example": "d280d915-7877-48c7-a0b2-faa15ec94717"
                                                    },
                                                    "name": {
                                                        "type": "string",
                                                        "example": "example-vm"
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 10.5263
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 89.4737
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/memory/vms?metricType=usage&reportType=rank": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the memory usage rank of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'usage'.",
                        "example": "ingress"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'rank'.",
                        "example": "rank"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the memory usage rank of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "properties": {
                                                    "id": {
                                                        "type": "string",
                                                        "example": "d280d915-7877-48c7-a0b2-faa15ec94717"
                                                    },
                                                    "name": {
                                                        "type": "string",
                                                        "example": "example-vm"
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 35.0991
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 64.9009
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/storage/vms?metricType={metricType}&reportType=rank": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the storage iops read or write rank of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'iopsRead' or 'iopsWrite'.",
                        "example": "ingress"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'rank'.",
                        "example": "rank"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the storage iops rank of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "properties": {
                                                    "id": {
                                                        "type": "string",
                                                        "example": "d280d915-7877-48c7-a0b2-faa15ec94717"
                                                    },
                                                    "name": {
                                                        "type": "string",
                                                        "example": "example-vm"
                                                    },
                                                    "device": {
                                                        "type": "string",
                                                        "example": "sda"
                                                    },
                                                    "usage": {
                                                        "type": "number",
                                                        "example": 184266.4594
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/metrics/network/vms?metricType={metricType}&reportType=rank": {
            "get": {
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the network ingress or egress rank of hosts",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "metricType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of metric to query, the value can be only 'ingress' or 'egress'.",
                        "example": "ingress"
                    },
                    {
                        "in": "query",
                        "name": "reportType",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The type of report to query, the value can be only 'rank'.",
                        "example": "rank"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the network ingress or egress rank of hosts successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 200
                                        },
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "properties": {
                                                    "id": {
                                                        "type": "string",
                                                        "example": "d280d915-7877-48c7-a0b2-faa15ec94717"
                                                    },
                                                    "name": {
                                                        "type": "string",
                                                        "example": "example-vm"
                                                    },
                                                    "device": {
                                                        "type": "string",
                                                        "example": "sda"
                                                    },
                                                    "usage": {
                                                        "type": "number",
                                                        "example": 45443.764
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "fetch metrics successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch metrics: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/nodes": {
            "get": {
                "operationId": "getNodes",
                "tags": [
                    "Nodes"
                ],
                "summary": "Retrieve the list of nodes",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "pageNum",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The page number of the event chunking to fetch (default is 1).",
                        "example": 1
                    },
                    {
                        "in": "query",
                        "name": "pageSize",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The size per page of the events to return (default is unlimit).",
                        "example": 10
                    },
                    {
                        "in": "query",
                        "name": "watch",
                        "required": false,
                        "schema": {
                            "type": "boolean"
                        },
                        "description": "The toggle to enable http chunked transfer for continues server push.",
                        "example": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of nodes successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetNodesResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Nodes",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "nodes": [
                                                    {
                                                        "id": "abc0005e",
                                                        "hostname": "example-node",
                                                        "role": "control-converged",
                                                        "address": "10.10.10.10:8001",
                                                        "managementIP": "10.10.10.10",
                                                        "license": {
                                                            "type": "trial",
                                                            "hostname": "example-node",
                                                            "serial": "1H2ZLG2",
                                                            "product": {
                                                                "name": "",
                                                                "features": null
                                                            },
                                                            "issue": {
                                                                "by": "Bigstack co., ltd.",
                                                                "to": "bigstack",
                                                                "hardware": "*",
                                                                "date": "2025-01-23 06:51:50 UTC"
                                                            },
                                                            "quantity": {
                                                                "type": "",
                                                                "vcpu": 0
                                                            },
                                                            "serviceLevelAgreement": {
                                                                "uptime": 0,
                                                                "period": "",
                                                                "meanTimeBetweenFailure": "",
                                                                "meanTimeToRecovery": ""
                                                            },
                                                            "expiry": {
                                                                "date": "2025-03-24 06:51:50 UTC",
                                                                "days": 44
                                                            }
                                                        },
                                                        "status": "up",
                                                        "vcpu": {
                                                            "totalCores": 48,
                                                            "usedCores": 38,
                                                            "usedPercent": 79.1666,
                                                            "freeCores": 10,
                                                            "freePercent": 20.8333
                                                        },
                                                        "memory": {
                                                            "totalMiB": 257822,
                                                            "usedMiB": 95400,
                                                            "usedPercent": 37.0022,
                                                            "freeMiB": 162422,
                                                            "freePercent": 62.9977
                                                        },
                                                        "storage": {
                                                            "totalMiB": 12571648,
                                                            "usedMiB": 716800,
                                                            "usedPercent": 5.7017,
                                                            "freeMiB": 11854848,
                                                            "freePercent": 94.2982
                                                        },
                                                        "uptimeSeconds": 5734686.37,
                                                        "labels": {
                                                            "isGpuEnabled": "false"
                                                        }
                                                    }
                                                ],
                                                "page": {
                                                    "total": 1,
                                                    "number": 1,
                                                    "size": 1
                                                }
                                            },
                                            "msg": "fetch nodes list successfully",
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to fetch nodes: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/datacenters/{dataCenter}/tokens": {
            "post": {
                "operationId": "getToken",
                "tags": [
                    "Tokens"
                ],
                "summary": "Retrieve the time-limited token for the data center",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "example-data-center"
                    }
                ],
                "requestBody": {
                    "description": "The user name and password to generate the token",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/GetTokensRequest"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Retrieve the time-limited token for the data center successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetTokensResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Token",
                                        "value": {
                                            "code": 201,
                                            "data": {
                                                "token": "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIwdDdGdWlJZC1lbnhVUWRZWGVZalZ6Q0pQZFRWMmxaU0NZanRkQW01S3djIn0.eyJleHAiOjE3Mzg5NjA2NzYsImlhdCI6MTczODk1MzQ3NiwianRpIjoiZjZjZTJlMjMtYzgwNC00N2YxLWE0OWEtZjdlNWZlYTgyYThjIiwiaXNzIjoiaHR0cHM6Ly8xMC4zMi4xMC4xODA6MTA0NDMvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjpbIm1hc3Rlci1yZWFsbSIsImFjY291bnQiXSwic3ViIjoiMDY1OTVlMDAtZWIzOS00YWUxLTkzYjctZDMyNzg1MDdiZTUzIiwidHlwIjoiQmVhcmVyIiwiYXpwIjoidG9rZW4tY29ubmVjdCIsInNlc3Npb25fc3RhdGUiOiJlOTE0MWVhZC01MTQ1LTRjZmQtODhiZC00ZjYxYzI0MzA0NzAiLCJhY3IiOiIxIiwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbImNyZWF0ZS1yZWFsbSIsImRlZmF1bHQtcm9sZXMtbWFzdGVyIiwib2ZmbGluZV9hY2Nlc3MiLCJhZG1pbiIsInVtYV9hdXRob3JpemF0aW9uIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctaWRlbnRpdHktcHJvdmlkZXJzIiwidmlldy1yZWFsbSIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6Im9wZW5pZCBlbWFpbCBwcm9maWxlIiwic2lkIjoiZTkxNDFlYWQtNTE0NS00Y2ZkLTg4YmQtNGY2MWMyNDMwNDcwIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJhZG1pbiJ9.LkywGnG2BhQ90n-f3vJvX-frq1wRS6DAop5MUcCFb5FSuGvv3oP_byeiuHvh9rh06QWcwhpiTU_L_hOaUzMXF9spJjos8s2iOwK-4o46Ka_4Q2gnlitjd8ZG5wGfuqKRp0vb25RsBpNr6rxGFzOxxmULkjEXOGXbcrqQJBWIaWdRldYunLFRVErIpmH23w4KHKWSB5XkwqpkD9HnLH_PqbaunGpp2acgZ826JpevC8vhEhIg6fxXcDnCKhkw7nNvfs3rCC08Qci32WtzujeKH6js1ASRkKWOUxVkfIkzZ5Yg4Tfdb79_Tgy2iXLT3teeNTepmLsaftn2JzLn_6vHwg",
                                                "refresh": "eyJhbGciOiJIUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICI1OGRkY2JhNC1hM2ZkLTQ2MTYtODgyYi1lMGY1ZjRlNzAyMTIifQ.eyJleHAiOjE3Mzg5NTUyNzYsImlhdCI6MTczODk1MzQ3NiwianRpIjoiNjkyZDkxMTAtMjExMy00NTU1LTgxMTctNWYxZTA2NzlmYmUxIiwiaXNzIjoiaHR0cHM6Ly8xMC4zMi4xMC4xODA6MTA0NDMvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiaHR0cHM6Ly8xMC4zMi4xMC4xODA6MTA0NDMvYXV0aC9yZWFsbXMvbWFzdGVyIiwic3ViIjoiMDY1OTVlMDAtZWIzOS00YWUxLTkzYjctZDMyNzg1MDdiZTUzIiwidHlwIjoiUmVmcmVzaCIsImF6cCI6InRva2VuLWNvbm5lY3QiLCJzZXNzaW9uX3N0YXRlIjoiZTkxNDFlYWQtNTE0NS00Y2ZkLTg4YmQtNGY2MWMyNDMwNDcwIiwic2NvcGUiOiJvcGVuaWQgZW1haWwgcHJvZmlsZSIsInNpZCI6ImU5MTQxZWFkLTUxNDUtNGNmZC04OGJkLTRmNjFjMjQzMDQ3MCJ9._B9rU8TG_YSvwsri48bfne3RxFhc4Yq83bzR43nL4-c",
                                                "expires": {
                                                    "access": 7200,
                                                    "refresh": 1800
                                                }
                                            },
                                            "msg": "create token successfully",
                                            "status": "created"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 401
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "invalid_grant: Invalid user credentials"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "unauthorized"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 500
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "failed to create token: internal server error"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "internal server error"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    },
    "components": {
        "schemas": {
            "GetMeResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer"
                    },
                    "data": {
                        "type": "object",
                        "required": [
                            "name"
                        ],
                        "properties": {
                            "name": {
                                "type": "string"
                            }
                        }
                    },
                    "msg": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            },
            "GetDataCentersResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "data": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "required": [
                                "name",
                                "version",
                                "virtualIp",
                                "isHaEnabled",
                                "isLocal"
                            ],
                            "properties": {
                                "name": {
                                    "type": "string",
                                    "example": "bigstack-data-center"
                                },
                                "version": {
                                    "type": "string",
                                    "example": "Cube Appliance 3.0.0"
                                },
                                "virtualIp": {
                                    "type": "string",
                                    "example": "10.10.10.10"
                                },
                                "isHaEnabled": {
                                    "type": "boolean",
                                    "example": false
                                },
                                "isLocal": {
                                    "type": "boolean",
                                    "example": true
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "fetch data center list successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetEventsResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "data": {
                        "type": "object",
                        "required": [
                            "events",
                            "page"
                        ],
                        "properties": {
                            "events": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "type",
                                        "severity",
                                        "id",
                                        "description",
                                        "host",
                                        "category",
                                        "service",
                                        "metadata",
                                        "time"
                                    ],
                                    "properties": {
                                        "type": {
                                            "type": "string",
                                            "example": "system"
                                        },
                                        "severity": {
                                            "type": "string",
                                            "example": "Info"
                                        },
                                        "id": {
                                            "type": "string",
                                            "example": "NET00003I"
                                        },
                                        "description": {
                                            "type": "string",
                                            "example": "instance 0125741a-7dbe-4309-bc1a-53d2880d2925 at 192.168.0.91 is reachable"
                                        },
                                        "host": {
                                            "type": "string",
                                            "example": "bigstack-host"
                                        },
                                        "category": {
                                            "type": "string",
                                            "example": "net"
                                        },
                                        "service": {
                                            "type": "string",
                                            "example": "neutron"
                                        },
                                        "metadata": {
                                            "type": "object",
                                            "required": [
                                                "id",
                                                "ip"
                                            ],
                                            "properties": {
                                                "id": {
                                                    "type": "string",
                                                    "example": "0125741a-7dbe-4309-bc1a-53d2880d2925"
                                                },
                                                "ip": {
                                                    "type": "string",
                                                    "example": "192.168.0.91"
                                                }
                                            }
                                        },
                                        "time": {
                                            "type": "string",
                                            "example": "2025-01-01T01:00:00Z"
                                        }
                                    }
                                }
                            },
                            "page": {
                                "type": "object",
                                "required": [
                                    "total",
                                    "number",
                                    "size"
                                ],
                                "properties": {
                                    "total": {
                                        "type": "integer",
                                        "example": 10
                                    },
                                    "number": {
                                        "type": "integer",
                                        "example": 1
                                    },
                                    "size": {
                                        "type": "integer",
                                        "example": 1
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "fetch events successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetAbstractedEventsResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer"
                    },
                    "data": {
                        "type": "object",
                        "properties": {
                            "events": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "type",
                                        "severity",
                                        "id",
                                        "description",
                                        "host",
                                        "category",
                                        "service",
                                        "metadata",
                                        "time"
                                    ],
                                    "properties": {
                                        "type": {
                                            "type": "string"
                                        },
                                        "severity": {
                                            "type": "string"
                                        },
                                        "id": {
                                            "type": "string"
                                        },
                                        "description": {
                                            "type": "string"
                                        },
                                        "host": {
                                            "type": "string"
                                        },
                                        "category": {
                                            "type": "string"
                                        },
                                        "service": {
                                            "type": "string"
                                        },
                                        "metadata": {
                                            "type": "object",
                                            "properties": {
                                                "id": {
                                                    "type": "string"
                                                },
                                                "ip": {
                                                    "type": "string"
                                                }
                                            }
                                        },
                                        "time": {
                                            "type": "string"
                                        }
                                    }
                                }
                            },
                            "limit": {
                                "type": "object",
                                "required": [
                                    "number",
                                    "description"
                                ],
                                "properties": {
                                    "number": {
                                        "type": "integer"
                                    },
                                    "description": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            },
            "GetHealthsResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "data": {
                        "type": "object",
                        "required": [
                            "overall",
                            "inUse",
                            "error",
                            "fixing"
                        ],
                        "properties": {
                            "overall": {
                                "type": "object",
                                "required": [
                                    "status"
                                ],
                                "properties": {
                                    "status": {
                                        "type": "object",
                                        "required": [
                                            "current",
                                            "description"
                                        ],
                                        "properties": {
                                            "current": {
                                                "type": "string",
                                                "example": "ng"
                                            },
                                            "description": {
                                                "type": "string",
                                                "example": "ceph has 2 ceph_osd down"
                                            }
                                        }
                                    }
                                }
                            },
                            "services": {
                                "type": "array",
                                "items": {
                                    "required": [
                                        "service",
                                        "category",
                                        "status",
                                        "modules"
                                    ],
                                    "properties": {
                                        "name": {
                                            "type": "string",
                                            "example": "clusterLink"
                                        },
                                        "category": {
                                            "type": "string",
                                            "example": "core"
                                        },
                                        "status": {
                                            "type": "object",
                                            "required": [
                                                "current"
                                            ],
                                            "properties": {
                                                "current": {
                                                    "type": "string",
                                                    "example": "ok"
                                                }
                                            }
                                        },
                                        "modules": {
                                            "type": "array",
                                            "required": [
                                                "name",
                                                "isAutoRepairable",
                                                "status"
                                            ],
                                            "items": {
                                                "type": "object",
                                                "required": [
                                                    "name",
                                                    "isAutoRepairable",
                                                    "status"
                                                ],
                                                "properties": {
                                                    "name": {
                                                        "type": "string",
                                                        "example": "link"
                                                    },
                                                    "isAutoRepairable": {
                                                        "type": "boolean",
                                                        "example": false
                                                    },
                                                    "status": {
                                                        "type": "object",
                                                        "required": [
                                                            "current"
                                                        ],
                                                        "properties": {
                                                            "current": {
                                                                "type": "string",
                                                                "example": "ok"
                                                            }
                                                        }
                                                    }
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "fetch health successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "PatchRepairResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer"
                    },
                    "msg": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            },
            "PatchModuleRepairResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 202
                    },
                    "msg": {
                        "type": "string",
                        "example": "the request of unhealthy module repair is accepted and repairing"
                    },
                    "status": {
                        "type": "string",
                        "example": "accepted"
                    }
                }
            },
            "GetModuleHealthHistoryResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer"
                    },
                    "data": {
                        "type": "object",
                        "required": [
                            "health",
                            "page"
                        ],
                        "properties": {
                            "health": {
                                "type": "object",
                                "required": [
                                    "category",
                                    "service",
                                    "module",
                                    "history"
                                ],
                                "properties": {
                                    "category": {
                                        "type": "string"
                                    },
                                    "service": {
                                        "type": "string"
                                    },
                                    "module": {
                                        "type": "string"
                                    },
                                    "history": {
                                        "type": "array",
                                        "items": {
                                            "type": "object",
                                            "required": [
                                                "time",
                                                "status"
                                            ],
                                            "properties": {
                                                "time": {
                                                    "type": "string",
                                                    "format": "date-time"
                                                },
                                                "status": {
                                                    "type": "string"
                                                },
                                                "error": {
                                                    "type": "object",
                                                    "properties": {
                                                        "type": {
                                                            "type": "string"
                                                        },
                                                        "nodes": {
                                                            "type": "array",
                                                            "items": {
                                                                "type": "string"
                                                            }
                                                        },
                                                        "description": {
                                                            "type": "string"
                                                        },
                                                        "details": {
                                                            "type": "string"
                                                        },
                                                        "log": {
                                                            "type": "string",
                                                            "format": "uri"
                                                        }
                                                    }
                                                }
                                            }
                                        }
                                    }
                                }
                            },
                            "page": {
                                "type": "object",
                                "properties": {
                                    "total": {
                                        "type": "integer"
                                    },
                                    "number": {
                                        "type": "integer"
                                    },
                                    "size": {
                                        "type": "integer"
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            },
            "GetIntegrationsResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "data": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "required": [
                                "name",
                                "isHeaderShortcutEnabled",
                                "description",
                                "isBuiltIn",
                                "url"
                            ],
                            "properties": {
                                "name": {
                                    "type": "string",
                                    "example": "openstack"
                                },
                                "isHeaderShortcutEnabled": {
                                    "type": "boolean",
                                    "example": true
                                },
                                "description": {
                                    "type": "string",
                                    "example": "openstack dashboard"
                                },
                                "isBuiltIn": {
                                    "type": "boolean",
                                    "example": true
                                },
                                "url": {
                                    "type": "string",
                                    "example": "https://10.10.10.10/skyline"
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "fetch integrations successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetLicensesResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "data": {
                        "type": "object",
                        "properties": {
                            "licenses": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "type",
                                        "hostname",
                                        "serial",
                                        "product",
                                        "issue",
                                        "quantity",
                                        "serviceLevelAgreement",
                                        "expiry"
                                    ],
                                    "properties": {
                                        "type": {
                                            "type": "string"
                                        },
                                        "hostname": {
                                            "type": "string"
                                        },
                                        "serial": {
                                            "type": "string"
                                        },
                                        "product": {
                                            "type": "object",
                                            "required": [
                                                "name",
                                                "features"
                                            ],
                                            "properties": {
                                                "name": {
                                                    "type": "string"
                                                },
                                                "features": {
                                                    "type": "object"
                                                }
                                            }
                                        },
                                        "issue": {
                                            "type": "object",
                                            "required": [
                                                "by",
                                                "to",
                                                "hardware",
                                                "date"
                                            ],
                                            "properties": {
                                                "by": {
                                                    "type": "string"
                                                },
                                                "to": {
                                                    "type": "string"
                                                },
                                                "hardware": {
                                                    "type": "string"
                                                },
                                                "date": {
                                                    "type": "string"
                                                }
                                            }
                                        },
                                        "quantity": {
                                            "type": "object",
                                            "required": [
                                                "type",
                                                "vcpu"
                                            ],
                                            "properties": {
                                                "type": {
                                                    "type": "string"
                                                },
                                                "vcpu": {
                                                    "type": "integer"
                                                }
                                            }
                                        },
                                        "serviceLevelAgreement": {
                                            "type": "object",
                                            "required": [
                                                "uptime",
                                                "period",
                                                "meanTimeBetweenFailure",
                                                "meanTimeToRecovery"
                                            ],
                                            "properties": {
                                                "uptime": {
                                                    "type": "number"
                                                },
                                                "period": {
                                                    "type": "string"
                                                },
                                                "meanTimeBetweenFailure": {
                                                    "type": "string"
                                                },
                                                "meanTimeToRecovery": {
                                                    "type": "string"
                                                }
                                            }
                                        },
                                        "expiry": {
                                            "type": "object",
                                            "required": [
                                                "date",
                                                "days"
                                            ],
                                            "properties": {
                                                "date": {
                                                    "type": "string"
                                                },
                                                "days": {
                                                    "type": "integer"
                                                }
                                            }
                                        }
                                    }
                                }
                            },
                            "page": {
                                "type": "object",
                                "required": [
                                    "total",
                                    "number",
                                    "size"
                                ],
                                "properties": {
                                    "total": {
                                        "type": "integer"
                                    },
                                    "number": {
                                        "type": "integer"
                                    },
                                    "size": {
                                        "type": "integer"
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            },
            "GetMetricsResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "data": {
                        "type": "object",
                        "required": [
                            "host",
                            "vm"
                        ],
                        "properties": {
                            "host": {
                                "type": "object",
                                "required": [
                                    "role",
                                    "status",
                                    "usage"
                                ],
                                "properties": {
                                    "role": {
                                        "type": "object",
                                        "required": [
                                            "controlConverged",
                                            "control",
                                            "compute",
                                            "storage",
                                            "others"
                                        ],
                                        "properties": {
                                            "controlConverged": {
                                                "type": "integer",
                                                "example": 10
                                            },
                                            "control": {
                                                "type": "integer",
                                                "example": 3
                                            },
                                            "compute": {
                                                "type": "integer",
                                                "example": 5
                                            },
                                            "storage": {
                                                "type": "integer",
                                                "example": 2
                                            },
                                            "others": {
                                                "type": "integer",
                                                "example": 0
                                            }
                                        }
                                    },
                                    "usage": {
                                        "type": "object",
                                        "required": [
                                            "vcpu",
                                            "memory",
                                            "storage"
                                        ],
                                        "properties": {
                                            "vcpu": {
                                                "type": "object",
                                                "required": [
                                                    "totalCores",
                                                    "usedCores",
                                                    "usedPercent",
                                                    "freeCores",
                                                    "freePercent"
                                                ],
                                                "properties": {
                                                    "totalCores": {
                                                        "type": "integer",
                                                        "example": 80
                                                    },
                                                    "usedCores": {
                                                        "type": "integer",
                                                        "example": 31
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 38.75
                                                    },
                                                    "freeCores": {
                                                        "type": "integer",
                                                        "example": 49
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 61.25
                                                    }
                                                }
                                            },
                                            "memory": {
                                                "type": "object",
                                                "required": [
                                                    "totalMiB",
                                                    "usedMiB",
                                                    "usedPercent",
                                                    "freeMiB",
                                                    "freePercent"
                                                ],
                                                "properties": {
                                                    "totalMiB": {
                                                        "type": "integer",
                                                        "example": 257371
                                                    },
                                                    "usedMiB": {
                                                        "type": "integer",
                                                        "example": 98255
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 38.2
                                                    },
                                                    "freeMiB": {
                                                        "type": "integer",
                                                        "example": 159116
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 61.8
                                                    }
                                                }
                                            },
                                            "storage": {
                                                "type": "object",
                                                "required": [
                                                    "totalMiB",
                                                    "usedMiB",
                                                    "usedPercent",
                                                    "freeMiB",
                                                    "freePercent"
                                                ],
                                                "properties": {
                                                    "totalMiB": {
                                                        "type": "integer",
                                                        "example": 102400
                                                    },
                                                    "usedMiB": {
                                                        "type": "integer",
                                                        "example": 51200
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 50.12
                                                    },
                                                    "freeMiB": {
                                                        "type": "integer",
                                                        "example": 51200
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 50.12
                                                    }
                                                }
                                            }
                                        }
                                    }
                                }
                            },
                            "vm": {
                                "type": "object",
                                "required": [
                                    "status",
                                    "usage"
                                ],
                                "properties": {
                                    "status": {
                                        "type": "object",
                                        "required": [
                                            "total",
                                            "running",
                                            "stopped",
                                            "paused",
                                            "suspend",
                                            "error",
                                            "unknown"
                                        ],
                                        "properties": {
                                            "total": {
                                                "type": "integer",
                                                "example": 10
                                            },
                                            "running": {
                                                "type": "integer",
                                                "example": 8
                                            },
                                            "stopped": {
                                                "type": "integer",
                                                "example": 2
                                            },
                                            "paused": {
                                                "type": "integer",
                                                "example": 0
                                            },
                                            "suspend": {
                                                "type": "integer",
                                                "example": 0
                                            },
                                            "error": {
                                                "type": "integer",
                                                "example": 0
                                            },
                                            "unknown": {
                                                "type": "integer",
                                                "example": 0
                                            }
                                        }
                                    },
                                    "usage": {
                                        "type": "object",
                                        "required": [
                                            "vcpu",
                                            "memory",
                                            "storage"
                                        ],
                                        "properties": {
                                            "vcpu": {
                                                "type": "object",
                                                "required": [
                                                    "totalCores",
                                                    "usedCores",
                                                    "freeCores",
                                                    "usedPercent",
                                                    "freePercent"
                                                ],
                                                "properties": {
                                                    "totalCores": {
                                                        "type": "integer",
                                                        "example": 80
                                                    },
                                                    "usedCores": {
                                                        "type": "integer",
                                                        "example": 31
                                                    },
                                                    "freeCores": {
                                                        "type": "integer",
                                                        "example": 49
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 38.75
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 61.25
                                                    }
                                                }
                                            },
                                            "memory": {
                                                "type": "object",
                                                "required": [
                                                    "totalMiB",
                                                    "usedMiB",
                                                    "freeMiB",
                                                    "usedPercent",
                                                    "freePercent"
                                                ],
                                                "properties": {
                                                    "totalMiB": {
                                                        "type": "integer",
                                                        "example": 257371
                                                    },
                                                    "usedMiB": {
                                                        "type": "integer",
                                                        "example": 98255
                                                    },
                                                    "freeMiB": {
                                                        "type": "integer",
                                                        "example": 159116
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 38.2
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 61.8
                                                    }
                                                }
                                            },
                                            "storage": {
                                                "type": "object",
                                                "required": [
                                                    "totalMiB",
                                                    "usedMiB",
                                                    "freeMiB",
                                                    "usedPercent",
                                                    "freePercent"
                                                ],
                                                "properties": {
                                                    "totalMiB": {
                                                        "type": "integer",
                                                        "example": 102400
                                                    },
                                                    "usedMiB": {
                                                        "type": "integer",
                                                        "example": 51200
                                                    },
                                                    "freeMiB": {
                                                        "type": "integer",
                                                        "example": 51200
                                                    },
                                                    "usedPercent": {
                                                        "type": "number",
                                                        "example": 50.1
                                                    },
                                                    "freePercent": {
                                                        "type": "number",
                                                        "example": 50.1
                                                    }
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "fetch summary successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetNodesResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "data": {
                        "type": "object",
                        "required": [
                            "nodes",
                            "page"
                        ],
                        "properties": {
                            "nodes": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "id",
                                        "hostname",
                                        "role",
                                        "address",
                                        "managementIP",
                                        "license",
                                        "status",
                                        "vcpu",
                                        "memory",
                                        "storage",
                                        "uptimeSeconds",
                                        "labels"
                                    ],
                                    "properties": {
                                        "id": {
                                            "type": "string"
                                        },
                                        "hostname": {
                                            "type": "string"
                                        },
                                        "role": {
                                            "type": "string"
                                        },
                                        "address": {
                                            "type": "string"
                                        },
                                        "managementIP": {
                                            "type": "string"
                                        },
                                        "license": {
                                            "type": "object",
                                            "required": [
                                                "status",
                                                "hostname",
                                                "serial",
                                                "product",
                                                "issue",
                                                "serviceLevelAgreement",
                                                "expire"
                                            ],
                                            "properties": {
                                                "status": {
                                                    "type": "string"
                                                },
                                                "hostname": {
                                                    "type": "string"
                                                },
                                                "serial": {
                                                    "type": "string"
                                                },
                                                "product": {
                                                    "type": "object",
                                                    "required": [
                                                        "name",
                                                        "features"
                                                    ],
                                                    "properties": {
                                                        "name": {
                                                            "type": "string"
                                                        },
                                                        "features": {
                                                            "type": "object",
                                                            "required": [
                                                                "name"
                                                            ],
                                                            "properties": {
                                                                "name": {
                                                                    "type": "string"
                                                                }
                                                            }
                                                        }
                                                    }
                                                },
                                                "issue": {
                                                    "type": "object",
                                                    "required": [
                                                        "by",
                                                        "to",
                                                        "hardware",
                                                        "date"
                                                    ],
                                                    "properties": {
                                                        "by": {
                                                            "type": "string"
                                                        },
                                                        "to": {
                                                            "type": "string"
                                                        },
                                                        "hardware": {
                                                            "type": "string"
                                                        },
                                                        "date": {
                                                            "type": "string"
                                                        }
                                                    }
                                                },
                                                "serviceLevelAgreement": {
                                                    "type": "object",
                                                    "required": [
                                                        "uptime",
                                                        "period",
                                                        "meanTimeBetweenFailure",
                                                        "meanTimeToRepair",
                                                        "expire"
                                                    ],
                                                    "properties": {
                                                        "uptime": {
                                                            "type": "number"
                                                        },
                                                        "period": {
                                                            "type": "string"
                                                        },
                                                        "meanTimeBetweenFailure": {
                                                            "type": "string"
                                                        },
                                                        "meanTimeToRepair": {
                                                            "type": "string"
                                                        }
                                                    }
                                                },
                                                "expire": {
                                                    "type": "object",
                                                    "required": [
                                                        "date",
                                                        "days"
                                                    ],
                                                    "properties": {
                                                        "date": {
                                                            "type": "string"
                                                        },
                                                        "days": {
                                                            "type": "integer"
                                                        }
                                                    }
                                                }
                                            }
                                        },
                                        "status": {
                                            "type": "string"
                                        },
                                        "vcpu": {
                                            "type": "object",
                                            "required": [
                                                "totalCores",
                                                "usedCores",
                                                "freeCores"
                                            ],
                                            "properties": {
                                                "totalCores": {
                                                    "type": "integer"
                                                },
                                                "usedCores": {
                                                    "type": "integer"
                                                },
                                                "freeCores": {
                                                    "type": "integer"
                                                }
                                            }
                                        },
                                        "memory": {
                                            "type": "object",
                                            "required": [
                                                "totalMiB",
                                                "usedMiB",
                                                "freeMiB"
                                            ],
                                            "properties": {
                                                "totalMiB": {
                                                    "type": "integer"
                                                },
                                                "usedMiB": {
                                                    "type": "integer"
                                                },
                                                "freeMiB": {
                                                    "type": "integer"
                                                }
                                            }
                                        },
                                        "storage": {
                                            "type": "object",
                                            "required": [
                                                "totalMiB",
                                                "usedMiB",
                                                "freeMiB"
                                            ],
                                            "properties": {
                                                "totalMiB": {
                                                    "type": "integer"
                                                },
                                                "usedMiB": {
                                                    "type": "integer"
                                                },
                                                "freeMiB": {
                                                    "type": "integer"
                                                }
                                            }
                                        },
                                        "uptimeSeconds": {
                                            "type": "number"
                                        },
                                        "labels": {
                                            "type": "object",
                                            "required": [
                                                "isGpuEnabled"
                                            ],
                                            "properties": {
                                                "isGpuEnabled": {
                                                    "type": "string"
                                                }
                                            }
                                        }
                                    }
                                }
                            },
                            "page": {
                                "type": "object",
                                "required": [
                                    "total",
                                    "number",
                                    "size"
                                ],
                                "properties": {
                                    "total": {
                                        "type": "integer",
                                        "example": 10
                                    },
                                    "number": {
                                        "type": "integer",
                                        "example": 1
                                    },
                                    "size": {
                                        "type": "integer",
                                        "example": 1
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "fetch nodes successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetTokensRequest": {
                "type": "object",
                "required": [
                    "name",
                    "password"
                ],
                "properties": {
                    "name": {
                        "type": "string",
                        "description": "the name of user to generate the token",
                        "example": "test-name"
                    },
                    "password": {
                        "type": "string",
                        "description": "the password of user to generate the token",
                        "example": "test-password"
                    }
                }
            },
            "GetTokensResponse": {
                "type": "object",
                "required": [
                    "code",
                    "data",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer"
                    },
                    "data": {
                        "type": "object",
                        "required": [
                            "token",
                            "refresh",
                            "expires"
                        ],
                        "properties": {
                            "token": {
                                "type": "string"
                            },
                            "refresh": {
                                "type": "string"
                            },
                            "expires": {
                                "type": "object",
                                "required": [
                                    "access",
                                    "refresh"
                                ],
                                "properties": {
                                    "access": {
                                        "type": "integer"
                                    },
                                    "refresh": {
                                        "type": "integer"
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            }
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Cube COS API",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
