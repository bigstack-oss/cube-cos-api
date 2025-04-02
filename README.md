<!-- PROJECT LOGO -->
<br />
<div align="center">
  <img src="assets/images/bigstack.png" alt="Logo" width="80" height="80">
  <h3 align="center">CubeCOS API</h3>
  <p align="center">
    A Central Communication Interface of CubeCOS
  </p>
</div>

<br/><br/>

#### ▎To Start Developing

<br/>

0). Get the build environment ready

We would need an `x86_64` / `amd64` based machine to build the rpm package.

We would need to have `golang` ready on the machine.

We would need rpm build tools.

For Fedora Linux based OS:

```bash
sudo dnf install -y rpmdevtools rpmlint
```

<br/>

1). Build rpm

```bash
task rpm:build
```

<br/>

2). Send the built rpm to a running CubeCOS

```
scp <path of rpm> <user>@<cubecos>:<path to place rpm>
```

<br/>

3). Prepare the running environment

```bash
systemctl stop cube-cos-api
systemctl disable cube-cos-api
dnf remove cube-cos-api
```

<br/>

4). Install the rpm and start the service

```bash
dnf install "<path to cube-cos-api rpm>"
cp -f /etc/cube/api/cube-cos-api.yaml.in /etc/cube/api/cube-cos-api.yaml
systemctl enable cube-cos-api
systemctl start cube-cos-api
```

<br/>

5). Clean up

```bash
systemctl stop cube-cos-api
systemctl disable cube-cos-api
dnf remove cube-cos-api
```

<br/>

---

<br/>

#### ▎License

Copyright (c) 2025 [Bigstack co., ltd](https://bigstack.co/)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
