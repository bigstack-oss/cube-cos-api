## ▎To Start Developing

The following guide is for anyone to write code which directly accesses the CubeCOS API.

Genearlly, there're two ways for CubeCOS API developing which help you to have a quick iteration

- RPM: A comprehensive way to develop with CubeCOS

- Binary replacement: A faster way to develop with CubeCOS (⚠️ the binary will be reset when CubeCOS get restarted)

<br/>

## ▎Get The Environment Ready

- Go >= 1.24.0

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

4). Log in to your CubeCOS

```bash
systemctl stop cube-cos-api
dnf -y remove cube-cos-api
dnf -y install "<path to cube-cos-api rpm>"
systemctl start cube-cos-api
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
