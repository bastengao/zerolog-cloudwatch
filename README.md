# zerlog AWS CloudWatch logs writer

Send logs from zerolog to AWS CloudWatch logs.

## Usage

Create log group and stream first, then specify group name and stream name.

Only need `logs:PutLogEvents` permission.

```go
import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/bastengao/zerolog-cloudwatch/writer"
)

func initWriter() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	w := writer.New(cfg, "my-group", "my-stream")
	log.Output(w)

  log.Info().Msg("hell world")
}
```
