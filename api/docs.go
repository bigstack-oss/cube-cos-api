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
                        "$ref": "#/components/parameters/dataCenter"
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
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Data centers",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "type": "cloud",
                                                    "name": "example-data-center",
                                                    "roles": [
                                                        "control-converged",
                                                        "control",
                                                        "compute",
                                                        "storage"
                                                    ],
                                                    "version": "Cube Appliance 3.0.0",
                                                    "virtualIp": "10.10.10.10",
                                                    "isLocal": true,
                                                    "isHaEnabled": false,
                                                    "utcTimeZone": "+00:00",
                                                    "additional": {
                                                        "helpUrl": "https://www.bigstack.co/contact-us",
                                                        "nodeLicenseStatus": {
                                                            "valid": 3,
                                                            "expired": 0,
                                                            "unlicense": 0
                                                        }
                                                    }
                                                }
                                            ],
                                            "msg": "fetch data center list successfully",
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
        "/api/v1/datacenters/{dataCenter}": {
            "get": {
                "operationId": "getDataCenter",
                "description": "Retrieve the data center info",
                "tags": [
                    "Data Centers"
                ],
                "summary": "Retrieve the data center info",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the data center info successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetDataCenterResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Cloud type data center",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "type": "cloud",
                                                "name": "example-data-center",
                                                "roles": [
                                                    "control-converged",
                                                    "control",
                                                    "compute",
                                                    "storage"
                                                ],
                                                "version": "Cube Appliance 3.0.0",
                                                "virtualIp": "10.10.10.10",
                                                "isLocal": true,
                                                "isHaEnabled": false,
                                                "utcTimeZone": "+00:00",
                                                "additional": {
                                                    "helpUrl": "https://www.bigstack.co/contact-us",
                                                    "nodeLicenseStatus": {
                                                        "valid": 3,
                                                        "expired": 0,
                                                        "unlicense": 0
                                                    }
                                                }
                                            },
                                            "msg": "fetch data center info successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example2": {
                                        "summary": "Edge type data center",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "type": "edge",
                                                "name": "example-data-center",
                                                "roles": [
                                                    "moderator",
                                                    "edge-core"
                                                ],
                                                "version": "Cube Appliance 3.0.0",
                                                "virtualIp": "10.10.10.10",
                                                "isLocal": true,
                                                "isHaEnabled": false,
                                                "utcTimeZone": "+00:00",
                                                "additional": {
                                                    "helpUrl": "https://www.bigstack.co/contact-us",
                                                    "nodeLicenseStatus": {
                                                        "valid": 3,
                                                        "expired": 0,
                                                        "unlicense": 0
                                                    }
                                                }
                                            },
                                            "msg": "fetch data center info successfully",
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
                                            "example": "failed to fetch data center: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/services": {
            "get": {
                "operationId": "getServices",
                "tags": [
                    "Services"
                ],
                "summary": "Retrieve the list of services",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of services successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetServicesResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Services",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "name": "clusterLink",
                                                    "category": "core",
                                                    "modules": [
                                                        {
                                                            "name": "link"
                                                        },
                                                        {
                                                            "name": "clock"
                                                        },
                                                        {
                                                            "name": "dns"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "clusterSys",
                                                    "category": "core",
                                                    "modules": [
                                                        {
                                                            "name": "bootstrap"
                                                        },
                                                        {
                                                            "name": "license"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "clusterSettings",
                                                    "category": "core",
                                                    "modules": [
                                                        {
                                                            "name": "etcd"
                                                        },
                                                        {
                                                            "name": "nodelist"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "haCluster",
                                                    "category": "core",
                                                    "modules": [
                                                        {
                                                            "name": "hacluster"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "msgQueue",
                                                    "category": "core",
                                                    "modules": [
                                                        {
                                                            "name": "rabbitmq"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "iaasDb",
                                                    "category": "core",
                                                    "modules": [
                                                        {
                                                            "name": "mysql"
                                                        },
                                                        {
                                                            "name": "mongodb"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "virtualIp",
                                                    "category": "core",
                                                    "modules": [
                                                        {
                                                            "name": "vip"
                                                        },
                                                        {
                                                            "name": "haproxy_ha"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "storage",
                                                    "category": "storage",
                                                    "modules": [
                                                        {
                                                            "name": "ceph"
                                                        },
                                                        {
                                                            "name": "ceph_mon"
                                                        },
                                                        {
                                                            "name": "ceph_mgr"
                                                        },
                                                        {
                                                            "name": "ceph_mds"
                                                        },
                                                        {
                                                            "name": "ceph_osd"
                                                        },
                                                        {
                                                            "name": "ceph_rgw"
                                                        },
                                                        {
                                                            "name": "rbd_target"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "apiService",
                                                    "category": "core",
                                                    "modules": [
                                                        {
                                                            "name": "haproxy"
                                                        },
                                                        {
                                                            "name": "httpd"
                                                        },
                                                        {
                                                            "name": "skyline"
                                                        },
                                                        {
                                                            "name": "lmi"
                                                        },
                                                        {
                                                            "name": "memcache"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "singleSignOn",
                                                    "category": "core",
                                                    "modules": [
                                                        {
                                                            "name": "k3s"
                                                        },
                                                        {
                                                            "name": "keycloak"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "network",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "neutron"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "compute",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "nova"
                                                        },
                                                        {
                                                            "name": "cyborg"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "bareMetal",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "ironic"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "image",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "glance"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "blockStor",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "cinder"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "fileStor",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "manila"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "objectStor",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "swift"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "orchestration",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "heat"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "lbaas",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "octavia"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "dnsaas",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "designate"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "k8saas",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "rancher"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "instanceHa",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "masakari"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "businessLogic",
                                                    "category": "cloud computing",
                                                    "modules": [
                                                        {
                                                            "name": "senlin"
                                                        },
                                                        {
                                                            "name": "watcher"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "dataPipe",
                                                    "category": "infrascope",
                                                    "modules": [
                                                        {
                                                            "name": "zookeeper"
                                                        },
                                                        {
                                                            "name": "kafka"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "metrics",
                                                    "category": "infrascope",
                                                    "modules": [
                                                        {
                                                            "name": "monasca"
                                                        },
                                                        {
                                                            "name": "telegraf"
                                                        },
                                                        {
                                                            "name": "grafana"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "logAnalytics",
                                                    "category": "infrascope",
                                                    "modules": [
                                                        {
                                                            "name": "filebeat"
                                                        },
                                                        {
                                                            "name": "auditbeat"
                                                        },
                                                        {
                                                            "name": "logstash"
                                                        },
                                                        {
                                                            "name": "opensearch"
                                                        },
                                                        {
                                                            "name": "opensearch-dashboards"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "notifications",
                                                    "category": "infrascope",
                                                    "modules": [
                                                        {
                                                            "name": "influxdb"
                                                        },
                                                        {
                                                            "name": "kapacitor"
                                                        }
                                                    ]
                                                }
                                            ],
                                            "msg": "fetch service details successfully",
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
                                            "example": "failed to fetch services: internal server error"
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
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "query",
                        "name": "type",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "system",
                                "host",
                                "instance"
                            ]
                        },
                        "description": "The type of event to query, the value can be only 'system', 'host', and 'instance'.",
                        "example": "system"
                    },
                    {
                        "in": "query",
                        "name": "keyword",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The keyword of the event to query.",
                        "example": "10.10.10.10 is reachable"
                    },
                    {
                        "in": "query",
                        "name": "id",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The id of the event to query.",
                        "example": "NET00003I"
                    },
                    {
                        "in": "query",
                        "name": "categories",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        },
                        "description": "The category of the event to query.",
                        "example": "NET"
                    },
                    {
                        "in": "query",
                        "name": "severities",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string",
                                "enum": [
                                    "Info",
                                    "Warning",
                                    "Error"
                                ]
                            }
                        },
                        "description": "The severity of the event to query, the value can be only 'Info', 'Warning', and 'Error'.",
                        "example": "Info"
                    },
                    {
                        "in": "query",
                        "name": "hosts",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        },
                        "description": "The host of the event to query.",
                        "example": "example-node-0"
                    },
                    {
                        "in": "query",
                        "name": "instances",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        },
                        "description": "The instance of the event to query.",
                        "example": "ccc449e4-a26c-47ac-afc1-c792ab1ed20a"
                    },
                    {
                        "in": "query",
                        "name": "past",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The past time of the event to query, use the unit of 's'(second), 'm'(minute), 'h'(hour), and 'd'(day) suffix to specify the time range.",
                        "example": "1d"
                    },
                    {
                        "$ref": "#/components/parameters/start"
                    },
                    {
                        "$ref": "#/components/parameters/stop"
                    },
                    {
                        "$ref": "#/components/parameters/pageSize"
                    },
                    {
                        "$ref": "#/components/parameters/pageNum"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
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
                                            "example": "invalid 'start' time: 2021-09-01T111:00:00+00:00"
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
        "/api/v1/datacenters/{dataCenter}/events/predefined": {
            "get": {
                "operationId": "getPredefinedEvents",
                "description": "Please note that if the 'types', 'categories', 'severities', or 'ids' query parameters are not specified, then 'NO' any predefined events will be returned.",
                "tags": [
                    "Events"
                ],
                "summary": "Retrieve the predefined events",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "query",
                        "name": "types",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string",
                                "enum": [
                                    "system",
                                    "host",
                                    "instance"
                                ]
                            }
                        },
                        "description": "The types of event to query, the value can be only 'system', 'host', and 'instance'.",
                        "example": "system"
                    },
                    {
                        "in": "query",
                        "name": "categories",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string",
                                "enum": [
                                    "PLC",
                                    "BSP",
                                    "RUG",
                                    "KSN",
                                    "CMP",
                                    "CEP",
                                    "NET",
                                    "SRV",
                                    "SDN",
                                    "VRT",
                                    "ETH",
                                    "CPU",
                                    "DSK",
                                    "MEM",
                                    "DEV"
                                ]
                            }
                        },
                        "description": "The categories of the event to query.",
                        "example": "CPU"
                    },
                    {
                        "in": "query",
                        "name": "severities",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string",
                                "enum": [
                                    "INFO",
                                    "WARNING",
                                    "CRITICAL"
                                ]
                            }
                        },
                        "description": "The severities of the event to query.",
                        "example": "Info"
                    },
                    {
                        "in": "query",
                        "name": "ids",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string",
                                "enum": [
                                    "PLC00001I",
                                    "PLC00002I",
                                    "BSP00001I",
                                    "RUG00001I",
                                    "KSN00001I",
                                    "KSN00002I",
                                    "CMP01001I",
                                    "CMP01002I",
                                    "CMP02001I",
                                    "CMP02002I",
                                    "CMP02003I",
                                    "CEP00001I",
                                    "CEP00002W",
                                    "NET00001I",
                                    "NET00002W",
                                    "NET00003I",
                                    "NET00004W",
                                    "SRV00001I",
                                    "SRV00002W",
                                    "SRV00003C",
                                    "SRV01001I",
                                    "SRV01002I",
                                    "SDN00001I",
                                    "SDN00002I",
                                    "VRT00001I",
                                    "VRT00002W",
                                    "VRT00003C",
                                    "ETH00001I",
                                    "ETH00002W",
                                    "CPU00001I",
                                    "CPU00002W",
                                    "CPU00003C",
                                    "DSK00001I",
                                    "DSK00002W",
                                    "DSK00003C",
                                    "MEM00001I",
                                    "MEM00002W",
                                    "MEM00003C",
                                    "DEV00001I",
                                    "DEV00002W",
                                    "CPU00004I",
                                    "CPU00005W",
                                    "CPU00006C",
                                    "CPU00007I",
                                    "MEM00004I",
                                    "MEM00005W",
                                    "MEM00006C",
                                    "MEM00007I"
                                ]
                            }
                        },
                        "description": "The ids of the event to query.",
                        "example": "CPU00004I"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the predefined events successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetPredefinedEventFilterResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Predefined event list",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "type": "system",
                                                    "id": "PLC00001I",
                                                    "severity": "INFO",
                                                    "category": "PLC",
                                                    "description": "CubeCOS node policy is successfully applied"
                                                },
                                                {
                                                    "type": "system",
                                                    "id": "PLC00002I",
                                                    "severity": "INFO",
                                                    "category": "PLC",
                                                    "description": "CubeCOS cluster policy is successfully applied"
                                                },
                                                {
                                                    "type": "system",
                                                    "id": "BSP00001I",
                                                    "severity": "INFO",
                                                    "category": "BSP",
                                                    "description": "CubeCOS node bootstrapping is started"
                                                }
                                            ],
                                            "msg": "fetch predefined events successfully",
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
                                            "example": "failed to fetch predefined events: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/events/filterConditions": {
            "get": {
                "operationId": "getEventFilterConditions",
                "tags": [
                    "Events"
                ],
                "summary": "Retrieve the event filter conditions",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "query",
                        "name": "past",
                        "required": false,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "1h",
                                "24h",
                                "7d",
                                "14d"
                            ]
                        },
                        "description": "The past time of the filter conditions to query, click 'try it out' to see a few options, but can specify with the 's'(second), 'm'(minute), 'h'(hour), and 'd'(day) suffix for other time ranges.",
                        "example": "1d"
                    },
                    {
                        "$ref": "#/components/parameters/start"
                    },
                    {
                        "$ref": "#/components/parameters/stop"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the event filter conditions successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetEventFilterConditionResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Event filter conditions",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "system": {
                                                    "severities": [
                                                        "CRITICAL",
                                                        "Info",
                                                        "Warning"
                                                    ],
                                                    "categories": [
                                                        "KSN",
                                                        "NET",
                                                        "SDN",
                                                        "SRV"
                                                    ]
                                                },
                                                "instance": {
                                                    "ids": [
                                                        "028952d3-c0ba-4494-96c0-2bf1bab407e5",
                                                        "02ecd121-c19b-41c7-8ed7-390745b01af4",
                                                        "12893f22-a353-4dce-a164-f889e0f39951",
                                                        "fbd5d7c2-38c8-436b-97f4-d9e5b322d02b"
                                                    ],
                                                    "categories": [
                                                        "CPU",
                                                        "MEM"
                                                    ]
                                                },
                                                "host": {
                                                    "names": [
                                                        "example-node-0"
                                                    ],
                                                    "categories": [
                                                        "DSK",
                                                        "MEM"
                                                    ]
                                                }
                                            },
                                            "msg": "fetch event filter conditions successfully",
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
                                            "example": "invalid 'start' time: 2021-09-01T111:00:00+00:00"
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
                                            "example": "failed to fetch event filter conditions: internal server error"
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
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "query",
                        "name": "type",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "system",
                                "host",
                                "instance"
                            ]
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
                        "$ref": "#/components/parameters/watch"
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
                                                        "time": "2025-02-04T06:05:08+00:00"
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
                                                        "time": "2025-02-03T19:30:15+00:00"
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
                                                        "time": "2025-02-03T19:15:15+00:00"
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
                                                        "time": "2025-02-03T17:32:09+00:00"
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
                                                        "time": "2025-02-03T17:32:09+00:00"
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
                                                        "time": "2025-02-03T17:30:15+00:00"
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
                                                        "time": "2025-02-03T17:30:15+00:00"
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
                                                        "time": "2025-02-03T17:15:15+00:00"
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
                                                        "time": "2025-02-03T16:50:15+00:00"
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
                                                        "time": "2025-02-03T16:35:15+00:00"
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
                                            "example": "invalid 'start' time: 2021-09-01T111:00:00+00:00"
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
        "/api/v1/datacenters/{dataCenter}/events/rank": {
            "get": {
                "operationId": "getRankedEvents",
                "tags": [
                    "Events"
                ],
                "summary": "Retrieve the ranked events",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "query",
                        "name": "type",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "system",
                                "host",
                                "instance"
                            ]
                        },
                        "description": "The type of event to query, the value can be only 'system', 'host', and 'instance'.",
                        "example": "system"
                    },
                    {
                        "in": "query",
                        "name": "category",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The category of the event to query.",
                        "example": "NET"
                    },
                    {
                        "in": "query",
                        "name": "categories",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        },
                        "description": "The categories of the event to query.",
                        "example": "CPU"
                    },
                    {
                        "in": "query",
                        "name": "severity",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The severity of the event to query, the value can be only 'Info', 'Warning', and 'Error'.",
                        "example": "Info"
                    },
                    {
                        "in": "query",
                        "name": "severities",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        },
                        "description": "The severities of the event to query.",
                        "example": "Info"
                    },
                    {
                        "in": "query",
                        "name": "host",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The host of the event to query.",
                        "example": "example-data-center"
                    },
                    {
                        "in": "query",
                        "name": "hosts",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        },
                        "description": "The hosts of the event to query.",
                        "example": "example-node-0"
                    },
                    {
                        "in": "query",
                        "name": "instance",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The instance of the event to query.",
                        "example": "ccc449e4-a26c-47ac-afc1-c792ab1ed20a"
                    },
                    {
                        "in": "query",
                        "name": "instances",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        },
                        "description": "The instances of the event to query.",
                        "example": "ccc449e4-a26c-47ac-afc1-c792ab1ed20a"
                    },
                    {
                        "in": "query",
                        "name": "past",
                        "required": false,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "1h",
                                "24h",
                                "7d",
                                "14d"
                            ]
                        },
                        "description": "The past time of the event rank to query, click 'try it out' to see a few options, but can specify with the 's'(second), 'm'(minute), 'h'(hour), and 'd'(day) suffix for other time ranges.",
                        "example": "1d"
                    },
                    {
                        "$ref": "#/components/parameters/start"
                    },
                    {
                        "$ref": "#/components/parameters/stop"
                    },
                    {
                        "in": "query",
                        "name": "limit",
                        "required": false,
                        "schema": {
                            "type": "integer"
                        },
                        "description": "The limit of the rank of event to return (default is 10).",
                        "example": 10
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the rank of event successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetRankedEventsResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Ranked system events",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "events": [
                                                    {
                                                        "id": "NET00003I",
                                                        "category": "NET",
                                                        "severity": "Info",
                                                        "percent": 98.2142,
                                                        "number": 440
                                                    },
                                                    {
                                                        "id": "NET00001I",
                                                        "category": "NET",
                                                        "severity": "Info",
                                                        "percent": 1.7857,
                                                        "number": 8
                                                    }
                                                ],
                                                "limit": {
                                                    "number": 25,
                                                    "description": "The top 2 event IDs with the highest proportion"
                                                }
                                            },
                                            "msg": "fetch event rank successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example2": {
                                        "summary": "Ranked host events",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "events": [
                                                    {
                                                        "id": "CPU00002W",
                                                        "category": "CPU",
                                                        "host": "example-host-0",
                                                        "percent": 75,
                                                        "number": 3
                                                    },
                                                    {
                                                        "id": "CPU00002W",
                                                        "category": "CPU",
                                                        "host": "example-host-1",
                                                        "percent": 25,
                                                        "number": 1
                                                    }
                                                ],
                                                "limit": {
                                                    "number": 25,
                                                    "description": "The top 2 event IDs with the highest proportion"
                                                }
                                            },
                                            "msg": "fetch event rank successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example3": {
                                        "summary": "Ranked instance events",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "events": [
                                                    {
                                                        "id": "CPU00007I",
                                                        "category": "CPU",
                                                        "instanceId": "74ac7cb9-800a-430f-809e-d96f0ce3e9a2",
                                                        "instanceName": "example-vm-0",
                                                        "percent": 75,
                                                        "number": 3
                                                    },
                                                    {
                                                        "id": "CPU00007I",
                                                        "category": "CPU",
                                                        "instanceId": "e8a44a60-b79f-42f8-bab6-761b1734299a",
                                                        "instanceName": "example-vm-1",
                                                        "percent": 25,
                                                        "number": 1
                                                    }
                                                ],
                                                "limit": {
                                                    "number": 25,
                                                    "description": "The top 2 event IDs with the highest proportion"
                                                }
                                            },
                                            "msg": "fetch event rank successfully",
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
                                            "example": "invalid 'start' time: 2021-09-01T111:00:00+00:00"
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
        "/api/v1/datacenters/{dataCenter}/notifications": {
            "get": {
                "operationId": "getNotifications",
                "tags": [
                    "Notifications"
                ],
                "summary": "Retrieve the notifications",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/start"
                    },
                    {
                        "$ref": "#/components/parameters/stop"
                    },
                    {
                        "in": "query",
                        "name": "past",
                        "required": false,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The past time of the notifications to query, click 'try it out' to see a few options, but can specify with the 's'(second), 'm'(minute), 'h'(hour), and 'd'(day) suffix for other time ranges.",
                        "example": "1h"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the notifications successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetNotificationsResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Notifications",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "id": "DEV00002I",
                                                    "nodeName": "example-node-0",
                                                    "time": "2025-07-21T04:09:30+08:00",
                                                    "additionalInfo": {
                                                        "class": "SSD",
                                                        "device": "/dev/sdb"
                                                    }
                                                },
                                                {
                                                    "id": "OSD00002E",
                                                    "nodeName": "example-node-1",
                                                    "time": "2025-07-21T04:09:30+08:00",
                                                    "additionalInfo": {
                                                        "osdId": "osd.1",
                                                        "reweight": 0.75
                                                    }
                                                }
                                            ],
                                            "msg": "fetch notifications successfully",
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
                                            "example": "failed to fetch notifications: internal server error"
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
                "operationId": "getHealths",
                "tags": [
                    "Health"
                ],
                "summary": "Retrieve the overall health status of all modules",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    },
                    {
                        "in": "query",
                        "name": "past",
                        "required": false,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "1h",
                                "24h",
                                "7d",
                                "14d",
                                "30d"
                            ]
                        },
                        "description": "The past time of the health to query, click 'try it out' to see a few options, but can specify with the 's'(second), 'm'(minute), 'h'(hour), and 'd'(day) suffix for other time ranges.",
                        "example": "1h"
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
                                                        "current": "ok",
                                                        "isFixing": false
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
                                                        "isFixing": false,
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
                "operationId": "repairAllModulesHealth",
                "tags": [
                    "Health"
                ],
                "summary": "Repair the health for all modules",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
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
        "/api/v1/datacenters/{dataCenter}/healths/services/{serviceType}": {
            "get": {
                "operationId": "getServiceHealthHistory",
                "tags": [
                    "Health"
                ],
                "summary": "Retrieve the all modules health status from a service",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "serviceType",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "clusterLink",
                                "clusterSys",
                                "clusterSettings",
                                "haCluster",
                                "msgQueue",
                                "iaasDb",
                                "virtualIp",
                                "storage",
                                "apiService",
                                "singleSignOn",
                                "network",
                                "compute",
                                "bareMetal",
                                "image",
                                "blockStor",
                                "fileStor",
                                "objectStor",
                                "orchestration",
                                "lbaas",
                                "dnsaas",
                                "k8saas",
                                "instanceHa",
                                "businessLogic",
                                "dataPipe",
                                "metrics",
                                "logAnalytics",
                                "notifications"
                            ]
                        },
                        "description": "The name of the service to retrieve health history. use GET /api/v1/datacenters/{dataCenter}/services to get the service list and their modules."
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    },
                    {
                        "$ref": "#/components/parameters/start"
                    },
                    {
                        "$ref": "#/components/parameters/stop"
                    },
                    {
                        "in": "query",
                        "name": "past",
                        "required": false,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "1h",
                                "24h",
                                "7d",
                                "14d",
                                "30d"
                            ]
                        },
                        "description": "The past time of the health history to query, click 'try it out' to see a few options, but can specify with the 's'(second), 'm'(minute), 'h'(hour), and 'd'(day) suffix for other time ranges.",
                        "example": "1d"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the health history of service successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetServiceHealthHistoryResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Health result of service",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "category": "cloud computing",
                                                    "service": "compute",
                                                    "module": "nova",
                                                    "isRepairable": true,
                                                    "history": [
                                                        {
                                                            "time": "2025-02-15T08:44:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ng",
                                                            "error": {
                                                                "type": "service down",
                                                                "reason": "1 node down",
                                                                "nodes": [
                                                                    "example-data-center"
                                                                ],
                                                                "description": "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
                                                                "details": "{ ... the best efforts of error summary / direction ...} ",
                                                                "log": "http://{dataCenter}:8888/log/nova/example-data-center-20250205113459-b3gc.log"
                                                            }
                                                        },
                                                        {
                                                            "time": "2025-02-15T08:49:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T08:54:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T08:59:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:04:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:09:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ng",
                                                            "error": {
                                                                "type": "service down",
                                                                "reason": "1 node down",
                                                                "nodes": [
                                                                    "example-data-center"
                                                                ],
                                                                "description": "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
                                                                "details": "{ ... the best efforts of error summary / direction ...} ",
                                                                "log": "http://{dataCenter}:8888/log/nova/example-data-center-20250205113459-b3gc.log"
                                                            }
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:14:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:19:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:24:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:29:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:34:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ng",
                                                            "error": {
                                                                "type": "service down",
                                                                "reason": "1 node down",
                                                                "nodes": [
                                                                    "example-data-center"
                                                                ],
                                                                "description": "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
                                                                "details": "{ ... the best efforts of error summary / direction ...} ",
                                                                "log": "http://{dataCenter}:8888/log/nova/example-data-center-20250205113459-b3gc.log"
                                                            }
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:39:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:44:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "category": "cloud computing",
                                                    "service": "compute",
                                                    "module": "cyborg",
                                                    "isRepairable": true,
                                                    "history": [
                                                        {
                                                            "time": "2025-02-15T08:44:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ng",
                                                            "error": {
                                                                "type": "service down",
                                                                "reason": "1 node down",
                                                                "nodes": [
                                                                    "example-data-center"
                                                                ],
                                                                "description": "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
                                                                "details": "{ ... the best efforts of error summary / direction ...} ",
                                                                "log": "http://{dataCenter}:8888/log/nova/example-data-center-20250205113459-b3gc.log"
                                                            }
                                                        },
                                                        {
                                                            "time": "2025-02-15T08:49:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T08:54:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T08:59:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:04:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:09:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ng",
                                                            "error": {
                                                                "type": "service down",
                                                                "reason": "1 node down",
                                                                "nodes": [
                                                                    "example-data-center"
                                                                ],
                                                                "description": "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
                                                                "details": "{ ... the best efforts of error summary / direction ...} ",
                                                                "log": "http://{dataCenter}:8888/log/nova/example-data-center-20250205113459-b3gc.log"
                                                            }
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:14:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:19:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:24:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:29:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:34:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ng",
                                                            "error": {
                                                                "type": "service down",
                                                                "reason": "1 node down",
                                                                "nodes": [
                                                                    "example-data-center"
                                                                ],
                                                                "description": "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
                                                                "details": "{ ... the best efforts of error summary / direction ...} ",
                                                                "log": "http://{dataCenter}:8888/log/nova/example-data-center-20250205113459-b3gc.log"
                                                            }
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:39:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        },
                                                        {
                                                            "time": "2025-02-15T09:44:36+00:00",
                                                            "hostname": "example-node-0",
                                                            "status": "ok"
                                                        }
                                                    ]
                                                }
                                            ],
                                            "msg": "fetch health successfully",
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
                                            "example": "invalid serviceType: example-service"
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
        "/api/v1/datacenters/{dataCenter}/healths/services/{serviceType}/modules/{moduleType}": {
            "get": {
                "operationId": "getHealthHistory",
                "tags": [
                    "Health"
                ],
                "summary": "Retrieve the health history of module",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "serviceType",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "clusterLink",
                                "clusterSys",
                                "clusterSettings",
                                "haCluster",
                                "msgQueue",
                                "iaasDb",
                                "virtualIp",
                                "storage",
                                "apiService",
                                "singleSignOn",
                                "network",
                                "compute",
                                "bareMetal",
                                "image",
                                "blockStor",
                                "fileStor",
                                "objectStor",
                                "orchestration",
                                "lbaas",
                                "dnsaas",
                                "k8saas",
                                "instanceHa",
                                "businessLogic",
                                "dataPipe",
                                "metrics",
                                "logAnalytics",
                                "notifications"
                            ]
                        },
                        "description": "The name of the service to retrieve health history. use GET /api/v1/datacenters/{dataCenter}/services to get the service list and their modules."
                    },
                    {
                        "in": "path",
                        "name": "moduleType",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "link",
                                "clock",
                                "dns",
                                "bootstrap",
                                "license",
                                "etcd",
                                "nodelist",
                                "hacluster",
                                "rabbitmq",
                                "mysql",
                                "mongodb",
                                "vip",
                                "haproxy_ha",
                                "ceph",
                                "ceph_mon",
                                "ceph_osd",
                                "ceph_rgw",
                                "ceph_mds",
                                "ceph_mgr",
                                "rbd_target",
                                "haproxy",
                                "httpd",
                                "skyline",
                                "api",
                                "memcache",
                                "k3s",
                                "keycloak",
                                "neutron",
                                "nova",
                                "cyborg",
                                "ironic",
                                "glance",
                                "cinder",
                                "manila",
                                "swift",
                                "heat",
                                "octavia",
                                "designate",
                                "rancher",
                                "masakari",
                                "zookeeper",
                                "kafka",
                                "monasca",
                                "telegraf",
                                "grafana",
                                "filebeat",
                                "auditbeat",
                                "logstash",
                                "opensearch",
                                "opensearch-dashboards",
                                "influxdb",
                                "kapacitor"
                            ]
                        },
                        "description": "The name of the module to retrieve health history. use GET /api/v1/datacenters/{dataCenter}/services/{serviceType} to get the module list."
                    },
                    {
                        "in": "query",
                        "name": "past",
                        "required": false,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "1h",
                                "24h",
                                "7d",
                                "14d",
                                "30d"
                            ]
                        },
                        "description": "The past time of the health history to query, click 'try it out' to see a few options, but can specify with the 's'(second), 'm'(minute), 'h'(hour), and 'd'(day) suffix for other time ranges.",
                        "example": "1d"
                    },
                    {
                        "$ref": "#/components/parameters/start"
                    },
                    {
                        "$ref": "#/components/parameters/stop"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    },
                    {
                        "in": "query",
                        "name": "aggregate",
                        "required": false,
                        "schema": {
                            "type": "boolean",
                            "default": false
                        },
                        "description": "If true, the health history will be aggregated by time and status(fixing related status > ng related status > ok related status), otherwise it will return the raw history data.",
                        "example": "false"
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
                                        "summary": "Health history of module",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "category": "cloud computing",
                                                "name": "compute",
                                                "module": "nova",
                                                "isRepairable": true,
                                                "history": [
                                                    {
                                                        "time": "2025-02-01T03:00:00+00:00",
                                                        "hostname": "example-node-0",
                                                        "status": "ok"
                                                    },
                                                    {
                                                        "time": "2025-02-01T02:55:00+00:00",
                                                        "hostname": "example-node-0",
                                                        "status": "ok"
                                                    },
                                                    {
                                                        "time": "2025-02-01T02:50:00+00:00",
                                                        "hostname": "example-node-0",
                                                        "status": "ng",
                                                        "error": {
                                                            "type": "service down",
                                                            "reason": "1 node down",
                                                            "nodes": [
                                                                "example-node-0"
                                                            ],
                                                            "description": "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
                                                            "details": "{ ... best effort error summary / direction ...}",
                                                            "log": "http://datacenter1:8888/log/nova/example-data-center-20250205113459-b3gc.log"
                                                        }
                                                    },
                                                    {
                                                        "time": "2025-02-01T02:45:00+00:00",
                                                        "hostname": "example-node-0",
                                                        "status": "ok"
                                                    }
                                                ],
                                                "status": {
                                                    "current": "ok",
                                                    "isFixing": false
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
                                            "example": "invalid serviceType: example-service"
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
            },
            "patch": {
                "operationId": "repairModuleHealth",
                "tags": [
                    "Health"
                ],
                "summary": "Repair the unhealthy module",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "serviceType",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "clusterLink",
                                "clusterSys",
                                "clusterSettings",
                                "haCluster",
                                "msgQueue",
                                "iaasDb",
                                "virtualIp",
                                "storage",
                                "apiService",
                                "singleSignOn",
                                "network",
                                "compute",
                                "bareMetal",
                                "image",
                                "blockStor",
                                "fileStor",
                                "objectStor",
                                "orchestration",
                                "lbaas",
                                "dnsaas",
                                "k8saas",
                                "instanceHa",
                                "businessLogic",
                                "dataPipe",
                                "metrics",
                                "logAnalytics",
                                "notifications"
                            ]
                        },
                        "description": "The name of the service which cover the module you tend to repair. use GET /api/v1/datacenters/{dataCenter}/services to get the service list and their modules."
                    },
                    {
                        "in": "path",
                        "name": "moduleType",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "link",
                                "clock",
                                "dns",
                                "bootstrap",
                                "license",
                                "etcd",
                                "nodelist",
                                "hacluster",
                                "rabbitmq",
                                "mysql",
                                "mongodb",
                                "vip",
                                "haproxy_ha",
                                "ceph",
                                "ceph_mon",
                                "ceph_osd",
                                "ceph_rgw",
                                "ceph_mds",
                                "ceph_mgr",
                                "rbd_target",
                                "haproxy",
                                "httpd",
                                "skyline",
                                "api",
                                "memcache",
                                "k3s",
                                "keycloak",
                                "neutron",
                                "nova",
                                "cyborg",
                                "ironic",
                                "glance",
                                "cinder",
                                "manila",
                                "swift",
                                "heat",
                                "octavia",
                                "designate",
                                "rancher",
                                "masakari",
                                "zookeeper",
                                "kafka",
                                "monasca",
                                "telegraf",
                                "grafana",
                                "filebeat",
                                "auditbeat",
                                "logstash",
                                "opensearch",
                                "opensearch-dashboards",
                                "influxdb",
                                "kapacitor"
                            ]
                        },
                        "description": "The name of the module to repair. use GET /api/v1/datacenters/{dataCenter}/services/{serviceType} to get the module list."
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
                                            "example": "invalid serviceType: example-service"
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
                "operationId": "getIntegrations",
                "tags": [
                    "Integrations"
                ],
                "summary": "Retrieve the list of integrated applications",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
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
                                                    "url": "https://example-datat-center.host:9999/base/overview"
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
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/keyword"
                    },
                    {
                        "$ref": "#/components/parameters/products"
                    },
                    {
                        "$ref": "#/components/parameters/listLicenseStatuses"
                    },
                    {
                        "$ref": "#/components/parameters/types"
                    },
                    {
                        "$ref": "#/components/parameters/pageSize"
                    },
                    {
                        "$ref": "#/components/parameters/pageNum"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
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
                                        "summary": "License list",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "licenses": [
                                                    {
                                                        "name": "example-license",
                                                        "type": "trial",
                                                        "hosts": [
                                                            "example-node-0"
                                                        ],
                                                        "serial": "1H2ZLG2",
                                                        "product": {
                                                            "name": "CubeCOS",
                                                            "feature": "kubernetes"
                                                        },
                                                        "issue": {
                                                            "by": "Bigstack co., ltd.",
                                                            "to": "bigstack",
                                                            "hardware": "*",
                                                            "date": "2025-03-16T17:31:21+08:00"
                                                        },
                                                        "quantity": "some of string about cpu capacity",
                                                        "supportPlan": "FMA",
                                                        "expiry": {
                                                            "date": "2025-05-15T17:31:21+08:00",
                                                            "days": 56
                                                        },
                                                        "status": {
                                                            "current": "valid",
                                                            "isExpiring": false
                                                        }
                                                    }
                                                ],
                                                "page": {
                                                    "total": 1,
                                                    "number": 1,
                                                    "size": 1,
                                                    "totalItemCount": 1
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
            },
            "post": {
                "operationId": "importClusterLicense",
                "tags": [
                    "Licenses"
                ],
                "summary": "Import licenses for the cluster",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "multipart/form-data": {
                            "schema": {
                                "$ref": "#/components/schemas/PostLicenseRequest"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Import licenses successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PostLicenseResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "License import result",
                                        "value": {
                                            "code": 200,
                                            "msg": "import license successfully",
                                            "status": "ok"
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
                                            "example": "failed to update licenses: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/licenses/verify": {
            "post": {
                "operationId": "verifyLicense",
                "tags": [
                    "Licenses"
                ],
                "summary": "Verify the license",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "multipart/form-data": {
                            "schema": {
                                "$ref": "#/components/schemas/PostLicenseRequest"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Verify license successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/VerifyLicenseResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "License verification result",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "license": {
                                                    "name": "example-license",
                                                    "type": "trial",
                                                    "product": {
                                                        "name": "CubeCOS",
                                                        "feature": "virtualization"
                                                    },
                                                    "issue": {
                                                        "by": "Bigstack Ltd.",
                                                        "to": "*",
                                                        "hardware": "*",
                                                        "date": "2025-04-09T19:16:13+08:00"
                                                    },
                                                    "quantity": "some of string about cpu capacity",
                                                    "supportPlan": "FMA",
                                                    "expiry": {
                                                        "date": "2025-05-09T19:16:13+08:00",
                                                        "days": 29
                                                    },
                                                    "status": {
                                                        "current": "valid",
                                                        "isExpiring": false
                                                    }
                                                },
                                                "effectNodes": [
                                                    {
                                                        "name": "example-node-1",
                                                        "role": "control-converged",
                                                        "expiry": {
                                                            "date": "2025-05-09T19:16:13+08:00",
                                                            "days": 30
                                                        },
                                                        "status": {
                                                            "current": "valid",
                                                            "isExpiring": false
                                                        }
                                                    },
                                                    {
                                                        "name": "example-node-2",
                                                        "role": "control-converged",
                                                        "expiry": {
                                                            "date": "2025-05-09T19:16:13+08:00",
                                                            "days": 30
                                                        },
                                                        "status": {
                                                            "current": "valid",
                                                            "isExpiring": false
                                                        }
                                                    },
                                                    {
                                                        "name": "example-node-3",
                                                        "role": "control-converged",
                                                        "expiry": {
                                                            "date": "2025-05-09T19:16:13+08:00",
                                                            "days": 30
                                                        },
                                                        "status": {
                                                            "current": "valid",
                                                            "isExpiring": false
                                                        }
                                                    }
                                                ]
                                            },
                                            "msg": "verify license successfully",
                                            "status": "ok"
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
                                            "example": "failed to verify license: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/licenses/hosts/{hostname}": {
            "post": {
                "operationId": "importNodeLicense",
                "tags": [
                    "Licenses"
                ],
                "summary": "Update licenses for specific node",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/hostname"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "multipart/form-data": {
                            "schema": {
                                "$ref": "#/components/schemas/PostLicenseRequest"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Update licenses successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PostLicenseResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "License import result",
                                        "value": {
                                            "code": 200,
                                            "msg": "import license successfully",
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
                                            "example": "failed to update licenses: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/licenses/attachments": {
            "get": {
                "operationId": "getLicenseAttachments",
                "tags": [
                    "Licenses"
                ],
                "summary": "Retrieve the list of license attachments",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/product"
                    },
                    {
                        "$ref": "#/components/parameters/keyword"
                    },
                    {
                        "$ref": "#/components/parameters/roles"
                    },
                    {
                        "$ref": "#/components/parameters/listLicenseAttachmentStatuses"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of license attachments successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetLicenseAttachmentsResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "License attachment list",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "serialNumber": "To Be Filled By O.E.M.",
                                                    "hostname": "example-node-0",
                                                    "role": "control-converged",
                                                    "product": "CubeCOS",
                                                    "status": "valid"
                                                },
                                                {
                                                    "serialNumber": "To Be Filled By O.E.M.",
                                                    "hostname": "example-node-1",
                                                    "role": "control-converged",
                                                    "product": "CubeCOS",
                                                    "status": "valid"
                                                },
                                                {
                                                    "serialNumber": "To Be Filled By O.E.M.",
                                                    "hostname": "example-node-2",
                                                    "role": "control-converged",
                                                    "product": "CubeCOS",
                                                    "status": "unlicense"
                                                }
                                            ],
                                            "msg": "fetch license attachments successfully",
                                            "status": "ok"
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
                                            "example": "failed to fetch license attachments: internal server error"
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
                "operationId": "getMetricsOverview",
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the metrics of data center",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the summary of data center successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetMetricsResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Overview of data center metrics",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "dataCenter": {
                                                    "usage": {
                                                        "cpu": {
                                                            "totalCores": 96,
                                                            "usedCores": 5.3846,
                                                            "usedPercent": 6.6012,
                                                            "freeCores": 90.6152,
                                                            "freePercent": 93.3986
                                                        },
                                                        "memory": {
                                                            "totalMiB": 515645.789,
                                                            "usedMiB": 269875.7773,
                                                            "usedPercent": 52.3374,
                                                            "freeMiB": 245770.0117,
                                                            "freePercent": 47.6625
                                                        }
                                                    }
                                                },
                                                "host": {
                                                    "role": {
                                                        "controlConverged": {
                                                            "count": 2,
                                                            "cpu": {
                                                                "totalCores": 96,
                                                                "usedCores": 5.3846,
                                                                "usedPercent": 6.6012,
                                                                "freeCores": 90.6152,
                                                                "freePercent": 93.3986
                                                            },
                                                            "memory": {
                                                                "totalMiB": 515645.789,
                                                                "usedMiB": 269875.7773,
                                                                "usedPercent": 52.3374,
                                                                "freeMiB": 245770.0117,
                                                                "freePercent": 47.6625
                                                            }
                                                        },
                                                        "control": {
                                                            "count": 0,
                                                            "cpu": {
                                                                "totalCores": 0,
                                                                "usedCores": 0,
                                                                "usedPercent": 0,
                                                                "freeCores": 0,
                                                                "freePercent": 0
                                                            },
                                                            "memory": {
                                                                "totalMiB": 0,
                                                                "usedMiB": 0,
                                                                "usedPercent": 0,
                                                                "freeMiB": 0,
                                                                "freePercent": 0
                                                            }
                                                        },
                                                        "compute": {
                                                            "count": 0,
                                                            "cpu": {
                                                                "totalCores": 0,
                                                                "usedCores": 0,
                                                                "usedPercent": 0,
                                                                "freeCores": 0,
                                                                "freePercent": 0
                                                            },
                                                            "memory": {
                                                                "totalMiB": 0,
                                                                "usedMiB": 0,
                                                                "usedPercent": 0,
                                                                "freeMiB": 0,
                                                                "freePercent": 0
                                                            }
                                                        },
                                                        "storage": {
                                                            "count": 0,
                                                            "cpu": {
                                                                "totalCores": 0,
                                                                "usedCores": 0,
                                                                "usedPercent": 0,
                                                                "freeCores": 0,
                                                                "freePercent": 0
                                                            },
                                                            "memory": {
                                                                "totalMiB": 0,
                                                                "usedMiB": 0,
                                                                "usedPercent": 0,
                                                                "freeMiB": 0,
                                                                "freePercent": 0
                                                            }
                                                        },
                                                        "edgeCore": {
                                                            "count": 0,
                                                            "cpu": {
                                                                "totalCores": 0,
                                                                "usedCores": 0,
                                                                "usedPercent": 0,
                                                                "freeCores": 0,
                                                                "freePercent": 0
                                                            },
                                                            "memory": {
                                                                "totalMiB": 0,
                                                                "usedMiB": 0,
                                                                "usedPercent": 0,
                                                                "freeMiB": 0,
                                                                "freePercent": 0
                                                            }
                                                        },
                                                        "moderator": {
                                                            "count": 0,
                                                            "cpu": {
                                                                "totalCores": 0,
                                                                "usedCores": 0,
                                                                "usedPercent": 0,
                                                                "freeCores": 0,
                                                                "freePercent": 0
                                                            },
                                                            "memory": {
                                                                "totalMiB": 0,
                                                                "usedMiB": 0,
                                                                "usedPercent": 0,
                                                                "freeMiB": 0,
                                                                "freePercent": 0
                                                            }
                                                        }
                                                    },
                                                    "usages": [
                                                        {
                                                            "role": "control-converged",
                                                            "name": "example-node-0",
                                                            "address": "10.10.10.10",
                                                            "cpu": {
                                                                "totalCores": 48,
                                                                "usedCores": 2.7194,
                                                                "usedPercent": 5.3006,
                                                                "freeCores": 45.2805,
                                                                "freePercent": 94.6993
                                                            },
                                                            "memory": {
                                                                "totalMiB": 257822.8945,
                                                                "usedMiB": 134933.4531,
                                                                "usedPercent": 52.3357,
                                                                "freeMiB": 122889.4414,
                                                                "freePercent": 47.6642
                                                            }
                                                        },
                                                        {
                                                            "role": "control-converged",
                                                            "name": "example-node-1",
                                                            "address": "10.10.10.11",
                                                            "cpu": {
                                                                "totalCores": 48,
                                                                "usedCores": 2.6652,
                                                                "usedPercent": 7.9019,
                                                                "freeCores": 45.3347,
                                                                "freePercent": 92.098
                                                            },
                                                            "memory": {
                                                                "totalMiB": 257822.8945,
                                                                "usedMiB": 134942.3242,
                                                                "usedPercent": 52.3391,
                                                                "freeMiB": 122880.5703,
                                                                "freePercent": 47.6608
                                                            }
                                                        }
                                                    ]
                                                },
                                                "vm": {
                                                    "status": {
                                                        "total": 15,
                                                        "running": 15,
                                                        "stopped": 0,
                                                        "suspend": 0,
                                                        "paused": 0,
                                                        "error": 0
                                                    },
                                                    "usage": {
                                                        "vcpu": {
                                                            "totalCores": 48,
                                                            "usedCores": 40,
                                                            "usedPercent": 0.8333,
                                                            "freeCores": 8,
                                                            "freePercent": 0.1666
                                                        },
                                                        "memory": {
                                                            "totalMiB": 257822,
                                                            "usedMiB": 100520,
                                                            "usedPercent": 0.3898,
                                                            "freeMiB": 157302,
                                                            "freePercent": 0.6101
                                                        },
                                                        "storage": {
                                                            "totalMiB": 12571648,
                                                            "usedMiB": 1126400,
                                                            "usedPercent": 0.0895,
                                                            "freeMiB": 11445248,
                                                            "freePercent": 0.9104
                                                        }
                                                    }
                                                }
                                            },
                                            "msg": "fetch summary successfully",
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
        "/api/v1/datacenters/{dataCenter}/metrics/{metricType}/{viewType}/{entityType}": {
            "get": {
                "operationId": "getMetricByTypes",
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the various metrics with different view from hosts or vms",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/metricType"
                    },
                    {
                        "$ref": "#/components/parameters/viewType"
                    },
                    {
                        "$ref": "#/components/parameters/entityType"
                    },
                    {
                        "$ref": "#/components/parameters/start"
                    },
                    {
                        "$ref": "#/components/parameters/stop"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the various metrics with different view from hosts or vms",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "oneOf": [
                                        {
                                            "$ref": "#/components/schemas/GetCpuUsageSummaryOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetCpuUsageRankOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetMemoryUsageSummaryOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetMemoryUsageRankOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetDiskUsageRankOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetDiskBandwidthHistoryOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetDiskIopsHistoryOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetDiskLatencyHistoryOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetNetworkTrafficInRankOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetNetworkTrafficOutRankOfHostsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetCpuUsageRankOfVmsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetMemoryUsageSummaryOfVmsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetMemoryUsageRankOfVmsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetDiskReadIopsRankOfVmsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetDiskWriteIopsRankOfVmsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetNetworkTrafficInRankOfVmsResponse"
                                        },
                                        {
                                            "$ref": "#/components/schemas/GetNetworkTrafficOutRankOfVmsResponse"
                                        }
                                    ]
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Cpu usage summary of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "totalCores": 48,
                                                "usedCores": 4.7778,
                                                "usedPercent": 6.6988,
                                                "freeCores": 43.2221,
                                                "freePercent": 93.3011
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example2": {
                                        "summary": "Cpu usage rank of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "percentage",
                                                "rank": [
                                                    {
                                                        "id": "example-node-0",
                                                        "name": "example-node-0",
                                                        "value": 12.1752,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:19:00+00:00",
                                                                "value": 10.3202
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:20:00+00:00",
                                                                "value": 10.1465
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:21:00+00:00",
                                                                "value": 14.5854
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:22:00+00:00",
                                                                "value": 11.9773
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "example-node-1",
                                                        "name": "example-node-1",
                                                        "value": 8.8248,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:19:00+00:00",
                                                                "value": 9.3202
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:20:00+00:00",
                                                                "value": 9.1465
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:21:00+00:00",
                                                                "value": 7.5854
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:22:00+00:00",
                                                                "value": 6.9773
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example3": {
                                        "summary": "Memory usage summary of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "totalMiB": 257371,
                                                "usedMiB": 98255,
                                                "usedPercent": 38.1752,
                                                "freeMiB": 159116,
                                                "freePercent": 61.8248
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example4": {
                                        "summary": "Memory usage rank of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "percentage",
                                                "rank": [
                                                    {
                                                        "id": "example-node-0",
                                                        "name": "example-node-0",
                                                        "value": 56.4807,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:23:00+00:00",
                                                                "value": 56.4997
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:24:00+00:00",
                                                                "value": 56.4998
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:25:00+00:00",
                                                                "value": 56.4963
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:26:00+00:00",
                                                                "value": 56.4758
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "example-node-1",
                                                        "name": "example-node-1",
                                                        "value": 37.1752,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:23:00+00:00",
                                                                "value": 36.4997
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:24:00+00:00",
                                                                "value": 36.4998
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:25:00+00:00",
                                                                "value": 36.4963
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:26:00+00:00",
                                                                "value": 36.4758
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example5": {
                                        "summary": "Disk usage rank of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "percentage",
                                                "rank": [
                                                    {
                                                        "id": "example-node-0",
                                                        "name": "example-node-0",
                                                        "value": 25.3645,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:26:00+00:00",
                                                                "value": 25.7166
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:27:00+00:00",
                                                                "value": 25.7188
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:28:00+00:00",
                                                                "value": 25.7209
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:29:00+00:00",
                                                                "value": 25.7243
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "example-node-1",
                                                        "name": "example-node-1",
                                                        "value": 15.3645,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:26:00+00:00",
                                                                "value": 15.7166
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:27:00+00:00",
                                                                "value": 15.7188
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:28:00+00:00",
                                                                "value": 15.7209
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:29:00+00:00",
                                                                "value": 15.7243
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example6": {
                                        "summary": "Disk bandwidth history of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "bytes",
                                                "read": [
                                                    {
                                                        "time": "2025-02-08T22:58:00+00:00",
                                                        "value": 0.7333
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:59:00+00:00",
                                                        "value": 0
                                                    },
                                                    {
                                                        "time": "2025-02-08T23:00:00+00:00",
                                                        "value": 955.7333
                                                    },
                                                    {
                                                        "time": "2025-02-08T23:01:00+00:00",
                                                        "value": 0
                                                    },
                                                    {
                                                        "time": "2025-02-08T23:02:00+00:00",
                                                        "value": 1434
                                                    }
                                                ],
                                                "write": [
                                                    {
                                                        "time": "2025-02-08T22:51:00+00:00",
                                                        "value": 81578.6666
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:52:00+00:00",
                                                        "value": 87929.15
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:53:00+00:00",
                                                        "value": 110865.0666
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:54:00+00:00",
                                                        "value": 110796.8
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:55:00+00:00",
                                                        "value": 110933.3333
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example7": {
                                        "summary": "Disk iops history of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "ops",
                                                "read": [
                                                    {
                                                        "time": "2025-02-08T22:55:00+00:00",
                                                        "value": 1.05
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:56:00+00:00",
                                                        "value": 1.0833
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:57:00+00:00",
                                                        "value": 1
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:58:00+00:00",
                                                        "value": 1.0166
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:59:00+00:00",
                                                        "value": 1
                                                    }
                                                ],
                                                "write": [
                                                    {
                                                        "time": "2025-02-08T22:55:00+00:00",
                                                        "value": 6.8333
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:56:00+00:00",
                                                        "value": 7.8833
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:57:00+00:00",
                                                        "value": 6.85
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:58:00+00:00",
                                                        "value": 4.9333
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:59:00+00:00",
                                                        "value": 6.2166
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example8": {
                                        "summary": "Disk latency history of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "millisecond",
                                                "read": [
                                                    {
                                                        "time": "2025-02-08T22:58:00+00:00",
                                                        "value": 4999.5
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:59:00+00:00",
                                                        "value": 21365.5666
                                                    },
                                                    {
                                                        "time": "2025-02-08T23:00:00+00:00",
                                                        "value": 39951.6333
                                                    },
                                                    {
                                                        "time": "2025-02-08T23:01:00+00:00",
                                                        "value": 6255.1333
                                                    },
                                                    {
                                                        "time": "2025-02-08T23:02:00+00:00",
                                                        "value": 5838730.85
                                                    }
                                                ],
                                                "write": [
                                                    {
                                                        "time": "2025-02-08T22:58:00+00:00",
                                                        "value": 12624417.95
                                                    },
                                                    {
                                                        "time": "2025-02-08T22:59:00+00:00",
                                                        "value": 29347351.6333
                                                    },
                                                    {
                                                        "time": "2025-02-08T23:00:00+00:00",
                                                        "value": 5219905.3
                                                    },
                                                    {
                                                        "time": "2025-02-08T23:01:00+00:00",
                                                        "value": 9249751.6666
                                                    },
                                                    {
                                                        "time": "2025-02-08T23:02:00+00:00",
                                                        "value": 10661656.35
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example9": {
                                        "summary": "Network inbound traffic rank of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "packets",
                                                "rank": [
                                                    {
                                                        "id": "example-node-0",
                                                        "name": "example-node-0",
                                                        "value": 7466.5592,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:32:00+00:00",
                                                                "value": 50064.9333
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:33:00+00:00",
                                                                "value": 7141170.1333
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:34:00+00:00",
                                                                "value": 39644.1333
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:35:00+00:00",
                                                                "value": 45140
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "example-node-1",
                                                        "name": "example-node-1",
                                                        "value": 1126.5592,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:32:00+00:00",
                                                                "value": 1467.9333
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:33:00+00:00",
                                                                "value": 1321.1333
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:34:00+00:00",
                                                                "value": 1211.1333
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:35:00+00:00",
                                                                "value": 1131.5145
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example10": {
                                        "summary": "Network outbound traffic rank of hosts",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "packets",
                                                "rank": [
                                                    {
                                                        "id": "example-node-0",
                                                        "name": "example-node-0",
                                                        "value": 3595.9166,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:35:00+00:00",
                                                                "value": 13138
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:36:00+00:00",
                                                                "value": 41106.6666
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:37:00+00:00",
                                                                "value": 149872.1333
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:38:00+00:00",
                                                                "value": 12152.6666
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "example-node-1",
                                                        "name": "example-node-1",
                                                        "value": 1595.9166,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:35:00+00:00",
                                                                "value": 3138
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:36:00+00:00",
                                                                "value": 11106.6666
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:37:00+00:00",
                                                                "value": 149872.1333
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:38:00+00:00",
                                                                "value": 12152.6666
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example11": {
                                        "summary": "Cpu usage rank of vms",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "percentage",
                                                "rank": [
                                                    {
                                                        "id": "example-vm-0",
                                                        "name": "example-vm-0",
                                                        "value": 22,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:34:52+00:00",
                                                                "value": 21
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:35:22+00:00",
                                                                "value": 23
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:35:52+00:00",
                                                                "value": 20
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:36:22+00:00",
                                                                "value": 22
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "example-vm-1",
                                                        "name": "example-vm-1",
                                                        "value": 16,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:34:52+00:00",
                                                                "value": 20
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:35:22+00:00",
                                                                "value": 18
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:35:52+00:00",
                                                                "value": 20
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:36:23+00:00",
                                                                "value": 18
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example12": {
                                        "summary": "Memory usage summary of vms",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "totalMiB": 257371,
                                                "usedMiB": 98255,
                                                "usedPercent": 38.1752,
                                                "freeMiB": 159116,
                                                "freePercent": 61.8248
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example13": {
                                        "summary": "Memory usage rank of vms",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "percentage",
                                                "rank": [
                                                    {
                                                        "id": "example-vm-0",
                                                        "name": "example-vm-0",
                                                        "value": 60.6915,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:35:52+00:00",
                                                                "value": 61.6362
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:36:23+00:00",
                                                                "value": 61.4931
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:36:53+00:00",
                                                                "value": 60.9485
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:37:23+00:00",
                                                                "value": 60.3947
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "example-vm-1",
                                                        "name": "example-vm-1",
                                                        "value": 36.6008,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:35:53+00:00",
                                                                "value": 36.6073
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:36:23+00:00",
                                                                "value": 36.6105
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:36:53+00:00",
                                                                "value": 36.6105
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:37:23+00:00",
                                                                "value": 36.6101
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example14": {
                                        "summary": "Disk read iops rank of vms",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "ops",
                                                "rank": [
                                                    {
                                                        "id": "9d56d601-85f0-4bfe-992d-ba72ab174552",
                                                        "name": "example-vm-0",
                                                        "device": "sda",
                                                        "value": 1640.3,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:38:53+00:00",
                                                                "value": 1345.3
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:39:23+00:00",
                                                                "value": 1123.3
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:39:53+00:00",
                                                                "value": 1345.3
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "301c55e3-5133-4c2c-9dda-4c31c382918d",
                                                        "name": "example-vm-1",
                                                        "device": "sdc",
                                                        "value": 235.35,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:38:53+00:00",
                                                                "value": 124.3
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:39:23+00:00",
                                                                "value": 214.3
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:39:53+00:00",
                                                                "value": 324.3
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example15": {
                                        "summary": "Disk write iops rank of vms",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "ops",
                                                "rank": [
                                                    {
                                                        "id": "9d56d601-85f0-4bfe-992d-ba72ab174552",
                                                        "name": "example-vm-0",
                                                        "device": "sda",
                                                        "value": 184158.3022,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:43:53+00:00",
                                                                "value": 182514.0885
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:44:23+00:00",
                                                                "value": 165440.9962
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:44:53+00:00",
                                                                "value": 784010.6095
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "301c55e3-5133-4c2c-9dda-4c31c382918d",
                                                        "name": "example-vm-1",
                                                        "device": "sdc",
                                                        "value": 39589.3872,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:43:53+00:00",
                                                                "value": 46107.319
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:44:23+00:00",
                                                                "value": 45280.7243
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:44:53+00:00",
                                                                "value": 42949.9957
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example16": {
                                        "summary": "Network inbound traffic rank of vms",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "packets",
                                                "rank": [
                                                    {
                                                        "id": "9d56d601-85f0-4bfe-992d-ba72ab174552",
                                                        "name": "example-vm-0",
                                                        "device": "tap8c5cff7e-b9",
                                                        "value": 49756.9502,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:45:53+00:00",
                                                                "value": 370067.1842
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:46:23+00:00",
                                                                "value": 409181.4363
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:46:53+00:00",
                                                                "value": 711269.641
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:47:23+00:00",
                                                                "value": 398196.6271
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "301c55e3-5133-4c2c-9dda-4c31c382918d",
                                                        "name": "example-vm-1",
                                                        "device": "tapdf5be3d9-c6",
                                                        "value": 25784.0787,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:45:53+00:00",
                                                                "value": 184756.5743
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:46:23+00:00",
                                                                "value": 207790.9312
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:46:53+00:00",
                                                                "value": 270692.8834
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:47:23+00:00",
                                                                "value": 200824.0072
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example17": {
                                        "summary": "Network outbound traffic rank of vms",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "packets",
                                                "rank": [
                                                    {
                                                        "id": "9d56d601-85f0-4bfe-992d-ba72ab174552",
                                                        "name": "example-vm-0",
                                                        "device": "tap8c5cff7e-b9",
                                                        "value": 54213.0697,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:46:53+00:00",
                                                                "value": 1027891.3544
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:47:23+00:00",
                                                                "value": 462342.2831
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:47:53+00:00",
                                                                "value": 435980.7214
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:48:23+00:00",
                                                                "value": 447617.6397
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "id": "301c55e3-5133-4c2c-9dda-4c31c382918d",
                                                        "name": "example-vm-1",
                                                        "device": "tapdf5be3d9-c6",
                                                        "value": 13561.1536,
                                                        "history": [
                                                            {
                                                                "time": "2025-02-14T00:46:53+00:00",
                                                                "value": 128726.0911
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:47:23+00:00",
                                                                "value": 109373.8789
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:47:53+00:00",
                                                                "value": 109977.441
                                                            },
                                                            {
                                                                "time": "2025-02-14T00:48:23+00:00",
                                                                "value": 111826.9696
                                                            }
                                                        ]
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
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
                                            "example": "metricType should be cpuUsage, memoryUsage, diskUsage, diskBandwidth, diskIops, diskLatency, diskReadIops, diskWriteIops, networkTrafficIn, or networkTrafficOut"
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
        "/api/v1/datacenters/{dataCenter}/metrics/{metricType}/{viewType}/{entityType}/{entityId or Name}": {
            "get": {
                "operationId": "getMetricByHostOrVm",
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve the various metrics with different view from single host or single vm",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/metricType"
                    },
                    {
                        "$ref": "#/components/parameters/viewType"
                    },
                    {
                        "$ref": "#/components/parameters/entityType"
                    },
                    {
                        "in": "path",
                        "name": "entityId or Name",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The id or name of the entity to query",
                        "example": "example-node-0"
                    },
                    {
                        "in": "query",
                        "name": "past",
                        "required": false,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "1h",
                                "24h",
                                "7d",
                                "14d"
                            ]
                        },
                        "description": "The past time of the health history to query, click 'try it out' to see a few options, but can specify with the 's'(second), 'm'(minute), 'h'(hour), and 'd'(day) suffix for other time ranges.",
                        "example": "1d"
                    },
                    {
                        "$ref": "#/components/parameters/start"
                    },
                    {
                        "$ref": "#/components/parameters/stop"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the various metrics with different view from single host or single vm",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/HostMetricHistoryResponse"
                                },
                                "examples": {
                                    "example1": {
                                        "summary": "Cpu usage history of host",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "percentage",
                                                "history": [
                                                    {
                                                        "time": "2025-03-24T09:39:00+08:00",
                                                        "value": 10.9059
                                                    },
                                                    {
                                                        "time": "2025-03-24T09:40:00+08:00",
                                                        "value": 10.5089
                                                    },
                                                    {
                                                        "time": "2025-03-24T09:41:00+08:00",
                                                        "value": 15.3512
                                                    },
                                                    {
                                                        "time": "2025-03-24T09:42:00+08:00",
                                                        "value": 12.3685
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
                                            "status": "ok"
                                        }
                                    },
                                    "example2": {
                                        "summary": "Memory size history of host",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "unit": "sizeMiB",
                                                "history": [
                                                    {
                                                        "time": "2025-04-15T05:42:00+08:00",
                                                        "value": 25918.1601
                                                    },
                                                    {
                                                        "time": "2025-04-15T05:43:00+08:00",
                                                        "value": 26017.2656
                                                    },
                                                    {
                                                        "time": "2025-04-15T05:44:00+08:00",
                                                        "value": 25927.0429
                                                    },
                                                    {
                                                        "time": "2025-04-15T05:45:00+08:00",
                                                        "value": 25998.707
                                                    }
                                                ]
                                            },
                                            "msg": "fetch metrics successfully",
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
                                            "example": "metricType should be cpuUsage, memoryUsage, diskUsage, diskBandwidth, diskIops, diskLatency, diskReadIops, diskWriteIops, networkTrafficIn, or networkTrafficOut"
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
        "/api/v1/datacenters/{dataCenter}/nodes": {
            "get": {
                "operationId": "getNodes",
                "tags": [
                    "Nodes"
                ],
                "summary": "Retrieve the list of nodes",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/keyword"
                    },
                    {
                        "$ref": "#/components/parameters/roles"
                    },
                    {
                        "$ref": "#/components/parameters/nodeLicenseStatuses"
                    },
                    {
                        "$ref": "#/components/parameters/products"
                    },
                    {
                        "$ref": "#/components/parameters/pageSize"
                    },
                    {
                        "$ref": "#/components/parameters/pageNum"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
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
                                                        "serialNumber": "1H2ZLG2",
                                                        "dataCenter": "example-data-center",
                                                        "hostname": "example-node-0",
                                                        "role": "control-converged",
                                                        "address": "10.10.10.10:8001",
                                                        "ip": "10.10.10.10",
                                                        "managementIP": "10.10.10.10",
                                                        "storageIP": "10.10.10.10",
                                                        "isVirtualIpOwner": true,
                                                        "license": {
                                                            "name": "example-license",
                                                            "type": "trial",
                                                            "hostname": "example-node-0",
                                                            "serial": "1H2ZLG2",
                                                            "product": {
                                                                "name": "CubeCOS",
                                                                "feature": "virtualization"
                                                            },
                                                            "issue": {
                                                                "by": "Bigstack co., ltd.",
                                                                "to": "bigstack",
                                                                "hardware": "*",
                                                                "date": "2025-01-23T14:51:50+08:00"
                                                            },
                                                            "quantity": "some of string about cpu capacity",
                                                            "supportPlan": "FMA",
                                                            "expiry": {
                                                                "date": "2025-03-24T14:51:50+08:00",
                                                                "days": 9
                                                            },
                                                            "status": {
                                                                "current": "valid",
                                                                "isExpiring": true
                                                            }
                                                        },
                                                        "status": "up",
                                                        "cpuSpec": "Intel(R) Xeon(R) CPU E5-2650 v4 @ 2.20GHz",
                                                        "networkInterfaces": [
                                                            {
                                                                "interface": "eth0",
                                                                "label": "IF.1",
                                                                "busIdSlaves": "0000:01:00.0",
                                                                "driver": "tg3",
                                                                "state": "UP",
                                                                "speed": "1000F/1000F"
                                                            },
                                                            {
                                                                "interface": "eth1",
                                                                "label": "IF.2",
                                                                "busIdSlaves": "0000:01:00.1",
                                                                "driver": "tg3",
                                                                "state": "DOWN",
                                                                "speed": "NA/1000F"
                                                            },
                                                            {
                                                                "interface": "eth2",
                                                                "label": "IF.3",
                                                                "busIdSlaves": "0000:02:00.0",
                                                                "driver": "tg3",
                                                                "state": "DOWN",
                                                                "speed": "NA/1000F"
                                                            },
                                                            {
                                                                "interface": "eth3",
                                                                "label": "IF.4",
                                                                "busIdSlaves": "0000:02:00.1",
                                                                "driver": "tg3",
                                                                "state": "DOWN",
                                                                "speed": "NA/1000F"
                                                            }
                                                        ],
                                                        "ipmi": {
                                                            "isSupported": true,
                                                            "isConnected": false,
                                                            "ip": "10.32.10.45"
                                                        },
                                                        "vcpu": {
                                                            "totalCores": 48,
                                                            "usedCores": 40,
                                                            "usedPercent": 83.3333,
                                                            "freeCores": 8,
                                                            "freePercent": 16.6666
                                                        },
                                                        "memory": {
                                                            "totalMiB": 257822,
                                                            "usedMiB": 100520,
                                                            "usedPercent": 38.9881,
                                                            "freeMiB": 157302,
                                                            "freePercent": 61.0118
                                                        },
                                                        "storage": {
                                                            "totalMiB": 12571648,
                                                            "usedMiB": 1126400,
                                                            "usedPercent": 8.9598,
                                                            "freeMiB": 11445248,
                                                            "freePercent": 91.0401
                                                        },
                                                        "uptimeSeconds": 2461586.74,
                                                        "labels": {
                                                            "isGpuEnabled": "false"
                                                        }
                                                    }
                                                ],
                                                "page": {
                                                    "total": 1,
                                                    "number": 1,
                                                    "size": 1,
                                                    "totalItemCount": 1
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
        "/api/v1/datacenters/{dataCenter}/nodes/{nodeName}": {
            "get": {
                "operationId": "getNode",
                "tags": [
                    "Nodes"
                ],
                "summary": "Retrieve the node details",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    },
                    {
                        "name": "nodeName",
                        "in": "path",
                        "description": "The name of the node",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "example": "example-node-0"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the node details successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetNodeResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Node",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "id": "abc0005e",
                                                "serialNumber": "1H2ZLG2",
                                                "dataCenter": "example-data-center",
                                                "hostname": "example-node-0",
                                                "role": "control-converged",
                                                "address": "10.10.10.10:8001",
                                                "ip": "10.10.10.10",
                                                "managementIP": "10.10.10.10",
                                                "storageIP": "10.10.10.10",
                                                "isVirtualIpOwner": true,
                                                "license": {
                                                    "name": "example-license",
                                                    "type": "trial",
                                                    "hostname": "example-node-0",
                                                    "serial": "1H2ZLG2",
                                                    "product": {
                                                        "name": "CubeCOS",
                                                        "feature": "virtualization"
                                                    },
                                                    "issue": {
                                                        "by": "Bigstack co., ltd.",
                                                        "to": "bigstack",
                                                        "hardware": "*",
                                                        "date": "2025-01-23T14:51:50+08:00"
                                                    },
                                                    "quantity": "some of string about cpu capacity",
                                                    "supportPlan": "FMA",
                                                    "expiry": {
                                                        "date": "2025-03-24T14:51:50+08:00",
                                                        "days": 9
                                                    },
                                                    "status": {
                                                        "current": "valid",
                                                        "isExpiring": true
                                                    }
                                                },
                                                "status": "up",
                                                "cpuSpec": "Intel(R) Xeon(R) CPU E5-2650 v4 @ 2.20GHz",
                                                "networkInterfaces": [
                                                    {
                                                        "interface": "eth0",
                                                        "label": "IF.1",
                                                        "busIdSlaves": "0000:01:00.0",
                                                        "driver": "tg3",
                                                        "state": "UP",
                                                        "speed": "1000F/1000F"
                                                    },
                                                    {
                                                        "interface": "eth1",
                                                        "label": "IF.2",
                                                        "busIdSlaves": "0000:01:00.1",
                                                        "driver": "tg3",
                                                        "state": "DOWN",
                                                        "speed": "NA/1000F"
                                                    },
                                                    {
                                                        "interface": "eth2",
                                                        "label": "IF.3",
                                                        "busIdSlaves": "0000:02:00.0",
                                                        "driver": "tg3",
                                                        "state": "DOWN",
                                                        "speed": "NA/1000F"
                                                    },
                                                    {
                                                        "interface": "eth3",
                                                        "label": "IF.4",
                                                        "busIdSlaves": "0000:02:00.1",
                                                        "driver": "tg3",
                                                        "state": "DOWN",
                                                        "speed": "NA/1000F"
                                                    }
                                                ],
                                                "ipmi": {
                                                    "isSupported": true,
                                                    "isConnected": false,
                                                    "ip": "10.32.10.45"
                                                },
                                                "vcpu": {
                                                    "totalCores": 48,
                                                    "usedCores": 40,
                                                    "usedPercent": 83.3333,
                                                    "freeCores": 8,
                                                    "freePercent": 16.6666
                                                },
                                                "memory": {
                                                    "totalMiB": 257822,
                                                    "usedMiB": 100520,
                                                    "usedPercent": 38.9881,
                                                    "freeMiB": 157302,
                                                    "freePercent": 61.0118
                                                },
                                                "storage": {
                                                    "totalMiB": 12571648,
                                                    "usedMiB": 1126400,
                                                    "usedPercent": 8.9598,
                                                    "freeMiB": 11445248,
                                                    "freePercent": 91.0401
                                                },
                                                "uptimeSeconds": 2461586.74,
                                                "labels": {
                                                    "isGpuEnabled": "false"
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
                                            "example": "failed to fetch node: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/nodes/{nodeName}/ipmi": {
            "post": {
                "operationId": "setNodeIpmi",
                "tags": [
                    "Nodes"
                ],
                "summary": "Set the node IPMI setting",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/NodeIpmiSettingRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Set node IPMI setting",
                                    "value": {
                                        "ip": "10.10.10.10",
                                        "port": 623,
                                        "username": "admin",
                                        "password": "example-password"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Node IPMI setting set successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/SetNodeIpmiSettingResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to set node IPMI setting: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/nodes/{nodeName}/ipmi/disconnect": {
            "delete": {
                "operationId": "disconnectNodeIpmi",
                "tags": [
                    "Nodes"
                ],
                "summary": "Disconnect the node IPMI control",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Disconnect the node IPMI successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/DisconnectNodeIpmiResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Disconnect node IPMI",
                                        "value": {
                                            "code": 200,
                                            "msg": "node IPMI is disconnected successfully",
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
                                            "example": "failed to disconnect node IPMI: bad request"
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
                                            "example": "failed to disconnect node IPMI: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/nodes/{nodeName}/ipmi/verify": {
            "post": {
                "operationId": "verifyNodeIpmi",
                "tags": [
                    "Nodes"
                ],
                "summary": "Verify the node IPMI setting",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/NodeIpmiSettingRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Verify node IPMI setting",
                                    "value": {
                                        "ip": "10.10.10.10",
                                        "port": 623,
                                        "username": "admin",
                                        "password": "example-password"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Node IPMI setting verified successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/VerifyNodeIpmiResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Node IPMI setting",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "board": {
                                                    "manufacturingDate": "2025-01-01T01:00:00+00:00",
                                                    "manufacturer": "DELL",
                                                    "product": "PowerEdge R630",
                                                    "serial": "CN747516CK0286",
                                                    "partNumber": "02C2CPA04"
                                                },
                                                "product": {
                                                    "name": "PowerEdge R630",
                                                    "manufacturer": "DELL",
                                                    "serial": "1MXXZH2",
                                                    "version": "01"
                                                }
                                            },
                                            "msg": "node IPMI setting verified successfully",
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to verify node IPMI setting: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/nodes/{nodeName}/ipmi/{operation}": {
            "post": {
                "operationId": "operateNodeIpmi",
                "tags": [
                    "Nodes"
                ],
                "summary": "Operate the node by IPMI",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    },
                    {
                        "name": "operation",
                        "in": "path",
                        "description": "The operation to perform on the node IPMI",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "enum": [
                                "poweron",
                                "poweroff",
                                "powercycle"
                            ]
                        },
                        "example": "poweroff"
                    }
                ],
                "responses": {
                    "202": {
                        "description": "the ipmi oepration request received",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/OperateNodeIpmiResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "node(example-node-0) is not powered on, cannot power cycle"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "status conflict"
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
                                            "example": "failed to operate node: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/nodes/{nodeName}/devices": {
            "get": {
                "operationId": "listNodeDevices",
                "tags": [
                    "Nodes"
                ],
                "summary": "Retrieve the node devices",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the node devices successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/ListNodeDevicesResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Node Devices",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "serial": "57T0A05UF5YE",
                                                    "device": "sda",
                                                    "type": "HDD",
                                                    "class": "HDD",
                                                    "osd": {
                                                        "pgs": 646,
                                                        "reweight": 1,
                                                        "daemons": [
                                                            {
                                                                "id": "osd.0",
                                                                "usagePercent": 21.9012,
                                                                "status": {
                                                                    "current": "up",
                                                                    "isProcessing": false
                                                                }
                                                            },
                                                            {
                                                                "id": "osd.1",
                                                                "usagePercent": 17.4514,
                                                                "status": {
                                                                    "current": "up",
                                                                    "isProcessing": false
                                                                }
                                                            }
                                                        ]
                                                    },
                                                    "sizeMiB": 533008.5754,
                                                    "availability": "in-use",
                                                    "status": {
                                                        "current": "ok",
                                                        "isPromotable": true,
                                                        "isDemotable": false,
                                                        "isProcessing": false
                                                    }
                                                },
                                                {
                                                    "serial": "183222E1E59B",
                                                    "device": "sdb",
                                                    "type": "SSD",
                                                    "class": "",
                                                    "osd": {
                                                        "pgs": 0,
                                                        "reweight": 0,
                                                        "daemons": []
                                                    },
                                                    "sizeMiB": 426387.7868,
                                                    "availability": "system",
                                                    "status": {
                                                        "current": "ok",
                                                        "isPromotable": true,
                                                        "isDemotable": false,
                                                        "isProcessing": false
                                                    }
                                                },
                                                {
                                                    "serial": "57P0A0EWF5YE",
                                                    "device": "sdc",
                                                    "type": "HDD",
                                                    "class": "HDD",
                                                    "osd": {
                                                        "pgs": 634,
                                                        "reweight": 1,
                                                        "daemons": [
                                                            {
                                                                "id": "osd.2",
                                                                "usagePercent": 15.748,
                                                                "status": {
                                                                    "current": "up",
                                                                    "isProcessing": false
                                                                }
                                                            },
                                                            {
                                                                "id": "osd.3",
                                                                "usagePercent": 21.4815,
                                                                "status": {
                                                                    "current": "up",
                                                                    "isProcessing": false
                                                                }
                                                            }
                                                        ]
                                                    },
                                                    "sizeMiB": 533008.5754,
                                                    "availability": "in-use",
                                                    "status": {
                                                        "current": "ok",
                                                        "isPromotable": true,
                                                        "isDemotable": false,
                                                        "isProcessing": false
                                                    }
                                                }
                                            ],
                                            "msg": "fetch node devices successfully",
                                            "status": "ok"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node device not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node device not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to fetch setting: internal server error"
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
            "post": {
                "operationId": "addNodeDevice",
                "tags": [
                    "Nodes"
                ],
                "summary": "Add a device to the node",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/AddNodeDeviceRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Add Node Device",
                                    "value": {
                                        "device": "sdb"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "the request to add node device is accepted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/AddNodeDeviceResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to add device: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/nodes/{nodeName}/devices/{deviceName}": {
            "delete": {
                "operationId": "removeNodeDevice",
                "tags": [
                    "Nodes"
                ],
                "summary": "Remove a device from the node",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    },
                    {
                        "$ref": "#/components/parameters/deviceName"
                    }
                ],
                "responses": {
                    "202": {
                        "description": "the request to remove node device is accepted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/RemoveNodeDeviceResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node or device not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node or device not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to remove device: internal server error"
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
                "operationId": "updateNodeDevice",
                "tags": [
                    "Nodes"
                ],
                "summary": "Update a device on the node",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    },
                    {
                        "$ref": "#/components/parameters/deviceName"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/UpdateNodeDeviceRequest"
                            },
                            "examples": {
                                "example1": {
                                    "summary": "Promote a device to SSD class",
                                    "value": {
                                        "class": "SSD"
                                    }
                                },
                                "example2": {
                                    "summary": "Demote a device to HDD class",
                                    "value": {
                                        "class": "HDD"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "the request to update node device is accepted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/UpdateNodeDeviceResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node or device not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node or device not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to update device: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/nodes/{nodeName}/osds/{osdId}": {
            "patch": {
                "operationId": "updateNodeOsd",
                "tags": [
                    "Nodes"
                ],
                "summary": "Update an OSD on the node",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    },
                    {
                        "$ref": "#/components/parameters/osdId"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/UpdateNodeOsdRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Reweight an OSD",
                                    "value": {
                                        "reweight": 0.75
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "the request to update node OSD is accepted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/UpdateNodeOsdResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node or OSD not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node or osd not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to update osd: internal server error"
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
            "delete": {
                "operationId": "deleteNodeOsd",
                "tags": [
                    "Nodes"
                ],
                "summary": "Delete an OSD on the node",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    },
                    {
                        "$ref": "#/components/parameters/osdId"
                    }
                ],
                "responses": {
                    "202": {
                        "description": "the request to delete node OSD is accepted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/DeleteNodeOsdResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node or OSD not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node or osd not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to delete osd: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/nodes/{nodeName}/osds/{osdId}/restart": {
            "post": {
                "operationId": "restartNodeOsd",
                "tags": [
                    "Nodes"
                ],
                "summary": "Restart an OSD on the node",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/nodeName"
                    },
                    {
                        "$ref": "#/components/parameters/osdId"
                    }
                ],
                "responses": {
                    "202": {
                        "description": "the request to restart node OSD is accepted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/RestartNodeOsdResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Node or OSD not found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "node or osd not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to restart osd: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/settings": {
            "get": {
                "operationId": "getSettings",
                "tags": [
                    "Settings"
                ],
                "summary": "Retrieve setting",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Setting retrieved successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetSettingResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Setting list",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "titlePrefix": {
                                                    "value": "example title prefix",
                                                    "status": {
                                                        "current": "ok",
                                                        "isUpdating": false
                                                    }
                                                },
                                                "email": {
                                                    "recipients": [
                                                        {
                                                            "address": "example.user.1@example.com",
                                                            "note": "example note 1",
                                                            "status": {
                                                                "current": "ok",
                                                                "isUpdating": false
                                                            }
                                                        },
                                                        {
                                                            "address": "example.user.2@example.com",
                                                            "note": "example note 2",
                                                            "status": {
                                                                "current": "updating",
                                                                "isUpdating": true
                                                            }
                                                        }
                                                    ],
                                                    "senders": [
                                                        {
                                                            "host": "email-smtp.example.mailserver.com",
                                                            "port": 587,
                                                            "username": "ABBBBBBMJM56DCCCCJR",
                                                            "from": "noreply@example.com",
                                                            "accessVerified": true,
                                                            "status": {
                                                                "current": "ok",
                                                                "isUpdating": false
                                                            }
                                                        }
                                                    ]
                                                },
                                                "slack": {
                                                    "channels": [
                                                        {
                                                            "name": "#example-alert-channel-1",
                                                            "url": "https://hooks.slack.com/services/example-token",
                                                            "description": "example alert channel 1",
                                                            "status": {
                                                                "current": "ok",
                                                                "isUpdating": false
                                                            }
                                                        },
                                                        {
                                                            "name": "#example-alert-channel-2",
                                                            "url": "https://hooks.slack.com/services/example-token",
                                                            "description": "example alert channel 2",
                                                            "status": {
                                                                "current": "ok",
                                                                "isUpdating": false
                                                            }
                                                        }
                                                    ]
                                                }
                                            },
                                            "msg": "all setting retrieved successfully",
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
                                            "example": "failed to fetch setting: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/settings/titlePrefix": {
            "put": {
                "operationId": "updateTitlePrefix",
                "tags": [
                    "Settings"
                ],
                "summary": "Update title prefix",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "type": "object",
                                "required": [
                                    "value"
                                ],
                                "properties": {
                                    "value": {
                                        "type": "string"
                                    }
                                }
                            },
                            "examples": {
                                "example": {
                                    "summary": "Title prefix",
                                    "value": {
                                        "value": "example title prefix"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Title prefix updated successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/UpdateTitlePrefixResponse"
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
                                            "example": "failed to update title prefix: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/settings/email/senders": {
            "post": {
                "operationId": "createEmailSender",
                "tags": [
                    "Settings"
                ],
                "summary": "Create an email sender",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/EmailSenderPostRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Email Sender",
                                    "value": {
                                        "host": "email-smtp.example.mailserver.com",
                                        "port": 587,
                                        "username": "example-user",
                                        "password": "example-password",
                                        "from": "noreply@example.com"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "201": {
                        "description": "Email sender created successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PostEmailSenderResponse"
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
                                            "example": "failed to fetch email sender: internal server error"
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
            "get": {
                "operationId": "getEmailSenders",
                "tags": [
                    "Settings"
                ],
                "summary": "Retrieve email senders",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Email senders retrieved successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetEmailSendersResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Email Senders List",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "host": "email-smtp.example.mailserver.com",
                                                    "port": 587,
                                                    "username": "example-user",
                                                    "from": "noreply@bigstack.co",
                                                    "accessVerified": true,
                                                    "status": {
                                                        "current": "ok",
                                                        "isUpdating": false
                                                    }
                                                }
                                            ],
                                            "msg": "email senders retrieved successfully",
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
                                            "example": "failed to fetch email sender: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/settings/email/senders/{senderHost}": {
            "post": {
                "operationId": "tryEmailSender",
                "tags": [
                    "Settings"
                ],
                "summary": "Try an email sender",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "senderHost",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The host of the email sender to operate",
                        "example": "email-smtp.example.mailserver.com"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/TryEmailSender"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Try Email Sender",
                                    "value": {
                                        "email": "example.user.1@example.com"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Email sender tried successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/TryEmailSenderResponse"
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
                                            "example": "failed to try email sender: internal server error"
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
                "operationId": "updateEmailSender",
                "tags": [
                    "Settings"
                ],
                "summary": "Update an email sender",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "senderHost",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The host of the email sender to operate",
                        "example": "email-smtp.example.mailserver.com"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/EmailSenderPatchRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Email Sender",
                                    "value": {
                                        "host": "email-smtp.example.mailserver.com",
                                        "port": 587,
                                        "username": "example-user",
                                        "password": "example-password",
                                        "from": "noreply@bigstack.co"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Email sender updated successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PutEmailSenderResponse"
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
                                            "example": "failed to fetch email sender: internal server error"
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
            "delete": {
                "operationId": "deleteEmailSender",
                "tags": [
                    "Settings"
                ],
                "summary": "Delete an email sender",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "senderHost",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The host of the email sender to operate",
                        "example": "email-smtp.example.mailserver.com"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Email sender deleted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/DeleteEmailSenderResponse"
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
                                            "example": "failed to fetch email sender: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/settings/email/recipients": {
            "post": {
                "operationId": "createEmailRecipient",
                "tags": [
                    "Settings"
                ],
                "summary": "Create an email recipient",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/EmailRecipientPostRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Email Recipient",
                                    "value": {
                                        "address": "example.user.1@example.com",
                                        "note": "example email recipient"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "201": {
                        "description": "Email recipient created successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PostEmailRecipientResponse"
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
                                            "example": "failed to create email recipients: internal server error"
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
            "get": {
                "operationId": "getEmailRecipients",
                "tags": [
                    "Settings"
                ],
                "summary": "Retrieve email recipients",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Email recipients retrieved successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetEmailRecipientsResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Email recipients list",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "address": "example.user.1@example.com",
                                                    "note": "example email recipient 1",
                                                    "status": {
                                                        "current": "ok",
                                                        "isUpdating": false
                                                    }
                                                },
                                                {
                                                    "address": "example.user.2@example.com",
                                                    "note": "example email recipient 2",
                                                    "status": {
                                                        "current": "ok",
                                                        "isUpdating": false
                                                    }
                                                }
                                            ],
                                            "msg": "email recipients retrieved successfully",
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
                                            "example": "failed to fetch email recipients: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/settings/email/recipients/{recipientEmail}": {
            "post": {
                "operationId": "tryEmailRecipient",
                "tags": [
                    "Settings"
                ],
                "summary": "Try an email recipient",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "recipientEmail",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "recipient email to operate"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Email recipient tried successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/TryEmailRecipientResponse"
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
                                            "example": "recipient not found"
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
                                            "example": "failed to try email recipient: internal server error"
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
            "put": {
                "operationId": "updateEmailRecipient",
                "tags": [
                    "Settings"
                ],
                "summary": "Update an email recipient",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "recipientEmail",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "recipient email to update"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/EmailRecipientPutRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Email Recipient",
                                    "value": {
                                        "address": "example.user.1@example.com",
                                        "note": "Updated recipients"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Email recipient updated successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PutEmailRecipientResponse"
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
                                            "example": "failed to update email recipient: internal server error"
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
            "delete": {
                "operationId": "deleteEmailRecipient",
                "tags": [
                    "Settings"
                ],
                "summary": "Delete an email recipient",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "recipientEmail",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "recipient email to delete"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Email recipient deleted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/DeleteEmailRecipientResponse"
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
                                            "example": "failed to delete email recipient: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/settings/slack/channels": {
            "post": {
                "operationId": "createSlackChannel",
                "tags": [
                    "Settings"
                ],
                "summary": "Create a slack channel",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/SlackChannelPostRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Slack Channel",
                                    "value": {
                                        "name": "#example-alert-channel-1",
                                        "url": "https://hooks.slack.com/services/example-token",
                                        "description": "example alert channel 1"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "201": {
                        "description": "Slack channel created successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PostSlackChannelResponse"
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
                                            "example": "failed to create slack channel: internal server error"
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
            "get": {
                "operationId": "getSlackChannels",
                "tags": [
                    "Settings"
                ],
                "summary": "Retrieve slack channels",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Slack channels retrieved successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetSlackChannelsResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Slack channel List",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "slackChannels": [
                                                    {
                                                        "name": "#example-alert-channel-1",
                                                        "url": "https://hooks.slack.com/services/example-token",
                                                        "description": "example alert channel 1",
                                                        "status": {
                                                            "current": "ok",
                                                            "isUpdating": false
                                                        }
                                                    },
                                                    {
                                                        "name": "#example-alert-channel-2",
                                                        "url": "https://hooks.slack.com/services/example-token",
                                                        "description": "example alert channel 2",
                                                        "status": {
                                                            "current": "ok",
                                                            "isUpdating": false
                                                        }
                                                    }
                                                ]
                                            },
                                            "msg": "slack channels retrieved successfully",
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
                                            "example": "failed to update slack channel: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/settings/slack/channels/{channelName}": {
            "post": {
                "operationId": "trySlackChannel",
                "tags": [
                    "Settings"
                ],
                "summary": "Try a slack channel",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "channelName",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "channel name to update"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Slack channel tried successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/TrySlackChannelResponse"
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
                                            "example": "failed to try slack channel: internal server error"
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
            "put": {
                "operationId": "updateSlackChannel",
                "tags": [
                    "Settings"
                ],
                "summary": "Update a slack channel",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "channelName",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "channel name to update"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/SlackChannelPutRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Slack channel",
                                    "value": {
                                        "name": "#example-alert-channel-1",
                                        "url": "https://hooks.slack.com/services/example-token",
                                        "description": "example alert channel 1"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Slack channel updated successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/PutSlackChannelResponse"
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
                                            "example": "failed to update slack channel: internal server error"
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
            "delete": {
                "operationId": "deleteSlackChannel",
                "tags": [
                    "Settings"
                ],
                "summary": "Delete a slack channel",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "channelName",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "channel name to delete"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Slack channel deleted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/DeleteSlackChannelResponse"
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
                                            "example": "failed to delete slack channel: internal server error"
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
                        "$ref": "#/components/parameters/dataCenter"
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
                                                "token": "example-token",
                                                "refresh": "example-refresh-token",
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
        },
        "/api/v1/datacenters/{dataCenter}/tunings/parameters": {
            "get": {
                "operationId": "listTunings",
                "tags": [
                    "Tunings"
                ],
                "summary": "Retrieve the list of tunings from a host or data center",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "query",
                        "name": "host",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        },
                        "description": "The name of the host to retrieve the tunings, can specify multiple hosts to retrieve the tunings, for example: host=example-node-0&host=example-node-1",
                        "example": "example-node-0"
                    },
                    {
                        "$ref": "#/components/parameters/keyword"
                    },
                    {
                        "$ref": "#/components/parameters/pageSize"
                    },
                    {
                        "$ref": "#/components/parameters/pageNum"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    },
                    {
                        "in": "query",
                        "name": "modified",
                        "required": false,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "boolean"
                            }
                        },
                        "description": "The flag to filter the modified or unmodified tunings",
                        "example": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of tunings from a host or data center successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/ListTuningResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Tuning list",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "tunings": [
                                                    {
                                                        "name": "neutron.debug.enabled",
                                                        "value": false,
                                                        "hosts": [
                                                            {
                                                                "name": "example-node-0",
                                                                "role": "control-converged",
                                                                "ip": "10.10.10.10"
                                                            }
                                                        ],
                                                        "description": "Set to true to enable neutron verbose log.",
                                                        "enabled": true,
                                                        "isModified": false,
                                                        "limitation": {
                                                            "type": "bool",
                                                            "default": false
                                                        },
                                                        "status": {
                                                            "current": "updating",
                                                            "updatedAt": "2025-03-12T19:22:00+08:00",
                                                            "isUpdating": true
                                                        }
                                                    },
                                                    {
                                                        "name": "cubesys.provider.extra",
                                                        "value": "",
                                                        "hosts": [
                                                            {
                                                                "name": "example-node-0",
                                                                "role": "control-converged",
                                                                "ip": "10.10.10.10"
                                                            }
                                                        ],
                                                        "description": "Set extra provider interfaces ('pvd-' prefix and <= 15 chars) [IF.2:pvd-xxx,eth2:pvd-yyy,...].",
                                                        "enabled": true,
                                                        "isModified": false,
                                                        "limitation": {
                                                            "type": "str",
                                                            "default": "",
                                                            "regex": ""
                                                        },
                                                        "status": {
                                                            "current": "ok",
                                                            "updatedAt": "2025-03-12T19:22:00+08:00",
                                                            "isUpdating": false
                                                        }
                                                    },
                                                    {
                                                        "name": "barbican.debug.enabled",
                                                        "value": false,
                                                        "hosts": [
                                                            {
                                                                "name": "example-node-0",
                                                                "role": "control-converged",
                                                                "ip": "10.10.10.10"
                                                            }
                                                        ],
                                                        "description": "Set to true to enable barbican verbose log.",
                                                        "enabled": true,
                                                        "isModified": false,
                                                        "limitation": {
                                                            "type": "bool",
                                                            "default": false
                                                        },
                                                        "status": {
                                                            "current": "ok",
                                                            "updatedAt": "2025-03-12T19:22:00+08:00",
                                                            "isUpdating": false
                                                        }
                                                    },
                                                    {
                                                        "name": "cinder.backup.endpoint",
                                                        "value": "",
                                                        "hosts": [
                                                            {
                                                                "name": "example-node-0",
                                                                "role": "control-converged",
                                                                "ip": "10.10.10.10"
                                                            }
                                                        ],
                                                        "description": "Set cinder backup storage endpoint.",
                                                        "enabled": true,
                                                        "isModified": false,
                                                        "limitation": {
                                                            "type": "str",
                                                            "default": "",
                                                            "regex": ""
                                                        },
                                                        "status": {
                                                            "current": "ok",
                                                            "updatedAt": "2025-03-12T19:22:00+08:00",
                                                            "isUpdating": false
                                                        }
                                                    },
                                                    {
                                                        "name": "influxdb.curator.rp",
                                                        "value": 7,
                                                        "hosts": [
                                                            {
                                                                "name": "example-node-0",
                                                                "role": "control-converged",
                                                                "ip": "10.10.10.10"
                                                            },
                                                            {
                                                                "name": "example-node-1",
                                                                "role": "control-converged",
                                                                "ip": "10.10.10.11"
                                                            }
                                                        ],
                                                        "description": "influxdb curator retention policy in days.",
                                                        "enabled": true,
                                                        "isModified": false,
                                                        "limitation": {
                                                            "type": "int",
                                                            "default": 7,
                                                            "min": 0,
                                                            "max": 365
                                                        },
                                                        "status": {
                                                            "current": "ok",
                                                            "updatedAt": "2025-03-12T19:22:00+08:00",
                                                            "isUpdating": false
                                                        }
                                                    },
                                                    {
                                                        "name": "influxdb.curator.rp",
                                                        "value": 23,
                                                        "hosts": [
                                                            {
                                                                "name": "example-node-0",
                                                                "role": "control-converged",
                                                                "ip": "10.10.10.10"
                                                            },
                                                            {
                                                                "name": "example-node-1",
                                                                "role": "control-converged",
                                                                "ip": "10.10.10.11"
                                                            }
                                                        ],
                                                        "description": "influxdb curator retention policy in days.",
                                                        "enabled": false,
                                                        "isModified": true,
                                                        "limitation": {
                                                            "type": "int",
                                                            "default": 7,
                                                            "min": 0,
                                                            "max": 365
                                                        },
                                                        "status": {
                                                            "current": "ok",
                                                            "updatedAt": "2025-03-12T19:22:00+08:00",
                                                            "isUpdating": false
                                                        }
                                                    }
                                                ],
                                                "page": {
                                                    "total": 15,
                                                    "number": 1,
                                                    "size": 5,
                                                    "totalItemCount:": 75
                                                }
                                            },
                                            "msg": "fetch tuning list successfully",
                                            "status": "ok"
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
                                            "example": "failed to list tuning: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/tunings/specs": {
            "get": {
                "operationId": "listTuningSpecs",
                "tags": [
                    "Tunings"
                ],
                "summary": "Retrieve the list of tuning spec",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve the list of tuning spec successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/ListTuningSpecResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Tuning spec list",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "name": "barbican.debug.enabled",
                                                    "description": "Set to true to enable barbican verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "ceph.debug.enabled",
                                                    "description": "Set to true to enable ceph debug logs.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "ceph.mirror.meta.sync",
                                                    "description": "Set to true to enable automatically volume metadata sync.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": true
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.backup.account",
                                                    "description": "Set cinder backup storage account.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.backup.endpoint",
                                                    "description": "Set cinder backup storage endpoint.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.backup.override",
                                                    "description": "Enable override cinder backup configurations.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.backup.pool",
                                                    "description": "Set cinder backup storage pool.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.backup.secret",
                                                    "description": "Set cinder backup storage account secret.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.backup.type",
                                                    "description": "Set cinder backup storage type <cube-storage|cube-swift>.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.debug.enabled",
                                                    "description": "Set to true to enable cinder verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.external.%d.account",
                                                    "description": "Set cinder external storage account.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.external.%d.driver",
                                                    "description": "Set cinder external storage type name <cube|purestorage>.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.external.%d.endpoint",
                                                    "description": "Set cinder external storage endpoint.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.external.%d.name",
                                                    "description": "Set cinder external storage rule name.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.external.%d.pool",
                                                    "description": "Set cinder external storage pool.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cinder.external.%d.secret",
                                                    "description": "Set cinder external storage account secret.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cubesys.alert.level",
                                                    "description": "Set health alert sensible level. (0: default, 1: highly sensitive)",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 0,
                                                        "min": 0,
                                                        "max": 2147483647
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cubesys.alert.level.%s",
                                                    "description": "Set health alert sensible level for service %s. (0: default, 1: highly sensitive)",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 0,
                                                        "min": 0,
                                                        "max": 2147483647
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cubesys.conntable.max",
                                                    "description": "Set max connection table size.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 262144,
                                                        "min": 0,
                                                        "max": 2147483647
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cubesys.log.default.retention",
                                                    "description": "Set log file retention policy in days.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 14,
                                                        "min": 0,
                                                        "max": 365
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cubesys.provider.extra",
                                                    "description": "Set extra provider interfaces ('pvd-' prefix and <= 15 chars) [IF.2:pvd-xxx,eth2:pvd-yyy,...].",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "cyborg.debug.enabled",
                                                    "description": "Set to true to enable cyborg verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "debug.enable_core_dump.%s",
                                                    "description": "Enable core dump for process %s",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "debug.enable_kdump",
                                                    "description": "Enable kdump to collect dump from kernel panic",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "debug.level.%s",
                                                    "description": "Set debug level for process %s",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 0,
                                                        "min": 0,
                                                        "max": 9
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "debug.max_core_dump",
                                                    "description": "Set the total number of core files before oldest are removed",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 0,
                                                        "min": 0,
                                                        "max": 999
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "designate.debug.enabled",
                                                    "description": "Set to true to enable designate verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "glance.debug.enabled",
                                                    "description": "Set to true to enable glance verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "glance.export.rp",
                                                    "description": "glance export retention policy in copies.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 3,
                                                        "min": 0,
                                                        "max": 255
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "heat.debug.enabled",
                                                    "description": "Set to true to enable heat verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "influxdb.curator.rp",
                                                    "description": "influxdb curator retention policy in days.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 7,
                                                        "min": 0,
                                                        "max": 365
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "ironic.debug.enabled",
                                                    "description": "Set to true to enable ironic verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "ironic.deploy.server",
                                                    "description": "Set to true to enable ironic deploy server (dhcp/tftp/pxe/http).",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "kapacitor.alert.check.enabled",
                                                    "description": "Set true to enable kapacitor alert check.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "kapacitor.alert.check.eventid",
                                                    "description": "Set kapacitor alert check eventid.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "SYS00002W"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "kapacitor.alert.check.interval",
                                                    "description": "Set kapacitor alert check interval (default to 60m).",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "60m"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "kapacitor.alert.extra.prefix",
                                                    "description": "Set kapacitor alert message prefix.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "Cube"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "kapacitor.alert.flow.base",
                                                    "description": "Set kapacitor alert base for abnormal flow.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "7d"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "kapacitor.alert.flow.threshold",
                                                    "description": "Set kapacitor alert threshold for abnormal flow.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 30,
                                                        "min": 0,
                                                        "max": 65535
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "kapacitor.alert.flow.unit",
                                                    "description": "Set kapacitor alert unit for abnormal flow.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "5m"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "keystone.debug.enabled",
                                                    "description": "Set to true to enable keystone verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "manila.debug.enabled",
                                                    "description": "Set to true to enable manila verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "manila.volume.type",
                                                    "description": "Set manila backend volume type.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "CubeStorage"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "masakari.host.evacuate_all",
                                                    "description": "Set to true to enable evacuate all instances when host goes down.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": true
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "masakari.wait.period",
                                                    "description": "Set wait period after service update",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 0,
                                                        "min": 0,
                                                        "max": 99999
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "monasca.debug.enabled",
                                                    "description": "Set to true to enable monasca verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "mysql.backup.curator.rp",
                                                    "description": "mysql backup retention policy in weeks.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 14,
                                                        "min": 0,
                                                        "max": 52
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "net.if.mtu.<name>",
                                                    "description": "Set interface MTU (MTU of parent interface must be greater than its VLAN interface).",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 1500,
                                                        "min": 68,
                                                        "max": 65536
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "net.ipv4.tcp_syncookies",
                                                    "description": "Turn on the Linux SYN cookies implementation.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": true
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "net.lacp.default.rate",
                                                    "description": "Set default LACP rate (fast/slow).",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "fast"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "net.lacp.default.xmit",
                                                    "description": "Set default LACP transmit hash policy (layer2/layer2+3/layer3+4).",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "layer3+4"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "neutron.debug.enabled",
                                                    "description": "Set to true to enable neutron verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "nova.control.host.memory",
                                                    "description": "Amount of memory in MB to reserve for the control host.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 0,
                                                        "min": 0,
                                                        "max": 524288
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "nova.control.host.vcpu",
                                                    "description": "Amount of vcpu to reserve for the control host.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 0,
                                                        "min": 0,
                                                        "max": 128
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "nova.debug.enabled",
                                                    "description": "Set to true to enable nova verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "nova.gpu.type",
                                                    "description": "Specify a supported gpu type instances would get.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": ""
                                                    },
                                                    "roles": []
                                                },
                                                {
                                                    "name": "nova.overcommit.cpu.ratio",
                                                    "description": "Specify an allowed CPU overcommitted ratio.",
                                                    "limitation": {
                                                        "type": "float",
                                                        "default": 16
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "nova.overcommit.disk.ratio",
                                                    "description": "Specify an allowed disk overcommitted ratio.",
                                                    "limitation": {
                                                        "type": "float",
                                                        "default": 1
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "nova.overcommit.ram.ratio",
                                                    "description": "Specify an allowed RAM overcommitted ratio.",
                                                    "limitation": {
                                                        "type": "float",
                                                        "default": 1.5
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "ntp.debug.enabled",
                                                    "description": "Set to true to enable ntp verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "octavia.debug.enabled",
                                                    "description": "Set to true to enable octavia verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "octavia.ha",
                                                    "description": "Set to true to enable octavia HA mode.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "opensearch.curator.rp",
                                                    "description": "opensearch curator retention policy in days.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 7,
                                                        "min": 0,
                                                        "max": 365
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "opensearch.heap.size",
                                                    "description": "Set opensearch heap size in MB.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 1024,
                                                        "min": 256,
                                                        "max": 65536
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "senlin.debug.enabled",
                                                    "description": "Set to true to enable senlin verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "skyline.debug.enabled",
                                                    "description": "Set to true to enable skyline verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "snapshot.apply.action",
                                                    "description": "Set snapshot apply action <apply|revert>.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "apply"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "snapshot.apply.policy.ignore",
                                                    "description": "Set snapshot apply policy ignore <true|false>.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "sshd.bind_to_all_interfaces",
                                                    "description": "Set to true to bind sshd to all interfaces.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "sshd.session.inactivity",
                                                    "description": "Set sshd session inactivity timeout in seconds.",
                                                    "limitation": {
                                                        "type": "int",
                                                        "default": 0,
                                                        "min": 0,
                                                        "max": 86400
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "time.timezone",
                                                    "description": "Set system timezone.",
                                                    "limitation": {
                                                        "type": "str",
                                                        "default": "UTC"
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "update.security.autoupdate",
                                                    "description": "Set to true to enable security autoupdate.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "name": "watcher.debug.enabled",
                                                    "description": "Set to true to enable watcher verbose log.",
                                                    "limitation": {
                                                        "type": "bool",
                                                        "default": false
                                                    },
                                                    "roles": [
                                                        {
                                                            "name": "control-converged",
                                                            "hosts": [
                                                                {
                                                                    "name": "example-node-0",
                                                                    "ip": "10.10.10.10"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                }
                                            ],
                                            "msg": "fetch tuning specs successfully",
                                            "status": "ok"
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
                                            "example": "failed to list tuning specs: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/tunings/parameters/{parameterName}": {
            "patch": {
                "operationId": "updateTuning",
                "tags": [
                    "Tunings"
                ],
                "summary": "Update value for a specific tuning parameter",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "parameterName",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the parameter to update. use GET /api/v1/datacenters/{dataCenter}/tunings/specs to get the list of parameters",
                        "example": "barbican.debug.enabled"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/UpdateTuningRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Tuning update request",
                                    "value": {
                                        "value": true,
                                        "hosts": [
                                            "example-node-0"
                                        ]
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Update the tuning successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/UpdateTuningResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Tuning update request received",
                                        "value": {
                                            "code": 202,
                                            "msg": "tuning update request received",
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
                                            "example": "invalid parameter name"
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
                                            "example": "failed to request tuning update: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/tunings/parameters/{parameterName}/enable": {
            "patch": {
                "operationId": "enableOrDisableTuning",
                "tags": [
                    "Tunings"
                ],
                "summary": "Enable or disable a specific tuning parameter",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "parameterName",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the parameter to update. use GET /api/v1/datacenters/{dataCenter}/tunings/specs to get the list of parameters",
                        "example": "barbican.debug.enabled"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/EnableOrDisableTuningRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Tuning update request",
                                    "value": {
                                        "enable": true,
                                        "hosts": [
                                            "example-node-0"
                                        ]
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Enable or disable the tuning successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/UpdateTuningResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Tuning update request received",
                                        "value": {
                                            "code": 202,
                                            "msg": "tuning update request received",
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
                                            "example": "invalid parameter name"
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
                                            "example": "failed to request tuning update: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/tunings/parameters/{parameterName}/reset": {
            "post": {
                "operationId": "resetTuning",
                "tags": [
                    "Tunings"
                ],
                "summary": "Reset a specific tuning parameter",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "parameterName",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The name of the parameter to reset. use GET /api/v1/datacenters/{dataCenter}/tunings/specs to get the list of parameters",
                        "example": "barbican.debug.enabled"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/ResetTuningRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Tuning reset request",
                                    "value": {
                                        "hosts": [
                                            "example-node-0"
                                        ]
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Reset the tuning successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/ResetTuningResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Tuning reset request received",
                                        "value": {
                                            "code": 202,
                                            "msg": "tuning reset request received",
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
                                            "example": "invalid parameter name"
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
                                            "example": "failed to request tuning update: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/triggers/materials": {
            "get": {
                "operationId": "getTriggerMaterials",
                "tags": [
                    "Triggers"
                ],
                "summary": "Retrieve all trigger materials",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve trigger materials successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetTriggerMaterialsResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Trigger materials list",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "attribute": {
                                                    "alertTypes": [
                                                        "system",
                                                        "host",
                                                        "instance"
                                                    ],
                                                    "severities": [
                                                        "INFO",
                                                        "WARNING",
                                                        "CTRITICAL"
                                                    ],
                                                    "categories": [
                                                        "KSN",
                                                        "CMP",
                                                        "ETH",
                                                        "CPU"
                                                    ],
                                                    "eventIds": [
                                                        "KSN00001I",
                                                        "CMP01001I",
                                                        "ETH00001I",
                                                        "CPU00002W",
                                                        "CPU00004I",
                                                        "CPU00006C"
                                                    ]
                                                },
                                                "response": {
                                                    "scriptType": {
                                                        "language": "Bash",
                                                        "environment": "Alpine Linux"
                                                    },
                                                    "emails": [
                                                        {
                                                            "address": "example-user@bigstack.co",
                                                            "note": "example email recipient"
                                                        }
                                                    ],
                                                    "slacks": [
                                                        {
                                                            "name": "Example slack channel",
                                                            "url": "https://example.slack.com/archives/exmpale-token-abc",
                                                            "description": "example slack channel"
                                                        }
                                                    ]
                                                }
                                            },
                                            "msg": "fetched trigger materials successfully",
                                            "status": "ok"
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
                                            "example": "failed to get trigger materials: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/triggers/materials/script/verify": {
            "post": {
                "operationId": "verifyTriggerScript",
                "tags": [
                    "Triggers"
                ],
                "summary": "Verify trigger script",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/VerifyMaterialScriptRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Trigger script verification request",
                                    "value": {
                                        "script": "base64-encoded-script-content"
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "Trigger script verified successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/VerifyMaterialScriptResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Trigger script verification result",
                                        "value": {
                                            "code": 200,
                                            "msg": "trigger script verified successfully",
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
                                            "example": "invalid script content"
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
                                            "example": "failed to verify trigger script: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/triggers": {
            "get": {
                "operationId": "getTriggers",
                "tags": [
                    "Triggers"
                ],
                "summary": "Retrieve all triggers",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/pageSize"
                    },
                    {
                        "$ref": "#/components/parameters/pageNum"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve triggers successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetTriggersResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Triggers list",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "triggers": [
                                                    {
                                                        "name": "Administrative Level Notification",
                                                        "isBuiltIn": true,
                                                        "description": "Configure how you are going to be notified for system events and host alerts, including levels 'warning', 'error', and 'critical'.",
                                                        "attribute": {
                                                            "alertTypes": [
                                                                "system",
                                                                "host",
                                                                "instance"
                                                            ],
                                                            "eventIds": [
                                                                "KSN00001I",
                                                                "ETH00001I",
                                                                "CPU00002W"
                                                            ],
                                                            "severities": [
                                                                "INFO",
                                                                "WARNING",
                                                                "CRITICAL"
                                                            ],
                                                            "categories": [
                                                                "DEV",
                                                                "CPU",
                                                                "DSK"
                                                            ]
                                                        },
                                                        "response": {
                                                            "types": [
                                                                "script",
                                                                "email",
                                                                "slack"
                                                            ],
                                                            "script": {
                                                                "name": "example-script",
                                                                "content": "IyEvYmluL2Jhc2gKCmN1cmwgaHR0cDovL2V4YW1wbGUuY29tICR7IHN0YXR1cyA9ICdzdWNjZXNzJzsgfQ=="
                                                            },
                                                            "slacks": [
                                                                {
                                                                    "name": "example-slack-channel-0",
                                                                    "url": "https://hooks.slack.com/services/<hookHash>/<hookHash>/<hookHash>",
                                                                    "description": "example slack channel 1"
                                                                },
                                                                {
                                                                    "name": "example-slack-channel-1",
                                                                    "url": "https://hooks.slack.com/services/<hookHash>/<hookHash>/<hookHash>",
                                                                    "description": "example slack channel 2"
                                                                }
                                                            ],
                                                            "emails": [
                                                                {
                                                                    "address": "example.user@example.com",
                                                                    "note": "example email recipient"
                                                                }
                                                            ]
                                                        },
                                                        "status": {
                                                            "current": "ok",
                                                            "isProcessing": false
                                                        },
                                                        "enabled": false
                                                    },
                                                    {
                                                        "name": "Instance Level Notification",
                                                        "isBuiltIn": true,
                                                        "description": "Configure how you are going to be notified for instance alerts, including levels 'warning', and 'critical'.",
                                                        "attribute": {
                                                            "alertTypes": [
                                                                "system",
                                                                "host",
                                                                "instance"
                                                            ],
                                                            "eventIds": [
                                                                "KSN00001I",
                                                                "ETH00001I",
                                                                "CPU00002W"
                                                            ],
                                                            "severities": [
                                                                "WARNING",
                                                                "Error",
                                                                "CRITICAL"
                                                            ],
                                                            "categories": [
                                                                "DEV",
                                                                "CPU",
                                                                "DSK"
                                                            ]
                                                        },
                                                        "response": {
                                                            "types": [
                                                                "script",
                                                                "email",
                                                                "slack"
                                                            ],
                                                            "script": {
                                                                "name": "example-script",
                                                                "content": "IyEvYmluL2Jhc2gKCmN1cmwgaHR0cDovL2V4YW1wbGUuY29tICR7IHN0YXR1cyA9ICdzdWNjZXNzJzsgfQ=="
                                                            },
                                                            "slacks": [
                                                                {
                                                                    "name": "example-slack-channel-0",
                                                                    "url": "https://hooks.slack.com/services/<hookHash>/<hookHash>/<hookHash>",
                                                                    "description": "example slack channel 1"
                                                                },
                                                                {
                                                                    "name": "example-slack-channel-1",
                                                                    "url": "https://hooks.slack.com/services/<hookHash>/<hookHash>/<hookHash>",
                                                                    "description": "example slack channel 2"
                                                                }
                                                            ],
                                                            "emails": [
                                                                {
                                                                    "address": "example.user@example.com",
                                                                    "note": "example email recipient"
                                                                }
                                                            ]
                                                        },
                                                        "status": {
                                                            "isProcessing": false
                                                        },
                                                        "enabled": false
                                                    }
                                                ],
                                                "page": {
                                                    "total": 1,
                                                    "number": 1,
                                                    "size": 2,
                                                    "totalItemCount": 2
                                                }
                                            },
                                            "msg": "fetch triggers successfully",
                                            "status": "ok"
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
                                            "example": "failed to list triggers: internal server error"
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
            "post": {
                "operationId": "createTrigger",
                "tags": [
                    "Triggers"
                ],
                "summary": "Create trigger",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/CreateTriggerRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Create trigger request",
                                    "value": {
                                        "name": "example-trigger",
                                        "description": "This is an example trigger",
                                        "attribute": {
                                            "alertTypes": [
                                                "system"
                                            ],
                                            "severities": [
                                                "CRITICAL"
                                            ],
                                            "categories": [
                                                "CPU"
                                            ],
                                            "eventIds": [
                                                "CPU00006C"
                                            ]
                                        },
                                        "response": {
                                            "script": {
                                                "name": "example-script",
                                                "content": "IyEvYmluL2Jhc2gKCmN1cmwgaHR0cDovL2V4YW1wbGUuY29tICR7IHN0YXR1cyA9ICdzdWNjZXNzJzsgfQ=="
                                            },
                                            "emails": [
                                                "example-user@bigstack.co"
                                            ],
                                            "slacks": [
                                                "https://hooks.slack.com/services/T00000000/B00000000/XXXXX"
                                            ]
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Trigger creation request received",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/CreateTriggerResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Trigger creation request received",
                                        "value": {
                                            "code": 202,
                                            "msg": "trigger creation request received",
                                            "status": "ok"
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
                                            "example": "failed to list triggers: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/triggers/{triggerName}": {
            "get": {
                "operationId": "getTrigger",
                "tags": [
                    "Triggers"
                ],
                "summary": "Retrieve a specific trigger",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/triggerName"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve trigger successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetTriggerResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Trigger",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "name": "Administrative Level Notification",
                                                "isBuiltIn": true,
                                                "description": "Configure how you are going to be notified for system events and host alerts, including levels 'warning', 'error', and 'critical'.",
                                                "attribute": {
                                                    "alertTypes": [
                                                        "system",
                                                        "host",
                                                        "instance"
                                                    ],
                                                    "eventIds": [
                                                        "KSN00001I",
                                                        "ETH00001I",
                                                        "CPU00002W"
                                                    ],
                                                    "severities": [
                                                        "WARNING",
                                                        "ERROR",
                                                        "CRITICAL"
                                                    ],
                                                    "categories": [
                                                        "DEV",
                                                        "CPU",
                                                        "DSK"
                                                    ]
                                                },
                                                "response": {
                                                    "types": [
                                                        "script",
                                                        "email",
                                                        "slack"
                                                    ],
                                                    "script": {
                                                        "name": "example-script",
                                                        "content": "IyEvYmluL2Jhc2gKCmN1cmwgaHR0cDovL2V4YW1wbGUuY29tICR7IHN0YXR1cyA9ICdzdWNjZXNzJzsgfQ=="
                                                    },
                                                    "slacks": [
                                                        {
                                                            "name": "example-slack-channel-0",
                                                            "url": "https://hooks.slack.com/services/<hookHash>/<hookHash>/<hookHash>",
                                                            "description": "example slack channel 1"
                                                        },
                                                        {
                                                            "name": "example-slack-channel-1",
                                                            "url": "https://hooks.slack.com/services/<hookHash>/<hookHash>/<hookHash>",
                                                            "description": "example slack channel 2"
                                                        }
                                                    ],
                                                    "emails": [
                                                        {
                                                            "address": "example.user@example.com",
                                                            "note": "example email recipient"
                                                        }
                                                    ]
                                                },
                                                "status": {
                                                    "current": "ok",
                                                    "isProcessing": false
                                                },
                                                "enabled": false
                                            },
                                            "msg": "fetch trigger successfully",
                                            "status": "ok"
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
                                            "example": "failed to list triggers: internal server error"
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
                "operationId": "updateTrigger",
                "tags": [
                    "Triggers"
                ],
                "summary": "Update a specific trigger",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/triggerName"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/UpdateTriggerRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Trigger update request",
                                    "value": {
                                        "attribute": {
                                            "alertTypes": [
                                                "system",
                                                "host"
                                            ],
                                            "severities": [
                                                "Warning",
                                                "Error"
                                            ],
                                            "categories": [
                                                "CPU",
                                                "DSK"
                                            ],
                                            "eventIds": [
                                                "CPU00006C"
                                            ]
                                        },
                                        "description": "example-update-description",
                                        "response": {
                                            "script": {
                                                "name": "example-script",
                                                "content": "IyEvYmluL2Jhc2gKCmN1cmwgaHR0cDovL2V4YW1wbGUuY29tICR7IHN0YXR1cyA9ICd1cGRhdGVkJzsgfQ=="
                                            },
                                            "slacks": [
                                                "https://hooks.slack.com/services/<hookHash>/<hookHash>/<hookHash>"
                                            ],
                                            "emails": [
                                                "example.user@example.com"
                                            ]
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Update trigger request received",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/UpdateTriggerResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Trigger update request received",
                                        "value": {
                                            "code": 202,
                                            "msg": "trigger update request received",
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
                                            "example": "invalid trigger name"
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
                                            "example": "failed to request trigger update: internal server error"
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
            "delete": {
                "operationId": "deleteTrigger",
                "tags": [
                    "Triggers"
                ],
                "summary": "Delete a specific trigger",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/triggerName"
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Delete trigger request received",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/DeleteTriggerResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Trigger deletion request received",
                                        "value": {
                                            "code": 202,
                                            "msg": "trigger deletion request received",
                                            "status": "ok"
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
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "trigger not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to delete trigger: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/triggers/{triggerName}/enable": {
            "patch": {
                "operationId": "enableOrDisableTrigger",
                "tags": [
                    "Triggers"
                ],
                "summary": "Enable or disable a specific trigger",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/triggerName"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/EnableOrDisableTriggerRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Trigger enablement request",
                                    "value": {
                                        "enable": true
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Enable or disable the trigger successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/UpdateTriggerResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Trigger update request received",
                                        "value": {
                                            "code": 202,
                                            "msg": "trigger update request received",
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
                                            "example": "invalid parameter name"
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
                                            "example": "failed to request trigger update: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/supportFiles": {
            "get": {
                "operationId": "getSupportFiles",
                "tags": [
                    "Support Files"
                ],
                "summary": "Retrieve all support files",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    },
                    {
                        "$ref": "#/components/parameters/roles"
                    },
                    {
                        "$ref": "#/components/parameters/start"
                    },
                    {
                        "$ref": "#/components/parameters/stop"
                    },
                    {
                        "$ref": "#/components/parameters/pageSize"
                    },
                    {
                        "$ref": "#/components/parameters/pageNum"
                    },
                    {
                        "$ref": "#/components/parameters/keyword"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve support files successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetSupportFilesResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Support file set list",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "supportFileSet": [
                                                    {
                                                        "name": "CUBE COS 3.0.0 Support File Set 2025-03-18T03:37:19+00:00",
                                                        "description": "example-description",
                                                        "files": [
                                                            {
                                                                "name": "CUBE_3.0.0_20250318-033723_example-node-0.support",
                                                                "group": "Cube Appliance 3.0.0 Support File Set 2025-03-18T03:37:19+00:00",
                                                                "description": "example-description",
                                                                "source": {
                                                                    "role": "control-converged",
                                                                    "host": "example-node-0"
                                                                },
                                                                "sizeMiB": 98.3954,
                                                                "url": "http://example-data-center/supportfiles/CUBE_3.0.0_20250318-215742_example-node-0.support?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Checksum-Mode=ENABLED&X-Amz-Credential=admin%2F20250318%2Fauto%2Fs3%2Faws4_request&X-Amz-Date=20250318T140312Z&X-Amz-Expires=604800&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=ae8c2a39d348709b51d6c0d4b731716f1dc33c172ba43c2919357106cbb15389",
                                                                "status": {
                                                                    "isCreating": false
                                                                }
                                                            },
                                                            {
                                                                "name": "CUBE_3.0.0_20250318-033723_example-node-1.support",
                                                                "group": "Cube Appliance 3.0.0 Support File Set 2025-03-18T03:37:19+00:00",
                                                                "description": "example-description",
                                                                "source": {
                                                                    "role": "control-converged",
                                                                    "host": "example-node-1"
                                                                },
                                                                "sizeMiB": 98.3954,
                                                                "url": "http://example-data-center/supportfiles/CUBE_3.0.0_20250318-215742_example-node-1.support?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Checksum-Mode=ENABLED&X-Amz-Credential=admin%2F20250318%2Fauto%2Fs3%2Faws4_request&X-Amz-Date=20250318T140312Z&X-Amz-Expires=604800&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=ae8c2a39d348709b51d6c0d4b731716f1dc33c172ba43c2919357106cbb15389",
                                                                "status": {
                                                                    "isCreating": false
                                                                }
                                                            }
                                                        ],
                                                        "sizeMiB": 196.7908,
                                                        "status": {
                                                            "current": "completed",
                                                            "createdAt": "2025-03-18T03:37:19+00:00",
                                                            "isCreating": false
                                                        }
                                                    }
                                                ],
                                                "page": {
                                                    "total": 1,
                                                    "number": 1,
                                                    "size": 1,
                                                    "totalItemCount": 1
                                                }
                                            },
                                            "msg": "retrieved support files successfully",
                                            "status": "ok"
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
                                            "example": "failed to list support files: internal server error"
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
            "post": {
                "operationId": "createSupportFiles",
                "tags": [
                    "Support Files"
                ],
                "summary": "Create one or more support files",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/CreateSupportFilesRequest"
                            },
                            "examples": {
                                "example": {
                                    "summary": "Support file creation request",
                                    "value": {
                                        "description": "example-description",
                                        "hosts": [
                                            "example-node-0"
                                        ]
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Support file creation request received",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/CreateSupportFilesResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Support file creation response",
                                        "value": {
                                            "code": 202,
                                            "msg": "support file creation request received",
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
                                            "example": "invalid host found"
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
                                            "example": "failed to request support file creation: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/supportFiles/{supportFileSet}": {
            "delete": {
                "operationId": "DeleteSupportFiles",
                "tags": [
                    "Support Files"
                ],
                "summary": "Delete support file set",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/supportFileSet"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Delete support files successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/DeleteSupportFileSetResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Delete support file set",
                                        "value": {
                                            "code": 200,
                                            "msg": "support file set deleted successfully",
                                            "status": "ok"
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
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "code": {
                                            "type": "integer",
                                            "example": 404
                                        },
                                        "msg": {
                                            "type": "string",
                                            "example": "support file set not found"
                                        },
                                        "status": {
                                            "type": "string",
                                            "example": "not found"
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
                                            "example": "failed to delete support files: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/supportFiles/hosts/{hostname}": {
            "get": {
                "operationId": "getHostSupportFile",
                "tags": [
                    "Support Files"
                ],
                "summary": "Retrieve host support files",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/hostname"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Retrieve host support files successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetHostSupportFilesResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Host support file list",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "name": "CUBE_3.0.0_20250318-215742_example-node-0.support",
                                                    "group": "Cube Appliance 3.0.0 Support File Set 2025-03-18T21:57:39+00:00",
                                                    "description": "",
                                                    "source": {
                                                        "role": "control-converged",
                                                        "host": "example-node-0"
                                                    },
                                                    "sizeMiB": 116.0698,
                                                    "url": "http://example-datacenter/supportfiles/CUBE_3.0.0_20250318-215742_example-node-0.support?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Checksum-Mode=ENABLED&X-Amz-Credential=admin%2F20250318%2Fauto%2Fs3%2Faws4_request&X-Amz-Date=20250318T140312Z&X-Amz-Expires=604800&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=ae8c2a39d348709b51d6c0d4b731716f1dc33c172ba43c2919357106cbb15389",
                                                    "status": {
                                                        "current": "completed",
                                                        "createdAt": "2025-03-18T21:57:39+00:00",
                                                        "isCreating": false
                                                    }
                                                }
                                            ],
                                            "msg": "retrieved support files successfully",
                                            "status": "ok"
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
                                            "example": "failed to list host support files: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/grafana/hosts/{hostname}": {
            "get": {
                "operationId": "getGrafanaHosts",
                "tags": [
                    "Grafana"
                ],
                "summary": "Get Grafana hosts dashboard",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/hostname"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Get Grafana hosts dashboard",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetGrafanaDashboardLinkResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Grafana hosts dashboard link",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "link": "http://example-data-center/grafana/d/i-R2q81iz/host?refresh=5m&kiosk=tv&orgId=1&var-HOST=example-node-0",
                                                "enabled": true
                                            },
                                            "msg": "fetch top host link successfully",
                                            "status": "ok"
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
                                            "example": "failed to get Grafana hosts dashboard: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/grafana/instances/{instanceId}": {
            "get": {
                "operationId": "getGrafanaInstances",
                "tags": [
                    "Grafana"
                ],
                "summary": "Get Grafana instances dashboard",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "in": "path",
                        "name": "instanceId",
                        "required": true,
                        "schema": {
                            "type": "string"
                        },
                        "description": "The instance ID of the instance to operate",
                        "example": "example-instance-id"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Get Grafana instances dashboard",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetGrafanaDashboardLinkResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Grafana instances dashboard link",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "link": "http://example-data-center/grafana/d/qzfq087Wk/instance?refresh=5m&orgId=1&var-TID=example-instance-id&var-TOP=50&var-TENANT=admin",
                                                "enabled": true
                                            },
                                            "msg": "fetch top instance link successfully",
                                            "status": "ok"
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
                                            "example": "failed to get Grafana instances dashboard: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/grafana/topHosts": {
            "get": {
                "operationId": "getGrafanaTopHosts",
                "tags": [
                    "Grafana"
                ],
                "summary": "Get Grafana top hosts dashboard",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Get Grafana top hosts dashboard",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetGrafanaDashboardLinkResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Grafana top hosts dashboard link",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "link": "http://example-data-center/grafana/d/M3ncw6lmk/top-hosts?refresh=5m&kiosk=tv&orgId=1",
                                                "enabled": true
                                            },
                                            "msg": "fetch top host link successfully",
                                            "status": "ok"
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
                                            "example": "failed to get Grafana top hosts dashboard: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/grafana/topInstances": {
            "get": {
                "operationId": "getGrafanaTopInstances",
                "tags": [
                    "Grafana"
                ],
                "summary": "Get Grafana top instances dashboard",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Get Grafana top instances dashboard",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetGrafanaDashboardLinkResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Grafana top instances dashboard link",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "link": "http://example-data-center/grafana/d/qzfq087Wk/top-instances?refresh=5m&orgId=1&var-TID=&var-TOP=50&var-TENANT=admin",
                                                "enabled": true
                                            },
                                            "msg": "fetch top instance link successfully",
                                            "status": "ok"
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
                                            "example": "failed to get Grafana top instances dashboard: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/grafana/networks": {
            "get": {
                "operationId": "getGrafanaNetworks",
                "tags": [
                    "Grafana"
                ],
                "summary": "Get Grafana networks dashboard",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Get Grafana networks dashboard",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetGrafanaDashboardLinkResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Grafana networks dashboard link",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "link": "http://example-data-center/grafana/d/Xx2kkftWk/network?orgId=1&refresh=5m",
                                                "enabled": true
                                            },
                                            "msg": "fetch top network link successfully",
                                            "status": "ok"
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
                                            "example": "failed to get Grafana networks dashboard: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/grafana/networkDevices": {
            "get": {
                "operationId": "getGrafanaNetworkDevices",
                "tags": [
                    "Grafana"
                ],
                "summary": "Get Grafana network devices dashboard",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Get Grafana network devices dashboard",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetGrafanaDashboardLinkResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Grafana network devices dashboard link",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "link": "http://example-data-center/grafana/d/i-device/device?refresh=5m&orgId=1",
                                                "enabled": true
                                            },
                                            "msg": "fetch network devices link successfully",
                                            "status": "ok"
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
                                            "example": "failed to get Grafana network devices dashboard: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/grafana/storages": {
            "get": {
                "operationId": "getGrafanaStorages",
                "tags": [
                    "Grafana"
                ],
                "summary": "Get Grafana storages dashboard",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Get Grafana storages dashboard",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetGrafanaDashboardLinkResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Grafana storages dashboard link",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "link": "http://example-data-center/grafana/d/QTc_sAxiw/storage?refresh=5m&kiosk=tv&orgId=1",
                                                "enabled": true
                                            },
                                            "msg": "fetch top storage link successfully",
                                            "status": "ok"
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
                                            "example": "failed to get Grafana storages dashboard: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/opensearch/requests/{requestId}": {
            "get": {
                "operationId": "getOpenSearchRequestDashboard",
                "tags": [
                    "OpenSearch"
                ],
                "summary": "Get OpenSearch dashboard by request id",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/requestId"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Get OpenSearch request id dashboard",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/GetOpenSearchDashboardLinkResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Opensearch request id dashboard link",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "link": "http://example-data-center/opensearch-dashboards/app/data-explorer/discover/#/view/req-aa95057d-aa8e-4c2c-af56-46ad614e6c33",
                                                "enabled": true
                                            },
                                            "msg": "fetch opensearch dashboard link of request id successfully",
                                            "status": "ok"
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
                                            "example": "failed to get OpenSearch dashboard of request id: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/images/materials": {
            "get": {
                "operationId": "listImageMaterials",
                "tags": [
                    "Images"
                ],
                "summary": "List image materials",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List image materials successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/ListImageMaterialsResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Image material list",
                                        "value": {
                                            "code": 200,
                                            "data": {
                                                "reservedImages": [
                                                    {
                                                        "file": "amphora-x64-haproxy-yoga.qcow2",
                                                        "name": "amphora-x64-haproxy",
                                                        "os": "Ubuntu",
                                                        "destination": "CubeStorage",
                                                        "domain": "default",
                                                        "sourceFromAnotherHypervisor": false,
                                                        "visibility": "private"
                                                    },
                                                    {
                                                        "file": "manila-service-image_yoga.qcow2",
                                                        "name": "manila-service-image",
                                                        "os": "Ubuntu",
                                                        "destination": "CubeStorage",
                                                        "domain": "default",
                                                        "sourceFromAnotherHypervisor": false,
                                                        "visibility": "private"
                                                    }
                                                ],
                                                "projects": [
                                                    {
                                                        "name": "admin",
                                                        "domain": "default",
                                                        "enabled": true,
                                                        "description": "Bootstrap project for initializing the cloud."
                                                    },
                                                    {
                                                        "name": "example-project",
                                                        "domain": "default",
                                                        "enabled": true,
                                                        "description": "Example project"
                                                    }
                                                ],
                                                "oses": [
                                                    "CentOS",
                                                    "Fedora",
                                                    "Ubuntu",
                                                    "Debian",
                                                    "Windows",
                                                    "Rocky",
                                                    "FreeBSD",
                                                    "CoreOS",
                                                    "Arch",
                                                    "Others"
                                                ],
                                                "destinations": [
                                                    "CubeStorage"
                                                ],
                                                "domains": [
                                                    "default"
                                                ],
                                                "visibility": [
                                                    "public",
                                                    "private"
                                                ]
                                            },
                                            "msg": "fetch image materials successfully",
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
                                            "example": "failed to list image materials: internal server error"
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
        "/api/v1/datacenters/{dataCenter}/images": {
            "get": {
                "operationId": "listImages",
                "tags": [
                    "Images"
                ],
                "summary": "List images",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    },
                    {
                        "$ref": "#/components/parameters/watch"
                    },
                    {
                        "$ref": "#/components/parameters/pageSize"
                    },
                    {
                        "$ref": "#/components/parameters/pageNum"
                    },
                    {
                        "$ref": "#/components/parameters/keyword"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List images successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/ListImagesResponse"
                                },
                                "examples": {
                                    "example": {
                                        "summary": "Image list",
                                        "value": {
                                            "code": 200,
                                            "data": [
                                                {
                                                    "id": "763655de-1cfc-45d4-b67c-5579ea43a6e1",
                                                    "name": "example image 0",
                                                    "os": "ubuntu",
                                                    "destination": "CubeStorage",
                                                    "domainName": "default",
                                                    "projectName": "admin",
                                                    "visibilityType": "public",
                                                    "createdAt": "2025-07-27T13:08:53+08:00",
                                                    "status": {
                                                        "current": "active",
                                                        "isProcessing": false
                                                    }
                                                },
                                                {
                                                    "id": "a5218f3c-47df-4ba2-ad41-5e9a50816e6c",
                                                    "name": "example image 2",
                                                    "os": "ubuntu",
                                                    "destination": "CubeStorage",
                                                    "domainName": "default",
                                                    "projectName": "admin",
                                                    "visibilityType": "public",
                                                    "createdAt": "2025-07-29T13:08:53+08:00",
                                                    "status": {
                                                        "current": "importing",
                                                        "isProcessing": true,
                                                        "processPercent": 50.98
                                                    }
                                                }
                                            ],
                                            "msg": "fetch image list successfully",
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
                                            "example": "failed to list images: internal server error"
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
            "post": {
                "operationId": "importImage",
                "tags": [
                    "Images"
                ],
                "summary": "Import an image",
                "parameters": [
                    {
                        "$ref": "#/components/parameters/dataCenter"
                    }
                ],
                "requestBody": {
                    "required": true,
                    "content": {
                        "multipart/form-data": {
                            "schema": {
                                "type": "object",
                                "required": [
                                    "image",
                                    "file",
                                    "name",
                                    "os",
                                    "destination",
                                    "domain",
                                    "project",
                                    "sourceFromAnotherHypervisor",
                                    "visibility"
                                ],
                                "properties": {
                                    "image": {
                                        "type": "string",
                                        "format": "binary",
                                        "description": "The image file to import"
                                    },
                                    "file": {
                                        "type": "string",
                                        "description": "The file name of the image to import"
                                    },
                                    "name": {
                                        "type": "string",
                                        "description": "The name of the image to import"
                                    },
                                    "os": {
                                        "type": "string",
                                        "description": "The OS of the image to import",
                                        "enum": [
                                            "CentOS",
                                            "Fedora",
                                            "Ubuntu",
                                            "Debian",
                                            "Windows",
                                            "Rocky",
                                            "FreeBSD",
                                            "CoreOS",
                                            "Arch",
                                            "Others"
                                        ]
                                    },
                                    "destination": {
                                        "type": "string",
                                        "description": "The destination of the image to import"
                                    },
                                    "domain": {
                                        "type": "string",
                                        "description": "The domain of the image to import"
                                    },
                                    "project": {
                                        "type": "string",
                                        "description": "The project of the image to import"
                                    },
                                    "sourceFromAnotherHypervisor": {
                                        "type": "boolean",
                                        "description": "Whether the image is from another hypervisor"
                                    },
                                    "visibility": {
                                        "type": "string",
                                        "description": "The visibility of the image to import",
                                        "enum": [
                                            "public",
                                            "private"
                                        ]
                                    }
                                }
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Image import accepted and under processing",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "required": [
                                        "code",
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
                                            "example": "image import accepted and under processing"
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
                                            "example": "failed to import image: internal server error"
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
        "securitySchemes": {
            "BearerAuth": {
                "type": "http",
                "scheme": "bearer",
                "bearerFormat": "JWT"
            }
        },
        "parameters": {
            "dataCenter": {
                "in": "path",
                "name": "dataCenter",
                "required": true,
                "schema": {
                    "type": "string"
                },
                "description": "The name of the data center to operate",
                "example": "example-data-center"
            },
            "nodeName": {
                "name": "nodeName",
                "in": "path",
                "description": "The name of the node",
                "required": true,
                "schema": {
                    "type": "string"
                },
                "example": "example-node-0"
            },
            "deviceName": {
                "in": "path",
                "name": "deviceName",
                "required": true,
                "schema": {
                    "type": "string"
                },
                "description": "The device name to remove from the node",
                "example": "sdb"
            },
            "osdId": {
                "in": "path",
                "name": "osdId",
                "required": true,
                "schema": {
                    "type": "string"
                },
                "description": "The OSD ID to operate",
                "example": "osd.0"
            },
            "instanceId": {
                "in": "path",
                "name": "instanceId",
                "required": true,
                "schema": {
                    "type": "string"
                },
                "description": "The instance ID of the instance to operate",
                "example": "example-instance-id"
            },
            "requestId": {
                "in": "path",
                "name": "requestId",
                "required": true,
                "schema": {
                    "type": "string"
                },
                "description": "The request id of the openstack instance operation",
                "example": "req-aa95057d-aa8e-4c2c-af56-46ad614e6c33"
            },
            "hostname": {
                "in": "path",
                "name": "hostname",
                "required": true,
                "schema": {
                    "type": "string"
                },
                "description": "The hostname of the host to operate",
                "example": "example-node-0"
            },
            "watch": {
                "in": "query",
                "name": "watch",
                "required": false,
                "schema": {
                    "type": "boolean"
                },
                "description": "The toggle to enable http chunked transfer for continuous server push.",
                "example": true
            },
            "pageSize": {
                "in": "query",
                "name": "pageSize",
                "required": false,
                "schema": {
                    "type": "integer"
                },
                "description": "The number of items per page (default is unlimit).",
                "example": 10
            },
            "pageNum": {
                "in": "query",
                "name": "pageNum",
                "required": false,
                "schema": {
                    "type": "integer"
                },
                "description": "The page number to retrieve",
                "example": 1
            },
            "keyword": {
                "in": "query",
                "name": "keyword",
                "required": false,
                "schema": {
                    "type": "string"
                },
                "description": "The keyword to search, can be any string",
                "example": "example-keyword"
            },
            "roles": {
                "in": "query",
                "name": "roles",
                "required": false,
                "schema": {
                    "type": "array",
                    "items": {
                        "type": "string",
                        "enum": [
                            "control-converged",
                            "control",
                            "compute",
                            "storage",
                            "edge-core",
                            "moderator"
                        ]
                    }
                },
                "description": "The role of the host",
                "example": "control-converged"
            },
            "listLicenseStatuses": {
                "in": "query",
                "name": "statuses",
                "required": false,
                "schema": {
                    "type": "array",
                    "items": {
                        "$ref": "#/components/schemas/ListLicenseCurrentStatus"
                    }
                },
                "description": "The status of the host",
                "example": "valid"
            },
            "listLicenseAttachmentStatuses": {
                "in": "query",
                "name": "statuses",
                "required": false,
                "schema": {
                    "type": "array",
                    "items": {
                        "$ref": "#/components/schemas/NodeLicenseCurrentStatus"
                    }
                },
                "description": "The status of the host",
                "example": "valid"
            },
            "types": {
                "in": "query",
                "name": "types",
                "required": false,
                "schema": {
                    "type": "array",
                    "items": {
                        "type": "string",
                        "enum": [
                            "trial",
                            "perpetual",
                            "community",
                            "enterprise"
                        ]
                    }
                },
                "description": "The type of the license to query, click 'try it out' to see a few options.",
                "example": "trial"
            },
            "nodeLicenseStatuses": {
                "in": "query",
                "name": "licenseStatuses",
                "required": false,
                "schema": {
                    "type": "array",
                    "items": {
                        "$ref": "#/components/schemas/NodeLicenseCurrentStatus"
                    }
                },
                "description": "The license status of the host",
                "example": "valid"
            },
            "product": {
                "in": "query",
                "name": "product",
                "required": false,
                "schema": {
                    "type": "string",
                    "enum": [
                        "CubeCOS",
                        "CubeCMP"
                    ]
                },
                "description": "The product of the host",
                "example": "CubeCOS"
            },
            "products": {
                "in": "query",
                "name": "products",
                "required": false,
                "schema": {
                    "type": "array",
                    "items": {
                        "type": "string",
                        "enum": [
                            "CubeCOS",
                            "CubeCMP"
                        ]
                    }
                },
                "description": "The products of the host",
                "example": "CubeCOS"
            },
            "metricType": {
                "in": "path",
                "name": "metricType",
                "required": true,
                "schema": {
                    "type": "string",
                    "enum": [
                        "cpuUsage",
                        "memoryUsage",
                        "diskUsage",
                        "diskBandwidth",
                        "diskIops",
                        "diskLatency",
                        "diskReadIops",
                        "diskWriteIops",
                        "networkTrafficIn",
                        "networkTrafficOut"
                    ]
                },
                "description": "The type of metric to query, the value can be 'cpuUsage', 'memoryUsage', 'diskUsage', 'diskBandwidth', 'diskIops', 'diskLatency', 'diskReadIops', 'diskWriteIops', 'networkTrafficIn', or 'networkTrafficOut'.",
                "example": 1
            },
            "viewType": {
                "in": "path",
                "name": "viewType",
                "required": true,
                "schema": {
                    "type": "string",
                    "enum": [
                        "summary",
                        "history",
                        "rank"
                    ]
                },
                "description": "The type of view to query, the value can be only 'summary', 'history', or 'rank'.",
                "example": 1
            },
            "entityType": {
                "in": "path",
                "name": "entityType",
                "required": true,
                "schema": {
                    "type": "string",
                    "enum": [
                        "hosts",
                        "vms"
                    ]
                },
                "description": "The type of entity to query, the value can be 'hosts' or 'vms'",
                "example": "hosts"
            },
            "start": {
                "in": "query",
                "name": "start",
                "required": false,
                "schema": {
                    "type": "string"
                },
                "description": "The start time of the event to query, the value should be in RFC3339 format (default is 24 hours ago).",
                "example": "2025-01-01T01:00:00+00:00"
            },
            "stop": {
                "in": "query",
                "name": "stop",
                "required": false,
                "schema": {
                    "type": "string"
                },
                "description": "The end time of the event to query, the value should be in RFC3339 format (default is now).",
                "example": "2025-01-01T01:00:00+00:00"
            },
            "supportFileSet": {
                "in": "path",
                "name": "supportFileSet",
                "required": true,
                "schema": {
                    "type": "string"
                },
                "description": "The name of the support file set to delete. (have to be done the http encode)",
                "example": "Cube Appliance 3.0.0 Support File Set 2025-06-18T15:54:48+00:00"
            },
            "triggerName": {
                "in": "path",
                "name": "triggerName",
                "required": true,
                "schema": {
                    "type": "string"
                },
                "description": "The name of the trigger to operate",
                "example": "Administrative Level Notification"
            }
        },
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
                                "type",
                                "name",
                                "roles",
                                "version",
                                "virtualIp",
                                "isHaEnabled",
                                "isLocal",
                                "utcTimeZone",
                                "additional"
                            ],
                            "properties": {
                                "type": {
                                    "type": "string",
                                    "enum": [
                                        "cloud",
                                        "edge"
                                    ]
                                },
                                "name": {
                                    "type": "string"
                                },
                                "roles": {
                                    "type": "array",
                                    "items": {
                                        "type": "string",
                                        "enum": [
                                            "control-converged",
                                            "control",
                                            "compute",
                                            "storage",
                                            "edge-core",
                                            "moderator"
                                        ]
                                    }
                                },
                                "version": {
                                    "type": "string"
                                },
                                "virtualIp": {
                                    "type": "string"
                                },
                                "isHaEnabled": {
                                    "type": "boolean"
                                },
                                "isLocal": {
                                    "type": "boolean"
                                },
                                "utcTimeZone": {
                                    "type": "string"
                                },
                                "additional": {
                                    "type": "object",
                                    "required": [
                                        "helpUrl",
                                        "nodeLicenseStatus"
                                    ],
                                    "properties": {
                                        "helpUrl": {
                                            "type": "string"
                                        },
                                        "nodeLicenseStatus": {
                                            "type": "object",
                                            "required": [
                                                "valid",
                                                "expired",
                                                "unlicense"
                                            ],
                                            "properties": {
                                                "valid": {
                                                    "type": "integer",
                                                    "example": 1
                                                },
                                                "expired": {
                                                    "type": "integer",
                                                    "example": 0
                                                },
                                                "unlicense": {
                                                    "type": "integer",
                                                    "example": 0
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
                        "example": "fetch data center list successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetDataCenterResponse": {
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
                            "type",
                            "name",
                            "roles",
                            "version",
                            "virtualIp",
                            "isHaEnabled",
                            "isLocal",
                            "utcTimeZone",
                            "additional"
                        ],
                        "properties": {
                            "type": {
                                "type": "string",
                                "enum": [
                                    "cloud",
                                    "edge"
                                ]
                            },
                            "name": {
                                "type": "string"
                            },
                            "roles": {
                                "type": "array",
                                "items": {
                                    "type": "string",
                                    "enum": [
                                        "control-converged",
                                        "control",
                                        "compute",
                                        "storage",
                                        "edge-core",
                                        "moderator"
                                    ]
                                }
                            },
                            "version": {
                                "type": "string"
                            },
                            "virtualIp": {
                                "type": "string"
                            },
                            "isHaEnabled": {
                                "type": "boolean"
                            },
                            "isLocal": {
                                "type": "boolean"
                            },
                            "utcTimeZone": {
                                "type": "string"
                            },
                            "additional": {
                                "type": "object",
                                "required": [
                                    "helpUrl",
                                    "nodeLicenseStatus"
                                ],
                                "properties": {
                                    "helpUrl": {
                                        "type": "string"
                                    },
                                    "nodeLicenseStatus": {
                                        "type": "object",
                                        "required": [
                                            "valid",
                                            "expired",
                                            "unlicense"
                                        ],
                                        "properties": {
                                            "valid": {
                                                "type": "integer",
                                                "example": 1
                                            },
                                            "expired": {
                                                "type": "integer",
                                                "example": 0
                                            },
                                            "unlicense": {
                                                "type": "integer",
                                                "example": 0
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "fetch data center successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetServicesResponse": {
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
                                "category",
                                "modules"
                            ],
                            "properties": {
                                "name": {
                                    "type": "string"
                                },
                                "category": {
                                    "type": "string"
                                },
                                "modules": {
                                    "type": "array",
                                    "items": {
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
                                            "example": "2025-01-01T01:00:00+00:00"
                                        }
                                    }
                                }
                            },
                            "page": {
                                "$ref": "#/components/schemas/Page"
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
            "GetPredefinedEventFilterResponse": {
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
                        "type": "array",
                        "items": {
                            "type": "object",
                            "required": [
                                "type",
                                "id",
                                "severity",
                                "category",
                                "description"
                            ],
                            "properties": {
                                "type": {
                                    "type": "string"
                                },
                                "id": {
                                    "type": "string"
                                },
                                "severity": {
                                    "type": "string"
                                },
                                "category": {
                                    "type": "string"
                                },
                                "description": {
                                    "type": "string"
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
                        "required": [
                            "events",
                            "limit"
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
            "GetRankedEventsResponse": {
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
                            "events",
                            "limit"
                        ],
                        "properties": {
                            "events": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "id",
                                        "percent",
                                        "number"
                                    ],
                                    "properties": {
                                        "id": {
                                            "type": "string"
                                        },
                                        "severity": {
                                            "type": "string"
                                        },
                                        "category": {
                                            "type": "string"
                                        },
                                        "host": {
                                            "type": "string"
                                        },
                                        "instanceId": {
                                            "type": "string"
                                        },
                                        "instanceName": {
                                            "type": "string"
                                        },
                                        "percent": {
                                            "type": "number"
                                        },
                                        "number": {
                                            "type": "integer"
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
            "GetEventFilterConditionResponse": {
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
                            "system",
                            "instance",
                            "host"
                        ],
                        "properties": {
                            "system": {
                                "type": "object",
                                "required": [
                                    "severities",
                                    "categories"
                                ],
                                "properties": {
                                    "severities": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    },
                                    "categories": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    }
                                }
                            },
                            "instance": {
                                "type": "object",
                                "required": [
                                    "ids",
                                    "categories"
                                ],
                                "properties": {
                                    "ids": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    },
                                    "categories": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    }
                                }
                            },
                            "host": {
                                "type": "object",
                                "required": [
                                    "names",
                                    "categories"
                                ],
                                "properties": {
                                    "names": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    },
                                    "categories": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
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
            "GetNotificationsResponse": {
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
                            "$ref": "#/components/schemas/Notification"
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "fetch notifications successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
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
                            "services"
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
                                            "isFixing"
                                        ],
                                        "properties": {
                                            "current": {
                                                "type": "string",
                                                "enum": [
                                                    "ok",
                                                    "ng"
                                                ]
                                            },
                                            "isFixing": {
                                                "type": "boolean"
                                            },
                                            "description": {
                                                "type": "string"
                                            }
                                        }
                                    }
                                }
                            },
                            "services": {
                                "type": "array",
                                "items": {
                                    "required": [
                                        "name",
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
                                                "current",
                                                "isFixing"
                                            ],
                                            "properties": {
                                                "current": {
                                                    "type": "string",
                                                    "enum": [
                                                        "ok",
                                                        "ng"
                                                    ]
                                                },
                                                "isFixing": {
                                                    "type": "boolean"
                                                },
                                                "description": {
                                                    "type": "string"
                                                }
                                            }
                                        },
                                        "modules": {
                                            "type": "array",
                                            "required": [
                                                "name",
                                                "status"
                                            ],
                                            "items": {
                                                "type": "object",
                                                "required": [
                                                    "name",
                                                    "status"
                                                ],
                                                "properties": {
                                                    "name": {
                                                        "type": "string",
                                                        "example": "link"
                                                    },
                                                    "status": {
                                                        "type": "object",
                                                        "required": [
                                                            "current",
                                                            "isFixing"
                                                        ],
                                                        "properties": {
                                                            "current": {
                                                                "type": "string",
                                                                "enum": [
                                                                    "ok",
                                                                    "ng"
                                                                ]
                                                            },
                                                            "isFixing": {
                                                                "type": "boolean"
                                                            },
                                                            "description": {
                                                                "type": "string"
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
            "GetServiceHealthHistoryResponse": {
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
                        "type": "array",
                        "items": {
                            "type": "object",
                            "required": [
                                "category",
                                "service",
                                "isRepairable",
                                "history",
                                "module"
                            ],
                            "properties": {
                                "category": {
                                    "type": "string"
                                },
                                "service": {
                                    "type": "string"
                                },
                                "isRepairable": {
                                    "type": "boolean"
                                },
                                "history": {
                                    "type": "array",
                                    "items": {
                                        "type": "object",
                                        "required": [
                                            "time",
                                            "hostname",
                                            "status"
                                        ],
                                        "properties": {
                                            "time": {
                                                "type": "string",
                                                "format": "date-time"
                                            },
                                            "hostname": {
                                                "type": "string",
                                                "example": "example-node-0"
                                            },
                                            "status": {
                                                "type": "string",
                                                "enum": [
                                                    "ok",
                                                    "fixing",
                                                    "ng"
                                                ]
                                            },
                                            "error": {
                                                "type": "object",
                                                "properties": {
                                                    "type": {
                                                        "type": "string"
                                                    },
                                                    "reason": {
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
                                },
                                "module": {
                                    "type": "string"
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
                            "category",
                            "name",
                            "module",
                            "isRepairable",
                            "history",
                            "status"
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
                            "isRepairable": {
                                "type": "boolean"
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
                                        "hostname": {
                                            "type": "string",
                                            "example": "example-node-0"
                                        },
                                        "status": {
                                            "type": "string",
                                            "enum": [
                                                "ok",
                                                "fixing",
                                                "ng"
                                            ]
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
                            },
                            "status": {
                                "type": "object",
                                "required": [
                                    "current",
                                    "isFixing"
                                ],
                                "properties": {
                                    "current": {
                                        "type": "string",
                                        "enum": [
                                            "ok",
                                            "ng"
                                        ]
                                    },
                                    "isFixing": {
                                        "type": "boolean"
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
                        "required": [
                            "licenses",
                            "page"
                        ],
                        "properties": {
                            "licenses": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "name",
                                        "type",
                                        "hosts",
                                        "serial",
                                        "product",
                                        "issue",
                                        "quantity",
                                        "supportPlan",
                                        "expiry",
                                        "status"
                                    ],
                                    "properties": {
                                        "name": {
                                            "type": "string"
                                        },
                                        "type": {
                                            "type": "string"
                                        },
                                        "hosts": {
                                            "type": "array",
                                            "items": {
                                                "type": "string"
                                            }
                                        },
                                        "serial": {
                                            "type": "string"
                                        },
                                        "product": {
                                            "type": "object",
                                            "required": [
                                                "name",
                                                "feature"
                                            ],
                                            "properties": {
                                                "name": {
                                                    "type": "string"
                                                },
                                                "feature": {
                                                    "type": "string"
                                                }
                                            }
                                        },
                                        "issue": {
                                            "$ref": "#/components/schemas/LicenseIssue"
                                        },
                                        "quantity": {
                                            "type": "string"
                                        },
                                        "supportPlan": {
                                            "type": "string"
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
                                        },
                                        "status": {
                                            "$ref": "#/components/schemas/ListLicenseStatus"
                                        }
                                    }
                                }
                            },
                            "page": {
                                "$ref": "#/components/schemas/Page"
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
            "PostLicenseRequest": {
                "type": "object",
                "required": [
                    "license"
                ],
                "properties": {
                    "license": {
                        "type": "string",
                        "format": "binary",
                        "description": "License file (must have a .license extension)"
                    }
                }
            },
            "VerifyLicenseResponse": {
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
                            "license",
                            "effectNodes"
                        ],
                        "properties": {
                            "license": {
                                "type": "object",
                                "required": [
                                    "name",
                                    "type",
                                    "product",
                                    "issue",
                                    "quantity",
                                    "supportPlan",
                                    "expiry",
                                    "status"
                                ],
                                "properties": {
                                    "name": {
                                        "type": "string"
                                    },
                                    "type": {
                                        "type": "string"
                                    },
                                    "product": {
                                        "type": "object",
                                        "required": [
                                            "name",
                                            "feature"
                                        ],
                                        "properties": {
                                            "name": {
                                                "type": "string"
                                            },
                                            "feature": {
                                                "type": "string"
                                            }
                                        }
                                    },
                                    "issue": {
                                        "$ref": "#/components/schemas/LicenseIssue"
                                    },
                                    "quantity": {
                                        "type": "string"
                                    },
                                    "supportPlan": {
                                        "type": "string"
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
                                    },
                                    "status": {
                                        "$ref": "#/components/schemas/VerifyLicenseStatus"
                                    }
                                }
                            },
                            "effectNodes": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "name",
                                        "role",
                                        "expiry",
                                        "status"
                                    ],
                                    "properties": {
                                        "name": {
                                            "type": "string"
                                        },
                                        "role": {
                                            "type": "string"
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
                                        },
                                        "status": {
                                            "type": "object",
                                            "required": [
                                                "current",
                                                "isExpiring"
                                            ],
                                            "properties": {
                                                "current": {
                                                    "type": "string",
                                                    "enum": [
                                                        "ok",
                                                        "expiring",
                                                        "expired"
                                                    ]
                                                },
                                                "isExpiring": {
                                                    "type": "boolean"
                                                }
                                            }
                                        }
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
            "PostLicenseResponse": {
                "type": "object",
                "required": [
                    "code",
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
            "GetLicenseAttachmentsResponse": {
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
                        "type": "array",
                        "items": {
                            "type": "object",
                            "required": [
                                "serialNumber",
                                "hostname",
                                "role",
                                "product",
                                "status"
                            ],
                            "properties": {
                                "serialNumber": {
                                    "type": "string"
                                },
                                "hostname": {
                                    "type": "string"
                                },
                                "role": {
                                    "type": "string"
                                },
                                "product": {
                                    "type": "string"
                                },
                                "status": {
                                    "$ref": "#/components/schemas/NodeLicenseCurrentStatus"
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
                            "dataCenter",
                            "host",
                            "vm"
                        ],
                        "properties": {
                            "dataCenter": {
                                "type": "object",
                                "required": [
                                    "usage"
                                ],
                                "properties": {
                                    "usage": {
                                        "type": "object",
                                        "required": [
                                            "cpu",
                                            "memory"
                                        ],
                                        "properties": {
                                            "cpu": {
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
                                            }
                                        }
                                    }
                                }
                            },
                            "host": {
                                "type": "object",
                                "required": [
                                    "role",
                                    "usages"
                                ],
                                "properties": {
                                    "role": {
                                        "type": "object",
                                        "required": [
                                            "controlConverged",
                                            "control",
                                            "compute",
                                            "storage",
                                            "edgeCore",
                                            "moderator"
                                        ],
                                        "properties": {
                                            "controlConverged": {
                                                "$ref": "#/components/schemas/RoleUsage"
                                            },
                                            "control": {
                                                "$ref": "#/components/schemas/RoleUsage"
                                            },
                                            "compute": {
                                                "$ref": "#/components/schemas/RoleUsage"
                                            },
                                            "storage": {
                                                "$ref": "#/components/schemas/RoleUsage"
                                            },
                                            "edgeCore": {
                                                "$ref": "#/components/schemas/RoleUsage"
                                            },
                                            "moderator": {
                                                "$ref": "#/components/schemas/RoleUsage"
                                            }
                                        }
                                    },
                                    "usages": {
                                        "type": "array",
                                        "items": {
                                            "type": "object",
                                            "required": [
                                                "role",
                                                "name",
                                                "address",
                                                "cpu",
                                                "memory"
                                            ],
                                            "properties": {
                                                "role": {
                                                    "type": "string"
                                                },
                                                "name": {
                                                    "type": "string"
                                                },
                                                "address": {
                                                    "type": "string"
                                                },
                                                "cpu": {
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
                                                            "type": "integer"
                                                        },
                                                        "usedCores": {
                                                            "type": "integer"
                                                        },
                                                        "usedPercent": {
                                                            "type": "number"
                                                        },
                                                        "freeCores": {
                                                            "type": "integer"
                                                        },
                                                        "freePercent": {
                                                            "type": "number"
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
                                                            "type": "integer"
                                                        },
                                                        "usedMiB": {
                                                            "type": "integer"
                                                        },
                                                        "usedPercent": {
                                                            "type": "number"
                                                        },
                                                        "freeMiB": {
                                                            "type": "integer"
                                                        },
                                                        "freePercent": {
                                                            "type": "number"
                                                        }
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
                                            "error"
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
            "GetCpuUsageSummaryOfHostsResponse": {
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
                            "totalCores",
                            "usedCores",
                            "usedPercent",
                            "freeCores",
                            "freePercent"
                        ],
                        "properties": {
                            "totalCores": {
                                "type": "integer"
                            },
                            "usedCores": {
                                "type": "integer"
                            },
                            "usedPercent": {
                                "type": "number"
                            },
                            "freeCores": {
                                "type": "integer"
                            },
                            "freePercent": {
                                "type": "number"
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
            "HostMetricHistoryResponse": {
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
                            "unit",
                            "history"
                        ],
                        "properties": {
                            "unit": {
                                "type": "string"
                            },
                            "history": {
                                "type": "array",
                                "items": {
                                    "$ref": "#/components/schemas/TimeValuePair"
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
            "GetCpuUsageRankOfHostsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetMemoryUsageSummaryOfHostsResponse": {
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
                            "totalMiB",
                            "usedMiB",
                            "usedPercent",
                            "freeMiB",
                            "freePercent"
                        ],
                        "properties": {
                            "totalMiB": {
                                "type": "integer"
                            },
                            "usedMiB": {
                                "type": "integer"
                            },
                            "usedPercent": {
                                "type": "number"
                            },
                            "freeMiB": {
                                "type": "integer"
                            },
                            "freePercent": {
                                "type": "number"
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
            "GetMemoryUsageRankOfHostsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetDiskUsageRankOfHostsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetDiskBandwidthHistoryOfHostsResponse": {
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
                        "$ref": "#/components/schemas/DiskReadWriteHistory"
                    },
                    "msg": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            },
            "GetDiskIopsHistoryOfHostsResponse": {
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
                        "$ref": "#/components/schemas/DiskReadWriteHistory"
                    },
                    "msg": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            },
            "GetDiskLatencyHistoryOfHostsResponse": {
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
                        "$ref": "#/components/schemas/DiskReadWriteHistory"
                    },
                    "msg": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            },
            "GetNetworkTrafficInRankOfHostsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetNetworkTrafficOutRankOfHostsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetCpuUsageRankOfVmsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetMemoryUsageSummaryOfVmsResponse": {
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
                            "totalMiB",
                            "usedMiB",
                            "usedPercent",
                            "freeMiB",
                            "freePercent"
                        ],
                        "properties": {
                            "totalMiB": {
                                "type": "integer"
                            },
                            "usedMiB": {
                                "type": "integer"
                            },
                            "usedPercent": {
                                "type": "number"
                            },
                            "freeMiB": {
                                "type": "integer"
                            },
                            "freePercent": {
                                "type": "number"
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
            "GetMemoryUsageRankOfVmsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetDiskReadIopsRankOfVmsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetDiskWriteIopsRankOfVmsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetNetworkTrafficInRankOfVmsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
            "GetNetworkTrafficOutRankOfVmsResponse": {
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
                        "$ref": "#/components/schemas/MetricRank"
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
                                    "$ref": "#/components/schemas/Node"
                                }
                            },
                            "page": {
                                "$ref": "#/components/schemas/Page"
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
            "GetNodeResponse": {
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
                        "$ref": "#/components/schemas/Node"
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
            "GetNodeIpmiSettingResponse": {
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
                            "ip",
                            "port",
                            "username",
                            "password"
                        ],
                        "properties": {
                            "ip": {
                                "type": "string"
                            },
                            "port": {
                                "type": "integer"
                            },
                            "username": {
                                "type": "string"
                            },
                            "password": {
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
            "VerifyNodeIpmiResponse": {
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
                            "board",
                            "product"
                        ],
                        "properties": {
                            "board": {
                                "type": "object",
                                "required": [
                                    "manufacturingDate",
                                    "manufacturer",
                                    "product",
                                    "serial",
                                    "partNumber"
                                ],
                                "properties": {
                                    "manufacturingDate": {
                                        "type": "string",
                                        "format": "date-time"
                                    },
                                    "manufacturer": {
                                        "type": "string"
                                    },
                                    "product": {
                                        "type": "string"
                                    },
                                    "serial": {
                                        "type": "string"
                                    },
                                    "partNumber": {
                                        "type": "string"
                                    }
                                }
                            },
                            "product": {
                                "type": "object",
                                "required": [
                                    "manufacturer",
                                    "name",
                                    "serial",
                                    "version"
                                ],
                                "properties": {
                                    "manufacturer": {
                                        "type": "string"
                                    },
                                    "name": {
                                        "type": "string"
                                    },
                                    "serial": {
                                        "type": "string"
                                    },
                                    "version": {
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
            "NodeIpmiSettingRequest": {
                "type": "object",
                "required": [
                    "ip",
                    "port",
                    "username",
                    "password"
                ],
                "properties": {
                    "ip": {
                        "type": "string",
                        "description": "IP address of the node's IPMI interface"
                    },
                    "port": {
                        "type": "integer",
                        "description": "Port number of the node's IPMI interface"
                    },
                    "username": {
                        "type": "string",
                        "description": "Username for the node's IPMI interface"
                    },
                    "password": {
                        "type": "string",
                        "description": "Password for the node's IPMI interface"
                    }
                }
            },
            "AddNodeDeviceRequest": {
                "type": "object",
                "required": [
                    "device"
                ],
                "properties": {
                    "device": {
                        "type": "string",
                        "description": "The device path, e.g., sdc"
                    }
                }
            },
            "SetNodeIpmiSettingResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "IPMI settings set successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "OperateNodeIpmiResponse": {
                "type": "object",
                "required": [
                    "code",
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
                        "example": "node is being operated successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "ListNodeDevicesResponse": {
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
                                "serial",
                                "device",
                                "type",
                                "class",
                                "sizeMiB",
                                "availability",
                                "osd",
                                "status"
                            ],
                            "properties": {
                                "serial": {
                                    "type": "string"
                                },
                                "device": {
                                    "type": "string"
                                },
                                "class": {
                                    "$ref": "#/components/schemas/DeviceType"
                                },
                                "type": {
                                    "type": "string"
                                },
                                "sizeMiB": {
                                    "type": "number"
                                },
                                "availability": {
                                    "type": "string",
                                    "enum": [
                                        "available",
                                        "in-use",
                                        "system"
                                    ]
                                },
                                "osd": {
                                    "type": "object",
                                    "required": [
                                        "pgs",
                                        "reweight",
                                        "daemons"
                                    ],
                                    "properties": {
                                        "pgs": {
                                            "type": "integer"
                                        },
                                        "reweight": {
                                            "type": "number"
                                        },
                                        "daemons": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "required": [
                                                    "id",
                                                    "usagePercent",
                                                    "status"
                                                ],
                                                "properties": {
                                                    "id": {
                                                        "type": "string"
                                                    },
                                                    "usagePercent": {
                                                        "type": "number"
                                                    },
                                                    "status": {
                                                        "type": "object",
                                                        "required": [
                                                            "current",
                                                            "isProcessing"
                                                        ],
                                                        "properties": {
                                                            "current": {
                                                                "type": "string",
                                                                "enum": [
                                                                    "up",
                                                                    "down",
                                                                    "warning",
                                                                    "error"
                                                                ]
                                                            },
                                                            "isProcessing": {
                                                                "type": "boolean"
                                                            }
                                                        }
                                                    }
                                                }
                                            }
                                        }
                                    }
                                },
                                "status": {
                                    "type": "object",
                                    "required": [
                                        "current",
                                        "description",
                                        "isPromotable",
                                        "isDemotable",
                                        "isProcessing"
                                    ],
                                    "properties": {
                                        "current": {
                                            "type": "string",
                                            "enum": [
                                                "ok",
                                                "warning",
                                                "fail"
                                            ]
                                        },
                                        "isPromotable": {
                                            "type": "boolean"
                                        },
                                        "isDemotable": {
                                            "type": "boolean"
                                        },
                                        "isProcessing": {
                                            "type": "boolean"
                                        },
                                        "description": {
                                            "type": "string"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "list node devices successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "AddNodeDeviceResponse": {
                "type": "object",
                "required": [
                    "code",
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
                        "example": "the request to add node device is accepted successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "UpdateNodeDeviceResponse": {
                "type": "object",
                "required": [
                    "code",
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
                        "example": "the request to update node device is accepted successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "RemoveNodeDeviceResponse": {
                "type": "object",
                "required": [
                    "code",
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
                        "example": "the request to remove node device is accepted successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "UpdateNodeOsdResponse": {
                "type": "object",
                "required": [
                    "code",
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
                        "example": "the request to update node OSD is accepted successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "RestartNodeOsdResponse": {
                "type": "object",
                "required": [
                    "code",
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
                        "example": "the request to restart node OSD is accepted successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "DeleteNodeOsdResponse": {
                "type": "object",
                "required": [
                    "code",
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
                        "example": "the request to delete node OSD is accepted successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "UpdateNodeDeviceRequest": {
                "type": "object",
                "required": [
                    "class"
                ],
                "properties": {
                    "class": {
                        "$ref": "#/components/schemas/DeviceType"
                    }
                }
            },
            "UpdateNodeOsdRequest": {
                "type": "object",
                "required": [
                    "reweight"
                ],
                "properties": {
                    "reweight": {
                        "type": "number",
                        "description": "The reweight value for the OSD. It's only allow to have two decimal places with a range of 0.0 to 1.0, for example, 0.5 or 0.75."
                    }
                }
            },
            "DisconnectNodeIpmiResponse": {
                "type": "object",
                "required": [
                    "code",
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
                        "example": "example-name"
                    },
                    "password": {
                        "type": "string",
                        "description": "the password of user to generate the token",
                        "example": "example-password"
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
            },
            "GetSupportFilesResponse": {
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
                            "supportFileSet",
                            "page"
                        ],
                        "properties": {
                            "supportFileSet": {
                                "type": "array",
                                "items": {
                                    "$ref": "#/components/schemas/SupportFileSet"
                                }
                            },
                            "page": {
                                "$ref": "#/components/schemas/Page"
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
            "DeleteSupportFileSetResponse": {
                "type": "object",
                "required": [
                    "code",
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
            "TitlePrefix": {
                "type": "object",
                "required": [
                    "value",
                    "status"
                ],
                "properties": {
                    "value": {
                        "type": "string"
                    },
                    "status": {
                        "$ref": "#/components/schemas/SettingStatus"
                    }
                }
            },
            "GetHostSupportFilesResponse": {
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
                                "group",
                                "description",
                                "source",
                                "sizeMiB",
                                "url",
                                "status"
                            ],
                            "properties": {
                                "name": {
                                    "type": "string"
                                },
                                "group": {
                                    "type": "string"
                                },
                                "description": {
                                    "type": "string"
                                },
                                "source": {
                                    "type": "object",
                                    "required": [
                                        "role",
                                        "host"
                                    ],
                                    "properties": {
                                        "role": {
                                            "type": "string"
                                        },
                                        "host": {
                                            "type": "string"
                                        }
                                    }
                                },
                                "sizeMiB": {
                                    "type": "integer"
                                },
                                "url": {
                                    "type": "string"
                                },
                                "status": {
                                    "type": "object",
                                    "required": [
                                        "current",
                                        "isCreating",
                                        "createdAt"
                                    ],
                                    "properties": {
                                        "current": {
                                            "type": "string",
                                            "enum": [
                                                "ok",
                                                "error"
                                            ]
                                        },
                                        "isCreating": {
                                            "type": "boolean"
                                        },
                                        "createdAt": {
                                            "type": "string"
                                        }
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
            "EmailSenderResponse": {
                "type": "object",
                "required": [
                    "host",
                    "port",
                    "username",
                    "from",
                    "accessVerified",
                    "status"
                ],
                "properties": {
                    "host": {
                        "type": "string"
                    },
                    "port": {
                        "type": "integer"
                    },
                    "username": {
                        "type": "string"
                    },
                    "from": {
                        "type": "string"
                    },
                    "accessVerified": {
                        "type": "boolean"
                    },
                    "status": {
                        "$ref": "#/components/schemas/SettingStatus"
                    }
                }
            },
            "EmailSenderPostRequest": {
                "type": "object",
                "required": [
                    "host",
                    "port",
                    "from",
                    "status"
                ],
                "properties": {
                    "host": {
                        "type": "string"
                    },
                    "port": {
                        "type": "integer"
                    },
                    "username": {
                        "type": "string"
                    },
                    "password": {
                        "type": "string"
                    },
                    "from": {
                        "type": "string"
                    }
                }
            },
            "EmailSenderPatchRequest": {
                "type": "object",
                "properties": {
                    "host": {
                        "type": "string"
                    },
                    "port": {
                        "type": "integer"
                    },
                    "username": {
                        "type": "string"
                    },
                    "password": {
                        "type": "string"
                    },
                    "from": {
                        "type": "string"
                    }
                }
            },
            "TryEmailSender": {
                "type": "object",
                "required": [
                    "email"
                ],
                "properties": {
                    "email": {
                        "type": "string"
                    }
                }
            },
            "EmailRecipientResponse": {
                "type": "object",
                "required": [
                    "address",
                    "note",
                    "status"
                ],
                "properties": {
                    "address": {
                        "type": "string"
                    },
                    "note": {
                        "type": "string"
                    },
                    "status": {
                        "$ref": "#/components/schemas/SettingStatus"
                    }
                }
            },
            "EmailRecipientPostRequest": {
                "type": "object",
                "required": [
                    "address",
                    "note"
                ],
                "properties": {
                    "address": {
                        "type": "string"
                    },
                    "note": {
                        "type": "string"
                    }
                }
            },
            "EmailRecipientPutRequest": {
                "type": "object",
                "properties": {
                    "address": {
                        "type": "string"
                    },
                    "note": {
                        "type": "string"
                    }
                }
            },
            "SlackChannelGetResponse": {
                "type": "object",
                "required": [
                    "name",
                    "url",
                    "description",
                    "status"
                ],
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "url": {
                        "type": "string"
                    },
                    "description": {
                        "type": "string"
                    },
                    "status": {
                        "$ref": "#/components/schemas/SettingStatus"
                    }
                }
            },
            "SlackChannelPostRequest": {
                "type": "object",
                "required": [
                    "name",
                    "url",
                    "description"
                ],
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "url": {
                        "type": "string"
                    },
                    "description": {
                        "type": "string"
                    }
                }
            },
            "SlackChannelPutRequest": {
                "type": "object",
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "url": {
                        "type": "string"
                    },
                    "description": {
                        "type": "string"
                    }
                }
            },
            "GetSettingResponse": {
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
                            "titlePrefix",
                            "email",
                            "slack"
                        ],
                        "properties": {
                            "titlePrefix": {
                                "type": "object",
                                "required": [
                                    "value",
                                    "status"
                                ],
                                "properties": {
                                    "value": {
                                        "type": "string"
                                    },
                                    "status": {
                                        "$ref": "#/components/schemas/SettingStatus"
                                    }
                                }
                            },
                            "email": {
                                "type": "object",
                                "required": [
                                    "senders",
                                    "recipients"
                                ],
                                "properties": {
                                    "senders": {
                                        "type": "array",
                                        "items": {
                                            "$ref": "#/components/schemas/EmailSenderResponse"
                                        }
                                    },
                                    "recipients": {
                                        "type": "array",
                                        "items": {
                                            "$ref": "#/components/schemas/EmailRecipientResponse"
                                        }
                                    }
                                }
                            },
                            "slack": {
                                "type": "object",
                                "properties": {
                                    "channels": {
                                        "type": "array",
                                        "items": {
                                            "$ref": "#/components/schemas/SlackChannelGetResponse"
                                        }
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
            "UpdateTitlePrefixResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "title prefix updated successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "PostEmailSenderResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 201
                    },
                    "msg": {
                        "type": "string",
                        "example": "email senders created successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "TryEmailSenderResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "email sender tried successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetEmailSendersResponse": {
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
                            "emailSenders": {
                                "type": "array",
                                "items": {
                                    "$ref": "#/components/schemas/EmailSenderResponse"
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "email senders retrieved successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "PutEmailSenderResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "email senders updated successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "DeleteEmailSenderResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "email senders deleted successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "PostEmailRecipientResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 201
                    },
                    "msg": {
                        "type": "string",
                        "example": "email recipient created successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetEmailRecipientsResponse": {
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
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/EmailRecipientResponse"
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
            "TryEmailRecipientResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "email recipient tried successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "PutEmailRecipientResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "email recipient updated successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "DeleteEmailRecipientResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "email recipient deleted successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "PostSlackChannelResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 201
                    },
                    "msg": {
                        "type": "string",
                        "example": "slack channel created successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetSlackChannelsResponse": {
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
                            "$ref": "#/components/schemas/SlackChannelGetResponse"
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "slack channels retrieved successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "TrySlackChannelResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "slack channel tried successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "PutSlackChannelResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "slack channel updated successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "DeleteSlackChannelResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "slack channel deleted successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "ListTuningResponse": {
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
                            "tunings",
                            "page"
                        ],
                        "properties": {
                            "tunings": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "name",
                                        "value",
                                        "hosts",
                                        "description",
                                        "enabled",
                                        "isModified",
                                        "limitation",
                                        "status"
                                    ],
                                    "properties": {
                                        "name": {
                                            "type": "string"
                                        },
                                        "value": {
                                            "type": "string"
                                        },
                                        "hosts": {
                                            "type": "array",
                                            "items": {
                                                "type": "object",
                                                "required": [
                                                    "name",
                                                    "role",
                                                    "ip"
                                                ],
                                                "properties": {
                                                    "name": {
                                                        "type": "string"
                                                    },
                                                    "role": {
                                                        "type": "string"
                                                    },
                                                    "ip": {
                                                        "type": "string"
                                                    }
                                                }
                                            }
                                        },
                                        "description": {
                                            "type": "string"
                                        },
                                        "enabled": {
                                            "type": "boolean"
                                        },
                                        "isModified": {
                                            "type": "boolean"
                                        },
                                        "limitation": {
                                            "type": "object",
                                            "required": [
                                                "type",
                                                "default",
                                                "min",
                                                "max"
                                            ],
                                            "properties": {
                                                "type": {
                                                    "$ref": "#/components/schemas/TuningLimitationType"
                                                },
                                                "default": {
                                                    "oneOf": [
                                                        {
                                                            "type": "string"
                                                        },
                                                        {
                                                            "type": "integer"
                                                        },
                                                        {
                                                            "type": "number"
                                                        },
                                                        {
                                                            "type": "boolean"
                                                        }
                                                    ]
                                                },
                                                "min": {
                                                    "oneOf": [
                                                        {
                                                            "type": "integer"
                                                        },
                                                        {
                                                            "type": "number"
                                                        }
                                                    ]
                                                },
                                                "max": {
                                                    "oneOf": [
                                                        {
                                                            "type": "integer"
                                                        },
                                                        {
                                                            "type": "number"
                                                        }
                                                    ]
                                                },
                                                "regex": {
                                                    "type": "string"
                                                }
                                            }
                                        },
                                        "status": {
                                            "type": "object",
                                            "required": [
                                                "current",
                                                "updatedAt",
                                                "isUpdating"
                                            ],
                                            "properties": {
                                                "current": {
                                                    "type": "string"
                                                },
                                                "updatedAt": {
                                                    "type": "string"
                                                },
                                                "isUpdating": {
                                                    "type": "boolean"
                                                }
                                            }
                                        }
                                    }
                                }
                            },
                            "page": {
                                "$ref": "#/components/schemas/Page"
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "tuning list retrieved successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "ListTuningSpecResponse": {
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
                                "description",
                                "limitation",
                                "roles",
                                "selector"
                            ],
                            "properties": {
                                "name": {
                                    "type": "string"
                                },
                                "description": {
                                    "type": "string"
                                },
                                "limitation": {
                                    "type": "object",
                                    "required": [
                                        "type",
                                        "default"
                                    ],
                                    "properties": {
                                        "type": {
                                            "$ref": "#/components/schemas/TuningLimitationType"
                                        },
                                        "default": {
                                            "oneOf": [
                                                {
                                                    "type": "string"
                                                },
                                                {
                                                    "type": "integer"
                                                },
                                                {
                                                    "type": "boolean"
                                                }
                                            ]
                                        },
                                        "min": {
                                            "oneOf": [
                                                {
                                                    "type": "integer"
                                                }
                                            ]
                                        },
                                        "max": {
                                            "oneOf": [
                                                {
                                                    "type": "integer"
                                                }
                                            ]
                                        },
                                        "regex": {
                                            "type": "string"
                                        }
                                    }
                                },
                                "roles": {
                                    "type": "array",
                                    "items": {
                                        "type": "object",
                                        "required": [
                                            "name",
                                            "hosts"
                                        ],
                                        "properties": {
                                            "name": {
                                                "type": "string"
                                            },
                                            "hosts": {
                                                "type": "array",
                                                "items": {
                                                    "type": "object",
                                                    "required": [
                                                        "name",
                                                        "ip"
                                                    ],
                                                    "properties": {
                                                        "name": {
                                                            "type": "string"
                                                        },
                                                        "ip": {
                                                            "type": "string"
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
                        "type": "string"
                    },
                    "status": {
                        "type": "string"
                    }
                }
            },
            "UpdateTuningRequest": {
                "type": "object",
                "required": [
                    "value",
                    "hosts"
                ],
                "properties": {
                    "value": {
                        "oneOf": [
                            {
                                "type": "string"
                            },
                            {
                                "type": "integer"
                            },
                            {
                                "type": "boolean"
                            }
                        ]
                    },
                    "hosts": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        }
                    }
                }
            },
            "EnableOrDisableTuningRequest": {
                "type": "object",
                "required": [
                    "enable",
                    "hosts"
                ],
                "properties": {
                    "enable": {
                        "type": "boolean"
                    },
                    "hosts": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        }
                    }
                }
            },
            "ResetTuningRequest": {
                "type": "object",
                "required": [
                    "hosts"
                ],
                "properties": {
                    "hosts": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        }
                    }
                }
            },
            "UpdateTuningResponse": {
                "type": "object",
                "required": [
                    "code",
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
            "ResetTuningResponse": {
                "type": "object",
                "required": [
                    "code",
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
            "GetTriggerMaterialsResponse": {
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
                            "attribute",
                            "response"
                        ],
                        "properties": {
                            "attribute": {
                                "type": "object",
                                "required": [
                                    "alertTypes",
                                    "severities",
                                    "categories",
                                    "eventIds"
                                ],
                                "properties": {
                                    "alertTypes": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    },
                                    "severities": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    },
                                    "categories": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    },
                                    "eventIds": {
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    }
                                }
                            },
                            "response": {
                                "type": "object",
                                "required": [
                                    "scriptType",
                                    "emails",
                                    "slacks"
                                ],
                                "properties": {
                                    "scriptType": {
                                        "type": "object",
                                        "required": [
                                            "language",
                                            "environment"
                                        ],
                                        "properties": {
                                            "language": {
                                                "type": "string"
                                            },
                                            "environment": {
                                                "type": "string"
                                            }
                                        }
                                    },
                                    "slacks": {
                                        "type": "array",
                                        "items": {
                                            "type": "object",
                                            "required": [
                                                "name",
                                                "url",
                                                "description"
                                            ],
                                            "properties": {
                                                "name": {
                                                    "type": "string"
                                                },
                                                "url": {
                                                    "type": "string"
                                                },
                                                "description": {
                                                    "type": "string"
                                                }
                                            }
                                        }
                                    },
                                    "emails": {
                                        "type": "array",
                                        "items": {
                                            "type": "object",
                                            "required": [
                                                "address",
                                                "note"
                                            ],
                                            "properties": {
                                                "address": {
                                                    "type": "string"
                                                },
                                                "note": {
                                                    "type": "string"
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
                        "example": "trigger materials retrieved successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "VerifyMaterialScriptResponse": {
                "type": "object",
                "required": [
                    "code",
                    "msg",
                    "status"
                ],
                "properties": {
                    "code": {
                        "type": "integer",
                        "example": 200
                    },
                    "msg": {
                        "type": "string",
                        "example": "trigger script verified successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "VerifyMaterialScriptRequest": {
                "type": "object",
                "required": [
                    "script"
                ],
                "properties": {
                    "script": {
                        "type": "string",
                        "example": "base64-encoded-script-content"
                    }
                }
            },
            "GetTriggersResponse": {
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
                            "triggers",
                            "page"
                        ],
                        "properties": {
                            "triggers": {
                                "type": "array",
                                "items": {
                                    "$ref": "#/components/schemas/Trigger"
                                }
                            },
                            "page": {
                                "$ref": "#/components/schemas/Page"
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "triggers retrieved successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "GetTriggerResponse": {
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
                        "$ref": "#/components/schemas/Trigger"
                    },
                    "msg": {
                        "type": "string",
                        "example": "triggers retrieved successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "CreateTriggerRequest": {
                "type": "object",
                "required": [
                    "name",
                    "attribute",
                    "response"
                ],
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "description": {
                        "type": "string"
                    },
                    "attribute": {
                        "type": "object",
                        "required": [
                            "alertTypes",
                            "severities",
                            "categories",
                            "eventIds"
                        ],
                        "properties": {
                            "alertTypes": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "severities": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "categories": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "eventIds": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "response": {
                        "type": "object",
                        "required": [
                            "script",
                            "emails",
                            "slacks"
                        ],
                        "properties": {
                            "script": {
                                "type": "object",
                                "required": [
                                    "name",
                                    "content"
                                ],
                                "properties": {
                                    "name": {
                                        "type": "string"
                                    },
                                    "content": {
                                        "type": "string",
                                        "description": "Base64 encoded script content"
                                    }
                                }
                            },
                            "emails": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "slacks": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                }
            },
            "CreateTriggerResponse": {
                "type": "object",
                "required": [
                    "code",
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
            "UpdateTriggerRequest": {
                "type": "object",
                "required": [
                    "attribute",
                    "response"
                ],
                "properties": {
                    "attribute": {
                        "type": "object",
                        "required": [
                            "alertTypes",
                            "severities",
                            "categories",
                            "eventIds"
                        ],
                        "properties": {
                            "alertTypes": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "severities": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "categories": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "eventIds": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "response": {
                        "type": "object",
                        "required": [
                            "script",
                            "slacks",
                            "emails"
                        ],
                        "properties": {
                            "script": {
                                "type": "object",
                                "required": [
                                    "name",
                                    "content"
                                ],
                                "properties": {
                                    "name": {
                                        "type": "string"
                                    },
                                    "content": {
                                        "type": "string",
                                        "description": "Base64 encoded script content"
                                    }
                                }
                            },
                            "slacks": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "emails": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "description": {
                        "type": "string"
                    }
                }
            },
            "EnableOrDisableTriggerRequest": {
                "type": "object",
                "required": [
                    "enable"
                ],
                "properties": {
                    "enable": {
                        "type": "boolean"
                    }
                }
            },
            "UpdateTriggerResponse": {
                "type": "object",
                "required": [
                    "code",
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
            "DeleteTriggerResponse": {
                "type": "object",
                "required": [
                    "code",
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
            "CreateSupportFilesRequest": {
                "type": "object",
                "required": [
                    "description",
                    "hosts"
                ],
                "properties": {
                    "description": {
                        "type": "string"
                    },
                    "hosts": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        }
                    }
                }
            },
            "CreateSupportFilesResponse": {
                "type": "object",
                "required": [
                    "code",
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
            "GetGrafanaDashboardLinkResponse": {
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
                            "link",
                            "enabled"
                        ],
                        "properties": {
                            "link": {
                                "type": "string"
                            },
                            "enabled": {
                                "type": "boolean"
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
            "GetOpenSearchDashboardLinkResponse": {
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
                            "link",
                            "enabled"
                        ],
                        "properties": {
                            "link": {
                                "type": "string"
                            },
                            "enabled": {
                                "type": "boolean"
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
            "ListImageMaterialsResponse": {
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
                            "reservedImages",
                            "projects",
                            "oses",
                            "destinations",
                            "domains",
                            "visibilities"
                        ],
                        "properties": {
                            "reservedImages": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "file",
                                        "name",
                                        "os",
                                        "destination",
                                        "domain",
                                        "project",
                                        "sourceFromAnotherHypervisor",
                                        "visibility"
                                    ],
                                    "properties": {
                                        "file": {
                                            "type": "string"
                                        },
                                        "name": {
                                            "type": "string"
                                        },
                                        "os": {
                                            "type": "string"
                                        },
                                        "destination": {
                                            "type": "string"
                                        },
                                        "domain": {
                                            "type": "string"
                                        },
                                        "project": {
                                            "type": "string"
                                        },
                                        "sourceFromAnotherHypervisor": {
                                            "type": "boolean"
                                        },
                                        "visibility": {
                                            "type": "string"
                                        }
                                    }
                                }
                            },
                            "projects": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "oses": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "destinations": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "domains": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "visibilities": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "materials retrieved successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "ListImagesResponse": {
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
                                "id",
                                "name",
                                "os",
                                "destination",
                                "domain",
                                "project",
                                "visibility",
                                "sizeMiB",
                                "createdAt",
                                "status"
                            ],
                            "properties": {
                                "id": {
                                    "type": "string"
                                },
                                "name": {
                                    "type": "string"
                                },
                                "os": {
                                    "type": "string"
                                },
                                "destination": {
                                    "type": "string"
                                },
                                "domain": {
                                    "type": "string"
                                },
                                "project": {
                                    "type": "string"
                                },
                                "visibility": {
                                    "type": "string",
                                    "enum": [
                                        "public",
                                        "private",
                                        "shared",
                                        "community",
                                        "unknown"
                                    ]
                                },
                                "createdAt": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "msg": {
                        "type": "string",
                        "example": "images retrieved successfully"
                    },
                    "status": {
                        "type": "string",
                        "example": "ok"
                    }
                }
            },
            "MetricRank": {
                "type": "object",
                "required": [
                    "unit",
                    "rank"
                ],
                "properties": {
                    "unit": {
                        "type": "string"
                    },
                    "rank": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "required": [
                                "id",
                                "name",
                                "value",
                                "history"
                            ],
                            "properties": {
                                "id": {
                                    "type": "string"
                                },
                                "name": {
                                    "type": "string"
                                },
                                "device": {
                                    "type": "string"
                                },
                                "value": {
                                    "type": "number"
                                },
                                "history": {
                                    "type": "array",
                                    "items": {
                                        "type": "object",
                                        "required": [
                                            "time",
                                            "value"
                                        ],
                                        "properties": {
                                            "time": {
                                                "type": "string",
                                                "format": "date-time"
                                            },
                                            "value": {
                                                "type": "number"
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            },
            "DiskReadWriteHistory": {
                "type": "object",
                "required": [
                    "unit",
                    "read",
                    "write"
                ],
                "properties": {
                    "unit": {
                        "type": "string"
                    },
                    "read": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "required": [
                                "time",
                                "value"
                            ],
                            "properties": {
                                "time": {
                                    "type": "string",
                                    "format": "date-time"
                                },
                                "value": {
                                    "type": "number"
                                }
                            }
                        }
                    },
                    "write": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "required": [
                                "time",
                                "value"
                            ],
                            "properties": {
                                "time": {
                                    "type": "string",
                                    "format": "date-time"
                                },
                                "value": {
                                    "type": "number"
                                }
                            }
                        }
                    }
                }
            },
            "RoleUsage": {
                "type": "object",
                "required": [
                    "count",
                    "cpu",
                    "memory"
                ],
                "properties": {
                    "count": {
                        "type": "integer"
                    },
                    "cpu": {
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
                                "type": "number"
                            },
                            "usedCores": {
                                "type": "number"
                            },
                            "freeCores": {
                                "type": "number"
                            },
                            "usedPercent": {
                                "type": "number"
                            },
                            "freePercent": {
                                "type": "number"
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
                                "type": "number"
                            },
                            "usedMiB": {
                                "type": "number"
                            },
                            "freeMiB": {
                                "type": "number"
                            },
                            "usedPercent": {
                                "type": "number"
                            },
                            "freePercent": {
                                "type": "number"
                            }
                        }
                    }
                }
            },
            "Page": {
                "type": "object",
                "required": [
                    "total",
                    "number",
                    "size",
                    "totalItemCount"
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
                    },
                    "totalItemCount": {
                        "type": "integer",
                        "example": 1
                    }
                }
            },
            "Node": {
                "type": "object",
                "required": [
                    "id",
                    "serialNumber",
                    "dataCenter",
                    "hostname",
                    "role",
                    "address",
                    "ip",
                    "managementIP",
                    "storageIP",
                    "isVirtualIpOwner",
                    "license",
                    "status",
                    "cpuSpec",
                    "networkInterfaces",
                    "ipmi",
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
                    "serialNumber": {
                        "type": "string"
                    },
                    "dataCenter": {
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
                    "ip": {
                        "type": "string"
                    },
                    "managementIP": {
                        "type": "string"
                    },
                    "storageIP": {
                        "type": "string"
                    },
                    "isVirtualIpOwner": {
                        "type": "boolean"
                    },
                    "license": {
                        "type": "object",
                        "required": [
                            "name",
                            "type",
                            "hostname",
                            "serial",
                            "product",
                            "issue",
                            "quantity",
                            "supportPlan",
                            "expiry",
                            "status"
                        ],
                        "properties": {
                            "name": {
                                "type": "string"
                            },
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
                                    "feature"
                                ],
                                "properties": {
                                    "name": {
                                        "type": "string"
                                    },
                                    "feature": {
                                        "type": "string"
                                    }
                                }
                            },
                            "issue": {
                                "$ref": "#/components/schemas/LicenseIssue"
                            },
                            "quantity": {
                                "type": "string"
                            },
                            "supportPlan": {
                                "type": "string"
                            },
                            "expiry": {
                                "type": "object",
                                "required": [
                                    "date",
                                    "days"
                                ],
                                "properties": {
                                    "date": {
                                        "type": "string",
                                        "format": "date-time"
                                    },
                                    "days": {
                                        "type": "integer"
                                    }
                                }
                            },
                            "status": {
                                "$ref": "#/components/schemas/NodeLicenseStatus"
                            }
                        }
                    },
                    "status": {
                        "type": "string",
                        "enum": [
                            "up",
                            "down",
                            "powering on",
                            "powering off",
                            "powering cycle",
                            "syncing",
                            "unknown"
                        ]
                    },
                    "cpuSpec": {
                        "type": "string"
                    },
                    "networkInterfaces": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "required": [
                                "interface",
                                "label",
                                "busIdSlaves",
                                "driver",
                                "state",
                                "speed"
                            ],
                            "properties": {
                                "interface": {
                                    "type": "string"
                                },
                                "label": {
                                    "type": "string"
                                },
                                "busIdSlaves": {
                                    "type": "string"
                                },
                                "driver": {
                                    "type": "string"
                                },
                                "state": {
                                    "type": "string"
                                },
                                "speed": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "ipmi": {
                        "type": "object",
                        "required": [
                            "isSupported",
                            "isConnected",
                            "ip"
                        ],
                        "properties": {
                            "isSupported": {
                                "type": "boolean"
                            },
                            "isConnected": {
                                "type": "boolean"
                            },
                            "ip": {
                                "type": "string"
                            }
                        }
                    },
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
                                "type": "integer"
                            },
                            "usedCores": {
                                "type": "integer"
                            },
                            "usedPercent": {
                                "type": "number"
                            },
                            "freeCores": {
                                "type": "integer"
                            },
                            "freePercent": {
                                "type": "number"
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
                                "type": "number"
                            },
                            "usedMiB": {
                                "type": "number"
                            },
                            "usedPercent": {
                                "type": "number"
                            },
                            "freeMiB": {
                                "type": "number"
                            },
                            "freePercent": {
                                "type": "number"
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
                                "type": "number"
                            },
                            "usedMiB": {
                                "type": "number"
                            },
                            "usedPercent": {
                                "type": "number"
                            },
                            "freeMiB": {
                                "type": "number"
                            },
                            "freePercent": {
                                "type": "number"
                            }
                        }
                    },
                    "uptimeSeconds": {
                        "type": "number"
                    },
                    "labels": {
                        "type": "object"
                    }
                }
            },
            "SupportFileSet": {
                "type": "object",
                "required": [
                    "name",
                    "description",
                    "files",
                    "sizeMiB",
                    "status"
                ],
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "description": {
                        "type": "string"
                    },
                    "files": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "$ref": "#/components/schemas/SupportFile"
                        }
                    },
                    "sizeMiB": {
                        "type": "number"
                    },
                    "status": {
                        "type": "object",
                        "required": [
                            "current",
                            "createdAt",
                            "isCreating"
                        ],
                        "properties": {
                            "current": {
                                "type": "string"
                            },
                            "createdAt": {
                                "type": "string"
                            },
                            "isCreating": {
                                "type": "boolean"
                            }
                        }
                    }
                }
            },
            "Trigger": {
                "type": "object",
                "required": [
                    "name",
                    "isBuiltIn",
                    "description",
                    "attribute",
                    "response",
                    "enabled"
                ],
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "isBuiltIn": {
                        "type": "boolean"
                    },
                    "description": {
                        "type": "string"
                    },
                    "attribute": {
                        "type": "object",
                        "required": [
                            "alertTypes",
                            "severities",
                            "categories",
                            "eventIds"
                        ],
                        "properties": {
                            "alertTypes": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "severities": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "categories": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "eventIds": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "response": {
                        "type": "object",
                        "required": [
                            "types",
                            "slacks",
                            "emails"
                        ],
                        "properties": {
                            "types": {
                                "type": "array",
                                "items": {
                                    "type": "string",
                                    "enum": [
                                        "script",
                                        "email",
                                        "slack"
                                    ]
                                }
                            },
                            "slacks": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "name",
                                        "url",
                                        "description"
                                    ],
                                    "properties": {
                                        "name": {
                                            "type": "string"
                                        },
                                        "url": {
                                            "type": "string"
                                        },
                                        "description": {
                                            "type": "string"
                                        }
                                    }
                                }
                            },
                            "emails": {
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "required": [
                                        "address",
                                        "note"
                                    ],
                                    "properties": {
                                        "address": {
                                            "type": "string"
                                        },
                                        "note": {
                                            "type": "string"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "status": {
                        "type": "object",
                        "required": [
                            "current",
                            "isProcessing"
                        ],
                        "properties": {
                            "current": {
                                "type": "string"
                            },
                            "isProcessing": {
                                "type": "boolean"
                            }
                        }
                    },
                    "enabled": {
                        "type": "boolean"
                    }
                }
            },
            "SupportFile": {
                "type": "object",
                "required": [
                    "name",
                    "group",
                    "description",
                    "source",
                    "url",
                    "sizeMiB",
                    "status"
                ],
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "sizeMiB": {
                        "type": "number"
                    },
                    "group": {
                        "type": "string"
                    },
                    "description": {
                        "type": "string"
                    },
                    "source": {
                        "type": "object",
                        "required": [
                            "role",
                            "host"
                        ],
                        "properties": {
                            "role": {
                                "type": "string"
                            },
                            "host": {
                                "type": "string"
                            }
                        }
                    },
                    "url": {
                        "type": "string"
                    },
                    "status": {
                        "type": "object",
                        "required": [
                            "current",
                            "createdAt",
                            "isCreating"
                        ],
                        "properties": {
                            "current": {
                                "type": "string"
                            },
                            "createdAt": {
                                "type": "string"
                            },
                            "isCreating": {
                                "type": "boolean"
                            }
                        }
                    }
                }
            },
            "TimeValuePair": {
                "type": "object",
                "required": [
                    "time",
                    "value"
                ],
                "properties": {
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "value": {
                        "type": "number"
                    }
                }
            },
            "TuningLimitationType": {
                "type": "string",
                "enum": [
                    "str",
                    "int",
                    "uint",
                    "bool"
                ]
            },
            "LicenseIssue": {
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
                        "type": "string",
                        "description": "this field will be the serial number(s) of the host(s) that the license is issued to. '*' means all hosts, genearlly, it's for trial license only. for the paid license, it will be the comma separated serial numbers of the hosts.\nexamples:\n  - \"*\"\n  - example-serial-number-1, example-serial-number-2, ...\n"
                    },
                    "date": {
                        "type": "string",
                        "format": "date-time"
                    }
                }
            },
            "ListLicenseCurrentStatus": {
                "type": "string",
                "enum": [
                    "valid",
                    "expired"
                ]
            },
            "ListLicenseStatus": {
                "type": "object",
                "required": [
                    "current",
                    "isExpiring"
                ],
                "properties": {
                    "current": {
                        "type": "string",
                        "enum": [
                            "valid",
                            "expired"
                        ]
                    },
                    "isExpiring": {
                        "type": "boolean"
                    }
                }
            },
            "NodeLicenseStatus": {
                "type": "object",
                "required": [
                    "current",
                    "isExpiring"
                ],
                "properties": {
                    "current": {
                        "$ref": "#/components/schemas/NodeLicenseCurrentStatus"
                    },
                    "isExpiring": {
                        "type": "boolean"
                    }
                }
            },
            "NodeLicenseCurrentStatus": {
                "type": "string",
                "enum": [
                    "unlicense",
                    "valid",
                    "expired"
                ]
            },
            "VerifyLicenseStatus": {
                "type": "object",
                "required": [
                    "current",
                    "isExpiring"
                ],
                "properties": {
                    "current": {
                        "type": "string",
                        "enum": [
                            "valid",
                            "unmatched hardware",
                            "invalid signature",
                            "system compromised",
                            "expired"
                        ]
                    },
                    "isExpiring": {
                        "type": "boolean"
                    }
                }
            },
            "SettingStatus": {
                "type": "object",
                "required": [
                    "current",
                    "isUpdating"
                ],
                "properties": {
                    "current": {
                        "type": "string",
                        "enum": [
                            "ok",
                            "updating",
                            "error"
                        ]
                    },
                    "isUpdating": {
                        "type": "boolean"
                    }
                }
            },
            "DeviceType": {
                "type": "string",
                "enum": [
                    "SSD",
                    "HDD"
                ]
            },
            "Notification": {
                "oneOf": [
                    {
                        "$ref": "#/components/schemas/DEV00001I"
                    },
                    {
                        "$ref": "#/components/schemas/DEV00001E"
                    },
                    {
                        "$ref": "#/components/schemas/DEV00002I"
                    },
                    {
                        "$ref": "#/components/schemas/DEV00002E"
                    },
                    {
                        "$ref": "#/components/schemas/DEV00003I"
                    },
                    {
                        "$ref": "#/components/schemas/DEV00003E"
                    },
                    {
                        "$ref": "#/components/schemas/DEV00004I"
                    },
                    {
                        "$ref": "#/components/schemas/DEV00004E"
                    },
                    {
                        "$ref": "#/components/schemas/OSD00001I"
                    },
                    {
                        "$ref": "#/components/schemas/OSD00001E"
                    },
                    {
                        "$ref": "#/components/schemas/OSD00002I"
                    },
                    {
                        "$ref": "#/components/schemas/OSD00002E"
                    },
                    {
                        "$ref": "#/components/schemas/OSD00003I"
                    },
                    {
                        "$ref": "#/components/schemas/OSD00003E"
                    }
                ],
                "discriminator": {
                    "propertyName": "id",
                    "mapping": {
                        "DEV00001I": "#/components/schemas/DEV00001I",
                        "DEV00001E": "#/components/schemas/DEV00001E",
                        "DEV00002I": "#/components/schemas/DEV00002I",
                        "DEV00002E": "#/components/schemas/DEV00002E",
                        "DEV00003I": "#/components/schemas/DEV00003I",
                        "DEV00003E": "#/components/schemas/DEV00003E",
                        "DEV00004I": "#/components/schemas/DEV00004I",
                        "DEV00004E": "#/components/schemas/DEV00004E",
                        "OSD00001I": "#/components/schemas/OSD00001I",
                        "OSD00001E": "#/components/schemas/OSD00001E",
                        "OSD00002I": "#/components/schemas/OSD00002I",
                        "OSD00002E": "#/components/schemas/OSD00002E",
                        "OSD00003I": "#/components/schemas/OSD00003I",
                        "OSD00003E": "#/components/schemas/OSD00003E"
                    }
                }
            },
            "DEV00001I": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^DEV"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "device"
                        ],
                        "properties": {
                            "device": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "DEV00001E": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^DEV"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "device"
                        ],
                        "properties": {
                            "device": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "DEV00002I": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^DEV"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "device",
                            "class"
                        ],
                        "properties": {
                            "device": {
                                "type": "string"
                            },
                            "class": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "DEV00002E": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^DEV"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "device",
                            "class"
                        ],
                        "properties": {
                            "device": {
                                "type": "string"
                            },
                            "class": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "DEV00003I": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^DEV"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "device",
                            "class"
                        ],
                        "properties": {
                            "device": {
                                "type": "string"
                            },
                            "class": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "DEV00003E": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^DEV"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "device",
                            "class"
                        ],
                        "properties": {
                            "device": {
                                "type": "string"
                            },
                            "class": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "DEV00004I": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^DEV"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "device"
                        ],
                        "properties": {
                            "device": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "DEV00004E": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^DEV"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "device"
                        ],
                        "properties": {
                            "device": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "OSD00001I": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^OSD"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "osdId"
                        ],
                        "properties": {
                            "osdId": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "OSD00001E": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^OSD"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "osdId"
                        ],
                        "properties": {
                            "osdId": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "OSD00002I": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^OSD"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "osdId",
                            "reweight"
                        ],
                        "properties": {
                            "osdId": {
                                "type": "string"
                            },
                            "reweight": {
                                "type": "number"
                            }
                        }
                    }
                }
            },
            "OSD00002E": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^OSD"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "osdId",
                            "reweight"
                        ],
                        "properties": {
                            "osdId": {
                                "type": "string"
                            },
                            "reweight": {
                                "type": "number"
                            }
                        }
                    }
                }
            },
            "OSD00003I": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^OSD"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "osdId"
                        ],
                        "properties": {
                            "osdId": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "OSD00003E": {
                "type": "object",
                "required": [
                    "id",
                    "nodeName",
                    "time",
                    "additionalInfo"
                ],
                "properties": {
                    "id": {
                        "type": "string",
                        "pattern": "^OSD"
                    },
                    "nodeName": {
                        "type": "string"
                    },
                    "time": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "additionalInfo": {
                        "type": "object",
                        "required": [
                            "osdId"
                        ],
                        "properties": {
                            "osdId": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        }
    },
    "security": [
        {
            "BearerAuth": []
        }
    ]
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
