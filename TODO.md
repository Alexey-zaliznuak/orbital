все логирование сделать через pkg/logger

возможно повыносить общие http штуки типо

```go
type ErrorResponse struct {
	Error string `json:"error"`
}
```
