# Testing the DigitalOcean triver

The tests for the DigitalOcean driver assume that you have access to a
DigitalOcean account that has a token that with read/write access. There are
[docs](https://www.digitalocean.com/community/tutorials/how-to-use-the-digitalocean-api-v2)
on how to do so if you do not already have one.

# Setting up an environment
The tests require a droplet running in a region that supports storage volumes.
You can find out which regions support volumes using the
[api](https://developers.digitalocean.com/documentation/v2/#regions) directly
or use the scripts included in the repo.

## Using Terraform
The scripts require that you have
[Terraform](https://github.com/hashicorp/terraform) available locally. You will
also to have access to a ssh key associated with your DigitalOcean account
([directions
here](https://www.digitalocean.com/community/tutorials/how-to-configure-ssh-key-based-authentication-on-a-linux-server#how-to-embed-your-public-key-when-creating-your-server)).
You will need to pass the ssh key id or fingerprint to the setup scripts so
that Terraform can spin up you droplet.

#### Starting an environment

```
cd drivers/storage/digitalocean/tests
./test-env-up $SSH_KEY_ID
```

#### Executing the tests
Once you have your environment set up, you can build and copy the tests to your running droplet.

Running the build comamand:
```
GOOS=linux GOARCH=amd64 BUILD_TAGS="gofig pflag libstorage_integration_docker libstorage_storage_driver libstorage_storage_executor libstorage_storage_driver_digitalocean libstorage_storage_executor_digitalocean" make build-tests
```
will create a `digitalocean.test` in the tests directory. You can scp that to
your droplet and then run the tests. You will also need to configure libstorage
to use the digitalocean driver by setting the following fields in
`/etc/libstorage/config.yaml`:

```
digitalocean:
  token: $YOUR_API_KEY
  # You can use other regions here
  region: sfo2
```

#### Deleting an environment
```
cd drivers/storage/digitalocean/tests
./test-env-down $SSH_KEY_ID
```
