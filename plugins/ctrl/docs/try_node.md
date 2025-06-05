### **Try Node**

The **Try Node** handles errors that may occur during packet processing. When a packet is processed and an error occurs,
it directs the error to the error output port for appropriate handling.

## **Specification**

- **None**: The Try Node operates by default without any additional configuration.

## **Ports**

- **in**: Receives and processes incoming packets.
- **out**: Outputs the original input packet if processed without error.
- **error**: Outputs any errors that occur during packet processing.

## **Example**

```yaml
- kind: try
  ports:
    out:
      - name: next
        port: in
    error:
      - name: catch
        port: in
```
