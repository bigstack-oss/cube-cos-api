Name:           cube-cos-api
Version:        %{version}
Release:        1%{?dist}.%{build_number}
Summary:        API service for CubeCOS

License:        Apache License 2.0
URL:            https://github.com/bigstack-oss/cube-cos-api
Source0:        https://github.com/bigstack-oss/cube-cos-api/tree/%{build_number}

BuildRequires: systemd golang

%description
The API service for CubeCOS.

%prep
rm -rf ./*
cp ../SOURCES/"cube-cos-api-%{version}.tar.gz" .
tar -xzf "cube-cos-api-%{version}.tar.gz"
rm "cube-cos-api-%{version}.tar.gz"
mv ./source/* .
rmdir source

%build
GOWORK=off go mod tidy -v
go clean -v
GOWORK=off GOOS=linux GOARCH=amd64 go build -o %{name} -v cmd/main.go

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/usr/local/bin
mv %{name} $RPM_BUILD_ROOT/usr/local/bin
mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/cube/api
cp ./configs/cube-cos-api.yaml.template $RPM_BUILD_ROOT/%{_sysconfdir}/cube/api/cube-cos-api.yaml
mkdir -p $RPM_BUILD_ROOT/%{_unitdir}
cp ./init/cube-cos-api.service $RPM_BUILD_ROOT/%{_unitdir}
mkdir -p $RPM_BUILD_ROOT/%{_datadir}/cube/api
cp LICENSE $RPM_BUILD_ROOT/%{_datadir}/cube/api

%files
/usr/local/bin/%{name}
%{_sysconfdir}/cube/api/cube-cos-api.yaml
%{_unitdir}/cube-cos-api.service
%{_datadir}/cube/api/LICENSE
