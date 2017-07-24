# Testing the DigitalOcean Block Storage (DOBS) Driver

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
GOOS=linux make test-dobs
```
will create a `dobs.test` in the tests directory. You can scp that to
your droplet and then run the tests. You will also need to configure libStorage
to use the dobs driver by exportiing the required config parameters:

```bash
export DOBS_TOKEN=<your API token>
export DOBS_REGION=<region you are testing in>

```

The tests may now be executed with the following command:

```bash
./dobs.test
```

An exit code of `0` means the tests completed successfully. If there are errors
then it may be useful to run the tests once more with increased logging:

```bash
LIBSTORAGE_LOGGING_LEVEL=debug ./dobs.test -test.v
```

#### Deleting an environment
```
cd drivers/storage/digitalocean/tests
./test-env-down $SSH_KEY_ID
```
