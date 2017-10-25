%define        _topdir  %{rpmbuild}
%define        _tmppath %{_topdir}/tmp

Summary: Tool for managing remote & local storage.
Name: %{prog_name}
Version: %{v_semver}
Release: 1
License: Apache License
Group: Applications/Storage
#Source: https://github.com/thecodeteam/rexray/archive/master.zip
URL: https://github.com/thecodeteam/rexray
Vendor: {code} by Dell EMC
Packager: Andrew Kutz <sakutz@gmail.com>
BuildArch: %{v_arch}
BuildRoot: %{_tmppath}/%{prog_name}-%{version}-%{release}

%description
A guest based storage introspection tool that
allows local visibility and management from cloud
and storage platforms.

%prep

%build

%install
install -D %{prog_path} $RPM_BUILD_ROOT/usr/bin/%{prog_name}

%post
/usr/bin/%{prog_name} install 1> /dev/null

%preun
/usr/bin/%{prog_name} uninstall --package 1> /dev/null

%clean
#rm -rf "$RPM_BUILD_ROOT"

%files
%attr(0755, root, root) /usr/bin/%{prog_name}
