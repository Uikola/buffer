package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/Uikola/buffer/internal/buffer"
	"github.com/Uikola/buffer/pkg/zlog"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Error().Msg(err.Error())
		os.Exit(1)
	}

	logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Error().Msg(err.Error())
		os.Exit(1)
	}
	// объявление логера
	log.Logger = zlog.Default(true, "dev", zerolog.Level(logLevel))

	if err = run(); err != nil {
		log.Error().Err(err).Msg(err.Error())
		os.Exit(1)
	}
}

func run() error {
	var wg sync.WaitGroup
	buf := buffer.New(1000)

	// горутина сохраняющая данные из буфера на сервере поштучно
	go func() {
		for {
			fact := buf.Pop()
			log.Info().Msg("Данные извлечены из буфера")
			wg.Add(1)
			go buffer.SaveData(fact, &wg)
			wg.Wait()
		}
	}()

	// обработчик, принимающий данные и отправляющий их в буфер
	http.HandleFunc("/buff-add", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			log.Error().Err(err).Msg("Не удалось распарсить форму")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"reason": "bad request"})
			return
		}

		form := r.MultipartForm
		data := form.Value

		buf.Data <- data
		log.Info().Msg("Данные отправлены в буфер")
	})

	if err := http.ListenAndServe(os.Getenv("PORT"), nil); err != nil {
		return err
	}
	return nil
}
