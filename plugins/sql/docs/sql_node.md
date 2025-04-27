# SQL Node

**The SQL Node** provides the functionality to interact with relational databases by executing SQL queries and
processing data. This node executes SQL queries on the database and returns the results as packets.

## Specification

- **driver**: The name of the database driver, such as `"sqlite3"`, `"postgres"`, etc.
- **source**: The database connection string, provided in the format appropriate for the driver.
- **isolation**: Sets the transaction isolation level. The default value is `0`. (Optional)

## Ports

- **in**: Receives SQL queries and parameters to send requests to the database.
- **out**: Returns packets containing the results of the query execution.
- **error**: Returns any errors that occur during query execution.

## Example

```yaml
- kind: snippet
  language: javascript
  code: |
    export default function (args) {
      return [
        'INSERT INTO USERS(name) VALUES (?)',
        ["foo", "bar"]
      ];
    }
  ports:
    out:
      - name: sql
        port: in

- kind: sql
  name: sql
  driver: sqlite3
  source: file::memory:?cache=shared
  ports:
    out:
      - name: next
        port: in
```
