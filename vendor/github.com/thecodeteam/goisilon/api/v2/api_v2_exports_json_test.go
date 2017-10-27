package v2

var getOneExportJSON = []byte(`{ "exports" :  [

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [ "10.50.0.111" ], "commit_asynchronous" : false,
"conflicting_paths" : [], "description" : "", "directory_transfer_size" :
131072, "encoding" : "DEFAULT", "id" : 24, "link_max" : 32767, "map_failure" :  {
"enabled" : false, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "map_full" : true, "map_lookup_uid" : false,
"map_non_root" :  { "enabled" : true, "primary_group" : {}, "secondary_groups" : [],
"user" :  { "id" : "USER:root" } }, "map_retry" : true, "map_root" :  {
"enabled" : true, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:root" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/testing" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" } ] }`)

var getAllExportsJSON = []byte(`{ "exports" :  [

{ "all_dirs" : true, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "Default export", "directory_transfer_size" : 131072, "encoding" :
"DEFAULT", "id" : 1, "link_max" : 32767, "map_failure" :  { "enabled" : false,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:nobody" } },
"map_full" : true, "map_lookup_uid" : false, "map_non_root" :  { "enabled" :
false, "primary_group" : {}, "secondary_groups" : [], "user" :  { "id" :
"USER:nobody" } }, "map_retry" : true, "map_root" :  { "enabled" : true,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:nobody" } },
"max_file_size" : 9223372036854775807, "name_max_size" : 255, "no_truncate" :
false, "paths" : [ "/ifs" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [],
"security_flavors" : [ "unix" ], "setattr_asynchronous" : false, "snapshot" : "-",
"symlinks" : true, "time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 16, "link_max" : 32767, "map_failure" :  { "enabled" : false,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:nobody" } },
"map_full" : true, "map_lookup_uid" : false, "map_non_root" :  { "enabled" :
false, "primary_group" : {}, "secondary_groups" : [], "user" :  { "id" :
"USER:nobody" } }, "map_retry" : true, "map_root" :  { "enabled" : true,
"primary_group" :  { "id" : "GROUP:wheel" }, "secondary_groups" : [], "user" :  {
"id" : "USER:root" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/admin_test" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [ "10.50.0.111" ], "commit_asynchronous" : false,
"conflicting_paths" : [], "description" : "", "directory_transfer_size" :
131072, "encoding" : "DEFAULT", "id" : 17, "link_max" : 32767, "map_failure" :  {
"enabled" : false, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "map_full" : true, "map_lookup_uid" : false,
"map_non_root" :  { "enabled" : true, "primary_group" : {}, "secondary_groups" : [],
"user" :  { "id" : "USER:root" } }, "map_retry" : true, "map_root" :  {
"enabled" : true, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:root" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/rexray_vol" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [ "10.50.0.111" ], "commit_asynchronous" : false,
"conflicting_paths" : [], "description" : "", "directory_transfer_size" :
131072, "encoding" : "DEFAULT", "id" : 18, "link_max" : 32767, "map_failure" :  {
"enabled" : false, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "map_full" : true, "map_lookup_uid" : false,
"map_non_root" :  { "enabled" : true, "primary_group" : {}, "secondary_groups" : [],
"user" :  { "id" : "USER:rexray" } }, "map_retry" : true, "map_root" :  {
"enabled" : true, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:rexray" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/clint" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [ "10.50.0.111" ], "commit_asynchronous" : false,
"conflicting_paths" : [], "description" : "", "directory_transfer_size" :
131072, "encoding" : "DEFAULT", "id" : 20, "link_max" : 32767, "map_failure" :  {
"enabled" : false, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "map_full" : true, "map_lookup_uid" : false,
"map_non_root" :  { "enabled" : false, "primary_group" : {}, "secondary_groups" : [],
"user" :  { "id" : "USER:nobody" } }, "map_retry" : true, "map_root" :  {
"enabled" : true, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/chris" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 23, "link_max" : 32767, "map_failure" :  { "enabled" : false,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:nobody" } },
"map_full" : true, "map_lookup_uid" : false, "map_non_root" :  { "enabled" :
false, "primary_group" : {}, "secondary_groups" : [], "user" :  { "id" :
"USER:nobody" } }, "map_retry" : true, "map_root" :  { "enabled" : true,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:root" } },
"max_file_size" : 9223372036854775807, "name_max_size" : 255, "no_truncate" :
false, "paths" : [ "/ifs/root_test" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [
"10.50.0.111" ], "security_flavors" : [ "unix" ], "setattr_asynchronous" :
false, "snapshot" : "-", "symlinks" : true, "time_delta" :
1.000000000000000e-09, "unresolved_clients" : [], "write_datasync_action" :
"DATASYNC", "write_datasync_reply" : "DATASYNC", "write_filesync_action" :
"FILESYNC", "write_filesync_reply" : "FILESYNC", "write_transfer_max_size" :
1048576, "write_transfer_multiple" : 512, "write_transfer_size" : 524288,
"write_unstable_action" : "UNSTABLE", "write_unstable_reply" : "UNSTABLE",
"zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [ "10.50.0.111" ], "commit_asynchronous" : false,
"conflicting_paths" : [], "description" : "", "directory_transfer_size" :
131072, "encoding" : "DEFAULT", "id" : 24, "link_max" : 32767, "map_failure" :  {
"enabled" : false, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "map_full" : true, "map_lookup_uid" : false,
"map_non_root" :  { "enabled" : true, "primary_group" : {}, "secondary_groups" : [],
"user" :  { "id" : "USER:root" } }, "map_retry" : true, "map_root" :  {
"enabled" : true, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:root" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/testing" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" } ],

"resume" : null, "total" : 7 }`)

var getAllExports2JSON = []byte(`{ "exports" :  [

{ "all_dirs" : true, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "Default export", "directory_transfer_size" : 131072, "encoding" :
"DEFAULT", "id" : 1, "link_max" : 32767, "map_failure" :  { "enabled" : false,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:nobody" } },
"map_full" : true, "map_lookup_uid" : false, "map_non_root" :  { "enabled" :
false, "primary_group" : {}, "secondary_groups" : [], "user" :  { "id" :
"USER:nobody" } }, "map_retry" : true, "map_root" :  { "enabled" : true,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:nobody" } },
"max_file_size" : 9223372036854775807, "name_max_size" : 255, "no_truncate" :
false, "paths" : [ "/ifs" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [],
"security_flavors" : [ "unix" ], "setattr_asynchronous" : false, "snapshot" : "-",
"symlinks" : true, "time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 16, "link_max" : 32767, "map_failure" :  { "enabled" : false,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:nobody" } },
"map_full" : true, "map_lookup_uid" : false, "map_non_root" :  { "enabled" :
false, "primary_group" : {}, "secondary_groups" : [], "user" :  { "id" :
"USER:nobody" } }, "map_retry" : true, "map_root" :  { "enabled" : true,
"primary_group" :  { "id" : "GROUP:wheel" }, "secondary_groups" : [], "user" :  {
"id" : "USER:root" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/admin_test" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [ "10.50.0.111" ], "commit_asynchronous" : false,
"conflicting_paths" : [], "description" : "", "directory_transfer_size" :
131072, "encoding" : "DEFAULT", "id" : 17, "link_max" : 32767, "map_failure" :  {
"enabled" : false, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "map_full" : true, "map_lookup_uid" : false,
"map_non_root" :  { "enabled" : true, "primary_group" : {}, "secondary_groups" : [],
"user" :  { "id" : "USER:root" } }, "map_retry" : true, "map_root" :  {
"enabled" : true, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:root" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/rexray_vol" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [ "10.50.0.111" ], "commit_asynchronous" : false,
"conflicting_paths" : [], "description" : "", "directory_transfer_size" :
131072, "encoding" : "DEFAULT", "id" : 18, "link_max" : 32767, "map_failure" :  {
"enabled" : false, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "map_full" : true, "map_lookup_uid" : false,
"map_non_root" :  { "enabled" : true, "primary_group" : {}, "secondary_groups" : [],
"user" :  { "id" : "USER:rexray" } }, "map_retry" : true, "map_root" :  {
"enabled" : true, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:rexray" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/clint" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" : true,
"clients" : [ "10.50.0.111" ], "commit_asynchronous" : false,
"conflicting_paths" : [], "description" : "", "directory_transfer_size" :
131072, "encoding" : "DEFAULT", "id" : 20, "link_max" : 32767, "map_failure" :  {
"enabled" : false, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "map_full" : true, "map_lookup_uid" : false,
"map_non_root" :  { "enabled" : false, "primary_group" : {}, "secondary_groups" : [],
"user" :  { "id" : "USER:nobody" } }, "map_retry" : true, "map_root" :  {
"enabled" : true, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/chris" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 23, "link_max" : 32767, "map_failure" :  { "enabled" : false,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:nobody" } },
"map_full" : true, "map_lookup_uid" : false, "map_non_root" :  { "enabled" :
false, "primary_group" : {}, "secondary_groups" : [], "user" :  { "id" :
"USER:nobody" } }, "map_retry" : true, "map_root" :  { "enabled" : true,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:root" } },
"max_file_size" : 9223372036854775807, "name_max_size" : 255, "no_truncate" :
false, "paths" : [ "/ifs/root_test" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [
"10.50.0.111" ], "security_flavors" : [ "unix" ], "setattr_asynchronous" :
false, "snapshot" : "-", "symlinks" : true, "time_delta" :
1.000000000000000e-09, "unresolved_clients" : [], "write_datasync_action" :
"DATASYNC", "write_datasync_reply" : "DATASYNC", "write_filesync_action" :
"FILESYNC", "write_filesync_reply" : "FILESYNC", "write_transfer_max_size" :
1048576, "write_transfer_multiple" : 512, "write_transfer_size" : 524288,
"write_unstable_action" : "UNSTABLE", "write_unstable_reply" : "UNSTABLE",
"zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [ "10.50.0.111" ], "commit_asynchronous" : false,
"conflicting_paths" : [], "description" : "", "directory_transfer_size" :
131072, "encoding" : "DEFAULT", "id" : 24, "link_max" : 32767, "map_failure" :  {
"enabled" : false, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:nobody" } }, "map_full" : true, "map_lookup_uid" : false,
"map_non_root" :  { "enabled" : true, "primary_group" : {}, "secondary_groups" : [],
"user" :  { "id" : "USER:root" } }, "map_retry" : true, "map_root" :  {
"enabled" : true, "primary_group" : {}, "secondary_groups" : [], "user" :  {
"id" : "USER:root" } }, "max_file_size" : 9223372036854775807, "name_max_size" :
255, "no_truncate" : false, "paths" : [ "/ifs/volumes/libstorage/testing" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true,
"case_insensitive" : false, "case_preserving" : true, "chown_restricted" :
false, "clients" : [], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 26, "link_max" : 32767, "map_failure" :  { "enabled" : false,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:nobody" } },
"map_full" : true, "map_lookup_uid" : false, "map_non_root" :  { "enabled" :
false, "primary_group" : {}, "secondary_groups" : [], "user" :  { "id" :
"USER:nobody" } }, "map_retry" : true, "map_root" :  { "enabled" : true,
"primary_group" : {}, "secondary_groups" : [], "user" :  { "id" : "USER:root" } },
"max_file_size" : 9223372036854775807, "name_max_size" : 255, "no_truncate" :
false, "paths" : [ "/ifs/volumes/libstorage/chris" ], "read_only" : false,
"read_only_clients" : [], "read_transfer_max_size" : 1048576,
"read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [ "10.50.0.220" ],
"security_flavors" : [ "unix" ], "setattr_asynchronous" : false, "snapshot" : "-",
"symlinks" : true, "time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE", "zone" : "System" } ], "resume" : null,
"total" : 8 }`)

var getAllExports3JSON = []byte(`{ "exports" : [

{ "all_dirs" : true, "block_size" : 8192, "can_set_time" : true, "clients" : [],
"commit_asynchronous" : false, "conflicting_paths" : [], "description" :
"Default export", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 1, "map_all" : {}, "map_full" : true, "map_lookup_uid" : false,
"map_retry" : true, "map_root" : { "groups" : [ "" ], "user" : "nobody" },
"max_file_size" : 9223372036854775807, "paths" : [ "/ifs" ], "read_only" :
false, "read_only_clients" : [], "read_transfer_max_size" : 1048576,
"read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true, "clients" : [],
"commit_asynchronous" : false, "conflicting_paths" : [], "description" : "",
"directory_transfer_size" : 131072, "encoding" : "DEFAULT", "id" : 16, "map_all" : {},
"map_full" : true, "map_lookup_uid" : false, "map_retry" : true, "map_root" : {
"groups" : [ "wheel" ], "user" : "root" }, "max_file_size" :
9223372036854775807, "paths" : [ "/ifs/volumes/libstorage/admin_test" ],
"read_only" : false, "read_only_clients" : [], "read_transfer_max_size" :
1048576, "read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [], "security_flavors" : [
"unix" ], "setattr_asynchronous" : false, "snapshot" : "-", "symlinks" : true,
"time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true, "clients" : [
"10.50.0.111" ], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 17, "map_all" : { "groups" : [ "" ], "user" : "root" }, "map_full" :
true, "map_lookup_uid" : false, "map_retry" : true, "map_root" : {},
"max_file_size" : 9223372036854775807, "paths" : [
"/ifs/volumes/libstorage/rexray_vol" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [],
"security_flavors" : [ "unix" ], "setattr_asynchronous" : false, "snapshot" : "-",
"symlinks" : true, "time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true, "clients" : [
"10.50.0.111" ], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 18, "map_all" : { "groups" : [ "" ], "user" : "rexray" }, "map_full" :
true, "map_lookup_uid" : false, "map_retry" : true, "map_root" : {},
"max_file_size" : 9223372036854775807, "paths" : [
"/ifs/volumes/libstorage/clint" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [],
"security_flavors" : [ "unix" ], "setattr_asynchronous" : false, "snapshot" : "-",
"symlinks" : true, "time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true, "clients" : [
"10.50.0.111" ], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 20, "map_all" : {}, "map_full" : true, "map_lookup_uid" : false,
"map_retry" : true, "map_root" : { "groups" : [ "" ], "user" : "nobody" },
"max_file_size" : 9223372036854775807, "paths" : [
"/ifs/volumes/libstorage/chris" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [],
"security_flavors" : [ "unix" ], "setattr_asynchronous" : false, "snapshot" : "-",
"symlinks" : true, "time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true, "clients" : [],
"commit_asynchronous" : false, "conflicting_paths" : [], "description" : "",
"directory_transfer_size" : 131072, "encoding" : "DEFAULT", "id" : 23, "map_all" : {},
"map_full" : true, "map_lookup_uid" : false, "map_retry" : true, "map_root" : {
"groups" : [ "" ], "user" : "root" }, "max_file_size" : 9223372036854775807,
"paths" : [ "/ifs/root_test" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [
"10.50.0.111" ], "security_flavors" : [ "unix" ], "setattr_asynchronous" :
false, "snapshot" : "-", "symlinks" : true, "time_delta" :
1.000000000000000e-09, "unresolved_clients" : [], "write_datasync_action" :
"DATASYNC", "write_datasync_reply" : "DATASYNC", "write_filesync_action" :
"FILESYNC", "write_filesync_reply" : "FILESYNC", "write_transfer_max_size" :
1048576, "write_transfer_multiple" : 512, "write_transfer_size" : 524288,
"write_unstable_action" : "UNSTABLE", "write_unstable_reply" : "UNSTABLE" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true, "clients" : [
"10.50.0.111" ], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 24, "map_all" : { "groups" : [ "" ], "user" : "root" }, "map_full" :
true, "map_lookup_uid" : false, "map_retry" : true, "map_root" : {},
"max_file_size" : 9223372036854775807, "paths" : [
"/ifs/volumes/libstorage/testing" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [],
"security_flavors" : [ "unix" ], "setattr_asynchronous" : false, "snapshot" : "-",
"symlinks" : true, "time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true, "clients" : [],
"commit_asynchronous" : false, "conflicting_paths" : [], "description" : "",
"directory_transfer_size" : 131072, "encoding" : "DEFAULT", "id" : 26, "map_all" : {},
"map_full" : true, "map_lookup_uid" : false, "map_retry" : true, "map_root" : {
"groups" : [ "" ], "user" : "root" }, "max_file_size" : 9223372036854775807,
"paths" : [ "/ifs/volumes/libstorage/chris" ], "read_only" : false,
"read_only_clients" : [], "read_transfer_max_size" : 1048576,
"read_transfer_multiple" : 512, "read_transfer_size" : 131072,
"read_write_clients" : [], "readdirplus" : true, "readdirplus_prefetch" : 10,
"return_32bit_file_ids" : false, "root_clients" : [ "10.50.0.220" ],
"security_flavors" : [ "unix" ], "setattr_asynchronous" : false, "snapshot" : "-",
"symlinks" : true, "time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE" },

{ "all_dirs" : false, "block_size" : 8192, "can_set_time" : true, "clients" : [
"10.50.0.112" ], "commit_asynchronous" : false, "conflicting_paths" : [],
"description" : "", "directory_transfer_size" : 131072, "encoding" : "DEFAULT",
"id" : 52, "map_all" : { "groups" : [ "" ], "user" : "root" }, "map_full" :
true, "map_lookup_uid" : false, "map_retry" : true, "map_root" : {},
"max_file_size" : 9223372036854775807, "paths" : [
"/ifs/volumes/libstorage/redis-01" ], "read_only" : false, "read_only_clients" : [],
"read_transfer_max_size" : 1048576, "read_transfer_multiple" : 512,
"read_transfer_size" : 131072, "read_write_clients" : [], "readdirplus" : true,
"readdirplus_prefetch" : 10, "return_32bit_file_ids" : false, "root_clients" : [],
"security_flavors" : [ "unix" ], "setattr_asynchronous" : false, "snapshot" : "-",
"symlinks" : true, "time_delta" : 1.000000000000000e-09, "unresolved_clients" : [],
"write_datasync_action" : "DATASYNC", "write_datasync_reply" : "DATASYNC",
"write_filesync_action" : "FILESYNC", "write_filesync_reply" : "FILESYNC",
"write_transfer_max_size" : 1048576, "write_transfer_multiple" : 512,
"write_transfer_size" : 524288, "write_unstable_action" : "UNSTABLE",
"write_unstable_reply" : "UNSTABLE" } ], "resume" : null, "total" : 9 }`)
