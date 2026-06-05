# ▎To Start Developing

The following guide is for anyone to write code which directly accesses the CubeCOS API.

Genearlly, there're two ways for CubeCOS API developing which help you to have a quick iteration

- RPM: A comprehensive way to develop with CubeCOS

- Binary replacement: A faster way to develop with CubeCOS (⚠️ the binary will be reset when CubeCOS get restarted)

## ▎Get The Environment Ready

- Go >= 1.24.2

- Taskfile

- rpmdevtools

- rpmlint

## ▎RPM

1). Go to cube-cos-api dir

2). Execute the task command

```bash
task rpm:build
```

3). Send the built rpm to a running CubeCOS

```bash
scp <path of rpm> <user>@<cubecos>:<path to place rpm>
```

4). Log in to your CubeCOS and reinstall the cube-cos-api by the RPM

```bash
systemctl stop cube-cos-api
dnf -y remove cube-cos-api
rm -f /usr/local/bin/cube-cos-api
dnf -y install "<path to cube-cos-api rpm>"
```

5). Commit the changes if you don't want to lost the binary after machine reboot

```bash
git add /usr/local/bin/cube-cos-api
hex_sdk git_push "change the cube-cos-api with dev version manually (YYYY-MM-DD)"
```

6). Restart the cube-cos-api by the hex sdk

```bash
hex_config bootstrap api
```

## ▎Binary Replacement

1). Go to cube-cos-api dir

2). Replace your CubeCOS IP in the Taskfile.yaml below

```sh
  devCosIp: "<REPLACE_WITH_YOUR_DEV_COS_IP>"
```

3). Execute the task command

```bash
task deployRemoteDevApi
```

## ▎API Documentation

The API spec lives in `api/cube-cos-openapi/docs.yaml` (OpenAPI 3.0 YAML) as the **single source of truth**.
This is a git submodule pointing to [bigstack-oss/cube-cos-openapi](https://github.com/bigstack-oss/cube-cos-openapi).

At build time, the YAML is converted to JSON and embedded into the Go binary via `//go:embed`, then served by gin-swagger at `/api/v1/datacenters/:dataCenter/apidocs/index.html`.

`api/docs.json` is **gitignored** — it is always generated fresh during local dev, CI, and RPM builds.

```txt
api/cube-cos-openapi/docs.yaml  →  (yq)  →  api/docs.json  →  (go:embed)  →  binary
            ↑                                                                    ↓
   single source of truth                                          gin-swagger serves UI
```

### Editing the API spec

1). Edit `api/cube-cos-openapi/docs.yaml` directly

2). Regenerate the embedded JSON

```bash
task generateApiDocs
```

NOTE: the `generateApiDocs` task is depended by `buildLocalBinary` and `runLocalDevApi`.
So we don't need to manually update it every time.

### Prerequisites

- [yq](https://github.com/mikefarah/yq) (for YAML → JSON conversion)

### Architecture

| Component                        | Role                                                                            |
| -------------------------------- | ------------------------------------------------------------------------------- |
| `api/cube-cos-openapi/docs.yaml` | Source of truth for the OpenAPI spec (validated by CI in cube-cos-openapi repo) |
| `api/docs.json`                  | Auto-generated JSON from the YAML (gitignored)                                  |
| `api/docs.go`                    | Uses `//go:embed docs.json` to bake the spec into the binary                    |
| `internal/apis/doc.go`           | Registers the gin-swagger handler on `/api/v1/apidocs/*any`                     |
| `github.com/swaggo/swag`         | Runtime library for spec registration (`swag.Register`)                         |
| `github.com/swaggo/gin-swagger`  | Gin middleware serving Swagger UI                                               |
| `github.com/swaggo/files`        | Embedded Swagger UI static assets                                               |

## | Update Jenkins Base Image

### Binary Pipeline Base Image

Jenkins `cube-cos-api-binary` pipeline uses the `localhost:5000/cube-cos-api-jail` base image.
If you want to update the base image, update the `./build/cube-cos-api-jail.Dockerfile` and execute `task deployJailImageToRemoteJenkins`, it generates a image tar file and loaded by jenkins build server.

### RPM Pipeline Base Image

Jenkins `cube-cos-api-rpm` pipeline uses the `localhost:5000/cube-cos-api-rpm` base image.
If you want to update the base image, update the `./build/cube-cos-api-rpm.Dockerfile` and execute `task deployRpmImageToRemoteJenkins`, it generates a image tar file and loaded by jenkins build server.
