package v1

type IsiVolume struct {
	Name         string `json:"name"`
	AttributeMap []struct {
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
	} `json:"attrs"`
}

// Isi PAPI volume JSON structs
type VolumeName struct {
	Name string `json:"name"`
}

type getIsiVolumesResp struct {
	Children []*VolumeName `json:"children"`
}

// Isi PAPI Volume ACL JSON structs
type Ownership struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AclRequest struct {
	Authoritative string     `json:"authoritative"`
	Action        string     `json:"action"`
	Owner         *Ownership `json:"owner"`
	Group         *Ownership `json:"group,omitempty"`
}

// Isi PAPI volume attributes JSON struct
type getIsiVolumeAttributesResp struct {
	AttributeMap []struct {
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
	} `json:"attrs"`
}

// Isi PAPI export path JSON struct
type ExportPathList struct {
	Paths  []string `json:"paths"`
	MapAll struct {
		User   string   `json:"user"`
		Groups []string `json:"groups,omitempty"`
	} `json:"map_all"`
}

// Isi PAPI export clients JSON struct
type ExportClientList struct {
	Clients []string `json:"clients"`
}

// Isi PAPI export Id JSON struct
type postIsiExportResp struct {
	Id int `json:"id"`
}

// Isi PAPI export attributes JSON structs
type IsiExport struct {
	Id      int      `json:"id"`
	Paths   []string `json:"paths"`
	Clients []string `json:"clients"`
}

type getIsiExportsResp struct {
	ExportList []*IsiExport `json:"exports"`
}

// Isi PAPI snapshot path JSON struct
type SnapshotPath struct {
	Path string `json:"path"`
	Name string `json:"name,omitempty"`
}

// Isi PAPI snapshot JSON struct
type IsiSnapshot struct {
	Created       int64   `json:"created"`
	Expires       int64   `json:"expires"`
	HasLocks      bool    `json:"has_locks"`
	Id            int64   `json:"id"`
	Name          string  `json:"name"`
	Path          string  `json:"path"`
	PctFilesystem float64 `json:"pct_filesystem"`
	PctReserve    float64 `json:"pct_reserve"`
	Schedule      string  `json:"schedule"`
	ShadowBytes   int64   `json:"shadow_bytes"`
	Size          int64   `json:"size"`
	State         string  `json:"state"`
	TargetId      int64   `json:"target_it"`
	TargetName    string  `json:"target_name"`
}

type getIsiSnapshotsResp struct {
	SnapshotList []*IsiSnapshot `json:"snapshots"`
	Total        int64          `json:"total"`
	Resume       string         `json:"resume"`
}

type isiThresholds struct {
	Advisory             int64       `json:"advisory"`
	AdvisoryExceeded     bool        `json:"advisory_exceeded"`
	AdvisoryLastExceeded interface{} `json:"advisory_last_exceeded"`
	Hard                 int64       `json:"hard"`
	HardExceeded         bool        `json:"hard_exceeded"`
	HardLastExceeded     interface{} `json:"hard_last_exceeded"`
	Soft                 int64       `json:"soft"`
	SoftExceeded         bool        `json:"soft_exceeded"`
	SoftLastExceeded     interface{} `json:"soft_last_exceeded"`
}

type IsiQuota struct {
	Container                 bool          `json:"container"`
	Enforced                  bool          `json:"enforced"`
	Id                        string        `json:"id"`
	IncludeSnapshots          bool          `json:"include_snapshots"`
	Linked                    interface{}   `json:"linked"`
	Notifications             string        `json:"notifications"`
	Path                      string        `json:"path"`
	Persona                   interface{}   `json:"persona"`
	Ready                     bool          `json:"ready"`
	Thresholds                isiThresholds `json:"thresholds"`
	ThresholdsIncludeOverhead bool          `json:"thresholds_include_overhead"`
	Type                      string        `json:"type"`
	Usage                     struct {
		Inodes   int64 `json:"inodes"`
		Logical  int64 `json:"logical"`
		Physical int64 `json:"physical"`
	} `json:"usage"`
}

type isiThresholdsReq struct {
	Advisory interface{} `json:"advisory"`
	Hard     interface{} `json:"hard"`
	Soft     interface{} `json:"soft"`
}

type IsiQuotaReq struct {
	Enforced                  bool             `json:"enforced"`
	IncludeSnapshots          bool             `json:"include_snapshots"`
	Path                      string           `json:"path"`
	Thresholds                isiThresholdsReq `json:"thresholds"`
	ThresholdsIncludeOverhead bool             `json:"thresholds_include_overhead"`
	Type                      string           `json:"type"`
}

type IsiUpdateQuotaReq struct {
	Enforced                  bool             `json:"enforced"`
	Thresholds                isiThresholdsReq `json:"thresholds"`
	ThresholdsIncludeOverhead bool             `json:"thresholds_include_overhead"`
}

type isiQuotaListResp struct {
	Quotas []IsiQuota `json:"quotas"`
}
