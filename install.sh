#!/bin/bash

if [ -n "$INSTALL_DIR" ]; then
    INSTALL_DIR="$INSTALL_DIR"
else
    INSTALL_DIR="/srv/go-ipxe-admin"
fi

const_arch=$(uname -m)
case "$const_arch" in
    386) const_arch="386" ;;
    amd64) const_arch="amd64" ;;
    arm) const_arch="arm" ;;
    arm64) const_arch="arm64" ;;
    *) echo "Unsupported architecture: $const_arch" ; exit 1 ;;
esac
echo "Installing for architecture $const_arch"

mkdir -p "$INSTALL_DIR"
curl -L "https://raw.githubusercontent.com/scolastico/go-ipxe-admin/main/bin/go-ipxe-admin-linux-$const_arch" -o "$INSTALL_DIR/go-ipxe-admin"
chmod +x "$INSTALL_DIR/go-ipxe-admin"

# check if systemd is used
if [ -f "/usr/bin/systemctl" ]; then
    cat << EOF > /etc/systemd/system/go-ipxe-admin.service
[Unit]
Description=go-ipxe-admin
After=network.target

[Service]
ExecStart=$INSTALL_DIR/go-ipxe-admin
Restart=always
User=root
Group=root

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable go-ipxe-admin
    systemctl start go-ipxe-admin
else
    cat << EOF > /etc/init.d/go-ipxe-admin
#!/bin/sh
### BEGIN INIT INFO
# Provides:          go-ipxe-admin
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: go-ipxe-admin
# Description:       go-ipxe-admin
### END INIT INFO

NAME=go-ipxe-admin
DAEMON="$INSTALL_DIR/go-ipxe-admin"
PIDFILE=/var/run/$NAME.pid
SCRIPTNAME="/etc/init.d/$NAME"

# Exit on not found
[ -x "$DAEMON" ] || exit 0

# Load the VERBOSE setting and other rcS variables
. /lib/init/vars.sh

# Load the functions
. /lib/lsb/init-functions

# Start the service
start_service() {
    start_daemon "$DAEMON"
}

# Stop the service
stop_service() {
    stop_daemon "$DAEMON"
}

case "$1" in
    start)
        [ "$VERBOSE" != no ] && log_daemon_msg "Starting $NAME" "$NAME"
        start_service
        log_end_msg $?
        ;;
    stop)
        [ "$VERBOSE" != no ] && log_daemon_msg "Stopping $NAME" "$NAME"
        stop_service
        log_end_msg $?
        ;;
    restart|force-reload)
        log_daemon_msg "Restarting $NAME" "$NAME"
        stop_service
        start_service
        log_end_msg $?
        ;;
    status)
        status_of_proc "$DAEMON" "$NAME" && exit 0 || exit 4
        ;;
    *)
        echo "Usage: $SCRIPTNAME {start|stop|restart|force-reload|status}"
        exit 1
        ;;
esac

exit 0
EOF

chmod +x /etc/init.d/go-ipxe-admin
update-rc.d go-ipxe-admin defaults
/etc/init.d/go-ipxe-admin start
fi

exit 0
