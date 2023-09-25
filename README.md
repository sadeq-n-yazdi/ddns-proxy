# Fetch It service

This service is used to update Dynamic DNS names using Google DNS service that is censored in Iran.

**Note:** To use this service you need have a domain with Google Domain service that DNS service set to use Google DNS service.

# How to build
This service is Go pure code, so you need go to compile it 
1. Install [Go 1.21.1+](https://go.dev/dl)
2. run `go get -v -u` to install and update dependencies
3. run `GOOS=linux GOARCH=amd64 go build -v -ldflags "-w -s" -o MY-SERVICE-NAME.app` to compile it for Linux on amd64 platform, check [Go Documentation](https://go.dev/doc/install/source#environment) for more information to build for other platforms
4. (optional) use UPX to reduce the size by running `upx MY-SERVICE-NAME.app`

# How Use

Step-by-step guide to deploy the service:

- Go to domains.Google.com and select your domain
- Add a DDNS record in DNS → Advanced Settings → Manage Dynamic DNS
- Save the subdomain name
- Copy credentials using DNS → Advanced Settings → Your domain has Dynamic DNS set up → View Credentials
- Update credential settings of service using `cred-sample.jsonc` as template and upload it to your VPS /etc/websites/YOUR_DOMAIN_NAME/cred.json
- Update the configuration of service using `config.ini` as you need
- Upload it to /etc/websites/YOUR_DOMAIN_NAME/config.ini
- run the service (or set up a service using systemd or a daemon, see `sample-service.service` for a sample systemd service implementation)

## TODO

This functionality will be added in the future:
1. Support listen to secure port
2. Add test codes 
3. Support generate and use certificate using Let's Encrypt services
4. add support for:
    - no-ip.com
    - dyndns.org
    - ...
5. write a man file

# Author

- Sadeq N. Yazdi <code>code[@]sadeq[dot]uk></code>

# Copyright
see [GPLv3](https://www.gnu.org/licenses/gpl-3.0.html)
https://www.gnu.org/licenses/gpl-3.0.txt

