%define        _topdir  ${RPMBUILD}
%define        _tmppath %{_topdir}/tmp

Summary: Tool for managing remote & local storage.
Name: rexray
Version: 1.0
Release: 1
License: Apache License
Group: Applications/Storage
#Source: https://github.com/emccode/rexray/archive/master.zip
URL: https://github.com/emccode/rexray
Vendor: EMC{code}
Packager: Andrew Kutz <sakutz@gmail.com>
BuildArch: x86_64
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}

%description
A guest based storage introspection tool that 
allows local visibility and management from cloud 
and storage platforms.

%prep

%build

%install
mkdir -p $RPM_BUILD_ROOT/usr/bin
cp -a ${GOPATH}/bin/rexray $RPM_BUILD_ROOT/usr/bin

%clean
rm -rf "$RPM_BUILD_ROOT"

%files
%defattr(-,root,root,-)
%{_bindir}/*