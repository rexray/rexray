package cli

import (
	"regexp"

	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
)

type lsVolumesResult struct {
	vols             []*apitypes.Volume
	iid              *apitypes.InstanceID
	volMatchType     map[*apitypes.Volume]matchTypes
	volMatchPatt     map[*apitypes.Volume]string
	matchTypeCount   map[matchTypes]int
	matchPattCount   map[string]int
	uniqStrMatchID   map[string]bool
	uniqStrMatchName map[string]bool
	volMatchVals     map[*apitypes.Volume]string
}

type volumeWithPath struct {
	*apitypes.Volume
	Path string
}

type matchTypes int

const (
	matchTypeGlob matchTypes = iota
	matchTypeExactID
	matchTypeExactIDIgnoreCase
	matchTypePartialID
	matchTypePartialIDIgnoreCase
	matchTypeExactName
	matchTypeExactNameIgnoreCase
	matchTypePartialName
	matchTypePartialNameIgnoreCase
	matchTypeRegexpID
	matchTypeRegexpIDIgnoreCase
	matchTypeRegexpName
	matchTypeRegexpNameIgnoreCase
)

type matchedVolume struct {
	*apitypes.Volume
	matchType matchTypes
}

type regexpPair struct {
	*regexp.Regexp
	ignoreCase *regexp.Regexp
}
