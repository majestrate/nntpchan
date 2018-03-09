package srnd

import (
	"log"
	"net/http"
	"strconv"
)

type CacheHandler interface {
	http.Handler
	GetI18N(r *http.Request) *I18N
}

type CacheInterface interface {
	RegenAll()
	RegenFrontPage()
	RegenOnModEvent(newsgroup, msgid, root string, page int)
	RegenerateBoard(group string)
	Regen(msg ArticleEntry)

	DeleteThreadMarkup(root_post_id string)
	DeleteBoardMarkup(group string)

	Start()
	Close()
	GetHandler() CacheHandler

	SetRequireCaptcha(required bool)
	InvertPagination()
}

//TODO only pass needed config
func NewCache(cache_type, host, port, user, password string, cache_config, config map[string]string, db Database, store ArticleStore) CacheInterface {
	prefix := config["prefix"]
	webroot := config["webroot"]
	translations := config["translations"]
	threads := mapGetInt(config, "regen_threads", 1)
	name := config["name"]
	attachments := mapGetInt(config, "allow_files", 1) == 1

	if cache_type == "file" {
		return NewFileCache(prefix, webroot, name, threads, attachments, db, store)
	}
	if cache_type == "null" {
		return NewNullCache(prefix, webroot, name, translations, attachments, db, store)
	}
	if cache_type == "varnish" {
		url := cache_config["url"]
		bind_addr := cache_config["bind"]
		workers, _ := strconv.Atoi(cache_config["workers"])
		if workers <= 0 {
			workers = 4
		}
		return NewVarnishCache(url, bind_addr, prefix, webroot, name, translations, workers, attachments, db, store)
	}

	log.Fatalf("invalid cache type: %s", cache_type)
	return nil
}
