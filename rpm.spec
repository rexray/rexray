%define        _topdir  %{rpmbuild}
%define        _tmppath %{_topdir}/tmp

Summary: Tool for managing remote & local storage.
Name: %{prog_name}
Version: %{v_semver}
Release: 1
License: Apache License
Group: Applications/Storage
#Source: https://github.com/thecodeteam/rexray/archive/master.zip
URL: https://github.com/AVENTER-UG/rexray
Vendor: {code} 
Packager: AVENTER UG (haftungsbeschraenkt) www.aventer.biz
BuildArch: %{v_arch}
BuildRoot: %{_tmppath}/%{prog_name}-%{version}-%{release}
Requires: systemd

%description
A guest based storage introspection tool that
allows local visibility and management from cloud
and storage platforms.

%prep

%build

%install
install -d -m 0755 %{buildroot}%{_unitdir}/
install -D %{prog_path} $RPM_BUILD_ROOT/usr/bin/%{prog_name}
install -D -m 644  %{prog_path}.service %{buildroot}%{_unitdir}/%{prog_name}.service

%post
/usr/bin/%{prog_name} install 1> /dev/null
%systemd_post %{prog_name}.service 

%preun
/usr/bin/%{prog_name} uninstall --package 1> /dev/null

%clean
#rm -rf "$RPM_BUILD_ROOT"

%files
%attr(0755, root, root) /usr/bin/%{prog_name}
%{_unitdir}/rexray.service
