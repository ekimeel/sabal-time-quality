### Building from Source
building the plugin requires all tags to be included

```bash
go build -gcflags="all=-N -l" -o tq-debug.so -buildmode=plugin
```