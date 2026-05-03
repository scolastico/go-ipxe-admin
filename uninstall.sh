#!/bin/bash

if [ -n "$INSTALL_DIR" ]; then
    INSTALL_DIR="$INSTALL_DIR"
else
    INSTALL_DIR="/srv/go-ipxe-admin"
fi

if [ -f "/usr/bin/systemctl" ]; then
    systemctl stop go-ipxe-admin
    systemctl disable go-ipxe-admin
    rm /etc/systemd/system/go-ipxe-admin.service
    systemctl daemon-reload
fi

if [ -f "/etc/init.d/go-ipxe-admin" ]; then
    /etc/init.d/go-ipxe-admin stop
    rm /etc/init.d/go-ipxe-admin
    update-rc.d go-ipxe-admin remove
fi

rm "$INSTALL_DIR/go-ipxe-admin"
echo "Do you want to remove the data directory? (y/n)"
read remove_data

if [ "$remove_data" = "y" ]; then
    rm -rf /srv/go-ipxe-admin/data
fi

exit 0
