variable region {
  description = <<EOF
The region to launch in

This *must* be a region that supports volumes!
EOF

  default = "sfo2"
}

variable size {
  description = "The size of the droplet"
  default     = "2gb"
}

variable image {
  description = "The image to use for the droplet"
  default     = "coreos-stable"
}

variable volume_size {
  description = "The size in GiB of the volume"
  default     = "10"
}

variable ssh_key {
  description = "A ssh key id or fingerprint of a key configured on your DigitalOcean account"
}
