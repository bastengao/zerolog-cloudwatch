package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/bastengao/zerolog-cloudwatch/writer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	w := writer.New(cfg, "my-group", "my-stream")
	log.Output(w)

	log.Info().Msg("hell world")
}
