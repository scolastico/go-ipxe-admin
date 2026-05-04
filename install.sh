#!/bin/sh

if [ -n "$INSTALL_DIR" ]; then
    INSTALL_DIR="$INSTALL_DIR"
else
    INSTALL_DIR="/srv/go-ipxe-admin"
fi

const_arch=$(uname -m)
case "$const_arch" in
    386|i386|i686) const_arch="386" ;;
    amd64|x86_64) const_arch="amd64" ;;
    armv7l) const_arch="armv7" ;;
    arm64|aarch64) const_arch="arm64" ;;
    *) echo "Unsupported architecture: $const_arch" ; exit 1 ;;
esac
echo "Detected architecture $const_arch"

mkdir -p "$INSTALL_DIR"
UPGRADE=0
if [ -f "$INSTALL_DIR/go-ipxe-admin" ]; then
    UPGRADE=1
    if [ -f "/etc/init.d/go-ipxe-admin" ]; then
        /etc/init.d/go-ipxe-admin stop
    elif [ -f "/usr/bin/systemctl" ] || [ -f "/bin/systemctl" ]; then
        systemctl stop go-ipxe-admin
    else
        /etc/init.d/go-ipxe-admin stop
    fi
    rm -f "$INSTALL_DIR/go-ipxe-admin"
fi

echo "Downloading go-ipxe-admin for architecture $const_arch..."
curl -L https://github.com/scolastico/go-ipxe-admin/raw/refs/heads/main/bin/go-ipxe-admin-linux-$const_arch -o "$INSTALL_DIR/go-ipxe-admin"
chmod +x "$INSTALL_DIR/go-ipxe-admin"

if [ "$UPGRADE" -eq 1 ]; then
    echo "Restarting go-ipxe-admin..."
    if [ -f "/etc/init.d/go-ipxe-admin" ]; then
        /etc/init.d/go-ipxe-admin start
    elif [ -f "/usr/bin/systemctl" ] || [ -f "/bin/systemctl" ]; then
        systemctl start go-ipxe-admin
    else
        /etc/init.d/go-ipxe-admin start
    fi
    exit 0
fi

printf "Enter admin username [admin]: "
read -r ADMIN_USER </dev/tty
ADMIN_USER=${ADMIN_USER:-admin}

printf "Enter admin password [admin]: "
read -r ADMIN_PASS </dev/tty
ADMIN_PASS=${ADMIN_PASS:-admin}

# save environment
cat << EOF > "$INSTALL_DIR/.env"
ADMIN_USER="$ADMIN_USER"
ADMIN_PASS="$ADMIN_PASS"
EOF

# check init system
if [ -f "/sbin/procd" ]; then
    # OpenWrt procd
    cat << EOF > /etc/init.d/go-ipxe-admin
#!/bin/sh /etc/rc.common
START=99
USE_PROCD=1
PROG="$INSTALL_DIR/go-ipxe-admin"

start_service() {
    procd_open_instance
    procd_set_param command "\$PROG"
    procd_set_param respawn
    procd_set_param env ADMIN_USER="$ADMIN_USER" ADMIN_PASS="$ADMIN_PASS"
    procd_close_instance
}
EOF
    chmod +x /etc/init.d/go-ipxe-admin
    /etc/init.d/go-ipxe-admin enable
    /etc/init.d/go-ipxe-admin start
elif [ -f "/usr/bin/systemctl" ] || [ -f "/bin/systemctl" ]; then
    # Systemd
    cat << EOF > /etc/systemd/system/go-ipxe-admin.service
[Unit]
Description=go-ipxe-admin
After=network.target

[Service]
EnvironmentFile=$INSTALL_DIR/.env
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
    # SysV init
    cat << EOF > /etc/init.d/go-ipxe-admin
#!/bin/sh
### BEGIN INIT INFO
# Provides:          go-ipxe-admin
# Required-Start:    \$remote_fs \$syslog
# Required-Stop:     \$remote_fs \$syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: go-ipxe-admin
# Description:       go-ipxe-admin
### END INIT INFO

NAME=go-ipxe-admin
DAEMON="$INSTALL_DIR/go-ipxe-admin"
PIDFILE=/var/run/\$NAME.pid
SCRIPTNAME="/etc/init.d/\$NAME"

# Exit on not found
[ -x "\$DAEMON" ] || exit 0

# Load the VERBOSE setting and other rcS variables
. /lib/init/vars.sh

# Load the functions
. /lib/lsb/init-functions

# Set environment
. "$INSTALL_DIR/.env"
export ADMIN_USER
export ADMIN_PASS

# Start the service
start_service() {
    cd "$INSTALL_DIR"
    start_daemon "\$DAEMON"
}

# Stop the service
stop_service() {
    stop_daemon "\$DAEMON"
}

case "\$1" in
    start)
        [ "\$VERBOSE" != no ] && log_daemon_msg "Starting \$NAME" "\$NAME"
        start_service
        log_end_msg \$?
        ;;
    stop)
        [ "\$VERBOSE" != no ] && log_daemon_msg "Stopping \$NAME" "\$NAME"
        stop_service
        log_end_msg \$?
        ;;
    restart|force-reload)
        log_daemon_msg "Restarting \$NAME" "\$NAME"
        stop_service
        start_service
        log_end_msg \$?
        ;;
    status)
        status_of_proc "\$DAEMON" "\$NAME" && exit 0 || exit 4
        ;;
    *)
        echo "Usage: \$SCRIPTNAME {start|stop|restart|force-reload|status}"
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
