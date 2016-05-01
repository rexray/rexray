package schema

const (
	// JSONSchema is the libStorage API JSON schema
	JSONSchema = `{
    "id": "https://github.com/emccode/libstorage",
    "$schema": "http://json-schema.org/draft-04/schema#",
    "definitions": {
        "volume": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string",
                    "description": "The volume ID."
                },
                "name": {
                    "type": "string",
                    "description": "The volume name."
                },
                "type": {
                    "type": "string",
                    "description": "The volume type."
                },
                "attachments": {
                    "type": "array",
                    "description": "The volume's attachments.",
                    "items": { "$ref": "#/definitions/volumeAttachment" }
                },
                "availabilityZone": {
                    "type": "string",
                    "description": "The zone for which the volume is available."
                },
                "iops": {
                    "type": "number",
                    "description": "The volume IOPs."
                },
                "networkName": {
                    "type": "string",
                    "description": "The name of the network on which the volume resides."
                },
                "size": {
                    "type": "number",
                    "description": "The volume size (GB)."
                },
                "status": {
                    "type": "string",
                    "description": "The volume status."
                },
                "fields": { "$ref": "#/definitions/fields" }
            },
            "required": [ "id", "name", "size" ],
            "additionalProperties": false
        },


        "volumeAttachment": {
            "type": "object",
            "properties": {
                "instanceID": { "$ref": "#/definitions/instanceID" },
                "deviceName": {
                    "type": "string",
                    "description": "The name of the device on to which the volume is mounted."
                },
                "status": {
                    "type": "string",
                    "description": "The status of the attachment."
                },
                "volumeID": {
                    "type": "string",
                    "description": "The ID of the volume to which the attachment belongs."
                },
                "mountPoint": {
                    "type": "string",
                    "description": "The file system path to which the volume is mounted."
                },
                "fields": { "$ref": "#/definitions/fields" }
            },
            "required": [ "instanceID", "deviceName", "volumeID" ],
            "additionalProperties": false
        },


        "instanceID": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string",
                    "description": "The instance ID."
                },
                "metadata": {
                    "type": "object",
                    "description": "Extra information about the instance ID."
                }
            },
            "required": [ "id" ],
            "additionalProperties": false
        },


        "instance": {
            "type": "object",
            "properties": {
                "instanceID": { "$ref": "#/definitions/instanceID" },
                "name": {
                    "type": "string",
                    "description": "The name of the instance."
                },
                "providerName": {
                    "type": "string",
                    "description": "The name of the provider that owns the object."
                },
                "region": {
                    "type": "string",
                    "description": "The region from which the object originates."
                },
                "fields": { "$ref": "#/definitions/fields" }
            },
            "required": [ "id" ],
            "additionalProperties": false
        },


        "snapshot": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string",
                    "description": "The snapshot's ID."
                },
                "name": {
                    "type": "string",
                    "description": "The name of the snapshot."
                },
                "description": {
                    "type": "string",
                    "description": "A description of the snapshot."
                },
                "startTime": {
                    "type": "number",
                    "description": "The time (epoch) at which the request to create the snapshot was submitted."
                },
                "status": {
                    "type": "string",
                    "description": "The status of the snapshot."
                },
                "volumeID": {
                    "type": "string",
                    "description": "The ID of the volume to which the snapshot belongs."
                },
                "volumeSize": {
                    "type": "number",
                    "description": "The size of the volume to which the snapshot belongs."
                },
                "fields": { "$ref": "#/definitions/fields" }
            },
            "required": [ "id" ],
            "additionalProperties": false
        },


        "task": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "number",
                    "description": "The task's unique identifier."
                },
                "name": {
                    "type": "string",
                    "description": "The name of the task."
                },
                "user": {
                    "type": "string",
                    "description": "The name of the user that created the task."
                },
                "completeTime": {
                    "type": "number",
                    "description": "The time stamp (epoch) when the task was completed."
                },
                "queueTime": {
                    "type": "number",
                    "description": "The time stamp (epoch) when the task was created."
                },
                "startTime": {
                    "type": "number",
                    "description": "The time stamp (epoch) when the task started running."
                },
                "result": {
                    "type": "object",
                    "description": "The result of the operation."
                },
                "error": {
                    "type": "object",
                    "description": "If the operation returned an error, this is it."
                },
                "fields": { "$ref": "#/definitions/fields" }
            },
            "required": [ "id", "name",  "user", "queueTime" ],
            "additionalProperties": false
        },


        "serviceInfo": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string",
                    "description": "Name is the service's name."
                },
                "instance": { "$ref": "#/definitions/instance" },
                "driver": { "$ref": "#/definitions/driverInfo" }
            },
            "required": [ "name", "driver" ],
            "additionalProperties": false
        },


        "driverInfo": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string",
                    "description": "Ignore is a flag that indicates whether the client logic should invoke the GetNextAvailableDeviceName function prior to submitting an AttachVolume request to the server."
                },
                "type": {
                    "type": "string",
                    "description": "Type is the type of storage the driver provides: block, nas, object."
                },
                "nextDevice": { "$ref": "#/definitions/nextDeviceInfo" }
            },
            "required": [ "name", "type" ],
            "additionalProperties": false
        },


        "executorInfo": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string",
                    "description": "The name of the executor."
                },
                "md5checksum": {
                    "type": "string",
                    "description": "The file's MD5 checksum. This can be used to determine if a local copy of the executor needs to be updated."
                },
                "size": {
                    "type": "number",
                    "description": "The size of the executor, in bytes."
                },
                "lastModified": {
                    "type": "number",
                    "description": "The time the executor was last modified as an epoch."
                }
            },
            "required": [ "name", "md5checksum", "size", "lastModified" ],
            "additionalProperties": false
        },


        "nextDeviceInfo": {
            "type": "object",
            "properties": {
                "ignore": {
                    "type": "boolean",
                    "description": "Ignore is a flag that indicates whether the client logic should invoke the GetNextAvailableDeviceName function prior to submitting an AttachVolume request to the server."
                },
                "prefix": {
                    "type": "string",
                    "description": "Prefix is the first part of a device path's value after the \"/dev/\" portion. For example, the prefix in \"/dev/xvda\" is \"xvd\"."
                },
                "pattern": {
                    "type": "string",
                    "description": "Pattern is the regex to match the part of a device path after the prefix."
                }
            },
            "additionalProperties": false
        },


        "fields": {
            "type": "object",
            "description": "Fields are additional properties that can be defined for this type.",
            "patternProperties": {
                ".+": { "type": "string" }
            },
            "additionalProperties": true
        },


        "volumeMap": {
            "type": "object",
            "patternProperties": {
                "^.+$": { "$ref": "#/definitions/volume" }
            },
            "additionalProperties": false
        },


        "snapshotMap": {
            "type": "object",
            "patternProperties": {
                "^.+$": { "$ref": "#/definitions/snapshot" }
            },
            "additionalProperties": false
        },


        "taskMap": {
            "type": "object",
            "patternProperties": {
                "^.+$": { "$ref": "#/definitions/task" }
            },
            "additionalProperties": false
        },


        "serviceVolumeMap": {
            "type": "object",
            "patternProperties": {
                "^.+$": { "$ref": "#/definitions/volumeMap" }
            },
            "additionalProperties": false
        },


        "serviceSnapshotMap": {
            "type": "object",
            "patternProperties": {
                "^.+$": { "$ref": "#/definitions/snapshotMap" }
            },
            "additionalProperties": false
        },


        "serviceTaskMap": {
            "type": "object",
            "patternProperties": {
                "^.+$": { "$ref": "#/definitions/taskMap" }
            },
            "additionalProperties": false
        },


        "serviceInfoMap": {
            "type": "object",
            "patternProperties": {
                "^.+$": { "$ref": "#/definitions/serviceInfo" }
            },
            "additionalProperties": false
        },


        "executorInfoMap": {
            "type": "object",
            "patternProperties": {
                "^.+$": { "$ref": "#/definitions/executorInfo" }
            },
            "additionalProperties": false
        },


        "driverInfoMap": {
            "type": "object",
            "patternProperties": {
                "^.+$": { "$ref": "#/definitions/driverInfo" }
            },
            "additionalProperties": false
        },


        "opts": {
            "type": "object",
            "description": "Opts are additional properties that can be defined for POST requests.",
            "patternProperties": {
                "^.+$": {
                    "anyOf": [
                        { "type": "array" },
                        { "type": "boolean" },
                        { "type": "integer" },
                        { "type": "number" },
                        { "type": "null" },
                        { "type": "string" },
                        { "$ref": "#/definitions/opts" }
                    ]
                }
            },
            "additionalProperties": true
        },


        "volumeCreateRequest": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "availabilityZone": {
                    "type": "string"
                },
                "iops": {
                    "type": "number"
                },
                "size": {
                    "type": "number"
                },
                "type": {
                    "type": "string"
                },
                "opts": { "$ref" : "#/definitions/opts" }
            },
            "required": [ "name" ],
            "additionalProperties": false
        },


        "volumeCopyRequest": {
            "type": "object",
            "properties": {
                "volumeName": {
                    "type": "string"
                },
                "opts": { "$ref" : "#/definitions/opts" }
            },
            "required": [ "volumeName" ],
            "additionalProperties": false
        },


        "volumeSnapshotRequest": {
            "type": "object",
            "properties": {
                "snapshotName": {
                    "type": "string"
                },
                "opts": { "$ref" : "#/definitions/opts" }
            },
            "required": [ "snapshotName" ],
            "additionalProperties": false
        },


        "volumeAttachRequest": {
            "type": "object",
            "properties": {
                "nextDeviceName": {
                    "type": "string"
                },
                "force": {
                    "type": "boolean"
                },
                "opts": { "$ref" : "#/definitions/opts" }
            },
            "additionalProperties": false
        },


        "volumeDetachRequest": {
            "type": "object",
            "properties": {
                "force": {
                    "type": "boolean"
                },
                "opts": { "$ref" : "#/definitions/opts" }
            },
            "additionalProperties": false
        },


        "snapshotCopyRequest": {
            "type": "object",
            "properties": {
                "snapshotName": {
                    "type": "string"
                },
                "destinationID": {
                    "type": "string"
                },
                "opts": { "$ref" : "#/definitions/opts" }
            },
            "required": [ "snapshotName", "destinationID" ],
            "additionalProperties": false
        },


        "snapshotRemoveRequest": {
            "type": "object",
            "properties": {
                "opts": { "$ref" : "#/definitions/opts" }
            },
            "additionalProperties": false
        },


        "error": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string",
                    "pattern": "^.{10,}|.*[Ee]rror$"
                },
                "status": {
                    "type": "number",
                    "minimum": 400,
                    "maximum": 599
                },
                "error": {
                    "type": "object",
                    "additionalProperties": true
                }
            },
            "required": [ "message", "status" ],
            "additionalProperties": false
        }
    }
}
`
)
