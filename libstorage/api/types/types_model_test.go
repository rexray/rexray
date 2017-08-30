package types

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestVolumeMarshalToYAML(t *testing.T) {

	v := &Volume{
		ID:   "vol-000",
		Name: "Volume 000",
		Attachments: []*VolumeAttachment{
			&VolumeAttachment{
				InstanceID: &InstanceID{
					ID:     "hi",
					Driver: "vfs",
				},
				VolumeID: "vol-000",
			},
		},
	}

	out, err := yaml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(out))

	i := &Instance{
		InstanceID: &InstanceID{
			ID:     "hi",
			Driver: "vfs",
		},
	}

	out, err = yaml.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(out))
}

func TestInstanceMarshalToYAML(t *testing.T) {

	iid := &InstanceID{
		ID:     "hi",
		Driver: "vfs",
	}
	iid.MarshalMetadata(map[string]interface{}{
		"key1": "val1",
		"key2": 2,
	})

	i := &Instance{
		Name:       "MyVFSInstance",
		InstanceID: iid,
	}

	out, err := yaml.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(out))
}

func TestInstanceWithOnlyInstanceIDMarshalToYAML(t *testing.T) {

	i := &Instance{
		InstanceID: &InstanceID{
			ID:     "hi",
			Driver: "vfs",
		},
	}

	out, err := yaml.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(out))
}
