resource digitalocean_volume "volume" {
  region = "${var.region}"
  size   = "${var.volume_size}"
  name   = "libstorage-volume"
}

data "template_file" "cloudinit" {
  template = "${file("${path.module}/templates/cloudinit.tpl")}"

  vars = {
    region = "${var.region}"
  }
}

resource digitalocean_droplet "droplet" {
  region = "${var.region}"
  size   = "${var.size}"
  image  = "${var.image}"
  name   = "libstorage-integration"

  user_data = "${data.template_file.cloudinit.rendered}"

  ssh_keys   = ["${var.ssh_key}"]
  volume_ids = ["${digitalocean_volume.volume.id}"]
}
