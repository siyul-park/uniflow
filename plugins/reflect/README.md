# 🔧 Reflect Plugin

The **Reflect Plugin** provides a standard `SQL` driver for controlling runtime resources. It enables you to query and
manage runtime resources using `SQL`-compatible nodes. This plugin allows you to write and execute `SQL` queries to
manage the state of runtime resources.

## Available Drivers

- **runtime**: A standard SQL driver for accessing and managing runtime resources. It uses the **system** data source to
  query and modify runtime tables such as `frames`, `processes`, and `symbols`.
