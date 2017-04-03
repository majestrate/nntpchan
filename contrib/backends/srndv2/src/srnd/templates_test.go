package srnd

import (
	"log"
	"os"
	"testing"
)

func makeBenchmarkDB() Database {
	return NewDatabase("postgres", "srnd", "/var/run/postgresql", "", "", "")
}

func BenchmarkRenderBoardPage(b *testing.B) {
	db := makeBenchmarkDB()
	db.CreateTables()
	defer db.Close()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wr, err := os.Create("boardpage.html")
			if err == nil {
				template.genBoardPage(true, true, "prefix", "test", "overchan.random", 0, wr, db, false)
			} else {
				log.Println("did not write", "boardpage.html", err)
			}
			wr.Close()
		}
	})
}

func BenchmarkRenderThread(b *testing.B) {
	db := makeBenchmarkDB()
	db.CreateTables()
	defer db.Close()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wr, err := os.Create("thread.html")
			if err == nil {
				template.genThread(true, true, ArticleEntry{"<c49be1451427261@nntp.nsfl.tk>", "overchan.random"}, "prefix", "frontend", wr, db, false)
			} else {
				log.Println("did not write", "thread.html", err)
			}
			wr.Close()
		}
	})
}
