#!/bin/bash

# DEFAULT SETTINGS
SERVER="user@example.sadeq.uk"
SERVER_SSH_PARAM=""
SERVER_SCP_PARAM=""
APP_DEPLOY_NAME="example.sadeq.uk"


[ -f "$HOME/.colors.sh" ] && . "$HOME/.colors.sh"

# USER DEPLOY CONFIG FILE
[ -f "$HOME/.deploy.conf" ] && . "$HOME/.deploy.conf"

# PROJECT DEPLOY CONFIG FILE
[ -f "./deploy.conf" ] && . "./deploy.conf"
APP_NAME=$(basename $(realpath .))

SERVICE_NAME="website@$APP_DEPLOY_NAME"
[ -f "$APP_NAME" ] && rm -f "$APP_NAME" 2>/dev/null

go get -u
go build -ldflags "-w -s" -o "$APP_NAME" && echo "Build completed." && \
upx "$APP_NAME" && echo "Compress app completed." && \
ssh $SERVER_SSH_PARAM "$SERVER" "systemctl stop '$SERVICE_NAME'; [ -f '/opt/websites/${APP_DEPLOY_NAME}/${APP_DEPLOY_NAME}.app' ] && mv '/opt/websites/${APP_DEPLOY_NAME}/${APP_DEPLOY_NAME}.app' '/opt/websites/${APP_DEPLOY_NAME}/${APP_DEPLOY_NAME}.app.old' && echo 'Backup old app completed'; exit 0"
scp $SERVER_SCP_PARAM "$APP_NAME" "$SERVER:/opt/websites/${APP_DEPLOY_NAME}/${APP_DEPLOY_NAME}.app"  && echo 'copy new service app file completed' && \
ssh $SERVER_SSH_PARAM "$SERVER" "systemctl daemon-reload; systemctl enable '${SERVICE_NAME}' ; systemctl start '${SERVICE_NAME}'" && echo 'start service completed' && \
echo -e "${BGREEN}Deploy successfully done${NC}" || echo -e "${BRED}Deploy failed${NC}"

