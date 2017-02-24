#cloud-config
write_files:
  - path: /home/core/.bashrc
    permissions: 0644
    owner: 'core:core'
    content: |
      export DIGITALOCEAN_REGION=${region}
