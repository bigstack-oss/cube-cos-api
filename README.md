<!-- PROJECT LOGO -->
<br/>
<div align="left">
  <img src="assets/images/bigstack.png" alt="Logo" width="" height="">
  </p>
</div>

[![License][License-Image]][License-Url] [![made-with-Go][Go-Made-Image]][Go-Made-Url] [![Go][Go-Report-Image]][Go-Report-Url] [![GitHub issues][Github-Issue-Image]][Github-Issue-Url] [![GitHub last commit (branch)][GitHub-Last-Commit-Image]][GitHub-Last-Commit-Url]

<br/><br/>

<h3 style="color: gray;">▎To Start Developing</h3>

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

3). Install the rpm and start the service

```bash
systemctl stop cube-cos-api
dnf -y remove cube-cos-api
dnf -y install "<path to cube-cos-api rpm>"
hex_config bootstrap api
```

<br/>

---

<br/>


<h3 style="color: gray;">▎License</h3>


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


[License-Url]: https://www.apache.org/licenses/LICENSE-2.0
[License-Image]: https://img.shields.io/badge/License-Apache2-blue.svg
[Go-Report-Url]: https://goreportcard.com/report/github.com/bigstack-oss/cube-cos-api
[Go-Report-Image]: https://goreportcard.com/badge/github.com/bigstack-oss/cube-cos-api
[Go-Made-Image]: https://img.shields.io/badge/Made%20with-Go-1f425f.svg
[Go-Made-Url]: https://go.dev/
[GitHub-Issue-Url]: https://github.com/bigstack-oss/cube-cos-api/issues
[GitHub-Issue-Image]: https://img.shields.io/github/issues/bigstack-oss/cube-cos-api?color=brightgreen
[GitHub-Last-Commit-Url]: https://github.com/bigstack-oss/cube-cos-api/issues
[GitHub-Last-Commit-Image]: https://img.shields.io/github/last-commit/bigstack-oss/cube-cos-api/develop
