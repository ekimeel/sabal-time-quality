### Building from Source
building the plugin requires all tags to be included

```bash
go build -gcflags="all=-N -l" -o tq-debbug.so -buildmode=plugin
```