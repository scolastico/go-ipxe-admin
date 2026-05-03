# go-ipxe-admin

A lightweight, Go-based iPXE deployment dashboard and management server. It provides an intuitive admin UI and REST API to manage iPXE boot templates, serve static assets, and track server deployments.

## Features

- **Admin Dashboard**: Web UI to manage active deployments, boot templates, and file assets.
- **REST API**: Full API for automation and integration.
- **iPXE Endpoint**: Dynamically serve iPXE boot scripts.
- **Asset Management**: Upload and serve static files (kernels, initrds, ISOs) for booting.
- **Security**: Protected by Basic Authentication.
- **Easy Installation**: Single script to install as a `systemd` or `sysvinit` service.
- **Lightweight**: Compiled to a single statically linked binary, compressed with UPX.

## Installation

The quickest way to install is using the provided `install.sh` script. It detects your system architecture, downloads the latest binary, and sets it up as a background service.

```bash
curl -sSL https://raw.githubusercontent.com/scolastico/go-ipxe-admin/main/install.sh | sudo bash
```

## Configuration and Usage

By default, the server runs on port `8080`.

### Environment Variables

You can configure the application using environment variables:

- `PORT`: The port the server runs on (default: `8080`)
- `ADMIN_USER`: The admin username (default: `admin`)
- `ADMIN_PASS`: The admin password (default: `admin`)

To access the dashboard, navigate to `http://<your-server-ip>:<port>/` in your web browser.

### API Endpoints

- `GET /ipxe`: The main endpoint for iPXE clients to boot from.
- `GET /a/:id/:filename`: Endpoint for serving uploaded static assets.
- `GET /api/deployments`, `POST /api/deployments`, `DELETE /api/deployments/:ip`: Manage deployments.
- `GET /api/templates`, `POST /api/templates`, `DELETE /api/templates/:id`: Manage boot templates.
- `GET /api/assets`, `POST /api/assets`, `DELETE /api/assets/:id`: Manage static assets.

## Uninstall

```bash
curl -sSL https://raw.githubusercontent.com/scolastico/go-ipxe-admin/main/uninstall.sh | sudo bash
```

## Building from Source

To build the project yourself, you will need Go installed.

```bash
# Clone the repository
git clone https://github.com/scolastico/go-ipxe-admin.git
cd go-ipxe-admin

# Build all binaries (requires Make and UPX)
make all

# Or run in development mode
make dev
```

This will output the compressed binaries to the `bin/` directory.

## License

This project is licensed under the terms found in the `LICENSE.txt` file.

## AI Usage Note

This software was created with Antigravity, a tool that uses AI to help
software development. AI-generated code may be imperfect.
Please review the code carefully for errors, vulnerabilities, or other
issues before using it. The developers do not guarantee the accuracy
or safety of AI-generated code and accept no liability for any
damage or harm caused by its use.
