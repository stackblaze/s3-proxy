Name:           s3-proxy
Version:        %{version}
Release:        1%{?dist}
Summary:        A lightweight S3-compatible proxy optimized for ZeroFS
License:        MIT
URL:            https://github.com/stackblaze/s3-proxy
Source0:        %{name}-%{version}.tar.gz

BuildArch:      noarch
BuildRequires:  systemd

%description
A lightweight S3-compatible proxy optimized for ZeroFS. Proxies S3 requests
to backend storage (Wasabi, AWS S3, Backblaze, etc.) with support for Range
requests, DELETE operations, and conditional writes.

%prep
%setup -q

%build
# Binary is pre-built, no compilation needed

%install
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/etc/systemd/system
mkdir -p %{buildroot}/usr/lib/systemd/system

install -m 755 s3-proxy %{buildroot}/usr/local/bin/s3-proxy

%files
%defattr(-,root,root,-)
/usr/local/bin/s3-proxy

%changelog
* %(date +"%a %b %d %Y") Stackblaze <noreply@stackblaze.com> - %{version}-1
- Initial package release

