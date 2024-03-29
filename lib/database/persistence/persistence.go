package persistence

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mendelgusmao/scoredb/lib/database"
	msgpack "github.com/vmihailenco/msgpack/v5"
)

type Persistence[T any] struct {
	database *database.Database[T]
	config   Configuration
	loading  bool
}

func NewPersistence[T any](database *database.Database[T], config Configuration) *Persistence[T] {
	return &Persistence[T]{
		database: database,
		config:   config,
	}
}

func (p *Persistence[T]) Loading() bool {
	return p.loading
}

func (p *Persistence[T]) Load() {
	go p.load()
}

func (p *Persistence[T]) load() {
	p.loading = true
	defer func() { p.loading = false }()

	if _, err := os.Stat(p.config.SnapshotPath); os.IsNotExist(err) {
		return
	}

	log.Println("[Persistence.Load] Loading database")

	file, err := os.Open(p.config.SnapshotPath)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	err = msgpack.NewDecoder(reader).Decode(p.database)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("[Persistence.Load] Database is loaded")
}

func (p *Persistence[T]) Save() error {
	buffer := bytes.NewBuffer(nil)
	err := msgpack.NewEncoder(buffer).Encode(p.database)

	if err != nil {
		return fmt.Errorf("[Persistence.Save] %v", err)
	}

	if err := os.WriteFile(p.config.SnapshotPath, buffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("[Persistence.Save] %v", err)
	}

	return nil
}

func (p *Persistence[T]) Work() {
	interval := p.config.SnapshotInterval
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				if p.loading {
					continue
				}

				err := p.Save()

				if err != nil {
					log.Println(err)
					continue
				}

				log.Printf("[Persistence.Work] Database saved")
			}
		}
	}()
}
