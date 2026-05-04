#!/bin/sh

if [ -n "$INSTALL_DIR" ]; then
    INSTALL_DIR="$INSTALL_DIR"
else
    INSTALL_DIR="/srv/go-ipxe-admin"
fi

if [ -f "/usr/bin/systemctl" ] || [ -f "/bin/systemctl" ]; then
    systemctl stop go-ipxe-admin
    systemctl disable go-ipxe-admin
    rm -f /etc/systemd/system/go-ipxe-admin.service
    systemctl daemon-reload
fi

if [ -f "/etc/init.d/go-ipxe-admin" ]; then
    /etc/init.d/go-ipxe-admin stop
    rm -f /etc/init.d/go-ipxe-admin
    if command -v update-rc.d >/dev/null 2>&1; then
        update-rc.d go-ipxe-admin remove
    fi
fi

rm -f "$INSTALL_DIR/go-ipxe-admin"
printf "Do you want to remove the data directory? (y/n) "
read -r remove_data </dev/tty

if [ "$remove_data" = "y" ]; then
    rm -rf "$INSTALL_DIR/data"
fi

exit 0
