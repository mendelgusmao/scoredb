package persistence

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mendelgusmao/scoredb/lib/database"
)

type Persistence struct {
	database *database.Database
	config   Configuration
}

func NewPersistence(database *database.Database, config Configuration) *Persistence {
	return &Persistence{
		database: database,
		config:   config,
	}
}

func (p *Persistence) Load() {
	if _, err := os.Stat(p.config.SnapshotPath); os.IsNotExist(err) {
		return
	}

	file, err := os.Open(p.config.SnapshotPath)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	err = gob.NewDecoder(reader).Decode(p.database)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("[Persistence.Load] Loaded database")
}

func (p *Persistence) Save() error {
	buffer := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buffer).Encode(p.database)

	if err != nil {
		return fmt.Errorf("[Persistence.Save] %v", err)
	}

	if err := os.WriteFile(p.config.SnapshotPath, buffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("[Persistence.Save] %v", err)
	}

	return nil
}

func (p *Persistence) Work() {
	interval := p.config.SnapshotInterval
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
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
