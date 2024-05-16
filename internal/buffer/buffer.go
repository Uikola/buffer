package buffer

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
)

// Buffer буфер для записи данных в бд поштучно.
type Buffer struct {
	Capacity int
	Data     chan map[string][]string
}

// New возвращает экземпляр буфера.
func New(capacity int) *Buffer {
	return &Buffer{
		Capacity: capacity,
		Data:     make(chan map[string][]string, capacity),
	}
}

// Push добавляет данные в буфер.
func (b *Buffer) Push(fact map[string][]string) {
	select {
	case b.Data <- fact:
	default:
		log.Info().Msg("Буфер переполнен")
	}
}

// Pop извлекает данные из буфера.
func (b *Buffer) Pop() map[string][]string {
	return <-b.Data
}

// SaveData сохраняет данные, делая запрос на стороннее апи.
func SaveData(fact map[string][]string, wg *sync.WaitGroup) {
	defer wg.Done()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, values := range fact {
		for _, val := range values {
			if err := writer.WriteField(key, val); err != nil {
				log.Error().Err(err).Msg(err.Error())
				return
			}
		}

	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, os.Getenv("KPI_URL"), body)
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		return
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKEN"))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		return
	}
	defer resp.Body.Close()

	log.Info().Msg(resp.Status)
}
