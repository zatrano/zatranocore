package main

import (
	"flag"
	"log"

	"zatrano/configs"
	"zatrano/database"
)

func main() {
	flag.Bool("migrate", false, "Veritabanı başlatma işlemini çalıştır (migrasyonları içerir)")
	flag.Bool("seed", false, "Veritabanı başlatma işlemini çalıştır (seederları içerir)")
	flag.Parse()

	configs.InitDB()
	defer configs.CloseDB()

	db := configs.GetDB()

	log.Println("Veritabanı başlatma işlemi çalıştırılıyor...")
	database.Initialize(db)

}
