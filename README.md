# Fetch It service

Using this, people can update Dynamic DNS names (Now, only Google DNS service) that are censored in Iran.

**Note:** 
To use this service, As well as a VPS, you need to have a domain with Google Domain service 
You have to set the DNS servers to use Google DNS service (default DNS server settings).

# How to build
This service is a Go pure code, so you just need to go to compile it. 
1. Install [Go 1.21.1+](https://go.dev/dl)
2. run `go get -v -u` to install and update dependencies
3. run `GOOS=linux GOARCH=amd64 go build -v -ldflags "-w -s" -o MY-SERVICE-NAME.app` to compile it for Linux on amd64 platforms, 
check [Go Documentation](https://go.dev/doc/install/source#environment) for more information to build for other platforms. 
4. (optional) use UPX to reduce the size by running `upx MY-SERVICE-NAME.app`

Also, You can use the `deploy.sh` bash script to compile and deploy the service on your VPS using SSH and SCP tools. 

# How to Use

Step-by-step guide to deploy the service:

- on Google Domain:
  - Go to [domains.Google.com](https://domains.google/) and select your domain
  - Add a subdomain with DDNS record by following these steps: `DNS` → `Advanced Settings` → `Manage Dynamic DNS`.
  - Save the subdomain name
  - Copy the credentials by following these steps: `DNS` → `Advanced Settings` → `Your domain has Dynamic DNS set up` → `View Credentials`.
- On the VPS:
  - Update the credential data file. Use `cred-sample.jsonc` as template.
    - file format is `jsonc`. Format is so like the `json` but can comment lines using `//` at the beginning of line.
    - every domain needs one entry in the config file with a unique name. that will be used as username for auth request.
    - the `password` is the password you need to use it in your auth request.
    - the `host` is the subdomain name.
    - the `dd-user` and `dd-pass` are the Google DNS DDNS service credential.  
  - Then upload the credential file to the VPS in `/etc/websites/YOUR_DOMAIN_NAME` folder
  - Update the configuration of service using `config.ini` as you need
  - Upload it to `/etc/websites/YOUR_DOMAIN_NAME/config.ini`
  - start the app to serve your requests (or set up a service using systemd or a daemon, see `sample-service.service` for a sample systemd service implementation)
- on the client:
  - you need to use `curl`, `wget` or any other get request to fetch the VPS service. 
    - to set the IP address manually, add `?myip=192.168.1.1`
    - if you do not set the IP address manually in the request, the service detects your public IP address.
    - the service checks the current IP address before requesting Google DDNS to update the IP of record. 
      - to force update, 
        - for all accounts: add the `force=yes` to the service config file
        - for only one of requests: add the `force=yes` to the request URL parameters (not implemented yet)
    - Using the TLS is strongly suggested for your safety

## TODO

This functionality will be added in the future:
1. [x] Do not try to update if the IP is already correct.
2. [x] Support listen to secure port
3. [ ] Add test codes 
4. [ ] Support generate and use certificate using Let's Encrypt services
5. [ ] add support for:
    - [ ] no-ip.com
    - [ ] dyndns.org
    - ...
6. [ ] write a man file
7. [ ] add support `force=yes` parameter in the request url
8. [ ] render `/about` page from MarkDown to HTML

# Author

- Sadeq N. Yazdi <code>code[@]sadeq[dot]uk></code>

# Copyright
see [GPLv3](https://www.gnu.org/licenses/gpl-3.0.html)
https://www.gnu.org/licenses/gpl-3.0.txt

