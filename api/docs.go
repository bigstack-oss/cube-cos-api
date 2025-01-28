package api

import "github.com/swaggo/swag"

const docTemplate = `{
    "openapi": "3.0.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/logout": {
            "post": {
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
        "/api/v1/datacenters": {
            "get": {
                "description": "Retrieve the list of data centers",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
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
                                                    "name": {
                                                        "type": "string",
                                                        "example": "bigstack-data-center"
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
                        "example": "test-data-center"
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
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of events successfully",
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
                                                "events": {
                                                    "type": "array",
                                                    "items": {
                                                        "type": "object",
                                                        "properties": {
                                                            "type": {
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
                "tags": [
                    "Health"
                ],
                "summary": "Retrieve the list of health",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "test-data-center"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of health successfully",
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
                                                "overall": {
                                                    "type": "object",
                                                    "properties": {
                                                        "status": {
                                                            "type": "object",
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
                                                "inUse": {
                                                    "type": "array",
                                                    "items": {
                                                        "properties": {
                                                            "service": {
                                                                "type": "string",
                                                                "example": "clusterLink"
                                                            },
                                                            "category": {
                                                                "type": "string",
                                                                "example": "core"
                                                            },
                                                            "status": {
                                                                "type": "object",
                                                                "properties": {
                                                                    "current": {
                                                                        "type": "string",
                                                                        "example": "ok"
                                                                    }
                                                                }
                                                            },
                                                            "module": {
                                                                "type": "array",
                                                                "items": {
                                                                    "type": "object",
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
                                                },
                                                "error": {
                                                    "type": "array",
                                                    "items": {
                                                        "properties": {
                                                            "service": {
                                                                "type": "string",
                                                                "example": "storage"
                                                            },
                                                            "category": {
                                                                "type": "string",
                                                                "example": "storage"
                                                            },
                                                            "status": {
                                                                "type": "object",
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
                                                            },
                                                            "module": {
                                                                "type": "array",
                                                                "items": {
                                                                    "type": "object",
                                                                    "properties": {
                                                                        "name": {
                                                                            "type": "string",
                                                                            "example": "ceph_osd"
                                                                        },
                                                                        "isAutoRepairable": {
                                                                            "type": "boolean",
                                                                            "example": true
                                                                        },
                                                                        "status": {
                                                                            "type": "object",
                                                                            "properties": {
                                                                                "current": {
                                                                                    "type": "string",
                                                                                    "example": "ng"
                                                                                },
                                                                                "description": {
                                                                                    "type": "string",
                                                                                    "example": "2 ceph_osd down"
                                                                                }
                                                                            }
                                                                        }
                                                                    }
                                                                }
                                                            }
                                                        }
                                                    }
                                                },
                                                "fixing": {
                                                    "type": "array"
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
        "/api/v1/datacenters/{dataCenter}/healths/{module}/repair": {
            "post": {
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
                        "example": "test-data-center"
                    },
                    {
                        "in": "path",
                        "name": "module",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the module to repair. The value can be 'all' to repair all modules under all services, or other module names like 'ceph_osd', 'nova', and so on.",
                        "example": "all, ceph_osd, or other module names"
                    }
                ],
                "responses": {
                    "202": {
                        "description": "The Request of the unhealthy module repair is accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
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
        "/api/v1/datacenters/{dataCenter}/integrations": {
            "get": {
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
                        "example": "test-data-center"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of integrated applications successfully",
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
        "/api/v1/datacenters/{dataCenter}/nodes": {
            "get": {
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
                        "example": "test-data-center"
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
                        "description": "Retrieve the list of nodes successfully",
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
                                                "nodes": {
                                                    "type": "array",
                                                    "items": {
                                                        "type": "object",
                                                        "properties": {
                                                            "id": {
                                                                "type": "string",
                                                                "example": "7a18177c"
                                                            },
                                                            "hostname": {
                                                                "type": "string",
                                                                "example": "bigstack-host"
                                                            },
                                                            "role": {
                                                                "type": "string",
                                                                "example": "control-converged"
                                                            },
                                                            "address": {
                                                                "type": "string",
                                                                "example": "10.10.10.10"
                                                            },
                                                            "managementIP": {
                                                                "type": "string",
                                                                "example": "192.168.10.10"
                                                            },
                                                            "license": {
                                                                "type": "object",
                                                                "properties": {
                                                                    "status": {
                                                                        "type": "string",
                                                                        "example": "trial"
                                                                    },
                                                                    "hostname": {
                                                                        "type": "string",
                                                                        "example": "bigstack-host"
                                                                    },
                                                                    "serial": {
                                                                        "type": "string",
                                                                        "example": "1N7L603"
                                                                    },
                                                                    "product": {
                                                                        "type": "object",
                                                                        "properties": {
                                                                            "name": {
                                                                                "type": "string",
                                                                                "example": "cubecos"
                                                                            },
                                                                            "features": {
                                                                                "type": "object",
                                                                                "properties": {
                                                                                    "name": {
                                                                                        "type": "string",
                                                                                        "example": "instance"
                                                                                    }
                                                                                }
                                                                            }
                                                                        }
                                                                    },
                                                                    "issue": {
                                                                        "type": "object",
                                                                        "properties": {
                                                                            "by": {
                                                                                "type": "string",
                                                                                "example": "Bigstack co., ltd."
                                                                            },
                                                                            "to": {
                                                                                "type": "string",
                                                                                "example": "bigstack"
                                                                            },
                                                                            "hardware": {
                                                                                "type": "string",
                                                                                "example": "*"
                                                                            },
                                                                            "date": {
                                                                                "type": "string",
                                                                                "example": "2025-01-01T01:00:00Z"
                                                                            }
                                                                        }
                                                                    },
                                                                    "serviceLevelAgreement": {
                                                                        "type": "object",
                                                                        "properties": {
                                                                            "uptime": {
                                                                                "type": "int",
                                                                                "example": 99.99
                                                                            },
                                                                            "period": {
                                                                                "type": "string",
                                                                                "example": "24x7"
                                                                            },
                                                                            "meanTimeBetweenFailure": {
                                                                                "type": "string",
                                                                                "example": "5 mins"
                                                                            },
                                                                            "meanTimeToRepair": {
                                                                                "type": "string",
                                                                                "example": "15 mins"
                                                                            }
                                                                        }
                                                                    },
                                                                    "expire": {
                                                                        "type": "object",
                                                                        "properties": {
                                                                            "date": {
                                                                                "type": "string",
                                                                                "example": "2025-01-01T01:00:00Z"
                                                                            },
                                                                            "days": {
                                                                                "type": "integer",
                                                                                "example": 30
                                                                            }
                                                                        }
                                                                    }
                                                                }
                                                            },
                                                            "status": {
                                                                "type": "string",
                                                                "example": "up"
                                                            },
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
                                                            },
                                                            "uptime": {
                                                                "type": "string",
                                                                "example": "26 days"
                                                            },
                                                            "labels": {
                                                                "type": "object",
                                                                "properties": {
                                                                    "isGpuEnabled": {
                                                                        "type": "string",
                                                                        "example": "true"
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
        "/api/v1/datacenters/{dataCenter}/summary": {
            "get": {
                "tags": [
                    "Summary"
                ],
                "summary": "Retrieve the summary of data center",
                "parameters": [
                    {
                        "in": "path",
                        "name": "dataCenter",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the data center to operate",
                        "example": "test-data-center"
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
                                            "example": "failed to fetch summary: internal server error"
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
                        "example": "test-data-center"
                    }
                ],
                "requestBody": {
                    "description": "The user name and password to generate the token",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "type": "object",
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
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 201
                                        },
                                        "data": {
                                            "type": "object",
                                            "properties": {
                                                "token": {
                                                    "type": "string",
                                                    "example": "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIwdDdGdWlJZC1lbnhVUWRZWGVZalZ6Q0pQZFRWMmxaU0NZanRkQW01S3djIn0.eyJleHAiOjE3MzgwODM0MTYsImlhdCI6MTczODA3NjIxNiwianRpIjoiZjg3MGQzNjAtNzhhZi00NGNmLWI2YjktMTNmMWM0NzhkMWU0IiwiaXNzIjoiaHR0cHM6Ly8xMC4zMi4xMC4xODA6MTA0NDMvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiYWNjb3VudCIsInN1YiI6IjNhNDUwOTU5LTYyYTctNDA2Mi04OGM0LWMyMGYxNTIxOTYxZiIsInR5cCI6IkJlYXJlciIsImF6cCI6InRva2VuLWNvbm5lY3QiLCJzZXNzaW9uX3N0YXRlIjoiMDViNDNhYjItNmRmMy00NjRkLWJlYTEtMGQxYmE2NzFiZWI5IiwiYWNyIjoiMSIsInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJkZWZhdWx0LXJvbGVzLW1hc3RlciIsIm9mZmxpbmVfYWNjZXNzIiwidW1hX2F1dGhvcml6YXRpb24iXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6Im9wZW5pZCBlbWFpbCBwcm9maWxlIiwic2lkIjoiMDViNDNhYjItNmRmMy00NjRkLWJlYTEtMGQxYmE2NzFiZWI5IiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJiaWdzdGFjayJ9.HC5CPPvRpkwdiufRg_Ous0k_3ifOWsgNoIqNCYaW3pX5yzizaOxzPXn9jDUkXEbirCf01OPiGtq1e5EXUg41pKHSL45MGJEQDn28fqrUBTN5Ixxwq83_54o7jQLqdV1PBaxw-SEZCa8_XArwtBsXRjm8A3cKnuzRU4xb5TGrOc1VDydQOLUFjCqMwV-V65CQ0Vt03NyiAjVeeBLiL5truT0F2ZgiuQEhDHaCgBR1wSeReYBYBhOGLiq0QA4GzgNlmTjdOC7RrXV1w7QPv2i_7IPbWCUNrFnZPGr2KJBLiIot72t2UmLsjSJ6a7jx7u1vxQ-Wx5TQQ_TiglGcghMwFg"
                                                },
                                                "refresh": {
                                                    "type": "string",
                                                    "example": "eyJhbGciOiJIUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICI1OGRkY2JhNC1hM2ZkLTQ2MTYtODgyYi1lMGY1ZjRlNzAyMTIifQ.eyJleHAiOjE3MzgwNzgwMTYsImlhdCI6MTczODA3NjIxNiwianRpIjoiN2Y0ZWU4MTYtMzdjZi00OGQyLTk2ZjktNTI5YjVkNDhjYzQzIiwiaXNzIjoiaHR0cHM6Ly8xMC4zMi4xMC4xODA6MTA0NDMvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiaHR0cHM6Ly8xMC4zMi4xMC4xODA6MTA0NDMvYXV0aC9yZWFsbXMvbWFzdGVyIiwic3ViIjoiM2E0NTA5NTktNjJhNy00MDYyLTg4YzQtYzIwZjE1MjE5NjFmIiwidHlwIjoiUmVmcmVzaCIsImF6cCI6InRva2VuLWNvbm5lY3QiLCJzZXNzaW9uX3N0YXRlIjoiMDViNDNhYjItNmRmMy00NjRkLWJlYTEtMGQxYmE2NzFiZWI5Iiwic2NvcGUiOiJvcGVuaWQgZW1haWwgcHJvZmlsZSIsInNpZCI6IjA1YjQzYWIyLTZkZjMtNDY0ZC1iZWExLTBkMWJhNjcxYmViOSJ9.yeFgPdQHu5Xp7CCpCeOiGoOGTf5Hesrad0VHtdWg2Vc"
                                                },
                                                "expires": {
                                                    "type": "object",
                                                    "properties": {
                                                        "access": {
                                                            "type": "integer",
                                                            "example": 7200
                                                        },
                                                        "refresh": {
                                                            "type": "integer",
                                                            "example": 1800
                                                        }
                                                    }
                                                }
                                            }
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "create token successfully"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "created"
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
