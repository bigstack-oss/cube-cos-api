## ▎To Start Developing

The following guide is for anyone to write code which directly accesses the CubeCOS API.

Genearlly, there're two ways for CubeCOS API developing which help you to have a quick iteration

- RPM: A comprehensive way to develop with CubeCOS

- Binary replacement: A faster way to develop with CubeCOS (⚠️ the binary will be reset when CubeCOS get restarted)

<br/>

## ▎Get The Environment Ready

- Go >= 1.24.2

- Taskfile

- rpmdevtools

- rpmlint

<br/>

## ▎RPM

1). Go to cube-cos-api dir

<br/>

2). Execute the task command

```bash
task rpm:build
```

<br/>

3). Send the built rpm to a running CubeCOS

```
scp <path of rpm> <user>@<cubecos>:<path to place rpm>
```

<br/>

4). Log in to your CubeCOS and reinstall the cube-cos-api by the RPM

```bash
systemctl stop cube-cos-api
dnf -y remove cube-cos-api
rm -f /usr/local/bin/cube-cos-api
dnf -y install "<path to cube-cos-api rpm>"
```

<br/>

5). Commit the changes if you don't want to lost the binary after machine reboot

```bash
git add /usr/local/bin/cube-cos-api
hex_sdk git_push "change the cube-cos-api with dev version manually (YYYY-MM-DD)"
```

<br/>

6). Restart the cube-cos-api by the hex sdk

```bash
hex_config bootstrap api
```

<br/>

## ▎Binary Replacement

1). Go to cube-cos-api dir

<br/>

2). Replace your CubeCOS IP in the Taskfile.yaml below

```sh
  devCosIp: "<REPLACE_WITH_YOUR_DEV_COS_IP>"
```

<br/>

3). Execute the task command

```bash
task deployRemoteDevApi
```

<br/>

## ▎API Documentation

The API spec lives in `cube-cos-openapi/docs.yaml` (OpenAPI 3.0 YAML) as the **single source of truth**.

At build time, the YAML is converted to JSON and embedded into the Go binary via `//go:embed`, then served by gin-swagger at `/api/v1/apidocs/index.html`.

```
cube-cos-openapi/docs.yaml  →  (yq)  →  api/docs.json  →  (go:embed)  →  binary
         ↑                                                                    ↓
   single source of truth                                          gin-swagger serves UI
```

<br/>

### Editing the API spec

1). Edit `cube-cos-openapi/docs.yaml` directly

<br/>

2). Regenerate the embedded JSON

```bash
task generateApiDocs
```

<br/>

3). Verify no uncommitted diff (CI runs this automatically)

```bash
task diffApiDocs
```

<br/>

### Prerequisites

- [yq](https://github.com/mikefarah/yq) (for YAML → JSON conversion)

<br/>

### Architecture

| Component                       | Role                                                                            |
| ------------------------------- | ------------------------------------------------------------------------------- |
| `cube-cos-openapi/docs.yaml`    | Source of truth for the OpenAPI spec (validated by CI in cube-cos-openapi repo) |
| `api/docs.json`                 | Auto-generated JSON from the YAML (committed, checked by `task diffApiDocs`)    |
| `api/docs.go`                   | Uses `//go:embed docs.json` to bake the spec into the binary                    |
| `internal/apis/doc.go`          | Registers the gin-swagger handler on `/api/v1/apidocs/*any`                     |
| `github.com/swaggo/swag`        | Runtime library for spec registration (`swag.Register`)                         |
| `github.com/swaggo/gin-swagger` | Gin middleware serving Swagger UI                                               |
| `github.com/swaggo/files`       | Embedded Swagger UI static assets                                               |
