# go-serverify

go-serverify is Go client library for [serverify](https://github.com/autopp/serverify).

## Installation

```sh
go get github.com/autopp/go-serverify
```

## Usage

```go
import (
	"testing"

	"github.com/autopp/go-serverify"
)

func TestExample(t *testing.T) {
	s := serverify.New("http://localhost:8080")

	// Create session for this test
	session, _ := s.CreateSession("test_session")
	defer session.Delete()

	// Act with serverify
	app := MyApplication(MyConfig { endpoint: session.BaseURL() })
	...

	// Assert occured request
	logs, _ := session.Logs()
	...
}
```

## License

[Apache-2.0](LICENSE)
